package service

import (
	"context"
	"strings"

	"bit-labs.cn/owl-admin/app-notification/contract"
)

// TextSender 是发一条短信模板消息的最小能力。
// 具体实现由短信子系统注册，通知中心不引用短信子系统的代码。
type TextSender interface {
	SendText(ctx context.Context, templateCode, phone string, variables map[string]string) (messageID string, err error)
}

var (
	defaultSender TextSender
)

// SetTextSender 由短信子系统在启动时注册发送实现。未注册时短信渠道返回「未安装」。
func SetTextSender(sender TextSender) { defaultSender = sender }

// SMSChannel 把通知投递转成一次短信发送。
type SMSChannel struct {
	sender TextSender
}

func NewSMSChannel() *SMSChannel { return &SMSChannel{} }

// SetSender 覆盖本实例的发送实现；未设置时回落到 SetTextSender 注册的全局实现。
func (c *SMSChannel) SetSender(sender TextSender) { c.sender = sender }

func (c *SMSChannel) resolveSender() TextSender {
	if c.sender != nil {
		return c.sender
	}
	return defaultSender
}

func (c *SMSChannel) Code() string { return contract.ChannelSMS }

func (c *SMSChannel) Send(ctx context.Context, req *contract.ChannelRequest) (*contract.ChannelResult, error) {
	sender := c.resolveSender()
	if sender == nil {
		return nil, ChannelMissing("短信子系统未安装")
	}
	if strings.TrimSpace(req.Address) == "" {
		return nil, Invalid("收件人没有手机号")
	}
	if strings.TrimSpace(req.SMSTemplate) == "" {
		return nil, Invalid("未配置短信模板编码")
	}
	messageID, err := sender.SendText(ctx, req.SMSTemplate, req.Address, req.Variables)
	if err != nil {
		return nil, err
	}
	return &contract.ChannelResult{ProviderMsgID: messageID}, nil
}
