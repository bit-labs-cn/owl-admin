package service

import (
	"context"
	"strings"

	"bit-labs.cn/owl/provider/setting"
)

const (
	settingSubsystem = "notify"

	keyEmailEnabled = "emailEnabled"
	keyWecomEnabled = "wecomEnabled"
	keyWecomWebhook = "wecomWebhook"
	keyBaseURL      = "baseUrl"
)

// NotifySettingCatalog 是通知中心写进系统设置的渠道开关。
func NotifySettingCatalog() setting.Catalog {
	return setting.Catalog{
		Title: "通知中心",
		Groups: []setting.Group{
			{
				Key:   "channel",
				Title: "渠道",
				Items: []setting.Item{
					{
						Key: keyEmailEnabled, Title: "邮件", Hint: "关闭后，新消息不再发邮件。",
						Type: setting.TypeBool, Default: "true",
					},
					{
						Key: keyWecomEnabled, Title: "企业微信", Hint: "关闭后，新消息不再发到企业微信群机器人。",
						Type: setting.TypeBool, Default: "false",
					},
					{
						Key: keyWecomWebhook, Title: "企业微信机器人地址", Hint: "群机器人的 Webhook。不要公开。",
						Type: setting.TypeString, Default: "", Max: 512,
					},
				},
			},
			{
				Key:   "general",
				Title: "通用",
				Items: []setting.Item{
					{
						Key: keyBaseURL, Title: "站点地址",
						Hint:    "把邮件里的站内链接和站内图片补成绝对地址，例如 https://admin.example.com。留空时相对链接在邮件里点不开。",
						Type:    setting.TypeString,
						Default: "",
						Max:     256,
					},
				},
			},
		},
	}
}

func (s *NotificationService) channelGates(ctx context.Context) (emailOn bool, wecomOn bool, webhook string) {
	emailOn = true
	if s.settings == nil {
		return true, false, ""
	}
	if v, err := s.settings.Bool(ctx, settingSubsystem, keyEmailEnabled); err == nil {
		emailOn = v
	}
	if v, err := s.settings.Bool(ctx, settingSubsystem, keyWecomEnabled); err == nil {
		wecomOn = v
	}
	if v, err := s.settings.String(ctx, settingSubsystem, keyWecomWebhook); err == nil {
		webhook = strings.TrimSpace(v)
	}
	return emailOn, wecomOn, webhook
}

// baseURL 返回站点地址，用于把外部渠道里的站内相对链接补成绝对地址。
// 未配置时返回空串，调用方按相对地址原样投递。
func (s *NotificationService) baseURL(ctx context.Context) string {
	if s.settings == nil {
		return ""
	}
	v, err := s.settings.String(ctx, settingSubsystem, keyBaseURL)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(v)
}
