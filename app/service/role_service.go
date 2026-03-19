package service

import (
	"context"

	"bit-labs.cn/owl-admin/app/event"
	"bit-labs.cn/owl-admin/app/model"
	"bit-labs.cn/owl-admin/app/repository"
	errContract "bit-labs.cn/owl/contract/errors"
	"bit-labs.cn/owl/contract/log"
	"bit-labs.cn/owl/provider/db"
	"bit-labs.cn/owl/provider/redis"
	"bit-labs.cn/owl/provider/router"
	"github.com/asaskevich/EventBus"
	"github.com/casbin/casbin/v2"
	validatorv10 "github.com/go-playground/validator/v10"
	"github.com/jinzhu/copier"
	"github.com/spf13/cast"
	"gorm.io/gorm"
)

const (
	CodeRoleExists   = "ROLE_EXISTS"
	CodeRoleNotFound = "ROLE_NOT_FOUND"
)

func RoleExists() *errContract.BizError {
	return errContract.NewBizError(CodeRoleExists, "角色已存在")
}

func RoleNotFound() *errContract.BizError {
	return errContract.NewBizError(CodeRoleNotFound, "角色不存在")
}

// CreateRoleReq 创建角色
type CreateRoleReq struct {
	Name   string `json:"name" validate:"required,min=2,max=32" label:"角色名称"`          // 角色名称
	Code   string `json:"code" validate:"required,alphanum,min=2,max=32" label:"角色编码"` // 角色编码
	Status int    `json:"status" validate:"omitempty,oneof=1 2" label:"状态"`            // 状态(1启用,2禁用)
	Remark string `json:"remark" validate:"omitempty,max=255" label:"备注"`              // 备注
}

// UpdateRoleReq 更新角色
type UpdateRoleReq struct {
	ID uint `json:"id,string" validate:"required" label:"角色ID"` // 角色ID
	CreateRoleReq
}

// AssignMenuToRole 分配菜单给角色, 菜单和按钮权限
type AssignMenuToRole struct {
	RoleID  uint     `json:"roleID,string" validate:"required" label:"角色ID"` // 角色ID
	MenuIDs []string `json:"menuIds" validate:"required" label:"菜单ID列表"`     // 菜单ID列表
}

// AssignRoleToUser 分配角色给用户
type AssignRoleToUser struct {
	UserID  uint     `json:"userID,string" validate:"required" label:"用户ID"` // 用户ID
	RoleIDs []string `json:"roleIDs" validate:"required" label:"角色ID列表"`     // 角色ID列表
}

type RetrieveRoleReq struct {
	router.PageReq
	NameLike string `json:"name" validate:"omitempty,max=32" label:"角色名称"`          // 名称模糊搜索
	CodeLike string `json:"code" validate:"omitempty,alphanum,max=32" label:"角色编码"` // 角色编码
	Status   uint8  `json:"status" validate:"omitempty,oneof=1 2" label:"状态"`       // 状态(1启用,2禁用)
}

// RoleService 角色服务
type RoleService struct {
	db.BaseRepository[model.Role]
	enforcer casbin.IEnforcer
	ctx      context.Context
	log      log.Logger
	validate *validatorv10.Validate

	roleRepo repository.RoleRepositoryInterface
	menuRepo *router.MenuRepository
	eventbus EventBus.Bus
	locker   redis.LockerFactory
}

func NewRoleService(
	menuManager *router.MenuRepository,
	roleRepo repository.RoleRepositoryInterface,
	enforcer casbin.IEnforcer,
	bus EventBus.Bus,
	locker redis.LockerFactory,
	validate *validatorv10.Validate,
	gdb *gorm.DB,
) *RoleService {
	return &RoleService{
		menuRepo:       menuManager,
		enforcer:       enforcer,
		roleRepo:       roleRepo,
		eventbus:       bus,
		locker:         locker,
		validate:       validate,
		BaseRepository: db.NewBaseRepository[model.Role](gdb),
	}
}
func (i *RoleService) WithContext(ctx context.Context) *RoleService {
	i.ctx = ctx
	return i
}
func (i *RoleService) CreateRole(ctx context.Context, req *CreateRoleReq) error {
	if err := i.validate.Struct(req); err != nil {
		return err
	}

	l := i.locker.New()
	if err := l.Lock("role:create"); err != nil {
		return err
	}
	defer l.Unlock()

	// 创建前唯一性校验：name 或 code 不能重复。
	if _, exists := i.roleRepo.WithContext(ctx).Unique(0, req.Name, req.Code); exists {
		return RoleExists()
	}

	var role model.Role
	err := copier.Copy(&role, req)
	if err != nil {
		return err
	}

	role.Enable()
	return i.roleRepo.WithContext(ctx).Save(&role)
}

func (i *RoleService) UpdateRole(ctx context.Context, req *UpdateRoleReq) error {
	if err := i.validate.Struct(req); err != nil {
		return err
	}

	l := i.locker.New()
	if err := l.Lock("role:update:" + cast.ToString(req.ID)); err != nil {
		return err
	}
	defer l.Unlock()

	// 更新前唯一性校验：排除自身记录后，name 或 code 不能重复。
	if _, exists := i.roleRepo.WithContext(ctx).Unique(req.ID, req.Name, req.Code); exists {
		return RoleExists()
	}

	role, err := i.roleRepo.WithContext(ctx).Detail(req.ID)
	if err != nil {
		return err
	}

	err = copier.Copy(&role, req)
	if err != nil {
		return err
	}

	return i.roleRepo.WithContext(ctx).Save(role)
}

// ChangeStatus 修改角色状态
func (i *RoleService) ChangeStatus(ctx context.Context, req *db.ChangeStatus) error {
	if err := i.validate.Struct(req); err != nil {
		return err
	}

	l := i.locker.New()
	if err := l.Lock("role:status:" + cast.ToString(req.ID)); err != nil {
		return err
	}
	defer l.Unlock()

	return i.BaseRepository.ChangeStatus(req)
}

func (i *RoleService) Options(ctx context.Context) (list []repository.RoleItem, err error) {
	return i.roleRepo.WithContext(ctx).Options()
}

// DeleteRole 删除角色
func (i *RoleService) DeleteRole(ctx context.Context, id uint) error {

	l := i.locker.New()
	if err := l.Lock("role:delete:" + cast.ToString(id)); err != nil {
		return err
	}
	defer l.Unlock()

	return i.BaseRepository.Delete(id)
}

func (i *RoleService) RetrieveRoles(ctx context.Context, req *RetrieveRoleReq) (count int64, list []model.Role, err error) {
	if err := i.validate.Struct(req); err != nil {
		return 0, nil, err
	}

	return i.roleRepo.WithContext(ctx).Retrieve(req.Page, req.PageSize, func(tx *gorm.DB) {
		db.AppendWhereFromStruct(tx, req)
		tx.Order("created_at desc")
	})
}

func (i *RoleService) AssignMenusToRole(ctx context.Context, req *AssignMenuToRole) error {
	if err := i.validate.Struct(req); err != nil {
		return err
	}

	l := i.locker.New()
	if err := l.Lock("role:assign-menus:" + cast.ToString(req.RoleID)); err != nil {
		return err
	}
	defer l.Unlock()

	role, err := i.roleRepo.WithContext(ctx).Detail(req.RoleID)
	if err != nil {
		return err
	}

	role.SetMenus(db.GetModelsByIDs[model.Menu](req.MenuIDs))

	err = i.roleRepo.WithContext(ctx).Save(role)

	i.eventbus.Publish(event.AssignMenuToRole, req)

	return err
}

// GetRolesMenuIDs 获取角色的菜单IDs
func (i *RoleService) GetRolesMenuIDs(ctx context.Context, ids ...string) (result []string) {

	ds, err := i.roleRepo.WithContext(ctx).GetRolesMenuIDs(ids...)
	if err != nil {
		i.log.Error("获取角色菜单IDs失败", err)
		return nil
	}
	return ds
}
