package model

import (
	"bit-labs.cn/owl/provider/db"
	"github.com/spf13/cast"
	"gorm.io/gorm/schema"
)

var _ schema.Tabler = (*UserGroup)(nil)
var _ schema.Tabler = (*UserGroupUser)(nil)

type UserGroup struct {
	db.BaseModel
	Name   string `gorm:"comment:用户组名称;type:string;size:128" json:"name"`
	Code   string `gorm:"comment:用户组编码;type:string;size:64" json:"code"`
	Status int    `gorm:"comment:状态(1启用,2禁用)" json:"status"`
	Remark string `gorm:"comment:用户组描述" json:"remark"`

	Users []User `gorm:"many2many:admin_user_group_user;joinForeignKey:user_group_id;References:id;JoinReferences:user_id" json:"users,omitempty"`
}

func (i *UserGroup) TableName() string {
	return "admin_user_group"
}

func (i *UserGroup) SetUsers(users []User) {
	i.Users = users
}

func (i *UserGroup) Enable()  { i.Status = 1 }
func (i *UserGroup) Disable() { i.Status = 2 }

func (i *UserGroup) GetUserIDs() []string {
	var ids []string
	for _, u := range i.Users {
		ids = append(ids, cast.ToString(u.ID))
	}
	return ids
}

// UserGroupUser 用户组与用户关联表
type UserGroupUser struct {
	UserGroupID uint `json:"userGroupID" gorm:"comment:用户组id;uniqueIndex:uk_group_user"`
	UserID      uint `json:"userID" gorm:"comment:用户id;uniqueIndex:uk_group_user"`
}

func (i *UserGroupUser) TableName() string {
	return "admin_user_group_user"
}
