package model

import "bit-labs.cn/owl/provider/db"

// Subscription 是主体对某个分类、某个渠道的开关。
type Subscription struct {
	db.BaseModel
	PrincipalType int    `gorm:"not null;default:1;uniqueIndex:uk_notify_sub,priority:1;comment:主体类型" json:"principalType"`
	PrincipalID   uint   `gorm:"not null;default:0;uniqueIndex:uk_notify_sub,priority:2;comment:主体 ID" json:"principalId,string"`
	Category      int    `gorm:"not null;default:1;uniqueIndex:uk_notify_sub,priority:3;comment:分类" json:"category"`
	Channel       string `gorm:"size:32;not null;default:'';uniqueIndex:uk_notify_sub,priority:4;comment:渠道" json:"channel"`
	Enabled       int    `gorm:"not null;default:1;comment:是否接收" json:"enabled"`
}

func (Subscription) TableName() string { return "notify_subscription" }

// Setting 是通知中心的运行配置。
type Setting struct {
	db.BaseModel
	Key   string `gorm:"size:64;not null;default:'';uniqueIndex:uk_notify_setting_key;comment:配置键" json:"key"`
	Value string `gorm:"type:text;comment:配置值" json:"value"`
}

func (Setting) TableName() string { return "notify_setting" }
