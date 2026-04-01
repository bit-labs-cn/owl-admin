// generate by auto_migrate Do not edit it
package database

import (
	. "bit-labs.cn/owl-admin/app/model"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) {
	_ = db.Migrator().AutoMigrate(

		&Dict{},
		&DictItem{},

		&Api{},

		&Menu{},
		&Role{},
		&User{},
		&RoleMenu{},
		&UserMenu{},
		&Dept{},
		&Area{},
		&Position{},
		&LoginLog{},
		&OperationLog{},

		&AppVersion{},
		&UserGroup{},
		&UserGroupUser{},
		&UserDept{},
	)
}
