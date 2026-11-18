package model

import "bit-labs.cn/owl/provider/db"

// NotifyType 是可被业务点名的消息类型。
type NotifyType struct {
	db.BaseModel
	Code            string `gorm:"size:64;not null;uniqueIndex:uk_notify_type_code;comment:类型编码" json:"code"`
	Name            string `gorm:"size:64;not null;default:'';comment:类型名称" json:"name"`
	Description     string `gorm:"size:255;not null;default:'';comment:说明" json:"description"`
	Kind            int    `gorm:"not null;default:1;comment:种类 1通知 2待办" json:"kind"`
	DefaultChannels string `gorm:"type:text;comment:默认渠道 JSON" json:"defaultChannels"`
	Status          int    `gorm:"not null;default:1;index:idx_notify_type_status;comment:1启用 2停用" json:"status"`
}

func (NotifyType) TableName() string { return "notification_type" }
