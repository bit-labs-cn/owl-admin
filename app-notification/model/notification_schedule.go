package model

import (
	"time"

	"bit-labs.cn/owl/provider/db"
)

// NotifySchedule 是一条定时发送任务：到 PlanAt 时按保存的受众与渠道发送。
// Audience 存的是选择条件（用户、角色、部门、用户组、外部人员），到点才解析成具体用户，
// 这样计划期间入职、调岗的人也能按发送当时的组织关系收到。
type NotifySchedule struct {
	db.BaseModel
	Kind     int        `gorm:"not null;default:1;comment:1通知 2待办" json:"kind"`
	TypeCode string     `gorm:"size:64;not null;default:'';comment:消息类型编码" json:"typeCode"`
	Title    string     `gorm:"size:255;not null;default:'';comment:标题" json:"title"`
	Content  string     `gorm:"type:text;comment:正文说明" json:"content"`
	Link     string     `gorm:"size:255;not null;default:'';comment:跳转链接" json:"link"`
	Channels string     `gorm:"type:text;comment:渠道 JSON" json:"channels"`
	Audience string     `gorm:"type:text;comment:受众选择条件 JSON" json:"audience"`
	DueAt    *time.Time `gorm:"comment:待办截止时间" json:"dueAt,omitempty"`
	PlanAt   time.Time  `gorm:"not null;index:idx_notify_schedule_plan;comment:计划发送时间" json:"planAt"`

	// AudienceSummary 与 RecipientCount 是计划时解析出的接收人快照，
	// 供管理端列表直接展示可读信息，不必回显裸 ID，也不必每次查询再解析一遍受众。
	AudienceSummary string `gorm:"size:512;not null;default:'';comment:接收人摘要" json:"audienceSummary"`
	RecipientCount  int    `gorm:"not null;default:0;comment:计划时解析出的接收人数" json:"recipientCount"`

	Status    int        `gorm:"not null;default:1;index:idx_notify_schedule_status;comment:1待发送 2发送中 3已发送 4已取消 5失败" json:"status"`
	Attempts  int        `gorm:"not null;default:0;comment:已尝试发送次数" json:"attempts"`
	LastError string     `gorm:"size:512;not null;default:'';comment:最后错误" json:"lastError"`
	MessageID uint       `gorm:"not null;default:0;comment:发送成功后生成的消息ID" json:"messageId,string"`
	SenderID  uint       `gorm:"not null;default:0;comment:发送人ID" json:"senderId,string"`
	SentAt    *time.Time `gorm:"comment:实际发送时间" json:"sentAt,omitempty"`
}

func (NotifySchedule) TableName() string { return "notification_schedule" }
