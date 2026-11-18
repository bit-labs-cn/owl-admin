package model

import (
	"time"

	"bit-labs.cn/owl/provider/db"
)

// Device 是一台可接收 App 推送的设备。
type Device struct {
	db.BaseModel
	PrincipalType int        `gorm:"not null;default:1;index:idx_notify_device_principal,priority:1;comment:主体类型" json:"principalType"`
	PrincipalID   uint       `gorm:"not null;default:0;index:idx_notify_device_principal,priority:2;comment:主体 ID" json:"principalId,string"`
	PrincipalName string     `gorm:"-" json:"principalName"`
	Platform      string     `gorm:"size:16;not null;default:'';comment:平台" json:"platform"`
	Provider      string     `gorm:"size:32;not null;default:'';comment:推送服务商" json:"provider"`
	Token         string     `gorm:"size:512;not null;default:'';uniqueIndex:uk_notify_device_token;comment:推送令牌" json:"token"`
	AppID         string     `gorm:"size:128;not null;default:'';comment:应用标识" json:"appId"`
	DeviceName    string     `gorm:"size:128;not null;default:'';comment:设备名称" json:"deviceName"`
	Status        int        `gorm:"not null;default:1;index:idx_notify_device_status;comment:状态" json:"status"`
	LastActiveAt  *time.Time `gorm:"comment:最近活跃时间" json:"lastActiveAt,omitempty"`
}

func (Device) TableName() string { return "notify_device" }
