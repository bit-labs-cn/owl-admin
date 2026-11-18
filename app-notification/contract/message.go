package contract

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

// 通知类型。告诉用户发生了什么，或需要做什么。
const (
	KindNotice  = 1 // 系统通知
	KindMessage = 2 // 业务消息
	KindTodo    = 3 // 待办
)

// 通知分类。决定能否退订，安全类强制送达。
const (
	CategorySystem    = 1
	CategorySecurity  = 2
	CategoryBiz       = 3
	CategoryMarketing = 4
)

// 渠道编码。新增渠道时实现 Channel 并用这些编码之外的新编码注册。
const (
	ChannelInApp = "inapp"
	ChannelEmail = "email"
	ChannelSMS   = "sms"
	ChannelApp   = "app"
)

// 动作类型。
const (
	ActionTypeLink     = 1 // 站内深链
	ActionTypeExternal = 2 // 外链
	ActionTypeAPI      = 3 // 预留：通知内直接调接口
)

// DeepLink 是跳到具体业务页的结构化地址，不从 HTML 里解析。
type DeepLink struct {
	App      string            `json:"app,omitempty"`
	Path     string            `json:"path,omitempty"`
	Query    map[string]string `json:"query,omitempty"`
	External bool              `json:"external,omitempty"`
	Target   string            `json:"target,omitempty"`
}

// Action 是通知上的操作按钮。
type Action struct {
	Key   string    `json:"key"`
	Label string    `json:"label"`
	Type  int       `json:"type"`
	Link  *DeepLink `json:"link,omitempty"`
	Style string    `json:"style,omitempty"`
}

// UintIDs 同时接受 JSON 数字和字符串，避免雪花 ID 在前端丢精度。
type UintIDs []uint

func (u *UintIDs) UnmarshalJSON(data []byte) error {
	text := strings.TrimSpace(string(data))
	if text == "" || text == "null" {
		return nil
	}
	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	out := make([]uint, 0, len(raw))
	for _, item := range raw {
		value := strings.Trim(string(item), `"`)
		n, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			continue
		}
		out = append(out, uint(n))
	}
	*u = out
	return nil
}

// Audience 描述接收范围。各 ID 列表取并集；All 表示全体有效用户。
type Audience struct {
	All        bool    `json:"all,omitempty"`
	UserIDs    UintIDs `json:"userIds,omitempty"`
	GroupIDs   UintIDs `json:"groupIds,omitempty"`
	DeptIDs    UintIDs `json:"deptIds,omitempty"`
	RoleIDs    UintIDs `json:"roleIds,omitempty"`
	ContactIDs UintIDs `json:"contactIds,omitempty"`
	ListIDs    UintIDs `json:"listIds,omitempty"`
}

// Message 是一次通知。HTML 负责展示，Link 负责跳转。
type Message struct {
	EventCode      string
	Source         string
	Kind           int
	Category       int
	Title          string
	ContentHTML    string
	ContentText    string
	Payload        map[string]string
	Link           *DeepLink
	Actions        []Action
	BizKey         string
	DueAt          *time.Time
	Audience       Audience
	Channels       []string
	ScheduleAt     *time.Time
	IdempotencyKey string
	Variables      map[string]string
	TemplateCode   string
}

// EventDescriptor 由业务子系统声明自己会发出的通知事件。
type EventDescriptor struct {
	Code            string   `json:"code"`
	Name            string   `json:"name"`
	Source          string   `json:"source"`
	Kind            int      `json:"kind"`
	Category        int      `json:"category"`
	DefaultChannels []string `json:"defaultChannels"`
	Description     string   `json:"description"`
}

// ChannelRequest 是渠道实现收到的一条投递。
type ChannelRequest struct {
	TaskID       uint
	DeliveryID   uint
	Title        string
	ContentHTML  string
	ContentText  string
	Address      string
	Variables    map[string]string
	Provider     string
	ConfigJSON   string
	SMSTemplate  string
	Payload      map[string]string
	Link         *DeepLink
	Kind         int
	DeviceTokens []DeviceTarget
}

// DeviceTarget 是一台已登记的推送设备。
type DeviceTarget struct {
	ID       uint
	Provider string
	Token    string
	Platform string
}

// ChannelResult 是渠道发送结果。InvalidTokens 会被标成失效，不再重试。
type ChannelResult struct {
	ProviderMsgID string
	InvalidTokens []string
}

// Channel 是可扩展的站外渠道。
type Channel interface {
	Code() string
	Send(ctx context.Context, req *ChannelRequest) (*ChannelResult, error)
}
