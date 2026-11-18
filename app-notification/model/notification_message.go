package model

import (
	"time"

	"bit-labs.cn/owl/provider/db"
)

// NotifyMessage 消息本体表（知会通知与行动待办共用）
type NotifyMessage struct {
	db.BaseModel
	Kind     int        `gorm:"not null;default:1;index:idx_notify_msg_kind;comment:1通知 2待办" json:"kind"`
	TypeCode string     `gorm:"size:64;not null;default:'';index:idx_notify_msg_type;comment:消息类型编码" json:"typeCode"`
	BizKey   string     `gorm:"size:128;not null;default:'';index:idx_notify_msg_biz;comment:业务主键" json:"bizKey"`
	Title    string     `gorm:"size:255;not null;default:'';comment:标题" json:"title"`
	Content  string     `gorm:"type:text;comment:正文说明" json:"content"`
	Link     string     `gorm:"size:255;not null;default:'';comment:跳转链接" json:"link"`
	DueAt    *time.Time `gorm:"index:idx_notify_msg_due;comment:待办截止时间" json:"dueAt,omitempty"`
	SenderID uint       `gorm:"not null;default:0;comment:发送人ID" json:"senderId,string"`
}

func (NotifyMessage) TableName() string { return "notification_message" }

// NotifyRecipient 接收流水与状态表
type NotifyRecipient struct {
	db.BaseModel
	MessageID  uint       `gorm:"not null;index:idx_notify_rec_msg;comment:消息ID" json:"messageId,string"`
	UserID     uint       `gorm:"not null;index:idx_notify_rec_user,priority:1;comment:接收人用户ID" json:"userId,string"`
	ReadStatus int        `gorm:"not null;default:0;index:idx_notify_rec_user,priority:2;comment:0未读 1已读" json:"readStatus"`
	ReadAt     *time.Time `gorm:"comment:阅读时间" json:"readAt,omitempty"`
	TodoStatus int        `gorm:"not null;default:0;index:idx_notify_rec_user,priority:3;comment:0无 1待处理 2已完成 3已过期" json:"todoStatus"`
	DoneAt     *time.Time `gorm:"comment:待办完成时间" json:"doneAt,omitempty"`
}

func (NotifyRecipient) TableName() string { return "notification_recipient" }
