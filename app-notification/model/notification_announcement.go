package model

import (
	"time"

	"bit-labs.cn/owl/provider/db"
)

// NotifyAnnouncement 站点公告本体表（读扩散存储）
type NotifyAnnouncement struct {
	db.BaseModel
	Title        string     `gorm:"size:255;not null;default:'';comment:公告标题" json:"title"`
	Content      string     `gorm:"type:text;comment:公告正文" json:"content"`
	Level        int        `gorm:"not null;default:1;comment:级别 1普通 2警告 3紧急" json:"level"`
	DisplayMode  int        `gorm:"not null;default:1;comment:展示方式 1仅公告中心 2横幅 3强提醒弹窗" json:"displayMode"`
	TargetScope  string     `gorm:"size:32;not null;default:'all';comment:受众范围 all/dept/role" json:"targetScope"`
	TargetValues string     `gorm:"type:text;comment:目标受众ID列表 JSON" json:"targetValues"`
	StartAt      *time.Time `gorm:"index:idx_notify_ann_time,priority:1;comment:生效开始时间" json:"startAt,omitempty"`
	EndAt        *time.Time `gorm:"index:idx_notify_ann_time,priority:2;comment:生效结束时间" json:"endAt,omitempty"`
	Status       int        `gorm:"not null;default:2;index:idx_notify_ann_status;comment:1草稿 2已发布 3已下线" json:"status"`
	CreatedBy    uint       `gorm:"not null;default:0;comment:创建人" json:"createdBy,string"`
}

func (NotifyAnnouncement) TableName() string { return "notification_announcement" }

// NotifyAnnouncementState 公告用户已读与关闭状态表
type NotifyAnnouncementState struct {
	db.BaseModel
	AnnouncementID uint       `gorm:"not null;uniqueIndex:uk_notify_ann_user,priority:1;comment:公告ID" json:"announcementId,string"`
	UserID         uint       `gorm:"not null;uniqueIndex:uk_notify_ann_user,priority:2;comment:用户ID" json:"userId,string"`
	IsDismissed    int        `gorm:"not null;default:0;comment:是否关闭横幅/弹窗" json:"isDismissed"`
	ReadAt         *time.Time `gorm:"comment:阅读时间" json:"readAt,omitempty"`
}

func (NotifyAnnouncementState) TableName() string { return "notification_announcement_state" }
