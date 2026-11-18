package model

import "bit-labs.cn/owl-admin/app-notification/contract"

const (
	KindNotice = contract.KindNotice
	KindTodo   = contract.KindTodo

	TodoStatusNone    = contract.TodoStatusNone
	TodoStatusPending = contract.TodoStatusPending
	TodoStatusDone    = contract.TodoStatusDone
	TodoStatusExpired = contract.TodoStatusExpired

	ReadStatusUnread = contract.ReadStatusUnread
	ReadStatusRead   = contract.ReadStatusRead

	ChannelInApp    = contract.ChannelInApp
	ChannelEmail    = contract.ChannelEmail
	ChannelSMS      = contract.ChannelSMS
	ChannelWeCom    = contract.ChannelWeCom
	ChannelDingTalk = contract.ChannelDingTalk

	RecipientTypeUser      = contract.RecipientTypeUser
	RecipientTypeExternal  = contract.RecipientTypeExternal
	RecipientTypeBroadcast = contract.RecipientTypeBroadcast

	TypeStatusEnabled  = contract.TypeStatusEnabled
	TypeStatusDisabled = contract.TypeStatusDisabled

	DeliveryStatusPending = contract.DeliveryStatusPending
	DeliveryStatusSuccess = contract.DeliveryStatusSuccess
	DeliveryStatusFailed  = contract.DeliveryStatusFailed
	DeliveryStatusSkipped = contract.DeliveryStatusSkipped

	DeliveryMaxAttempts = contract.DeliveryMaxAttempts

	ScheduleStatusPending  = contract.ScheduleStatusPending
	ScheduleStatusSending  = contract.ScheduleStatusSending
	ScheduleStatusSent     = contract.ScheduleStatusSent
	ScheduleStatusCanceled = contract.ScheduleStatusCanceled
	ScheduleStatusFailed   = contract.ScheduleStatusFailed
	ScheduleMaxAttempts    = contract.ScheduleMaxAttempts

	LevelInfo    = contract.LevelInfo
	LevelWarning = contract.LevelWarning
	LevelDanger  = contract.LevelDanger

	DisplayModeInbox  = contract.DisplayModeInbox
	DisplayModeBanner = contract.DisplayModeBanner
	DisplayModePopup  = contract.DisplayModePopup

	ScopeAll  = contract.ScopeAll
	ScopeDept = contract.ScopeDept
	ScopeRole = contract.ScopeRole

	AnnouncementStatusDraft     = 1
	AnnouncementStatusPublished = 2
	AnnouncementStatusOffline   = 3
)
