// Package subscribe 注册 owl-admin 对领域事件的订阅。
// 一个事件一个文件，文件名与事件对应，避免所有监听堆在同一个文件里。
package subscribe

import (
	"bit-labs.cn/owl"
	"bit-labs.cn/owl-admin/app/service"
	contractEvent "bit-labs.cn/owl/contract/event"
	"bit-labs.cn/owl/contract/foundation"
	"bit-labs.cn/owl/contract/log"
	"bit-labs.cn/owl/provider/router"
	"github.com/casbin/casbin/v3"
)

// Init 从容器取出事件总线与依赖，逐个注册订阅。
// 订阅失败直接 panic：权限同步或登录日志缺失属于启动期错误，不应带病运行。
func Init(app foundation.Application) {
	err := app.Invoke(func(
		bus contractEvent.EventBus,
		enforcer *casbin.SyncedEnforcer,
		logger log.Logger,
		menuRepo *router.MenuRepository,
		logSvc *service.LogService,
	) {
		owl.PanicIf(subscribeAssignRoleToUser(bus, enforcer, logger))
		owl.PanicIf(subscribeAssignMenuToRole(bus, enforcer, menuRepo, logger))
		owl.PanicIf(subscribeUserLoginSuccess(bus, logSvc))
	})
	owl.PanicIf(err)
}
