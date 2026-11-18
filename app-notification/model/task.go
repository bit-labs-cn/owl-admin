package model

import (
	"time"

	"bit-labs.cn/owl/provider/db"
)

// Task 是一次通知发送，手工发起和子系统调用共用。
type Task struct {
	db.BaseModel
	BizNo          string     `gorm:"size:64;not null;default:'';index:idx_notify_task_biz;comment:业务号" json:"bizNo"`
	Title          string     `gorm:"size:255;not null;default:'';comment:标题" json:"title"`
	EventCode      string     `gorm:"size:64;not null;default:'';index:idx_notify_task_event;comment:事件编码" json:"eventCode"`
	Source         string     `gorm:"size:64;not null;default:'';comment:来源子系统" json:"source"`
	Category       int        `gorm:"not null;default:1;comment:分类" json:"category"`
	Kind           int        `gorm:"not null;default:1;index:idx_notify_task_kind;comment:通知类型" json:"kind"`
	ChannelsJSON   string     `gorm:"type:text;comment:渠道 JSON" json:"channelsJson"`
	AudienceJSON   string     `gorm:"type:text;comment:受众 JSON" json:"audienceJson"`
	TemplateCode   string     `gorm:"size:64;not null;default:'';comment:模板编码" json:"templateCode"`
	ContentHTML    string     `gorm:"type:text;comment:HTML 正文" json:"contentHtml"`
	ContentText    string     `gorm:"type:text;comment:纯文本正文" json:"contentText"`
	VariableJSON   string     `gorm:"type:text;comment:变量 JSON" json:"variableJson"`
	PayloadJSON    string     `gorm:"type:text;comment:附加内容 JSON" json:"payloadJson"`
	LinkJSON       string     `gorm:"type:text;comment:深链 JSON" json:"linkJson"`
	ActionJSON     string     `gorm:"type:text;comment:动作 JSON" json:"actionJson"`
	SMSTemplate    string     `gorm:"size:64;not null;default:'';comment:短信模板编码" json:"smsTemplate"`
	BizKey         string     `gorm:"size:128;not null;default:'';index:idx_notify_task_biz_key;comment:业务主键" json:"bizKey"`
	DueAt          *time.Time `gorm:"comment:待办截止时间" json:"dueAt,omitempty"`
	ScheduleAt     *time.Time `gorm:"index:idx_notify_task_schedule;comment:定时发送时间" json:"scheduleAt,omitempty"`
	Status         int        `gorm:"not null;default:1;index:idx_notify_task_status;comment:任务状态" json:"status"`
	IdempotencyKey string     `gorm:"size:128;not null;default:'';index:idx_notify_task_idem;comment:幂等键" json:"idempotencyKey"`
	SuccessCount   int        `gorm:"not null;default:0;comment:成功数" json:"successCount"`
	FailCount      int        `gorm:"not null;default:0;comment:失败数" json:"failCount"`
	SkipCount      int        `gorm:"not null;default:0;comment:跳过数" json:"skipCount"`
	Origin         string     `gorm:"size:255;not null;default:'';comment:站点公网地址" json:"origin"`
}

func (Task) TableName() string { return "notify_task" }
