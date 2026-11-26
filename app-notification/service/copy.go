package service

import (
	"encoding/json"
	"strings"
	"time"

	"bit-labs.cn/owl-admin/app-notification/contract"
	"bit-labs.cn/owl-admin/app-notification/model"
	"bit-labs.cn/owl-admin/app-notification/repository"
)

// channelCopy 是一个渠道渲染完成后的文案。
type channelCopy struct {
	Title   string
	Content string
	Link    string
}

// renderedSet 是一条消息的模板渲染结果，正文统一按 HTML 保存。
// 模板不再区分渠道：同一份文案在投递时按渠道转换，站内信和邮件直接用 HTML。
type renderedSet struct {
	title string
	body  string
	link  string
}

// renderTemplate 渲染消息类型的正文模板，模板为空时回退到调用方给的标题、正文和链接。
func renderTemplate(tpl *model.NotifyTemplate, vars map[string]any, title, content, link string) (renderedSet, error) {
	out := renderedSet{title: title, body: content, link: link}
	if tpl == nil {
		return out, nil
	}
	renderedTitle, err := applyText(tpl.Title, title, vars)
	if err != nil {
		return out, err
	}
	// 模板正文是富文本，按 HTML 模板渲染，变量按 HTML 转义
	renderedBody, err := applyHTML(tpl.Body, content, vars)
	if err != nil {
		return out, err
	}
	renderedLink, err := applyText(tpl.Link, link, vars)
	if err != nil {
		return out, err
	}
	return renderedSet{title: renderedTitle, body: ContentHTML(renderedBody), link: renderedLink}, nil
}

// forChannel 取某个渠道的文案。
// 站内信和邮件承载 HTML；企业微信、短信、钉钉只吃纯文本，统一从 HTML 转过来。
func (s renderedSet) forChannel(channel string) channelCopy {
	body := s.body
	switch channel {
	case model.ChannelWeCom, model.ChannelSMS, model.ChannelDingTalk:
		body = HTMLToText(s.body)
	}
	return channelCopy{Title: s.title, Content: body, Link: s.link}
}

func pendingDelivery(messageID, userID uint, channel string, copy channelCopy) model.NotifyDelivery {
	return model.NotifyDelivery{
		MessageID:     messageID,
		UserID:        userID,
		Channel:       channel,
		Title:         copy.Title,
		Content:       copy.Content,
		Link:          copy.Link,
		Status:        model.DeliveryStatusPending,
		RecipientType: model.RecipientTypeUser,
	}
}

// inAppDelivery 是站内信的投递记录。站内信一写库就算送达，所以直接记成功。
func inAppDelivery(userID uint, copy channelCopy, contact repository.UserContact, sentAt time.Time) model.NotifyDelivery {
	row := model.NotifyDelivery{
		UserID:        userID,
		Channel:       model.ChannelInApp,
		Title:         copy.Title,
		Content:       copy.Content,
		Link:          copy.Link,
		Status:        model.DeliveryStatusSuccess,
		Attempts:      1,
		RecipientType: model.RecipientTypeUser,
		SentAt:        &sentAt,
	}
	if contact.ID > 0 {
		row.RecipientName = contactName(contact)
		row.RecipientEmail = contact.Email
	}
	return row
}

func skippedDelivery(messageID, userID uint, channel string, copy channelCopy, reason string) model.NotifyDelivery {
	return model.NotifyDelivery{
		MessageID:     messageID,
		UserID:        userID,
		Channel:       channel,
		Title:         copy.Title,
		Content:       copy.Content,
		Link:          copy.Link,
		Status:        model.DeliveryStatusSkipped,
		LastError:     reason,
		RecipientType: model.RecipientTypeUser,
	}
}

// externalDeliveries 为外部人员生成投递记录。外部人员没有站内信，只能走邮件。
func externalDeliveries(externals []contract.ExternalRecipient, rendered renderedSet, channels []string, emailOn bool) []model.NotifyDelivery {
	if len(externals) == 0 {
		return nil
	}
	copy := rendered.forChannel(model.ChannelEmail)
	emailRequested := containsChannel(channels, model.ChannelEmail)
	out := make([]model.NotifyDelivery, 0, len(externals))
	for _, external := range externals {
		delivery := model.NotifyDelivery{
			Channel:        model.ChannelEmail,
			Title:          copy.Title,
			Content:        copy.Content,
			Link:           copy.Link,
			RecipientType:  model.RecipientTypeExternal,
			RecipientName:  external.Name,
			RecipientEmail: external.Email,
			RecipientPhone: external.Phone,
		}
		switch {
		case !emailOn:
			delivery.Status = model.DeliveryStatusSkipped
			delivery.LastError = "邮件渠道已关闭"
		case !emailRequested:
			delivery.Status = model.DeliveryStatusSkipped
			delivery.LastError = "外部人员只能通过邮件接收，请勾选邮件渠道"
		case external.Email == "":
			delivery.Status = model.DeliveryStatusSkipped
			delivery.LastError = "外部人员没有邮箱"
		default:
			delivery.Status = model.DeliveryStatusPending
		}
		out = append(out, delivery)
	}

	return out
}

func decodeChannels(raw string) ([]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	var list []string
	if err := json.Unmarshal([]byte(raw), &list); err != nil {
		return nil, err
	}
	return list, nil
}

func encodeChannels(list []string) string {
	list = normalizeChannels(list)
	if len(list) == 0 {
		return "[]"
	}
	raw, err := json.Marshal(list)
	if err != nil {
		return "[]"
	}
	return string(raw)
}
