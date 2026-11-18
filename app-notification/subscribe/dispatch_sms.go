package subscribe

import (
	"context"

	notifEvent "bit-labs.cn/owl-admin/app-notification/event"
	"bit-labs.cn/owl-admin/app-notification/model"
	"bit-labs.cn/owl-admin/app-notification/service"
	contractEvent "bit-labs.cn/owl/contract/event"
	"bit-labs.cn/owl/contract/log"
)

// subscribeDispatchSMS 监听短信投递事件，异步调用短信子系统发送。
func subscribeDispatchSMS(bus contractEvent.EventBus, taskSvc *service.TaskService, logger log.Logger) error {
	return bus.Subscribe(notifEvent.TopicDispatchSMS, contractEvent.AsHandler(func(ctx context.Context, evt *notifEvent.DispatchPayload) error {
		if evt == nil || evt.DeliveryID == 0 {
			return nil
		}
		if err := taskSvc.DispatchOne(ctx, evt.DeliveryID, model.ChannelSMS); err != nil {
			logger.Error("notify: dispatch sms delivery failed: ", evt.DeliveryID, err)
			return err
		}
		return nil
	}))
}
