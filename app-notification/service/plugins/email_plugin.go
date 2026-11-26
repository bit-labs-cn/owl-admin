package plugins

import (
	"context"
	"strings"

	"bit-labs.cn/owl-admin/app-notification/contract"
	mailcontract "bit-labs.cn/owl/contract/mailer"
	"bit-labs.cn/owl/provider/mailer"
)

type EmailPlugin struct {
	mailer *mailer.MailerManager
}

func NewEmailPlugin(m *mailer.MailerManager) *EmailPlugin {
	return &EmailPlugin{mailer: m}
}

func (p *EmailPlugin) ChannelCode() string {
	return contract.ChannelEmail
}

func (p *EmailPlugin) Send(ctx context.Context, req *contract.ChannelRequest) error {
	if p.mailer == nil || strings.TrimSpace(req.Email) == "" {
		return nil
	}

	msg := &mailcontract.Message{
		To:      []string{req.Email},
		Subject: req.Title,
		Body:    req.ContentText,
		HTML:    req.Content,
	}

	return p.mailer.Send(ctx, msg)
}
