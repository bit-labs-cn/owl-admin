package model

import (
	"time"

	"bit-labs.cn/owl/provider/db"
)

// Delivery 是一个收件人在一个渠道上的一次投递。
type Delivery struct {
	db.BaseModel
	TaskID        uint       `gorm:"not null;default:0;uniqueIndex:uk_notify_delivery,priority:1;index:idx_notify_delivery_task;comment:任务 ID" json:"taskId,string"`
	TaskTitle     string     `gorm:"-" json:"taskTitle"`
	RecipientType int        `gorm:"not null;default:1;uniqueIndex:uk_notify_delivery,priority:2;comment:收件人类型" json:"recipientType"`
	RecipientID   uint       `gorm:"not null;default:0;uniqueIndex:uk_notify_delivery,priority:3;comment:收件人 ID" json:"recipientId,string"`
	RecipientName string     `gorm:"size:128;not null;default:'';comment:收件人名称" json:"recipientName"`
	Channel       string     `gorm:"size:32;not null;default:'';uniqueIndex:uk_notify_delivery,priority:4;index:idx_notify_delivery_channel;comment:渠道" json:"channel"`
	Provider      string     `gorm:"size:32;not null;default:'';comment:服务商" json:"provider"`
	Address       string     `gorm:"size:255;not null;default:'';comment:投递地址" json:"address"`
	Status        int        `gorm:"not null;default:1;index:idx_notify_delivery_status;comment:投递状态" json:"status"`
	ErrorCode     string     `gorm:"size:64;not null;default:'';comment:错误码" json:"errorCode"`
	ErrorMsg      string     `gorm:"size:500;not null;default:'';comment:错误信息" json:"errorMsg"`
	RetryCount    int        `gorm:"not null;default:0;comment:已重试次数" json:"retryCount"`
	ProviderMsgID string     `gorm:"size:128;not null;default:'';comment:服务商消息 ID" json:"providerMsgId"`
	SentAt        *time.Time `gorm:"comment:发送时间" json:"sentAt,omitempty"`
	NextRetryAt   *time.Time `gorm:"index:idx_notify_delivery_retry;comment:下次重试时间" json:"nextRetryAt,omitempty"`
}

func (Delivery) TableName() string { return "notify_delivery" }
