package service

import (
	"testing"

	"bit-labs.cn/owl-admin/app-notification/model"
)

func TestChooseChannelsUsesCallerListAndSkipsUnconnected(t *testing.T) {
	plan := chooseChannels(
		[]uint{7},
		[]string{model.ChannelSMS, model.ChannelWeCom, model.ChannelDingTalk, model.ChannelEmail},
		[]string{model.ChannelInApp},
		false,
		true,
		true,
		func(channel string) bool {
			return channel == model.ChannelEmail || channel == model.ChannelWeCom || channel == model.ChannelInApp
		},
	)
	if len(plan.Users) != 1 {
		t.Fatalf("users = %d", len(plan.Users))
	}
	user := plan.Users[0]
	if !user.Email || !plan.WeCom || user.InApp {
		t.Fatalf("choice = %+v wecom=%v", user, plan.WeCom)
	}
	if len(user.Skips) != 2 {
		t.Fatalf("skips = %+v", user.Skips)
	}
}

func TestChooseChannelsExplicitChannelBypassesTypeDefault(t *testing.T) {
	plan := chooseChannels(
		[]uint{9},
		[]string{model.ChannelInApp, model.ChannelEmail},
		[]string{model.ChannelInApp},
		true,
		true,
		true,
		func(string) bool { return true },
	)
	user := plan.Users[0]
	if !user.InApp || !user.Email {
		t.Fatalf("explicit channels should both be sent, choice = %+v", user)
	}
}

func TestChooseChannelsRecordsGlobalGateSkips(t *testing.T) {
	plan := chooseChannels(
		[]uint{5},
		[]string{model.ChannelInApp, model.ChannelEmail, model.ChannelWeCom},
		nil,
		true,
		false,
		false,
		func(string) bool { return true },
	)
	user := plan.Users[0]
	if !user.InApp || user.Email || user.WeCom || plan.WeCom {
		t.Fatalf("choice = %+v wecom=%v", user, plan.WeCom)
	}
	if len(user.Skips) != 2 {
		t.Fatalf("skips = %+v", user.Skips)
	}
	for _, skip := range user.Skips {
		if skip.Reason == "" {
			t.Fatalf("skip reason missing: %+v", skip)
		}
	}
}
