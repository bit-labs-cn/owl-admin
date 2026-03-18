package repository

import (
	"context"
	"errors"

	"bit-labs.cn/owl-admin/app/model"
	"bit-labs.cn/owl/contract"
	"bit-labs.cn/owl/provider/db"
	"gorm.io/gorm"
)

var ErrRoleNotExists = errors.New("角色不存在")
var ErrRoleExists = errors.New("角色已存在")

type RoleRepositoryInterface interface {
	Save(role *model.Role) error
	Detail(id any) (*model.Role, error)
	Options() (list []RoleItem, err error)
	GetRolesMenuIDs(roleID ...string) ([]string, error)
	Unique(id uint, name string, code string) (*model.Role, bool)
	Retrieve(page, pageSize int, fn func(db *gorm.DB)) (count int64, list []model.Role, err error)
	contract.WithContext[RoleRepositoryInterface]
}

type RoleRepository struct {
	db  *gorm.DB
	ctx context.Context
	db.BaseRepository[model.Role]
}

var _ RoleRepositoryInterface = (*RoleRepository)(nil)

func NewRoleRepository(d *gorm.DB) RoleRepositoryInterface {
	return &RoleRepository{
		db:             d,
		BaseRepository: db.NewBaseRepository[model.Role](d),
	}
}

// Save 保存角色
func (i *RoleRepository) Save(role *model.Role) error {
	if _, b := i.Unique(role.ID, role.Name, role.Code); b {
		return ErrRoleExists
	}

	err := i.db.Save(&role).Error
	if err != nil {
		return err
	}
	err = i.db.Model(&role).Association("Menus").Replace(&role.Menus)

	return err
}

// WithContext 设置上下文
func (i *RoleRepository) WithContext(ctx context.Context) RoleRepositoryInterface {
	i.db = i.db.WithContext(ctx)
	i.ctx = ctx
	return i
}

// Detail 角色详情
func (i *RoleRepository) Detail(id any) (*model.Role, error) {
	var role model.Role
	err := i.db.Where("id = ?", id).Preload("Menus").First(&role).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &role, ErrRoleNotExists
	}
	return &role, err
}

func (i *RoleRepository) Retrieve(page, pageSize int, fn func(db *gorm.DB)) (count int64, list []model.Role, err error) {
	return i.BaseRepository.Retrieve(page, pageSize, fn)
}

// Unique 唯一性校验
func (i *RoleRepository) Unique(id uint, name string, code string) (*model.Role, bool) {
	return i.BaseRepository.Unique(id, func(db *gorm.DB) {
		db.Where("name = ? or code = ?", name, code)
	})
}

// GetRolesMenuIDs 获取多个角色所拥有的菜单ID
func (i *RoleRepository) GetRolesMenuIDs(roleIDs ...string) (result []string, err error) {
	err = i.db.Where("role_id in ?", roleIDs).Model(&model.RoleMenu{}).Select("menu_id").Scan(&result).Error
	return
}

type RoleItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Options 角色下拉选择列表
func (i *RoleRepository) Options() (result []RoleItem, err error) {
	i.db.Model(&model.Role{}).Scan(&result)
	return
}
