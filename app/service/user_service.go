package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/cast"

	"bit-labs.cn/owl-admin/app/event"
	"bit-labs.cn/owl-admin/app/model"
	"bit-labs.cn/owl-admin/app/provider/jwt"
	"bit-labs.cn/owl-admin/app/repository"
	errContract "bit-labs.cn/owl/contract/errors"
	mailerContract "bit-labs.cn/owl/contract/mailer"
	"bit-labs.cn/owl/provider/captcha"
	"bit-labs.cn/owl/provider/conf"
	"bit-labs.cn/owl/provider/db"
	"bit-labs.cn/owl/provider/mailer"
	owlredis "bit-labs.cn/owl/provider/redis"
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
	CodeUserExists        = "USER_EXISTS"
	CodeUserNotFound      = "USER_NOT_FOUND"
	CodeEmailCodeInvalid  = "EMAIL_CODE_INVALID"
	CodeEmailCodeTooOften = "EMAIL_CODE_TOO_OFTEN"
	CodeEmailSendFailed   = "EMAIL_SEND_FAILED"
	CodeCaptchaInvalid    = "CAPTCHA_INVALID"

	emailRegisterSource    = "email"
	registerCodeKeyPrefix  = "register:code:"
	registerCooldownPrefix = "register:cooldown:"
	registerCodeTTL        = 10 * time.Minute
	registerCooldownTTL    = 60 * time.Second
	registerCodeLength     = 6
	trustedDeviceTTL       = 7 * 24 * time.Hour
)

func UserExists() *errContract.BizError {
	return errContract.NewBizError(CodeUserExists, "用户已存在")
}

func UserNotFound() *errContract.BizError {
	return errContract.NewBizError(CodeUserNotFound, "用户不存在")
}

func EmailCodeInvalid() *errContract.BizError {
	return errContract.NewBizError(CodeEmailCodeInvalid, "邮箱验证码错误或已过期")
}

func EmailCodeTooOften() *errContract.BizError {
	return errContract.NewBizError(CodeEmailCodeTooOften, "发送过于频繁，请稍后再试")
}

func EmailSendFailed() *errContract.BizError {
	return errContract.NewBizError(CodeEmailSendFailed, "邮件发送失败")
}

func CaptchaInvalid() *errContract.BizError {
	return errContract.NewBizError(CodeCaptchaInvalid, "验证码错误或已过期")
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
	DeptID   string `json:"parentId" validate:"omitempty" label:"归属部门"`            // 归属部门ID
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
	// Keyword 综合搜索：登录名、昵称同时模糊匹配（与 UsernameLike 互斥，优先 Keyword）
	Keyword      string `json:"keyword" form:"keyword" validate:"omitempty,max=64" label:"关键词"`
	UsernameLike string `json:"username" form:"username" validate:"omitempty,max=64" label:"用户名"` // 用户名模糊
	PhoneLike    string `json:"phone" form:"phone" validate:"omitempty,max=32" label:"手机号"`       // 手机号模糊
	Status       int    `json:"status" form:"status" validate:"omitempty,oneof=1 2" label:"状态"`   // 状态
	DeptID       string `json:"deptId" form:"deptId" validate:"omitempty"`                        // 部门筛选
}

type LoginReq struct {
	Username string         `json:"username" validate:"required" label:"用户名"` // 用户名
	Password string         `json:"password" validate:"required" label:"密码"`  // 密码
	DeviceID string         `json:"deviceId" validate:"omitempty,max=64" label:"设备指纹"`
	Captcha  *CaptchaAnswer `json:"captcha" validate:"omitempty" label:"验证码"`
}

// CaptchaAnswer 登录验证码答案（字段随 type 不同而使用）
type CaptchaAnswer struct {
	CaptchaId string               `json:"captchaId" validate:"required" label:"验证码ID"`
	X         int                  `json:"x" label:"滑块X坐标"`
	Y         int                  `json:"y" label:"滑块Y坐标"`
	Angle     int                  `json:"angle" label:"旋转角度"`
	Points    []captcha.ClickPoint `json:"points" label:"点选坐标"`
}

// LoginContext 登录请求上下文（由 handle 注入）
type LoginContext struct {
	IP        string
	UserAgent string
}

// SendRegisterCodeReq 发送注册验证码
type SendRegisterCodeReq struct {
	Email string `json:"email" validate:"required,email" label:"邮箱"`
}

// EmailRegisterReq 邮箱注册
type EmailRegisterReq struct {
	Email    string `json:"email" validate:"required,email" label:"邮箱"`
	Code     string `json:"code" validate:"required,len=6" label:"验证码"`
	Password string `json:"password" validate:"required,min=6,max=64" label:"密码"`
	NickName string `json:"nickName" validate:"omitempty,max=32" label:"昵称"`
}

type LoginResp struct {
	*model.User
	AccessToken string `json:"accessToken,omitempty"` // 访问令牌
	NeedCaptcha bool   `json:"needCaptcha,omitempty"` // 是否需要验证码
	CaptchaType string `json:"captchaType,omitempty"` // 验证码类型
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
	roleSvc           *RoleService
	userRepo          repository.UserRepositoryInterface
	trustedDeviceRepo repository.TrustedDeviceRepositoryInterface
	captchaSvc        *captcha.Service
	eventBus          EventBus.Bus
	configure         *conf.Configure
	locker            owlredis.LockerFactory
	validate          *validatorv10.Validate
	redisClient       redis.UniversalClient
	mailer            *mailer.MailerManager
}

func NewUserService(
	roleSvc *RoleService,
	userRepo repository.UserRepositoryInterface,
	trustedDeviceRepo repository.TrustedDeviceRepositoryInterface,
	captchaSvc *captcha.Service,
	tx *gorm.DB,
	eventBus EventBus.Bus,
	configure *conf.Configure,
	jwtSvc *jwt.JWTService,
	menuManager *router.MenuRepository,
	locker owlredis.LockerFactory,
	validate *validatorv10.Validate,
	redisClient redis.UniversalClient,
	mailerMgr *mailer.MailerManager,
) *UserService {
	return &UserService{
		db:                tx,
		roleSvc:           roleSvc,
		userRepo:          userRepo,
		trustedDeviceRepo: trustedDeviceRepo,
		captchaSvc:        captchaSvc,
		BaseRepository:    db.NewBaseRepository[model.User](tx),
		eventBus:          eventBus,
		configure:         configure,
		jwtSvc:            jwtSvc,
		menuManger:        menuManager,
		locker:            locker,
		validate:          validate,
		redisClient:       redisClient,
		mailer:            mailerMgr,
	}
}

// Login 用户登录
func (i *UserService) Login(ctx context.Context, req *LoginReq, loginCtx *LoginContext) (resp *LoginResp, err error) {
	if err := i.validate.Struct(req); err != nil {
		return nil, err
	}

	user, err := i.GetUserByName(ctx, req.Username)
	if err != nil {
		return nil, err
	}

	ip := ""
	ua := ""
	if loginCtx != nil {
		ip = loginCtx.IP
		ua = loginCtx.UserAgent
	}

	needCaptcha := i.evaluateRisk(ctx, user.ID, req.DeviceID, ip)
	captchaVerified := false
	if needCaptcha && i.captchaSvc.Enabled() {
		captchaType := i.captchaSvc.DefaultType()
		if req.Captcha == nil || req.Captcha.CaptchaId == "" {
			return &LoginResp{
				NeedCaptcha: true,
				CaptchaType: captchaType,
			}, nil
		}
		ok, verifyErr := i.captchaSvc.Verify(ctx, &captcha.VerifyReq{
			Type:      captchaType,
			CaptchaId: req.Captcha.CaptchaId,
			X:         req.Captcha.X,
			Y:         req.Captcha.Y,
			Angle:     req.Captcha.Angle,
			Points:    req.Captcha.Points,
		})
		if verifyErr != nil {
			return nil, verifyErr
		}
		if !ok {
			return nil, CaptchaInvalid()
		}
		captchaVerified = true
	}

	if ok := utils.BcryptCheck(req.Password, user.Password); !ok {
		return nil, ErrLogin
	}

	token, err := i.jwtSvc.GenerateToken(user)
	if err != nil {
		return nil, err
	}

	if req.DeviceID != "" {
		i.recordTrustedDevice(ctx, user.ID, req.DeviceID, ip, ua, captchaVerified)
	}

	return &LoginResp{
		User:        user,
		AccessToken: token,
	}, nil
}

// evaluateRisk 判定当前登录是否需要验证码
func (i *UserService) evaluateRisk(ctx context.Context, userID uint, deviceID, ip string) bool {
	if deviceID == "" {
		return true
	}

	device, err := i.trustedDeviceRepo.WithContext(ctx).FindByUserAndDevice(userID, deviceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return true
		}
		return true
	}

	if !sameIPSegment(device.LastIP, ip) {
		return true
	}

	if time.Since(device.VerifiedAt) > trustedDeviceTTL {
		return true
	}

	return false
}

// recordTrustedDevice 登录成功后更新可信设备记录
func (i *UserService) recordTrustedDevice(ctx context.Context, userID uint, deviceID, ip, ua string, captchaVerified bool) {
	now := time.Now()
	record := &model.TrustedDevice{
		UserID:      userID,
		DeviceID:    deviceID,
		LastIP:      ip,
		LastUA:      ua,
		LastLoginAt: now,
	}

	if captchaVerified {
		record.VerifiedAt = now
	} else {
		existing, err := i.trustedDeviceRepo.WithContext(ctx).FindByUserAndDevice(userID, deviceID)
		if err == nil {
			record.VerifiedAt = existing.VerifiedAt
		} else {
			record.VerifiedAt = now
		}
	}

	_ = i.trustedDeviceRepo.WithContext(ctx).Upsert(record)
}

// sameIPSegment 比较两个 IP 是否处于同一网段（IPv4 /24）
func sameIPSegment(a, b string) bool {
	if a == b {
		return true
	}
	partsA := strings.Split(a, ".")
	partsB := strings.Split(b, ".")
	if len(partsA) == 4 && len(partsB) == 4 {
		return partsA[0] == partsB[0] && partsA[1] == partsB[1] && partsA[2] == partsB[2]
	}
	return false
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

	deptID := req.DeptID
	req.DeptID = ""

	kw := strings.TrimSpace(req.Keyword)
	type userListFilter struct {
		router.PageReq
		UsernameLike string
		PhoneLike    string
		Status       int
		DeptID       string
	}
	flt := userListFilter{
		PageReq:      req.PageReq,
		UsernameLike: req.UsernameLike,
		PhoneLike:    req.PhoneLike,
		Status:       req.Status,
		DeptID:       req.DeptID,
	}
	if kw != "" {
		// 由 Keyword 统一匹配登录名+昵称，避免仅 username 时搜不到“显示名”
		flt.UsernameLike = ""
	}

	c, u, e := i.userRepo.WithContext(ctx).Retrieve(req.Page, req.PageSize, func(tx *gorm.DB) {
		if kw != "" {
			like := "%" + kw + "%"
			tx.Where("(admin_user.username LIKE ? OR admin_user.nickname LIKE ?)", like, like)
		}
		db.AppendWhereFromStruct(tx, &flt)
		tx.Preload("Roles")
		tx.Preload("Depts")
		tx.Order("created_at desc")

		if deptID != "" {
			deptIDs := i.getDescendantDeptIDs(ctx, cast.ToUint(deptID))
			tx.Where("admin_user.id IN (?)",
				i.db.Table("admin_user_dept").Select("user_id").Where("dept_id IN ?", deptIDs))
		}
	})

	req.DeptID = deptID
	return cast.ToInt(c), u, e
}

// getDescendantDeptIDs 收集指定部门及所有子孙部门的 ID 集合
func (i *UserService) getDescendantDeptIDs(ctx context.Context, deptID uint) []uint {
	var allDepts []model.Dept
	if err := i.db.WithContext(ctx).Select("id, parent_id").Find(&allDepts).Error; err != nil {
		return []uint{deptID}
	}

	childrenMap := make(map[uint][]uint)
	for _, d := range allDepts {
		childrenMap[uint(d.ParentId)] = append(childrenMap[uint(d.ParentId)], d.ID)
	}

	result := []uint{deptID}
	queue := []uint{deptID}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, childID := range childrenMap[current] {
			result = append(result, childID)
			queue = append(queue, childID)
		}
	}
	return result
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

	if req.DeptID != "" {
		user.Depts = db.GetModelsByIDs[model.Dept]([]string{req.DeptID})
	}

	if err = i.userRepo.WithContext(ctx).Save(&user); err != nil {
		return err
	}
	return err
}

// SendRegisterCode 发送邮箱注册验证码
func (i *UserService) SendRegisterCode(ctx context.Context, req *SendRegisterCodeReq) error {
	if err := i.validate.Struct(req); err != nil {
		return err
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	if i.userRepo.WithContext(ctx).UniqueByEmail(0, email) {
		return UserExists()
	}

	cooldownKey := registerCooldownPrefix + email
	ok, err := i.redisClient.SetNX(ctx, cooldownKey, "1", registerCooldownTTL).Result()
	if err != nil {
		return err
	}
	if !ok {
		return EmailCodeTooOften()
	}

	code, err := generateNumericCode(registerCodeLength)
	if err != nil {
		return err
	}

	codeKey := registerCodeKeyPrefix + email
	if err = i.redisClient.Set(ctx, codeKey, code, registerCodeTTL).Err(); err != nil {
		return err
	}

	subject := "注册验证码"
	body := fmt.Sprintf("您的注册验证码为：%s，%d 分钟内有效。如非本人操作请忽略。", code, int(registerCodeTTL.Minutes()))
	htmlBody := fmt.Sprintf("<p>您的注册验证码为：<strong>%s</strong>，%d 分钟内有效。</p><p>如非本人操作请忽略。</p>", code, int(registerCodeTTL.Minutes()))

	if err = i.mailer.Send(ctx, &mailerContract.Message{
		To:      []string{email},
		Subject: subject,
		Body:    body,
		HTML:    htmlBody,
	}); err != nil {
		_ = i.redisClient.Del(ctx, codeKey).Err()
		return EmailSendFailed()
	}

	return nil
}

// RegisterByEmail 邮箱验证码注册
func (i *UserService) RegisterByEmail(ctx context.Context, req *EmailRegisterReq, registerIP string) error {
	if err := i.validate.Struct(req); err != nil {
		return err
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	code := strings.TrimSpace(req.Code)

	l := i.locker.New()
	if err := l.Lock("user:register:" + email); err != nil {
		return err
	}
	defer l.Unlock()

	codeKey := registerCodeKeyPrefix + email
	stored, err := i.redisClient.Get(ctx, codeKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return EmailCodeInvalid()
		}
		return err
	}
	if stored != code {
		return EmailCodeInvalid()
	}

	if i.userRepo.WithContext(ctx).UniqueByEmail(0, email) {
		return UserExists()
	}
	if i.userRepo.WithContext(ctx).Unique(0, email, emailRegisterSource) {
		return UserExists()
	}

	now := time.Now()
	user := model.User{
		Username:   email,
		Email:      email,
		Nickname:   req.NickName,
		Status:     1,
		Source:     emailRegisterSource,
		RegisterIP: registerIP,
		VerifiedAt: &now,
	}
	if user.Nickname == "" {
		user.Nickname = email
	}
	user.SetPassword(req.Password)

	if err = i.userRepo.WithContext(ctx).Save(&user); err != nil {
		return err
	}

	_ = i.redisClient.Del(ctx, codeKey).Err()
	return nil
}

// Register 注册用户（保留兼容，请使用 RegisterByEmail）
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

func generateNumericCode(length int) (string, error) {
	const digits = "0123456789"
	var sb strings.Builder
	for n := 0; n < length; n++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		sb.WriteByte(digits[num.Int64()])
	}
	return sb.String(), nil
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

	if req.DeptID != "" {
		user.Depts = db.GetModelsByIDs[model.Dept]([]string{req.DeptID})
	} else {
		user.Depts = []model.Dept{}
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
