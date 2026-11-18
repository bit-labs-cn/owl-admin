package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"bit-labs.cn/owl-admin/app-notification/contract"
	notifEvent "bit-labs.cn/owl-admin/app-notification/event"
	"bit-labs.cn/owl-admin/app-notification/model"
	contractEvent "bit-labs.cn/owl/contract/event"
	"gorm.io/gorm"
)

type deliverySkip struct {
	reason string
}

func (e *deliverySkip) Error() string { return e.reason }

// DeliverChannel 消费一条外部渠道发送事件，并回写发送记录。
func (s *NotificationService) DeliverChannel(ctx context.Context, deliveryID uint) error {
	if deliveryID == 0 {
		return nil
	}
	row, err := s.catalog.FindDelivery(ctx, deliveryID)
	if err != nil {
		return err
	}
	if row == nil || row.Status == model.DeliveryStatusSuccess || row.Status == model.DeliveryStatusSkipped {
		return nil
	}
	if row.Status == model.DeliveryStatusFailed && row.Attempts >= model.DeliveryMaxAttempts {
		return nil
	}

	sendErr := s.dispatchDelivery(ctx, row)
	if sendErr == nil {
		now := time.Now()
		row.Status = model.DeliveryStatusSuccess
		row.LastError = ""
		row.SentAt = &now
		return s.catalog.SaveDelivery(ctx, row)
	}
	var skipped *deliverySkip
	if errors.As(sendErr, &skipped) {
		row.Status = model.DeliveryStatusSkipped
		row.LastError = skipped.reason
		return s.catalog.SaveDelivery(ctx, row)
	}

	row.Attempts++
	row.LastError = sendErr.Error()
	if row.Attempts >= model.DeliveryMaxAttempts {
		row.Status = model.DeliveryStatusFailed
		if err := s.catalog.SaveDelivery(ctx, row); err != nil {
			return err
		}
		return nil
	}
	row.Status = model.DeliveryStatusPending
	if err := s.catalog.SaveDelivery(ctx, row); err != nil {
		return err
	}
	return sendErr
}

func (s *NotificationService) dispatchDelivery(ctx context.Context, row *model.NotifyDelivery) error {
	switch row.Channel {
	case model.ChannelEmail:
		return s.sendEmail(ctx, row)
	case model.ChannelWeCom:
		return s.sendWeCom(ctx, row)
	case model.ChannelSMS, model.ChannelDingTalk:
		return &deliverySkip{reason: skipReasonNotConnected}
	default:
		return &deliverySkip{reason: skipReasonNotConnected}
	}
}

func (s *NotificationService) sendEmail(ctx context.Context, row *model.NotifyDelivery) error {
	sender := s.getChannel(model.ChannelEmail)
	if sender == nil {
		return &deliverySkip{reason: skipReasonNotConnected}
	}
	email := strings.TrimSpace(row.RecipientEmail)
	if email == "" {
		contacts, err := s.catalog.ListContacts(ctx, []uint{row.UserID})
		if err != nil {
			return err
		}
		if len(contacts) > 0 {
			email = strings.TrimSpace(contacts[0].Email)
		}
	}
	if email == "" {
		return &deliverySkip{reason: "用户没有邮箱"}
	}
	var payload map[string]any
	fromJSON(row.Payload, &payload)
	return sender.Send(ctx, &contract.ChannelRequest{
		UserID:      row.UserID,
		Email:       email,
		Title:       row.Title,
		Content:     row.Content,
		ContentText: HTMLToText(row.Content),
		Link:        row.Link,
		Payload:     payload,
	})
}

func (s *NotificationService) sendWeCom(ctx context.Context, row *model.NotifyDelivery) error {
	sender := s.getChannel(model.ChannelWeCom)
	if sender == nil {
		return &deliverySkip{reason: skipReasonNotConnected}
	}
	_, enabled, webhook := s.channelGates(ctx)
	if !enabled || webhook == "" {
		return &deliverySkip{reason: "企业微信机器人未配置"}
	}
	var payload map[string]any
	fromJSON(row.Payload, &payload)
	return sender.Send(ctx, &contract.ChannelRequest{
		UserID:  row.UserID,
		Title:   row.Title,
		Content: row.Content,
		Link:    row.Link,
		Webhook: webhook,
		Payload: payload,
	})
}

// RetryDelivery 把失败或仍在等待的发送记录重新交给事件总线。
func (s *NotificationService) RetryDelivery(ctx context.Context, id uint) error {
	row, err := s.catalog.FindDelivery(ctx, id)
	if err != nil {
		return err
	}
	if row == nil {
		return NotFound("发送记录不存在")
	}
	if row.Status != model.DeliveryStatusFailed {
		return Invalid("只有失败的发送记录可以重试")
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		row.Status = model.DeliveryStatusPending
		row.Attempts = 0
		row.LastError = ""
		row.SentAt = nil
		if err := tx.Model(&model.NotifyDelivery{}).Where("id = ?", row.ID).Updates(map[string]any{
			"status":     row.Status,
			"attempts":   0,
			"last_error": "",
			"sent_at":    nil,
		}).Error; err != nil {
			return err
		}
		evt, err := contractEvent.NewEvent(notifEvent.TopicNotifyChannelSend, notifEvent.ChannelSendPayload{
			DeliveryID: row.ID,
		}, contractEvent.WithSource("notify"), contractEvent.WithSubject(fmt.Sprintf("delivery-retry:%d:%d", row.ID, time.Now().UnixNano())))
		if err != nil {
			return err
		}
		return s.bus.PublishTx(ctx, tx, evt)
	})
}
