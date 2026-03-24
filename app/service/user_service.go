package service

import (
	"context"
	"errors"

	"github.com/spf13/cast"

	"bit-labs.cn/owl-admin/app/event"
	"bit-labs.cn/owl-admin/app/model"
	"bit-labs.cn/owl-admin/app/provider/jwt"
	"bit-labs.cn/owl-admin/app/repository"
	errContract "bit-labs.cn/owl/contract/errors"
	"bit-labs.cn/owl/provider/conf"
	"bit-labs.cn/owl/provider/db"
	"bit-labs.cn/owl/provider/redis"
	"bit-labs.cn/owl/provider/router"
	"bit-labs.cn/owl/utils"
	"github.com/asaskevich/EventBus"
	validatorv10 "github.com/go-playground/validator/v10"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

var (
	ErrLogin = errors.New("用户名或密码错误")
)

const (
	CodeUserExists   = "USER_EXISTS"
	CodeUserNotFound = "USER_NOT_FOUND"
)

func UserExists() *errContract.BizError {
	return errContract.NewBizError(CodeUserExists, "用户已存在")
}

func UserNotFound() *errContract.BizError {
	return errContract.NewBizError(CodeUserNotFound, "用户不存在")
}

type UserBatchFields struct {
	Username string `json:"username" validate:"required,min=2,max=32" label:"用户名"` // 用户名
	NickName string `json:"nickName" validate:"omitempty,max=32" label:"昵称"`       // 昵称
	Email    string `json:"email" validate:"omitempty,email" label:"邮箱"`           // 邮箱
	Phone    string `json:"phone" validate:"omitempty,numeric" label:"手机号"`        // 手机号
	Remark   string `json:"remark" validate:"omitempty,max=255" label:"备注"`        // 备注
	Status   int    `json:"status" validate:"omitempty,oneof=1 2 3" label:"状态"`    // 状态(1启用,2禁用,3待审核)
	Sex      int    `json:"sex" validate:"omitempty,oneof=1 2 3" label:"性别"`       // 性别(1男,2女,3未知)
	Source   string `json:"source" validate:"omitempty,max=32" label:"来源"`         // 来源
	SourceID string `json:"sourceId" validate:"omitempty,max=64" label:"来源ID"`     // 来源ID
}

// CreateUserReq 创建用户
type CreateUserReq struct {
	UserBatchFields
	Password string `json:"password" validate:"required,min=6,max=64" label:"密码"` // 密码
}

type UpdateUserReq struct {
	ID uint `json:"id,string,omitempty"` // 用户ID
	UserBatchFields
}

type RetrieveUserReq struct {
	router.PageReq
	UsernameLike string `json:"username" form:"username" validate:"omitempty,max=64" label:"用户名"` // 用户名模糊
	PhoneLike    string `json:"phone" form:"phone" validate:"omitempty,max=32" label:"手机号"`       // 手机号模糊
	Status       int    `json:"status" form:"status" validate:"omitempty,oneof=1 2" label:"状态"`   // 状态
}

type LoginReq struct {
	Username string `json:"username" validate:"required" label:"用户名"` // 用户名
	Password string `json:"password" validate:"required" label:"密码"`  // 密码
}

type LoginResp struct {
	*model.User
	AccessToken string `json:"accessToken"` // 访问令牌
}

type ChangePasswordReq struct {
	UserID      uint   `json:"userId,string"`                                            // 用户ID
	OldPassword string `json:"oldPassword" validate:"required,min=6,max=64" label:"旧密码"` // 旧密码
	NewPassword string `json:"newPassword" validate:"required,min=6,max=64" label:"新密码"` // 新密码
}

// ResetPasswordReq 超管重置用户密码
type ResetPasswordReq struct {
	UserID      uint   `json:"userId,string" validate:"required" label:"用户ID"`           // 用户ID
	NewPassword string `json:"newPassword" validate:"required,min=6,max=64" label:"新密码"` // 新密码
}

type AssignPermissionReq struct {
	PermissionIds []uint `json:"permissionIds"`             // 权限ID列表
	UserId        uint   `json:"userId" binding:"required"` // 用户ID
}

// AssignMenuToUser 分配菜单给用户
type AssignMenuToUser struct {
	UserID  uint     `json:"userId,string"` // 用户id
	MenuIDs []string `json:"menuIds"`       // 菜单列表
}

type AssignRolesReq struct {
	UserID  uint   `json:"userId,string"` // 用户id
	RoleIDs []uint `json:"roleIds"`       // 角色ids
}

type UserService struct {
	db         *gorm.DB
	menuManger *router.MenuRepository
	jwtSvc     *jwt.JWTService
	db.BaseRepository[model.User]
	roleSvc   *RoleService
	userRepo  repository.UserRepositoryInterface
	eventBus  EventBus.Bus
	configure *conf.Configure
	locker    redis.LockerFactory
	validate  *validatorv10.Validate
}

func NewUserService(
	roleSvc *RoleService,
	userRepo repository.UserRepositoryInterface,
	tx *gorm.DB,
	eventBus EventBus.Bus,
	configure *conf.Configure,
	jwtSvc *jwt.JWTService,
	menuManager *router.MenuRepository,
	locker redis.LockerFactory,
	validate *validatorv10.Validate,
) *UserService {
	return &UserService{
		db:             tx,
		roleSvc:        roleSvc,
		userRepo:       userRepo,
		BaseRepository: db.NewBaseRepository[model.User](tx),
		eventBus:       eventBus,
		configure:      configure,
		jwtSvc:         jwtSvc,
		menuManger:     menuManager,
		locker:         locker,
		validate:       validate,
	}
}

// Login 用户登录
func (i *UserService) Login(ctx context.Context, req *LoginReq) (resp *LoginResp, err error) {
	if err := i.validate.Struct(req); err != nil {
		return nil, err
	}

	user, err := i.GetUserByName(ctx, req.Username)
	if err != nil {
		return nil, err
	}

	// 校验密码是否正确
	if ok := utils.BcryptCheck(req.Password, user.Password); !ok {
		return nil, ErrLogin
	}

	token, err := i.jwtSvc.GenerateToken(user)
	return &LoginResp{
		User:        user,
		AccessToken: token,
	}, err
}

// LoginByThirdParty 第三方用户登录
func (i *UserService) LoginByThirdParty(ctx context.Context, username, provider string) (resp *LoginResp, err error) {

	thirdProvider, err := i.userRepo.WithContext(ctx).GetByNameAndThirdProvider(username, provider)
	if err != nil {
		return nil, err
	}

	token, err := i.jwtSvc.GenerateToken(&thirdProvider)
	return &LoginResp{
		User:        &thirdProvider,
		AccessToken: token,
	}, err
}

// GetUserByName 根据用户名查找用户
func (i *UserService) GetUserByName(ctx context.Context, name string) (*model.User, error) {

	var adminLoginReq LoginReq
	err := i.configure.GetConfig("app.admin", &adminLoginReq)
	if err != nil {
		return nil, err
	}

	// 查找用户，优先从配置里面找admin
	var user model.User
	if name == adminLoginReq.Username {
		user = model.NewSuperUser()
		user.Password = adminLoginReq.Password
	} else {
		// 从数据库查找用户
		user, err = i.userRepo.WithContext(ctx).GetByName(name)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrLogin
		}
	}

	return &user, nil
}

// GetMyPermissions 获取当前用户的权限标识列表
func (i *UserService) GetMyPermissions(ctx context.Context, userID uint, isSuperAdmin bool) ([]string, error) {
	if isSuperAdmin {
		return []string{"*:*:*"}, nil
	}

	user, err := i.userRepo.WithContext(ctx).FindById(userID)
	if err != nil {
		return nil, err
	}

	roleIDs := user.GetRoleIDs()
	menuIDs := i.roleSvc.GetRolesMenuIDs(ctx, roleIDs...)
	return i.menuManger.GetPermissionsByMenuIDs(menuIDs...), nil
}

// GetUserMenus 获取用户菜单
func (i *UserService) GetUserMenus(ctx context.Context, userID uint) []*router.Menu {

	user, err := i.userRepo.WithContext(ctx).FindById(userID)
	if err != nil {
		return nil
	}

	menuIDs := i.roleSvc.GetRolesMenuIDs(ctx, user.GetRoleIDs()...)

	menus := i.menuManger.GetMenuByMenuIDs(menuIDs...)
	return menus
}

// AssignRoleToUser 分配角色给用户
func (i *UserService) AssignRoleToUser(ctx context.Context, req *AssignRoleToUser) error {
	if err := i.validate.Struct(req); err != nil {
		return err
	}
	l := i.locker.New()
	if err := l.Lock("user:assign-roles:" + cast.ToString(req.UserID)); err != nil {
		return err
	}
	defer l.Unlock()

	roles := db.GetModelsByIDs[model.Role](req.RoleIDs)

	user, err := i.userRepo.WithContext(ctx).FindById(req.UserID)
	if err != nil {
		return err
	}

	user.SetRoles(roles)
	err = i.userRepo.WithContext(ctx).Save(user)

	i.eventBus.Publish(event.AssignRoleToUser, req)
	return err
}

// GetUserRoleIDs 获取用户的角色IDs
func (i *UserService) GetUserRoleIDs(ctx context.Context, id uint) ([]string, error) {

	user, err := i.userRepo.WithContext(ctx).FindById(id)
	if err != nil {
		return nil, err
	}

	return user.GetRoleIDs(), nil
}

// ChangeUserPassword 修改用户密码
func (i *UserService) ChangeUserPassword(ctx context.Context, req *ChangePasswordReq) error {
	if err := i.validate.Struct(req); err != nil {
		return err
	}

	l := i.locker.New()
	if err := l.Lock("user:change-pwd:" + cast.ToString(req.UserID)); err != nil {
		return err
	}
	defer l.Unlock()

	user, err := i.userRepo.WithContext(ctx).FindById(req.UserID)
	if err != nil {
		return err
	}

	err = user.ChangePassword(req.OldPassword, req.NewPassword)
	if err != nil {
		return err
	}
	return i.userRepo.WithContext(ctx).Save(user)
}

// ResetUserPassword 超管重置用户密码（不校验旧密码）
func (i *UserService) ResetUserPassword(ctx context.Context, req *ResetPasswordReq) error {
	if err := i.validate.Struct(req); err != nil {
		return err
	}

	l := i.locker.New()
	if err := l.Lock("user:reset-pwd:" + cast.ToString(req.UserID)); err != nil {
		return err
	}
	defer l.Unlock()

	user, err := i.userRepo.WithContext(ctx).FindById(req.UserID)
	if err != nil {
		return err
	}
	user.SetPassword(req.NewPassword)

	return i.userRepo.WithContext(ctx).Save(user)
}

type ChangeAvatarReq struct {
	UserID uint   `json:"userId,string" validate:"required" label:"用户ID"`
	Avatar string `json:"avatar" validate:"required,url" label:"头像"`
}

// ChangeUserAvatar 修改用户头像
func (i *UserService) ChangeUserAvatar(ctx context.Context, req *ChangeAvatarReq) error {
	if err := i.validate.Struct(req); err != nil {
		return err
	}

	l := i.locker.New()
	if err := l.Lock("user:avatar:" + cast.ToString(req.UserID)); err != nil {
		return err
	}
	defer l.Unlock()

	user, err := i.userRepo.WithContext(ctx).FindById(req.UserID)
	if err != nil {
		return err
	}
	user.SetAvatar(req.Avatar)

	return i.userRepo.WithContext(ctx).Save(user)
}

// RetrieveUsers 获取用户列表
func (i *UserService) RetrieveUsers(ctx context.Context, req *RetrieveUserReq) (count int, list []model.User, err error) {
	if err = i.validate.Struct(req); err != nil {
		return 0, nil, err
	}

	c, u, e := i.userRepo.WithContext(ctx).Retrieve(req.Page, req.PageSize, func(tx *gorm.DB) {
		db.AppendWhereFromStruct(tx, req)
		tx.Preload("Roles")
		tx.Order("created_at desc")
	})

	return cast.ToInt(c), u, e
}

// RetrieveUsersByRoleID 按角色获取用户列表
func (i *UserService) RetrieveUsersByRoleID(ctx context.Context, roleID uint, page, pageSize int) (count int, list []model.User, err error) {
	c, u, e := i.userRepo.WithContext(ctx).Retrieve(page, pageSize, func(tx *gorm.DB) {
		tx.Joins("JOIN admin_user_role aur ON aur.user_id = admin_user.id AND aur.role_id = ?", roleID)
		tx.Preload("Roles")
		tx.Order("created_at desc")
	})

	return cast.ToInt(c), u, e
}

// CreateUser 创建用户
func (i *UserService) CreateUser(ctx context.Context, req *CreateUserReq) error {
	if err := i.validate.Struct(req); err != nil {
		return err
	}

	l := i.locker.New()
	if err := l.Lock("user:create"); err != nil {
		return err
	}
	defer l.Unlock()

	// 创建前做唯一性校验，避免 repository 直接返回“业务已存在”错误。
	if i.userRepo.WithContext(ctx).Unique(0, req.Username, req.Source) {
		return UserExists()
	}

	var user model.User
	err := copier.Copy(&user, req)
	if err != nil {
		return err
	}
	user.SetPassword(req.Password)

	if err = i.userRepo.WithContext(ctx).Save(&user); err != nil {
		return err
	}
	return err
}

// Register 注册用户
func (i *UserService) Register(ctx context.Context, req *model.User) error {
	if err := i.validate.Struct(req); err != nil {
		return err
	}

	l := i.locker.New()
	if err := l.Lock("user:register"); err != nil {
		return err
	}
	defer l.Unlock()

	var user model.User
	err := copier.Copy(&user, req)
	if err != nil {
		return err
	}

	return i.userRepo.WithContext(ctx).Save(&user)
}

// UpdateUser 更新用户
func (i *UserService) UpdateUser(ctx context.Context, req *UpdateUserReq) error {
	if err := i.validate.Struct(req); err != nil {
		return err
	}

	l := i.locker.New()
	if err := l.Lock("user:update:" + cast.ToString(req.ID)); err != nil {
		return err
	}
	defer l.Unlock()

	// 更新前做唯一性校验：当 (username, source) 存在且不是自身记录时返回“已存在”业务异常。
	if i.userRepo.WithContext(ctx).Unique(req.ID, req.Username, req.Source) {
		return UserExists()
	}

	user, err := i.userRepo.WithContext(ctx).FindById(req.ID)
	if err != nil {
		return err
	}

	err = copier.Copy(&user, req)
	if err != nil {
		return err
	}

	return i.userRepo.WithContext(ctx).Save(user)
}

// ChangeUserStatus 修改用户状态
func (i *UserService) ChangeUserStatus(ctx context.Context, req *db.ChangeStatus) error {
	if err := i.validate.Struct(req); err != nil {
		return err
	}

	l := i.locker.New()
	if err := l.Lock("user:status:" + cast.ToString(req.ID)); err != nil {
		return err
	}
	defer l.Unlock()

	return i.BaseRepository.ChangeStatus(req)
}

func (i *UserService) DeleteUser(ctx context.Context, id uint) error {
	l := i.locker.New()
	if err := l.Lock("user:delete:" + cast.ToString(id)); err != nil {
		return err
	}
	defer l.Unlock()

	err := i.BaseRepository.Delete(id)
	return err
}
