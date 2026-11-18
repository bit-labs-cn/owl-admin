package model

import (
	"time"

	"bit-labs.cn/owl/provider/db"
)

// NotifyDelivery 是一条外部渠道投递结果。UserID 为 0 且接收人类型为群发时表示群机器人。
type NotifyDelivery struct {
	db.BaseModel
	MessageID uint   `gorm:"not null;index:idx_notify_delivery_msg;comment:消息ID" json:"messageId,string"`
	UserID    uint   `gorm:"not null;default:0;index:idx_notify_delivery_user;comment:接收人用户ID，0 表示群发或外部人员" json:"userId,string"`
	Channel   string `gorm:"size:32;not null;default:'';index:idx_notify_delivery_channel;comment:渠道" json:"channel"`
	Title     string `gorm:"size:255;not null;default:'';comment:渲染后的标题" json:"title"`
	Content   string `gorm:"type:text;comment:渲染后的正文" json:"content"`
	Link      string `gorm:"size:255;not null;default:'';comment:渲染后的链接" json:"link"`
	Status    int    `gorm:"not null;default:1;index:idx_notify_delivery_status;comment:1待发送 2成功 3失败 4跳过" json:"status"`
	Attempts  int    `gorm:"not null;default:0;comment:已尝试次数" json:"attempts"`
	LastError string `gorm:"size:512;not null;default:'';comment:最后错误" json:"lastError"`
	Payload   string `gorm:"type:text;comment:附加业务元数据 JSON" json:"-"`

	// 接收人快照：外部人员没有 admin_user 记录，投递时只能依赖这里
	RecipientType  int    `gorm:"not null;default:1;comment:1站内用户 2外部人员 3群机器人" json:"recipientType"`
	RecipientName  string `gorm:"size:128;not null;default:'';comment:接收人名称" json:"recipientName"`
	RecipientEmail string `gorm:"size:255;not null;default:'';comment:接收人邮箱" json:"recipientEmail"`
	RecipientPhone string `gorm:"size:32;not null;default:'';comment:接收人手机号" json:"recipientPhone"`

	SentAt *time.Time `gorm:"comment:发送成功时间" json:"sentAt,omitempty"`
}

func (NotifyDelivery) TableName() string { return "notification_delivery" }
