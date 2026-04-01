package repository

import (
	"context"

	"bit-labs.cn/owl-admin/app/model"
	"bit-labs.cn/owl/contract"
	"bit-labs.cn/owl/provider/db"
	"gorm.io/gorm"
)

type UserGroupRepositoryInterface interface {
	Save(group *model.UserGroup) error
	Detail(id any) (*model.UserGroup, error)
	Retrieve(page, pageSize int, fn func(db *gorm.DB)) (count int64, list []model.UserGroup, err error)
	Options() (list []UserGroupItem, err error)
	Unique(id uint, name string, code string) (*model.UserGroup, bool)
	GetUsersByGroupID(groupID uint, page, pageSize int) (count int64, list []model.User, err error)
	GetUserIDsByGroupID(groupID uint) ([]string, error)
	contract.WithContext[UserGroupRepositoryInterface]
}

type UserGroupRepository struct {
	db  *gorm.DB
	ctx context.Context
	db.BaseRepository[model.UserGroup]
}

var _ UserGroupRepositoryInterface = (*UserGroupRepository)(nil)

func NewUserGroupRepository(d *gorm.DB) UserGroupRepositoryInterface {
	return &UserGroupRepository{
		db:             d,
		BaseRepository: db.NewBaseRepository[model.UserGroup](d),
	}
}

func (i *UserGroupRepository) WithContext(ctx context.Context) UserGroupRepositoryInterface {
	i.db = i.db.WithContext(ctx)
	i.ctx = ctx
	return i
}

func (i *UserGroupRepository) Save(group *model.UserGroup) error {
	err := i.db.Omit("Users").Save(&group).Error
	if err != nil {
		return err
	}
	err = i.db.Model(&group).Association("Users").Replace(&group.Users)
	return err
}

func (i *UserGroupRepository) Detail(id any) (*model.UserGroup, error) {
	var group model.UserGroup
	err := i.db.Where("id = ?", id).First(&group).Error
	return &group, err
}

func (i *UserGroupRepository) Retrieve(page, pageSize int, fn func(db *gorm.DB)) (count int64, list []model.UserGroup, err error) {
	return i.BaseRepository.Retrieve(page, pageSize, fn)
}

func (i *UserGroupRepository) Unique(id uint, name string, code string) (*model.UserGroup, bool) {
	return i.BaseRepository.Unique(id, func(db *gorm.DB) {
		db.Where("name = ? or code = ?", name, code)
	})
}

func (i *UserGroupRepository) GetUsersByGroupID(groupID uint, page, pageSize int) (count int64, list []model.User, err error) {
	q := i.db.Model(&model.User{}).
		Joins("JOIN admin_user_group_user ugu ON ugu.user_id = admin_user.id AND ugu.user_group_id = ?", groupID).
		Preload("Roles").
		Order("admin_user.created_at desc")

	err = q.Count(&count).Error
	if err != nil {
		return
	}
	if page > 0 && pageSize > 0 {
		q = q.Offset((page - 1) * pageSize).Limit(pageSize)
	}
	err = q.Find(&list).Error
	return
}

type UserGroupItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (i *UserGroupRepository) Options() (result []UserGroupItem, err error) {
	err = i.db.Model(&model.UserGroup{}).Where("status = 1").Scan(&result).Error
	return
}

func (i *UserGroupRepository) GetUserIDsByGroupID(groupID uint) ([]string, error) {
	var ids []string
	err := i.db.Model(&model.UserGroupUser{}).
		Where("user_group_id = ?", groupID).
		Pluck("user_id", &ids).Error
	return ids, err
}
