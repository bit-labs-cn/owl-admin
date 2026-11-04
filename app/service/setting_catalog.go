package service

import (
	"bit-labs.cn/owl/provider/setting"
)

const (
	subsystemAdmin = "admin"

	keySiteName                = "siteName"
	keySiteLogo                = "siteLogo"
	keyCopyright               = "copyright"
	keyRegisterEnabled         = "registerEnabled"
	keyRegisterCooldownSeconds = "registerCooldownSeconds"
)

// AdminSettingCatalog 是 admin 子系统写死的系统配置。
func AdminSettingCatalog() setting.Catalog {
	return setting.Catalog{
		Title: "系统管理",
		Groups: []setting.Group{
			{
				Key:   "site",
				Title: "站点",
				Items: []setting.Item{
					{
						Key: keySiteName, Title: "站点名称", Hint: "显示在侧栏和登录页。留空沿用平台配置里的名称。",
						Type: setting.TypeString, Default: "", Max: 64, Public: true,
					},
					{
						Key: keySiteLogo, Title: "站点标志", Hint: "侧栏左上角的图片，以 base64 保存，不能超过 200KB。留空使用内置标志。",
						Type: setting.TypeImage, Default: "", Max: setting.ImageMaxBytes, Public: true,
					},
					{
						Key: keyCopyright, Title: "版权信息", Hint: "显示在页面底部。留空则显示 Copyright © 2020-present 和站点名称。",
						Type: setting.TypeString, Default: "", Max: 128, Public: true,
					},
				},
			},
			{
				Key:   "account",
				Title: "账号",
				Items: []setting.Item{
					{
						Key: keyRegisterEnabled, Title: "开放注册", Hint: "关闭后不能自行注册。",
						Type: setting.TypeBool, Default: "true", Public: true,
					},
					{
						Key: keyRegisterCooldownSeconds, Title: "验证码发送间隔", Hint: "单位是秒。",
						Type: setting.TypeInt, Default: "60", Min: 30, Max: 300,
					},
				},
			},
		},
	}
}
