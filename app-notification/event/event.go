package event

import "bit-labs.cn/owl-admin/app-notification/contract"

const (
	// TopicNotifyDirect 定向消息发送事件（通知或待办）
	TopicNotifyDirect = "notify.direct.send"

	// TopicNotifyTodoComplete 业务完成勾销待办事件
	TopicNotifyTodoComplete = "notify.todo.complete"

	// TopicNotifyAnnouncement 全员/定向公告广播事件
	TopicNotifyAnnouncement = "notify.announcement.broadcast"

	// TopicNotifyChannelSend 外部渠道发送事件。载荷只带发送记录 ID。
	TopicNotifyChannelSend = "notify.channel.send"
)

// ChannelSendPayload 是外部渠道发送事件的载荷。
type ChannelSendPayload struct {
	DeliveryID uint `json:"deliveryId"`
}

// DirectMessagePayload 定向消息事件载荷
type DirectMessagePayload = contract.DirectMessage

// TodoCompletePayload 待办完成事件载荷
type TodoCompletePayload struct {
	BizKey string `json:"bizKey"`
}

// AnnouncementPayload 公告广播事件载荷
type AnnouncementPayload = contract.AnnouncementMessage
