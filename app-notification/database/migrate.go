package database

import "bit-labs.cn/owl-admin/app-notification/model"

// Models 返回通知中心待自动迁移的表。
func Models() []any {
	return []any{
		&model.Template{},
		&model.Task{},
		&model.Delivery{},
		&model.Inbox{},
		&model.Device{},
		&model.Announcement{},
		&model.AnnouncementState{},
		&model.Subscription{},
		&model.Setting{},
		&model.ContactList{},
		&model.Contact{},
		&model.ChannelAccount{},
		&model.Event{},
	}
}
