package contract

import "context"

// ChannelSender 是外部渠道插件（如邮件、企业微信、钉钉等）的统一扩展接口 (SPI)。
type ChannelSender interface {
	// ChannelCode 渠道编码，如 "email", "sms", "wecom", "dingtalk"
	ChannelCode() string

	// Send 执行向单人发送
	Send(ctx context.Context, req *ChannelRequest) error
}
