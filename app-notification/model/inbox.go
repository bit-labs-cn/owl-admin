package model

import (
	"time"

	"bit-labs.cn/owl/provider/db"
)

// Inbox 是某个主体收到的一条站内信。
type Inbox struct {
	db.BaseModel
	PrincipalType int        `gorm:"not null;default:1;uniqueIndex:uk_notify_inbox,priority:1;index:idx_notify_inbox_principal,priority:1;comment:主体类型" json:"principalType"`
	PrincipalID   uint       `gorm:"not null;default:0;uniqueIndex:uk_notify_inbox,priority:2;index:idx_notify_inbox_principal,priority:2;comment:主体 ID" json:"principalId,string"`
	TaskID        uint       `gorm:"not null;default:0;uniqueIndex:uk_notify_inbox,priority:3;index:idx_notify_inbox_task;comment:任务 ID" json:"taskId,string"`
	Source        string     `gorm:"size:64;not null;default:'';comment:来源子系统" json:"source"`
	Kind          int        `gorm:"not null;default:1;index:idx_notify_inbox_kind;comment:通知类型" json:"kind"`
	Category      int        `gorm:"not null;default:1;comment:分类" json:"category"`
	Title         string     `gorm:"size:255;not null;default:'';comment:标题" json:"title"`
	ContentHTML   string     `gorm:"type:text;comment:HTML 正文" json:"contentHtml"`
	ContentText   string     `gorm:"type:text;comment:纯文本正文" json:"contentText"`
	PayloadJSON   string     `gorm:"type:text;comment:附加内容 JSON" json:"payloadJson"`
	LinkJSON      string     `gorm:"type:text;comment:深链 JSON" json:"linkJson"`
	ActionJSON    string     `gorm:"type:text;comment:动作 JSON" json:"actionJson"`
	BizKey        string     `gorm:"size:128;not null;default:'';index:idx_notify_inbox_biz;comment:业务主键" json:"bizKey"`
	TodoStatus    int        `gorm:"not null;default:0;index:idx_notify_inbox_todo;comment:待办状态" json:"todoStatus"`
	DueAt         *time.Time `gorm:"comment:截止时间" json:"dueAt,omitempty"`
	DoneAt        *time.Time `gorm:"comment:完成时间" json:"doneAt,omitempty"`
	ReadStatus    int        `gorm:"not null;default:1;comment:阅读状态" json:"readStatus"`
	ReadAt        *time.Time `gorm:"comment:阅读时间" json:"readAt,omitempty"`
}

func (Inbox) TableName() string { return "notify_inbox" }
