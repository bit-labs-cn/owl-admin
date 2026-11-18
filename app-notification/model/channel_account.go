package model

import "bit-labs.cn/owl/provider/db"

// ChannelAccount 是一个渠道的发送账号。敏感字段由调用方加密后写入。
type ChannelAccount struct {
	db.BaseModel
	Channel    string `gorm:"size:32;not null;default:'';index:idx_notify_channel_account;comment:渠道" json:"channel"`
	Provider   string `gorm:"size:32;not null;default:'';comment:服务商" json:"provider"`
	Name       string `gorm:"size:128;not null;default:'';comment:名称" json:"name"`
	ConfigJSON string `gorm:"type:text;comment:配置 JSON" json:"configJson"`
	IsDefault  int    `gorm:"not null;default:2;comment:是否默认" json:"isDefault"`
	Status     int    `gorm:"not null;default:1;comment:状态" json:"status"`
}

func (ChannelAccount) TableName() string { return "notify_channel_account" }

// Event 是业务方登记的通知事件。
type Event struct {
	db.BaseModel
	Code         string `gorm:"size:64;not null;default:'';uniqueIndex:uk_notify_event_code;comment:事件编码" json:"code"`
	Name         string `gorm:"size:128;not null;default:'';comment:事件名称" json:"name"`
	Source       string `gorm:"size:64;not null;default:'';index:idx_notify_event_source;comment:来源子系统" json:"source"`
	Kind         int    `gorm:"not null;default:1;comment:默认类型" json:"kind"`
	Category     int    `gorm:"not null;default:1;comment:默认分类" json:"category"`
	ChannelsJSON string `gorm:"type:text;comment:默认渠道 JSON" json:"channelsJson"`
	Description  string `gorm:"size:500;not null;default:'';comment:说明" json:"description"`
	Status       int    `gorm:"not null;default:1;comment:状态" json:"status"`
}

func (Event) TableName() string { return "notify_event" }
