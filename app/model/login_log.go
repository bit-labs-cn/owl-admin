package model

import "time"

type LoginLog struct {
	Id        int        `json:"id,string" gorm:"primaryKey"`
	Ip        string     `json:"ip" gorm:"column:ip"`
	LoginTime int        `json:"loginTime" gorm:"column:login_time"`
	UserId    int        `json:"userId" gorm:"column:user_id"`
	UserName  string     `json:"userName" gorm:"column:user_name"`
	UserType  string     `json:"userType" gorm:"column:user_type"`
	UserAgent string     `json:"userAgent" gorm:"column:user_agent"`
	CreatedAt *time.Time `json:"createdAt,omitempty"`
}

func (i LoginLog) TableName() string {
	return "admin_login_log"
}
