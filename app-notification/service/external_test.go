package service

import (
	"testing"

	"bit-labs.cn/owl-admin/app-notification/contract"
	"bit-labs.cn/owl-admin/app-notification/model"
)

func TestNormalizeExternalsDropsEmptyAndDedupes(t *testing.T) {
	out := normalizeExternals([]contract.ExternalRecipient{
		{Name: "张三", Email: "a@example.com"},
		{Name: "张三重复", Email: "A@example.com"},
		{Name: "无联系方式"},
		{Phone: "13800000000"},
	})
	if len(out) != 2 {
		t.Fatalf("externals = %+v", out)
	}
	if out[0].Name != "张三" || out[0].Email != "a@example.com" {
		t.Fatalf("first = %+v", out[0])
	}
	if out[1].Name != "13800000000" {
		t.Fatalf("phone-only name = %q", out[1].Name)
	}
}

func TestExternalDeliveriesRequireEmailChannel(t *testing.T) {
	externals := []contract.ExternalRecipient{{Name: "李四", Email: "li@example.com"}}
	rendered := renderedSet{title: "标题", body: "正文"}

	withoutEmail := externalDeliveries(externals, rendered, []string{model.ChannelInApp}, true)
	if len(withoutEmail) != 1 || withoutEmail[0].Status != model.DeliveryStatusSkipped {
		t.Fatalf("without email channel = %+v", withoutEmail)
	}
	if withoutEmail[0].RecipientType != model.RecipientTypeExternal {
		t.Fatalf("recipient type = %d", withoutEmail[0].RecipientType)
	}

	withEmail := externalDeliveries(externals, rendered, []string{model.ChannelEmail}, true)
	if len(withEmail) != 1 || withEmail[0].Status != model.DeliveryStatusPending {
		t.Fatalf("with email channel = %+v", withEmail)
	}
	if withEmail[0].RecipientEmail != "li@example.com" {
		t.Fatalf("email = %q", withEmail[0].RecipientEmail)
	}
}

func TestExternalDeliveriesSkipWhenEmailDisabled(t *testing.T) {
	externals := []contract.ExternalRecipient{{Name: "王五", Email: "wang@example.com"}}
	out := externalDeliveries(externals, renderedSet{}, []string{model.ChannelEmail}, false)
	if len(out) != 1 || out[0].Status != model.DeliveryStatusSkipped {
		t.Fatalf("disabled gate = %+v", out)
	}
}

func TestEffectiveChannels(t *testing.T) {
	if got := effectiveChannels(nil, []string{model.ChannelEmail}, true); len(got) != 1 || got[0] != model.ChannelEmail {
		t.Fatalf("typed default = %v", got)
	}
	if got := effectiveChannels(nil, []string{model.ChannelEmail}, false); len(got) != 1 || got[0] != model.ChannelInApp {
		t.Fatalf("untyped default = %v", got)
	}
	if got := effectiveChannels([]string{model.ChannelSMS}, []string{model.ChannelEmail}, true); len(got) != 1 || got[0] != model.ChannelSMS {
		t.Fatalf("caller wins = %v", got)
	}
}
