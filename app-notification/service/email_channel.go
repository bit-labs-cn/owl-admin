package service

import (
	"context"
	"encoding/json"
	"strings"

	"bit-labs.cn/owl-admin/app-notification/contract"
	mailcontract "bit-labs.cn/owl/contract/mailer"
	"bit-labs.cn/owl/provider/mailer"
)

// EmailChannel 通过框架邮件驱动发信。
type EmailChannel struct {
	mailer *mailer.MailerManager
}

func NewEmailChannel() *EmailChannel { return &EmailChannel{} }

func (c *EmailChannel) SetMailer(m *mailer.MailerManager) { c.mailer = m }

func (c *EmailChannel) Code() string { return contract.ChannelEmail }

func (c *EmailChannel) Send(ctx context.Context, req *contract.ChannelRequest) (*contract.ChannelResult, error) {
	if c.mailer == nil {
		return nil, ChannelMissing("邮件驱动未安装")
	}
	if strings.TrimSpace(req.Address) == "" {
		return nil, Invalid("收件人没有邮箱")
	}
	msg := &mailcontract.Message{
		To:      []string{req.Address},
		Subject: req.Title,
		Body:    req.ContentText,
		HTML:    wrapEmailHTML(req.ContentHTML),
	}
	if err := c.mailer.Send(ctx, msg); err != nil {
		return nil, err
	}
	return &contract.ChannelResult{}, nil
}

func wrapEmailHTML(content string) string {
	return `<div style="font-family:sans-serif;font-size:14px;line-height:1.6;color:#222">` + content + `</div>`
}

func decodeConfig[T any](raw string) (T, error) {
	var out T
	if strings.TrimSpace(raw) == "" {
		return out, Invalid("渠道账号未配置")
	}
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return out, Invalid("渠道配置不是合法 JSON")
	}
	return out, nil
}
