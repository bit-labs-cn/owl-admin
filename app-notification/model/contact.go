package model

import (
	"time"

	"bit-labs.cn/owl/provider/db"
)

// ContactList 是外部联系人名单。
type ContactList struct {
	db.BaseModel
	Name   string `gorm:"size:128;not null;default:'';comment:名单名称" json:"name"`
	Remark string `gorm:"size:255;not null;default:'';comment:备注" json:"remark"`
	Status int    `gorm:"not null;default:1;comment:状态" json:"status"`
}

func (ContactList) TableName() string { return "notify_contact_list" }

// Contact 是名单里的一个外部联系人。
type Contact struct {
	db.BaseModel
	ListID         uint       `gorm:"not null;default:0;index:idx_notify_contact_list;comment:名单 ID" json:"listId,string"`
	Name           string     `gorm:"size:128;not null;default:'';comment:名称" json:"name"`
	Email          string     `gorm:"size:255;not null;default:'';index:idx_notify_contact_email;comment:邮箱" json:"email"`
	Phone          string     `gorm:"size:32;not null;default:'';index:idx_notify_contact_phone;comment:手机号" json:"phone"`
	Status         int        `gorm:"not null;default:1;comment:订阅状态" json:"status"`
	Source         string     `gorm:"size:64;not null;default:'';comment:来源" json:"source"`
	UnsubscribedAt *time.Time `gorm:"comment:退订时间" json:"unsubscribedAt,omitempty"`
}

func (Contact) TableName() string { return "notify_contact" }
