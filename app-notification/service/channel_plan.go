package service

import "bit-labs.cn/owl-admin/app-notification/model"

const (
	skipReasonNotConnected = "未接入"
	skipReasonEmailOff     = "邮件渠道已在系统设置中关闭"
	skipReasonWeComOff     = "企业微信渠道已在系统设置中关闭"
)

// channelChoice 是一个接收人最终要走的渠道。
type channelChoice struct {
	UserID uint
	InApp  bool
	Email  bool
	WeCom  bool
	Skips  []skipChoice
}

// skipChoice 是没发出去的渠道，仍要留下一条跳过记录。
type skipChoice struct {
	Channel string
	Reason  string
}

// sendPlan 是一条消息在全部接收人上的渠道决定。
type sendPlan struct {
	Users []channelChoice
	WeCom bool
}

// effectiveChannels 决定这条消息实际要尝试的渠道：调用方指定优先，其次类型默认，最后站内信。
func effectiveChannels(requested, defaults []string, typed bool) []string {
	requested = normalizeChannels(requested)
	if len(requested) > 0 {
		return requested
	}
	defaults = normalizeChannels(defaults)
	if typed && len(defaults) > 0 {
		return defaults
	}
	return []string{model.ChannelInApp}
}

// chooseChannels 按调用方指定、类型默认和系统渠道开关决定渠道。
// 没发出去的渠道一律记入 Skips，让发送记录能解释原因。
func chooseChannels(userIDs []uint, requested, defaults []string, typed bool, emailOn, wecomOn bool, hasSender func(string) bool) sendPlan {
	requested = effectiveChannels(requested, defaults, typed)

	plan := sendPlan{Users: make([]channelChoice, 0, len(userIDs))}
	for _, uid := range userIDs {
		choice := channelChoice{UserID: uid}
		for _, channel := range requested {
			allowed, reason := channelAllowed(channel, emailOn, wecomOn)
			if !allowed {
				choice.Skips = append(choice.Skips, skipChoice{Channel: channel, Reason: reason})
				continue
			}

			switch channel {
			case model.ChannelInApp:
				choice.InApp = true
			case model.ChannelEmail:
				if hasSender != nil && hasSender(channel) {
					choice.Email = true
				} else {
					choice.Skips = append(choice.Skips, skipChoice{Channel: channel, Reason: skipReasonNotConnected})
				}
			case model.ChannelWeCom:
				if hasSender != nil && hasSender(channel) {
					choice.WeCom = true
					plan.WeCom = true
				} else {
					choice.Skips = append(choice.Skips, skipChoice{Channel: channel, Reason: skipReasonNotConnected})
				}
			default:
				choice.Skips = append(choice.Skips, skipChoice{Channel: channel, Reason: skipReasonNotConnected})
			}
		}
		plan.Users = append(plan.Users, choice)
	}

	return plan
}

// channelAllowed 判断这个渠道是否可用。第二个返回值是跳过的原因。
func channelAllowed(channel string, emailOn, wecomOn bool) (bool, string) {
	if channel == model.ChannelEmail && !emailOn {
		return false, skipReasonEmailOff
	}
	if channel == model.ChannelWeCom && !wecomOn {
		return false, skipReasonWeComOff
	}
	return true, ""
}

func normalizeChannels(codes []string) []string {
	if len(codes) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(codes))
	out := make([]string, 0, len(codes))
	for _, code := range codes {
		if code == "" {
			continue
		}
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}
		out = append(out, code)
	}

	return out
}

func containsChannel(codes []string, channel string) bool {
	for _, code := range codes {
		if code == channel {
			return true
		}
	}
	return false
}

func dedupeUsers(ids []uint) []uint {
	seen := make(map[uint]struct{}, len(ids))
	out := make([]uint, 0, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}

	return out
}
