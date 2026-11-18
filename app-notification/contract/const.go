package contract

// 核心类型定义：通知与待办
const (
	// KindNotice 知会通知（发生了什么：只读，阅后即终结）
	KindNotice = 1
	// KindTodo 行动待办（该做什么：待处理/已完成/已过期，必须闭环）
	KindTodo = 2
)

// 待办状态
const (
	TodoStatusNone    = 0
	TodoStatusPending = 1 // 待处理
	TodoStatusDone    = 2 // 已完成
	TodoStatusExpired = 3 // 已过期
)

// 阅读状态
const (
	ReadStatusUnread = 0 // 未读
	ReadStatusRead   = 1 // 已读
)

// 常用渠道标识
const (
	ChannelInApp    = "inapp"    // 站内信与实时推送 (默认必备)
	ChannelEmail    = "email"    // 邮件
	ChannelSMS      = "sms"      // 短信
	ChannelWeCom    = "wecom"    // 企业微信机器人
	ChannelDingTalk = "dingtalk" // 钉钉机器人
)

// 接收人类型
const (
	RecipientTypeUser      = 1 // 站内用户
	RecipientTypeExternal  = 2 // 外部人员
	RecipientTypeBroadcast = 3 // 群机器人
)

// 消息类型状态
const (
	TypeStatusEnabled  = 1
	TypeStatusDisabled = 2
)

// 外部渠道发送记录状态
const (
	DeliveryStatusPending = 1
	DeliveryStatusSuccess = 2
	DeliveryStatusFailed  = 3
	DeliveryStatusSkipped = 4
)

// DeliveryMaxAttempts 是外部渠道暂时失败后的最多发送次数。
const DeliveryMaxAttempts = 5

// 定时发送任务状态
const (
	ScheduleStatusPending  = 1 // 待发送，到点后由调度协程取走
	ScheduleStatusSending  = 2 // 发送中，已被某个实例领取
	ScheduleStatusSent     = 3 // 已发送
	ScheduleStatusCanceled = 4 // 已取消，不再发送
	ScheduleStatusFailed   = 5 // 发送失败，可手动重试
)

// ScheduleMaxAttempts 是定时任务取走后发送失败的最多尝试次数。
const ScheduleMaxAttempts = 3

// 公告级别
const (
	LevelInfo    = 1 // 普通知会
	LevelWarning = 2 // 警告
	LevelDanger  = 3 // 紧急
)

// 公告展示方式
const (
	DisplayModeInbox  = 1 // 仅公告中心展示
	DisplayModeBanner = 2 // 顶部横幅曝光
	DisplayModePopup  = 3 // 强提醒弹窗
)

// 公告受众范围
const (
	ScopeAll  = "all"  // 全员
	ScopeDept = "dept" // 指定部门
	ScopeRole = "role" // 指定角色
)
