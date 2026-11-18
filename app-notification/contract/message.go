package contract

import (
	"time"
)

// DirectMessage 定向消息（通知或待办）
type DirectMessage struct {
	Kind         int            `json:"kind"`                   // KindNotice(1) 或 KindTodo(2)
	Title        string         `json:"title"`                  // 标题
	Content      string         `json:"content"`                // 详细说明或正文
	Link         string         `json:"link,omitempty"`         // 跳转深链（例如 /inspection/task/1024）
	BizKey       string         `json:"bizKey,omitempty"`       // 业务唯一主键（例如 inspection:1024，用于自动销项）
	DueAt        *time.Time     `json:"dueAt,omitempty"`        // 待办截止时间
	UserIDs      []uint         `json:"userIds"`                // 目标接收人列表
	Channels     []string       `json:"channels,omitempty"`     // 本条指定渠道，如 ["sms","wecom","dingtalk"]；空则用类型默认，再空则站内信
	TypeCode     string         `json:"typeCode,omitempty"`     // 消息类型编码，空则不套用模板
	Vars         map[string]any `json:"vars,omitempty"`         // 模板变量
	SkipTemplate bool           `json:"skipTemplate,omitempty"` // 跳过类型模板，直接用 Title、Content、Link（人工发送使用）
	SenderID     uint           `json:"senderId,omitempty"`     // 发送人（0 表示系统）
	Payload      map[string]any `json:"payload,omitempty"`      // 附加业务元数据

	// ExternalRecipients 是没有系统账号的外部人员，只能走邮件等外部渠道
	ExternalRecipients []ExternalRecipient `json:"externalRecipients,omitempty"`
}

// ExternalRecipient 是没有系统账号的外部人员。
type ExternalRecipient struct {
	Name  string `json:"name"`            // 称呼，用于发送记录展示
	Email string `json:"email,omitempty"` // 邮箱
	Phone string `json:"phone,omitempty"` // 手机号
}

// AnnouncementMessage 公告消息（全员或定向广播）
type AnnouncementMessage struct {
	Title        string         `json:"title"`                  // 公告标题
	Content      string         `json:"content"`                // 公告正文 (Markdown/富文本)
	Level        int            `json:"level"`                  // 级别: LevelInfo(1), LevelWarning(2), LevelDanger(3)
	DisplayMode  int            `json:"displayMode"`            // 展示方式: DisplayModeInbox(1), DisplayModeBanner(2), DisplayModePopup(3)
	Scope        string         `json:"scope"`                  // ScopeAll("all"), ScopeDept("dept"), ScopeRole("role")
	TargetValues []uint         `json:"targetValues,omitempty"` // 目标受众ID列表（如部门ID或角色ID）
	StartAt      *time.Time     `json:"startAt,omitempty"`      // 生效开始时间
	EndAt        *time.Time     `json:"endAt,omitempty"`        // 生效结束时间
	SenderID     uint           `json:"senderId,omitempty"`     // 发布人ID
	Payload      map[string]any `json:"payload,omitempty"`
}

// ChannelRequest 传递给外部渠道插件的单条外发请求
type ChannelRequest struct {
	UserID  uint           `json:"userId"`
	Email   string         `json:"email,omitempty"`
	Phone   string         `json:"phone,omitempty"`
	Title   string         `json:"title"`
	Content string         `json:"content"`
	// ContentText 是同一份正文的纯文本版，给邮件的 text/plain 备用部分用。
	ContentText string         `json:"contentText,omitempty"`
	Link        string         `json:"link,omitempty"`
	Kind        int            `json:"kind"`
	Webhook     string         `json:"webhook,omitempty"`
	Payload     map[string]any `json:"payload,omitempty"`
}
