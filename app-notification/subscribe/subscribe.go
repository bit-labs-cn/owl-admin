package subscribe

import (
	"context"

	notifEvent "bit-labs.cn/owl-admin/app-notification/event"
	"bit-labs.cn/owl-admin/app-notification/service"
	contractEvent "bit-labs.cn/owl/contract/event"
	"bit-labs.cn/owl/contract/foundation"
	"bit-labs.cn/owl/contract/log"
)

// Init 注册通知系统对事件总线（Event Bus）的订阅与消费。
func Init(app foundation.Application) {
	_ = app.Invoke(func(
		bus contractEvent.EventBus,
		svc *service.NotificationService,
		logger log.Logger,
	) {
		// 1. 订阅定向消息（通知/待办）
		_ = bus.Subscribe(notifEvent.TopicNotifyDirect, contractEvent.AsHandler(func(ctx context.Context, evt *notifEvent.DirectMessagePayload) error {
			if evt == nil {
				return nil
			}
			if err := svc.HandleDirectEvent(ctx, evt); err != nil {
				logger.Error("notify: handle direct message event failed: ", err)
				return err
			}
			return nil
		}))

		// 2. 订阅业务办结勾销待办
		_ = bus.Subscribe(notifEvent.TopicNotifyTodoComplete, contractEvent.AsHandler(func(ctx context.Context, evt *notifEvent.TodoCompletePayload) error {
			if evt == nil || evt.BizKey == "" {
				return nil
			}
			if err := svc.CompleteByBizKey(ctx, evt.BizKey); err != nil {
				logger.Error("notify: complete todo by bizKey failed: ", evt.BizKey, err)
				return err
			}
			return nil
		}))

		// 3. 订阅外部渠道发送
		_ = bus.Subscribe(notifEvent.TopicNotifyChannelSend, contractEvent.AsHandler(func(ctx context.Context, evt *notifEvent.ChannelSendPayload) error {
			if evt == nil || evt.DeliveryID == 0 {
				return nil
			}
			if err := svc.DeliverChannel(ctx, evt.DeliveryID); err != nil {
				logger.Error("notify: deliver channel failed: ", err)
				return err
			}
			return nil
		}))

		// 4. 订阅公告广播
		_ = bus.Subscribe(notifEvent.TopicNotifyAnnouncement, contractEvent.AsHandler(func(ctx context.Context, evt *notifEvent.AnnouncementPayload) error {
			if evt == nil {
				return nil
			}
			if err := svc.HandleAnnouncementEvent(ctx, evt); err != nil {
				logger.Error("notify: handle announcement broadcast event failed: ", err)
				return err
			}
			return nil
		}))
	})
}
