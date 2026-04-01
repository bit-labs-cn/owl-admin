package model

import (
	"bit-labs.cn/owl/utils"
	"errors"
	"time"

	"bit-labs.cn/owl/provider/db"

	"github.com/spf13/cast"
)

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
	Source     string     `gorm:"comment:用户来源" json:"source"`
	SourceID   string     `gorm:"comment:第三方用户唯一标识" json:"sourceID"`

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
		BaseModel:    db.BaseModel{ID: 19941996},
		Username:     "glen",
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
