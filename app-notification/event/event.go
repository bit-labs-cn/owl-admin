package event

// 通知子系统的事件主题定义。
// 通过统一事件总线（owl Event Bus）进行跨协程与跨实例的异步削峰、扇出与投递。
const (
	// TopicTaskSubmit 通知任务提交事件（触发异步受众解析与扇出）
	TopicTaskSubmit = "notify.task.submit"

	// TopicDispatchEmail 邮件渠道单次投递事件
	TopicDispatchEmail = "notify.dispatch.email"

	// TopicDispatchSMS 短信渠道单次投递事件
	TopicDispatchSMS = "notify.dispatch.sms"

	// TopicDispatchApp 移动推送渠道单次投递事件（APNs / FCM）
	TopicDispatchApp = "notify.dispatch.app"

	// TopicTodoComplete 业务完成勾销待办事件
	TopicTodoComplete = "notify.todo.complete"
)

// TaskSubmitPayload 任务提交事件载荷
type TaskSubmitPayload struct {
	TaskID uint `json:"taskId"`
}

// DispatchPayload 单条渠道投递事件载荷
type DispatchPayload struct {
	DeliveryID    uint   `json:"deliveryId"`
	TaskID        uint   `json:"taskId"`
	RecipientType int    `json:"recipientType"`
	RecipientID   uint   `json:"recipientId"`
	Channel       string `json:"channel"`
	Address       string `json:"address"`
}

// TodoCompletePayload 待办完成事件载荷
type TodoCompletePayload struct {
	BizKey string `json:"bizKey"`
}
