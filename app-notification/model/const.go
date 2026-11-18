package model

import "bit-labs.cn/owl-admin/app-notification/contract"

const (
	PrincipalTypeAdmin  = 1 // 后台用户
	PrincipalTypePortal = 2 // 门户 / C 端用户

	RecipientTypeUser    = 1
	RecipientTypeContact = 2

	StatusEnabled  = 1
	StatusDisabled = 2

	TaskStatusDraft         = 1
	TaskStatusPendingReview = 2
	TaskStatusPending       = 3
	TaskStatusSending       = 4
	TaskStatusPartial       = 5
	TaskStatusSuccess       = 6
	TaskStatusFailed        = 7
	TaskStatusCancelled     = 8

	DeliveryStatusPending = 1
	DeliveryStatusSending = 2
	DeliveryStatusSent    = 3
	DeliveryStatusFailed  = 4
	DeliveryStatusSkipped = 5

	TodoStatusPending = 1
	TodoStatusDone    = 2
	TodoStatusExpired = 3

	DeviceStatusActive  = 1
	DeviceStatusInvalid = 2

	ContactStatusSubscribed   = 1
	ContactStatusUnsubscribed = 2
	ContactStatusBounced      = 3

	AnnouncementStatusDraft     = 1
	AnnouncementStatusPublished = 2
	AnnouncementStatusOffline   = 3

	DisplayModeBanner = 1
	DisplayModePopup  = 2
	DisplayModeInbox  = 3

	FrequencyOnce    = 1
	FrequencySession = 2
	FrequencyLogin   = 3

	LevelInfo    = 1
	LevelSuccess = 2
	LevelWarning = 3
	LevelDanger  = 4

	ReadStatusUnread = 1
	ReadStatusRead   = 2
)

const (
	KindNotice  = contract.KindNotice
	KindMessage = contract.KindMessage
	KindTodo    = contract.KindTodo

	CategorySystem    = contract.CategorySystem
	CategorySecurity  = contract.CategorySecurity
	CategoryBiz       = contract.CategoryBiz
	CategoryMarketing = contract.CategoryMarketing

	ChannelInApp = contract.ChannelInApp
	ChannelEmail = contract.ChannelEmail
	ChannelSMS   = contract.ChannelSMS
	ChannelApp   = contract.ChannelApp
)
