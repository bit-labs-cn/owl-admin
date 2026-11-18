package subscribe

import (
	"context"

	notifEvent "bit-labs.cn/owl-admin/app-notification/event"
	"bit-labs.cn/owl-admin/app-notification/service"
	contractEvent "bit-labs.cn/owl/contract/event"
	"bit-labs.cn/owl/contract/log"
)

// subscribeTaskSubmit 监听任务提交事件，异步执行受众解析与渠道扇出。
func subscribeTaskSubmit(bus contractEvent.EventBus, taskSvc *service.TaskService, logger log.Logger) error {
	return bus.Subscribe(notifEvent.TopicTaskSubmit, contractEvent.AsHandler(func(ctx context.Context, evt *notifEvent.TaskSubmitPayload) error {
		if evt == nil || evt.TaskID == 0 {
			return nil
		}
		if err := taskSvc.FanoutAsync(ctx, evt.TaskID); err != nil {
			logger.Error("notify: async fanout task failed: ", evt.TaskID, err)
			return err
		}
		return nil
	}))
}
