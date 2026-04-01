package model

import "gorm.io/gorm/schema"

var _ schema.Tabler = (*UserDept)(nil)

type UserDept struct {
	UserID uint `json:"userID,string" gorm:"comment:用户id;uniqueIndex:uk_user_dept"`
	DeptID uint `json:"deptID,string" gorm:"comment:部门id;uniqueIndex:uk_user_dept"`
}

func (UserDept) TableName() string {
	return "admin_user_dept"
}
