// Package subscribe 注册通知中心子系统对事件总线（Event Bus）的订阅与消费。
// 遵循与 app/subscribe 一致的设计：一个事件一个文件，清晰隔离不同渠道与任务流。
package subscribe

import (
	"bit-labs.cn/owl"
	"bit-labs.cn/owl-admin/app-notification/service"
	contractEvent "bit-labs.cn/owl/contract/event"
	"bit-labs.cn/owl/contract/foundation"
	"bit-labs.cn/owl/contract/log"
)

// Init 从容器取出事件总线、任务服务与收件箱服务，注册全部事件监听。
func Init(app foundation.Application) {
	err := app.Invoke(func(
		bus contractEvent.EventBus,
		taskSvc *service.TaskService,
		inboxSvc *service.InboxService,
		logger log.Logger,
	) {
		owl.PanicIf(subscribeTaskSubmit(bus, taskSvc, logger))
		owl.PanicIf(subscribeDispatchEmail(bus, taskSvc, logger))
		owl.PanicIf(subscribeDispatchSMS(bus, taskSvc, logger))
		owl.PanicIf(subscribeDispatchApp(bus, taskSvc, logger))
		owl.PanicIf(subscribeTodoComplete(bus, inboxSvc, logger))
	})
	owl.PanicIf(err)
}
