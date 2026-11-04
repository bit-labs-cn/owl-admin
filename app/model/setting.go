package model

import "bit-labs.cn/owl/provider/db"

// Setting 保存某个子系统改过的配置值。没有记录时使用代码里的默认值。
type Setting struct {
	db.BaseModel
	Subsystem string `gorm:"size:64;not null;uniqueIndex:uk_admin_setting_sub_key;comment:子系统" json:"subsystem"`
	Key       string `gorm:"size:64;not null;uniqueIndex:uk_admin_setting_sub_key;comment:配置键" json:"key"`
	// 300000 盖住 200KB 图片的 base64。MySQL 在长度超过 65535 时会建成 mediumtext。
	Value string `gorm:"size:300000;not null;comment:配置值" json:"value"`
}

func (Setting) TableName() string { return "admin_setting" }
