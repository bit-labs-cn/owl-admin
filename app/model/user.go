package model

import (
	"errors"
	"strings"
	"time"

	"bit-labs.cn/owl/utils"

	"bit-labs.cn/owl/provider/db"

	"github.com/spf13/cast"
)

// 用户状态。与字典 user_status 一致。
const (
	UserStatusEnabled  = 1 // 启用
	UserStatusDisabled = 2 // 禁用
	UserStatusPending  = 3 // 待审核
)

// 用户性别。与字典 gender 一致。
const (
	UserSexMale    = 1 // 男
	UserSexFemale  = 2 // 女
	UserSexUnknown = 3 // 未知
)

// 用户来源，表示账号是怎么来的，不是联系邮箱或手机号。
// 空字符串是后台添加和批量导入。邮箱注册、手机号注册、Oauth2 提供方各自写入对应标识。
const (
	UserSourceLocal         = ""      // 后台添加、批量导入
	UserSourceEmailRegister = "email" // 邮箱验证码注册
	UserSourcePhoneRegister = "phone" // 手机号注册
	UserSourceGitHub        = "github"
	UserSourceGitee         = "gitee"
	UserSourceGoogle        = "google"
	UserSourceOwl           = "owl"

	// UserSourceQueryLocal 列表筛选后台添加用户时的查询值。
	// 库中这类用户的 source 为空，空查询参数表示不按来源筛选。
	UserSourceQueryLocal = "local"
)

// UserSourceLabel 把库存的 source 转成给人看的名称。
func UserSourceLabel(source string) string {
	switch source {
	case UserSourceLocal:
		return "后台添加"
	case UserSourceEmailRegister:
		return "邮箱注册"
	case UserSourcePhoneRegister:
		return "手机号注册"
	case UserSourceGitHub:
		return "GitHub"
	case UserSourceGitee:
		return "Gitee"
	case UserSourceGoogle:
		return "Google"
	case UserSourceOwl:
		return "统一身份认证"
	default:
		if strings.TrimSpace(source) == "" {
			return "后台添加"
		}
		return source
	}
}

// UserSourceQueryValue 把库存 source 转成列表查询参数。
func UserSourceQueryValue(source string) string {
	if strings.TrimSpace(source) == "" {
		return UserSourceQueryLocal
	}
	return source
}

// ParseUserSourceQuery 解析列表查询参数。空字符串表示不筛选。
func ParseUserSourceQuery(query string) (stored string, ok bool) {
	query = strings.TrimSpace(query)
	if query == "" {
		return "", false
	}
	if query == UserSourceQueryLocal {
		return UserSourceLocal, true
	}
	return query, true
}

type User struct {
	db.BaseModel
	Avatar     string     `gorm:"comment:用户头像" json:"avatar"`
	Username   string     `gorm:"comment:用户名称;type:string;size:512" json:"username"`
	Nickname   string     `gorm:"comment:用户昵称;type:string;size:128" json:"nickname"`
	Password   string     `gorm:"comment:用户密码" json:"-"`
	Remark     string     `gorm:"comment:remark" json:"remark"`
	Phone      string     `gorm:"comment:手机;type:string;size:32" json:"phone"`
	Email      string     `gorm:"comment:邮箱" json:"email"`
	Status     int        `gorm:"comment:状态" json:"status"`
	Sex        int        `gorm:"comment:性别" json:"sex"`
	VerifiedAt *time.Time `gorm:"comment:验证时间" json:"verified_at"`
	RegisterIP string     `gorm:"comment:注册IP;type:string;size:64" json:"register_ip"`
	Source     string     `gorm:"comment:用户来源" json:"source"`
	SourceID   string     `gorm:"comment:第三方用户唯一标识" json:"sourceID"`
	SourceRaw  string     `gorm:"comment:第三方登录原始信息;type:text" json:"sourceRaw,omitempty"`

	Roles  []Role      `gorm:"many2many:admin_user_role;joinForeignKey:user_id;References:id;JoinReferences:role_id" json:"roles"`
	Menus  []Menu      `gorm:"many2many:admin_user_menu;joinForeignKey:user_id;References:id;JoinReferences:menu_id" json:"menus"`
	Groups []UserGroup `gorm:"many2many:admin_user_group_user;joinForeignKey:user_id;References:id;JoinReferences:user_group_id" json:"groups,omitempty"`
	Depts  []Dept      `gorm:"many2many:admin_user_dept;joinForeignKey:user_id;References:id;JoinReferences:dept_id" json:"depts,omitempty"`

	Permissions  []string `json:"permissions" gorm:"-"`
	IsSuperAdmin bool     `json:"isSuperAdmin" gorm:"-"`
	update       map[string]bool
}

func (i *User) TableName() string {
	return "admin_user"
}

func (i *User) SetRoles(roles []Role) {
	i.Roles = roles
}
func (i *User) ChangePassword(old, new string) error {
	if i.Password != old {
		return errors.New("旧密码错误")
	}
	i.Password = new
	return nil
}

func (i *User) SetAvatar(avatar string) {
	i.Avatar = avatar
}

func (i *User) SetPassword(newPassword string) {
	i.Password = utils.BcryptHash(newPassword)
}

func (i *User) GetRoleIDs() []string {
	var roleIDs []string
	for _, role := range i.Roles {
		roleIDs = append(roleIDs, cast.ToString(role.ID))
	}
	return roleIDs
}

func (i *User) SetGroups(groups []UserGroup) {
	i.Groups = groups
}

func (i *User) GetGroupIDs() []string {
	var ids []string
	for _, g := range i.Groups {
		ids = append(ids, cast.ToString(g.ID))
	}
	return ids
}

func (i *User) SetDepts(depts []Dept) {
	i.Depts = depts
}

func (i *User) GetDeptIDs() []string {
	var ids []string
	for _, d := range i.Depts {
		ids = append(ids, cast.ToString(d.ID))
	}
	return ids
}

func NewSuperUser() User {
	return User{
		BaseModel:    db.BaseModel{ID: 761297884},
		Username:     "admin",
		Nickname:     "超级管理员",
		IsSuperAdmin: true,
		Permissions:  []string{"*:*:*"},
		Roles:        []Role{{Name: "superAdmin"}},
	}
}

// UserMenu 用户菜单
type UserMenu struct {
	UserID uint   `json:"userID" gorm:"comment:角色id;index"`
	MenuID string `json:"menuID" gorm:"comment:菜单id;index"`
}

func (i *UserMenu) TableName() string {
	return "admin_user_menu"
}
