package model

import "bit-labs.cn/owl/provider/db"

// NotifyTemplate 是某个消息类型的文案模板。
// 模板不带渠道：一条消息要发到哪些渠道在发送时决定，模板只描述标题、正文和链接。
type NotifyTemplate struct {
	db.BaseModel
	TypeCode string `gorm:"size:64;not null;uniqueIndex:uk_notify_tpl;comment:类型编码" json:"typeCode"`
	Title    string `gorm:"size:255;not null;default:'';comment:标题模板" json:"title"`
	Body     string `gorm:"type:text;comment:正文模板（富文本）" json:"body"`
	Link     string `gorm:"size:255;not null;default:'';comment:跳转链接模板" json:"link"`
}

func (NotifyTemplate) TableName() string { return "notification_template" }
