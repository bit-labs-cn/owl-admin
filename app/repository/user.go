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
	// Unique 用于唯一性判断：当存在与 (username, source) 匹配的其他记录时返回 true。
	// id > 0 时会排除自身（id != ?），用于 update 场景。
	Unique(id uint, username string, source string) bool
	Save(user *model.User) error
	Delete(ids ...any) error
	Retrieve(page, pageSize int, fn func(db *gorm.DB)) (count int64, list []model.User, err error)
	GetByName(name string) (model.User, error)
	GetByNameAndThirdProvider(name string, provider string) (model.User, error)
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
	err := i.db.Save(&user).Error
	if err != nil {
		return err
	}
	err = i.db.Model(&user).Association("Roles").Replace(&user.Roles)

	return err
}

func (i *UserRepository) Unique(id uint, username string, source string) bool {
	_, exists := i.BaseRepository.Unique(id, func(db *gorm.DB) {
		db.Where("username", username).Where("source", source)
	})
	return exists
}
func (i *UserRepository) FindById(id any) (*model.User, error) {
	var user model.User
	err := i.db.Where("id = ?", id).Preload("Roles").First(&user).Error
	return &user, err
}
func (i *UserRepository) GetByName(name string) (model.User, error) {
	var user model.User
	err := i.db.Where("username = ?", name).Preload("Roles").First(&user).Error
	return user, err
}

func (i *UserRepository) GetByNameAndThirdProvider(name string, provider string) (model.User, error) {
	var user model.User
	err := i.db.Where("username = ?", name).Where("source = ?", provider).Preload("Roles").First(&user).Error
	return user, err
}
