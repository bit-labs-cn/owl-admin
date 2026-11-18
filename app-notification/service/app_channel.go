package service

import (
	"context"
	"strconv"
	"strings"

	"bit-labs.cn/owl-admin/app-notification/contract"
	"bit-labs.cn/owl-admin/app-notification/service/push"
)

// AppChannel 向用户已登记的设备发送 APNs 或 FCM 推送。
type AppChannel struct {
	apns *push.APNsClient
	fcm  *push.FCMClient
}

func NewAppChannel() *AppChannel {
	return &AppChannel{apns: push.NewAPNsClient(), fcm: push.NewFCMClient()}
}

func (c *AppChannel) Code() string { return contract.ChannelApp }

func (c *AppChannel) Send(ctx context.Context, req *contract.ChannelRequest) (*contract.ChannelResult, error) {
	if len(req.DeviceTokens) == 0 {
		return nil, Invalid("用户没有可推送的设备")
	}
	cfgRaw := req.ConfigJSON
	var apnsCfg push.APNsConfig
	var fcmCfg push.FCMConfig
	if strings.Contains(cfgRaw, "teamId") {
		parsed, err := decodeConfig[push.APNsConfig](cfgRaw)
		if err != nil {
			return nil, err
		}
		apnsCfg = parsed
	}
	if strings.Contains(cfgRaw, "projectId") {
		parsed, err := decodeConfig[push.FCMConfig](cfgRaw)
		if err != nil {
			return nil, err
		}
		fcmCfg = parsed
	}
	data := map[string]string{
		"taskId": strconv.FormatUint(uint64(req.TaskID), 10),
		"kind":   strconv.Itoa(req.Kind),
	}
	if req.Link != nil {
		data["app"] = req.Link.App
		data["path"] = req.Link.Path
		data["link"] = toJSON(req.Link)
	}
	for k, v := range req.Payload {
		data[k] = v
	}
	result := &contract.ChannelResult{}
	var firstErr error
	sent := 0
	for _, device := range req.DeviceTokens {
		var one push.SendResult
		var err error
		switch device.Provider {
		case "apns":
			one, err = c.apns.Send(ctx, apnsCfg, device.Token, req.Title, req.ContentText, data)
		case "fcm":
			one, err = c.fcm.Send(ctx, fcmCfg, device.Token, req.Title, req.ContentText, data)
		default:
			err = Invalid("不支持的推送服务商 " + device.Provider)
		}
		if one.InvalidToken {
			result.InvalidTokens = append(result.InvalidTokens, device.Token)
		}
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		sent++
		if result.ProviderMsgID == "" {
			result.ProviderMsgID = one.MessageID
		}
	}
	if sent == 0 {
		if firstErr != nil {
			return result, firstErr
		}
		return result, Invalid("推送全部失败")
	}
	return result, nil
}
