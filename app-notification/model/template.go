package model

import "bit-labs.cn/owl/provider/db"

// Template 是按渠道准备的通知模板。
type Template struct {
	db.BaseModel
	Code         string `gorm:"size:64;not null;default:'';uniqueIndex:uk_notify_template_code;comment:模板编码" json:"code"`
	Name         string `gorm:"size:128;not null;default:'';comment:模板名称" json:"name"`
	EventCode    string `gorm:"size:64;not null;default:'';index:idx_notify_template_event;comment:事件编码" json:"eventCode"`
	Category     int    `gorm:"not null;default:1;comment:分类" json:"category"`
	Kind         int    `gorm:"not null;default:1;comment:通知类型" json:"kind"`
	Title        string `gorm:"size:255;not null;default:'';comment:标题" json:"title"`
	ContentHTML  string `gorm:"type:text;comment:HTML 正文" json:"contentHtml"`
	ContentText  string `gorm:"type:text;comment:纯文本正文" json:"contentText"`
	SMSTemplate  string `gorm:"size:64;not null;default:'';comment:短信模板编码" json:"smsTemplate"`
	VariableJSON string `gorm:"type:text;comment:变量定义 JSON" json:"variableJson"`
	LinkJSON     string `gorm:"type:text;comment:默认深链 JSON" json:"linkJson"`
	ActionJSON   string `gorm:"type:text;comment:默认动作 JSON" json:"actionJson"`
	Status       int    `gorm:"not null;default:1;comment:状态" json:"status"`
	Remark       string `gorm:"size:255;not null;default:'';comment:备注" json:"remark"`
}

func (Template) TableName() string { return "notify_template" }
