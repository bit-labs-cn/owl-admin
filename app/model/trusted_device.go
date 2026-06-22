package model

import (
	"time"

	"bit-labs.cn/owl/provider/db"
)

// TrustedDevice 用户可信设备（通过验证码验证后登记）
type TrustedDevice struct {
	db.BaseModel
	UserID      uint      `gorm:"comment:用户ID;index:idx_user_device,unique,priority:1" json:"userId,string"`
	DeviceID    string    `gorm:"comment:设备指纹;type:string;size:64;index:idx_user_device,unique,priority:2" json:"deviceId"`
	LastIP      string    `gorm:"comment:最近登录IP;type:string;size:64" json:"lastIp"`
	LastUA      string    `gorm:"comment:最近UserAgent;type:string;size:512" json:"lastUa"`
	VerifiedAt  time.Time `gorm:"comment:最近验证时间" json:"verifiedAt"`
	LastLoginAt time.Time `gorm:"comment:最近登录时间" json:"lastLoginAt"`
}

func (TrustedDevice) TableName() string {
	return "admin_trusted_device"
}
