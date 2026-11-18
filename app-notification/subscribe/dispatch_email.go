package subscribe

import (
	"context"

	notifEvent "bit-labs.cn/owl-admin/app-notification/event"
	"bit-labs.cn/owl-admin/app-notification/model"
	"bit-labs.cn/owl-admin/app-notification/service"
	contractEvent "bit-labs.cn/owl/contract/event"
	"bit-labs.cn/owl/contract/log"
)

// subscribeDispatchEmail 监听邮件投递事件，异步调用邮件驱动发送。
func subscribeDispatchEmail(bus contractEvent.EventBus, taskSvc *service.TaskService, logger log.Logger) error {
	return bus.Subscribe(notifEvent.TopicDispatchEmail, contractEvent.AsHandler(func(ctx context.Context, evt *notifEvent.DispatchPayload) error {
		if evt == nil || evt.DeliveryID == 0 {
			return nil
		}
		if err := taskSvc.DispatchOne(ctx, evt.DeliveryID, model.ChannelEmail); err != nil {
			logger.Error("notify: dispatch email delivery failed: ", evt.DeliveryID, err)
			return err
		}
		return nil
	}))
}
