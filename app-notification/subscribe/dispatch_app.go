package subscribe

import (
	"context"

	notifEvent "bit-labs.cn/owl-admin/app-notification/event"
	"bit-labs.cn/owl-admin/app-notification/model"
	"bit-labs.cn/owl-admin/app-notification/service"
	contractEvent "bit-labs.cn/owl/contract/event"
	"bit-labs.cn/owl/contract/log"
)

// subscribeDispatchApp 监听移动 App 推送事件，异步分发到 APNs 与 FCM。
func subscribeDispatchApp(bus contractEvent.EventBus, taskSvc *service.TaskService, logger log.Logger) error {
	return bus.Subscribe(notifEvent.TopicDispatchApp, contractEvent.AsHandler(func(ctx context.Context, evt *notifEvent.DispatchPayload) error {
		if evt == nil || evt.DeliveryID == 0 {
			return nil
		}
		if err := taskSvc.DispatchOne(ctx, evt.DeliveryID, model.ChannelApp); err != nil {
			logger.Error("notify: dispatch app push delivery failed: ", evt.DeliveryID, err)
			return err
		}
		return nil
	}))
}
