package subscribe

import (
	"context"

	notifEvent "bit-labs.cn/owl-admin/app-notification/event"
	"bit-labs.cn/owl-admin/app-notification/service"
	contractEvent "bit-labs.cn/owl/contract/event"
	"bit-labs.cn/owl/contract/log"
)

// subscribeTodoComplete 监听待办完成事件，将匹配业务主键的待办自动标记为已完成。
func subscribeTodoComplete(bus contractEvent.EventBus, inboxSvc *service.InboxService, logger log.Logger) error {
	return bus.Subscribe(notifEvent.TopicTodoComplete, contractEvent.AsHandler(func(ctx context.Context, evt *notifEvent.TodoCompletePayload) error {
		if evt == nil || evt.BizKey == "" {
			return nil
		}
		if err := inboxSvc.CompleteByBizKey(ctx, evt.BizKey); err != nil {
			logger.Error("notify: complete todo by bizKey failed: ", evt.BizKey, err)
			return err
		}
		return nil
	}))
}
