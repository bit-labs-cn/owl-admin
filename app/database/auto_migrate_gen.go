// generate by auto_migrate Do not edit it
package database

import (
	. "bit-labs.cn/owl-admin/app/model"
)

// Models 返回待自动迁移的 model 列表。
func Models() []any {
	return []any{
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
		&TrustedDevice{},
	}
}
