package repository

import (
	"context"

	"bit-labs.cn/owl-admin/app/model"
	"bit-labs.cn/owl/contract"
	"bit-labs.cn/owl/provider/db"
	"gorm.io/gorm"
)

type UserRepositoryInterface interface {
	FindById(id any) (*model.User, error)
	// Unique 用于唯一性判断：当存在与 (username, source, source_id) 匹配的其他记录时返回 true。
	// id > 0 时会排除自身（id != ?），用于 update 场景。source_id 为空时与本地账号、邮箱注册一致。
	Unique(id uint, username string, source string, sourceID string) bool
	Save(user *model.User) error
	// SaveBatch 在同一事务中批量保存用户及其角色/部门关联。
	SaveBatch(users []*model.User) error
	// ExistingUsernames 返回 (username, source) 已存在的用户名集合。
	ExistingUsernames(usernames []string, source string) (map[string]struct{}, error)
	// UpdateStatusByIDs 按 ID 列表批量更新状态。
	UpdateStatusByIDs(ids []uint, status int) error
	Delete(ids ...any) error
	Retrieve(page, pageSize int, fn func(db *gorm.DB)) (count int64, list []model.User, err error)
	GetByName(name string) (model.User, error)
	GetByEmail(email string) (model.User, error)
	UniqueByEmail(id uint, email string) bool
	GetByNameAndThirdProvider(name string, provider string, sourceID string) (model.User, error)
	// UpdateSourceRaw 写入第三方登录带回的原始用户资料。
	UpdateSourceRaw(id uint, raw string) error
	// ListSources 按 source 分组，返回已经出现过的来源。空字符串表示本地用户。
	ListSources() ([]string, error)
	contract.WithContext[UserRepositoryInterface]
}

var _ UserRepositoryInterface = (*UserRepository)(nil)

type UserRepository struct {
	db  *gorm.DB
	ctx context.Context
	db.BaseRepository[model.User]
}

func NewUserRepository(tx *gorm.DB) UserRepositoryInterface {
	return &UserRepository{
		db:             tx,
		BaseRepository: db.NewBaseRepository[model.User](tx),
	}
}

func (i *UserRepository) WithContext(ctx context.Context) UserRepositoryInterface {
	i.db = i.db.WithContext(ctx)
	i.ctx = ctx
	return i
}

func (i *UserRepository) Save(user *model.User) error {
	err := i.db.Omit("Roles", "Groups", "Depts", "Menus").Save(user).Error
	if err != nil {
		return err
	}
	if err = db.ReplaceJoinTable(i.db, user, "Roles", user.Roles); err != nil {
		return err
	}
	if err = db.ReplaceJoinTable(i.db, user, "Groups", user.Groups); err != nil {
		return err
	}
	return db.ReplaceJoinTable(i.db, user, "Depts", user.Depts)
}

func (i *UserRepository) SaveBatch(users []*model.User) error {
	if len(users) == 0 {
		return nil
	}
	return i.db.Transaction(func(tx *gorm.DB) error {
		repo := &UserRepository{
			db:             tx,
			ctx:            i.ctx,
			BaseRepository: db.NewBaseRepository[model.User](tx),
		}
		for _, user := range users {
			if err := repo.Save(user); err != nil {
				return err
			}
		}
		return nil
	})
}

func (i *UserRepository) ExistingUsernames(usernames []string, source string) (map[string]struct{}, error) {
	out := map[string]struct{}{}
	if len(usernames) == 0 {
		return out, nil
	}
	var found []string
	err := i.db.Model(&model.User{}).
		Where("username IN ? AND source = ?", usernames, source).
		Pluck("username", &found).Error
	if err != nil {
		return nil, err
	}
	for _, name := range found {
		out[name] = struct{}{}
	}
	return out, nil
}

func (i *UserRepository) UpdateStatusByIDs(ids []uint, status int) error {
	if len(ids) == 0 {
		return nil
	}
	return i.db.Model(&model.User{}).Where("id IN ?", ids).Update("status", status).Error
}

func (i *UserRepository) Unique(id uint, username string, source string, sourceID string) bool {
	_, exists := i.BaseRepository.Unique(id, func(db *gorm.DB) {
		db.Where("username", username).Where("source", source).Where("source_id", sourceID)
	})
	return exists
}
func (i *UserRepository) FindById(id any) (*model.User, error) {
	var user model.User
	err := i.db.Where("id = ?", id).Preload("Roles").Preload("Groups").Preload("Depts").First(&user).Error
	return &user, err
}
func (i *UserRepository) GetByName(name string) (model.User, error) {
	var user model.User
	err := i.db.Where("username = ?", name).Preload("Roles").Omit("Avatar").Preload("Depts").First(&user).Error
	return user, err
}

func (i *UserRepository) GetByEmail(email string) (model.User, error) {
	var user model.User
	err := i.db.Where("email = ?", email).Preload("Roles").Preload("Depts").First(&user).Error
	return user, err
}

func (i *UserRepository) UniqueByEmail(id uint, email string) bool {
	if email == "" {
		return false
	}
	_, exists := i.BaseRepository.Unique(id, func(db *gorm.DB) {
		db.Where("email = ?", email)
	})
	return exists
}

func (i *UserRepository) GetByNameAndThirdProvider(name string, provider string, sourceID string) (model.User, error) {
	var user model.User
	err := i.db.Where("username = ?", name).Where("source = ?", provider).Where("source_id = ?", sourceID).Preload("Roles").First(&user).Error
	return user, err
}

func (i *UserRepository) UpdateSourceRaw(id uint, raw string) error {
	return i.db.Model(&model.User{}).Where("id = ?", id).Update("source_raw", raw).Error
}

func (i *UserRepository) ListSources() ([]string, error) {
	type sourceRow struct {
		Source string `gorm:"column:source"`
	}
	var rows []sourceRow
	err := i.db.Model(&model.User{}).
		Select("COALESCE(source, '') AS source").
		Group("COALESCE(source, '')").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	sources := make([]string, 0, len(rows))
	for _, row := range rows {
		sources = append(sources, row.Source)
	}
	return sources, nil
}
