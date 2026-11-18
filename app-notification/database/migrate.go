package database

import (
	"bit-labs.cn/owl-admin/app-notification/model"
)

// Models 返回企业级通知中心 4 张核心表模型。
func Models() []any {
	return []any{
		&model.NotifyMessage{},
		&model.NotifyRecipient{},
		&model.NotifyAnnouncement{},
		&model.NotifyAnnouncementState{},
		&model.NotifyType{},
		&model.NotifyTemplate{},
		&model.NotifyDelivery{},
		&model.NotifySchedule{},
	}
}
