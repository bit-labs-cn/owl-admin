package model

import (
	"time"

	"bit-labs.cn/owl/provider/db"
)

// Announcement 是站点公告，和每人一条的站内信分开管理。
type Announcement struct {
	db.BaseModel
	Title        string     `gorm:"size:255;not null;default:'';comment:标题" json:"title"`
	ContentHTML  string     `gorm:"type:text;comment:HTML 正文" json:"contentHtml"`
	ContentText  string     `gorm:"type:text;comment:纯文本正文" json:"contentText"`
	Level        int        `gorm:"not null;default:1;comment:级别" json:"level"`
	DisplayMode  int        `gorm:"not null;default:1;comment:展示方式" json:"displayMode"`
	AudienceJSON string     `gorm:"type:text;comment:受众 JSON" json:"audienceJson"`
	StartAt      *time.Time `gorm:"comment:生效开始" json:"startAt,omitempty"`
	EndAt        *time.Time `gorm:"comment:生效结束" json:"endAt,omitempty"`
	Closable     int        `gorm:"not null;default:1;comment:可否关闭" json:"closable"`
	Frequency    int        `gorm:"not null;default:1;comment:展示频率" json:"frequency"`
	Status       int        `gorm:"not null;default:1;index:idx_notify_announcement_status;comment:状态" json:"status"`
	Sort         int        `gorm:"not null;default:0;comment:排序" json:"sort"`
}

func (Announcement) TableName() string { return "notify_announcement" }

// AnnouncementState 记录某个主体关闭或看过某条公告。
type AnnouncementState struct {
	db.BaseModel
	PrincipalType  int        `gorm:"not null;default:1;uniqueIndex:uk_notify_ann_state,priority:1;comment:主体类型" json:"principalType"`
	PrincipalID    uint       `gorm:"not null;default:0;uniqueIndex:uk_notify_ann_state,priority:2;comment:主体 ID" json:"principalId,string"`
	AnnouncementID uint       `gorm:"not null;default:0;uniqueIndex:uk_notify_ann_state,priority:3;comment:公告 ID" json:"announcementId,string"`
	DismissedAt    *time.Time `gorm:"comment:关闭时间" json:"dismissedAt,omitempty"`
	LastShownAt    *time.Time `gorm:"comment:最近展示时间" json:"lastShownAt,omitempty"`
}

func (AnnouncementState) TableName() string { return "notify_announcement_state" }
