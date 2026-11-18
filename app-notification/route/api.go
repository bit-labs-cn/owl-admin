package route

import (
	"bit-labs.cn/owl"
	v1 "bit-labs.cn/owl-admin/app-notification/handle/v1"
	adminmw "bit-labs.cn/owl-admin/app/middleware"
	"bit-labs.cn/owl-admin/app/provider/jwt"
	adminsvc "bit-labs.cn/owl-admin/app/service"
	"bit-labs.cn/owl/contract/foundation"
	"bit-labs.cn/owl/provider/router"
	"github.com/casbin/casbin/v3"
	"github.com/gin-gonic/gin"
)

var (
	menuRoot         *router.Menu
	menuAnnouncement *router.Menu
	menuInbox        *router.Menu
	menuType         *router.Menu
	menuTemplate     *router.Menu
	menuDelivery     *router.Menu
	menuSend         *router.Menu
	menuMessage      *router.Menu
	menuSchedule     *router.Menu
)

// InitMenu 返回通知中心菜单。
func InitMenu() []*router.Menu {
	return []*router.Menu{menuRoot}
}

// InitAPI 注册通知中心路由。
func InitAPI(app foundation.Application, appName string) {
	err := app.Invoke(func(
		engine *gin.Engine,
		enforcer *casbin.SyncedEnforcer,
		jwtSvc *jwt.JWTService,
		logSvc *adminsvc.LogService,
		announcement *v1.AnnouncementHandle,
		message *v1.MessageHandle,
		notifyType *v1.TypeHandle,
		template *v1.TemplateHandle,
		delivery *v1.DeliveryHandle,
		send *v1.SendHandle,
		schedule *v1.ScheduleHandle,
		messageAdmin *v1.MessageAdminHandle,
	) {
		gv1 := engine.Group("/api/v1", adminmw.PermissionCheck(enforcer, jwtSvc))
		gv1.Use(adminmw.OperationLog(logSvc))

		// 1. 公告管理与用户拉取
		{
			r := router.NewRouteInfoBuilder(appName, announcement, gv1, router.MenuOption{ComponentName: "NotifyAnnouncement", Path: "/system/notify/announcement/index", Icon: "ep:bell"})
			r.Get("/notify/announcements", router.AccessAuthorized, announcement.Retrieve).Name("公告列表").Build()
			r.Post("/notify/announcements", router.AccessAuthorized, announcement.Create).Name("创建公告").Build()
			r.Put("/notify/announcements/:id", router.AccessAuthorized, announcement.Update).Name("更新公告").Build()
			r.Delete("/notify/announcements/:id", router.AccessAuthorized, announcement.Delete).Name("删除公告").Build()
			r.Get("/notify/announcements/:id", router.AccessAuthorized, announcement.Detail).Name("公告详情").Build()
			r.Get("/notify/announcements/:id/read-state", router.AccessAuthorized, announcement.ReadState).Name("公告已读情况").Build()
			r.Post("/notify/announcements/:id/publish", router.AccessAuthorized, announcement.Publish).Name("发布公告").Build()
			r.Post("/notify/announcements/:id/offline", router.AccessAuthorized, announcement.Offline).Name("下线公告").Build()

			r.Get("/notify/my/announcements", router.AccessAuthenticated, announcement.Mine).Name("我的公告").Build()
			r.Post("/notify/my/announcements/:id/dismiss", router.AccessAuthenticated, announcement.Dismiss).Name("关闭公告").Build()
			menuAnnouncement = r.GetMenu()
		}

		// 2. 消息与待办（收件箱）
		{
			r := router.NewRouteInfoBuilder(appName, message, gv1, router.MenuOption{ComponentName: "NotifyInbox", Path: "/system/notify/inbox/index", Icon: "ep:message"})
			r.Get("/notify/my/counts", router.AccessAuthenticated, message.Counts).Name("我的未读数").Build()
			r.Get("/notify/my/inbox", router.AccessAuthenticated, message.Retrieve).Name("我的消息列表").Build()
			r.Post("/notify/my/inbox/read", router.AccessAuthenticated, message.Read).Name("标记消息已读").Build()
			r.Post("/notify/my/inbox/read-all", router.AccessAuthenticated, message.ReadAll).Name("全部标记已读").Build()
			r.Delete("/notify/my/inbox/:id", router.AccessAuthenticated, message.Delete).Name("删除我的消息").Build()
			r.Post("/notify/my/todos/:id/done", router.AccessAuthenticated, message.Done).Name("完成待办").Build()
			r.Post("/notify/my/todos/:id/reopen", router.AccessAuthenticated, message.Reopen).Name("重开待办").Build()
			r.Post("/notify/my/todos/done", router.AccessAuthenticated, message.Complete).Name("办结待办").Build()
			r.Post("/notify/my/stream-ticket", router.AccessAuthenticated, message.Ticket).Name("实时连接票据").Build()
			menuInbox = r.GetMenu()
		}

		{
			r := router.NewRouteInfoBuilder(appName, send, gv1, router.MenuOption{ComponentName: "NotifySend", Path: "/system/notify/send/index", Icon: "ep:position"})
			r.Get("/notify/recipient-options", router.AccessAuthenticated, send.Recipients).Name("收件人选项").Build()
			r.Get("/notify/audience-options", router.AccessAuthenticated, send.Audience).Name("受众选项").Build()
			r.Post("/notify/audience-estimate", router.AccessAuthenticated, send.EstimateAudience).Name("接收人预估").Build()
			r.Post("/notify/messages", router.AccessAuthorized, send.Send).Name("发送通知").Build()
			r.Get("/notify/type-options", router.AccessAuthenticated, notifyType.Options).Name("消息类型选项").Build()
			menuSend = r.GetMenu()
		}

		{
			r := router.NewRouteInfoBuilder(appName, schedule, gv1, router.MenuOption{ComponentName: "NotifySchedule", Path: "/system/notify/schedule/index", Icon: "ep:timer"})
			r.Get("/notify/schedules", router.AccessAuthorized, schedule.Retrieve).Name("定时发送列表").Build()
			r.Post("/notify/schedules/:id/cancel", router.AccessAuthorized, schedule.Cancel).Name("取消定时发送").Build()
			r.Post("/notify/schedules/:id/send-now", router.AccessAuthorized, schedule.SendNow).Name("立即发送定时任务").Build()
			r.Post("/notify/schedules/:id/retry", router.AccessAuthorized, schedule.Retry).Name("重试定时发送").Build()
			menuSchedule = r.GetMenu()
		}

		{
			r := router.NewRouteInfoBuilder(appName, messageAdmin, gv1, router.MenuOption{ComponentName: "NotifyMessage", Path: "/system/notify/message/index", Icon: "ep:list"})
			r.Get("/notify/messages", router.AccessAuthorized, messageAdmin.Retrieve).Name("消息记录列表").Build()
			r.Get("/notify/messages/:id", router.AccessAuthorized, messageAdmin.Detail).Name("消息接收明细").Build()
			menuMessage = r.GetMenu()
		}

		{
			r := router.NewRouteInfoBuilder(appName, notifyType, gv1, router.MenuOption{ComponentName: "NotifyType", Path: "/system/notify/type/index", Icon: "ep:collection-tag"})
			r.Get("/notify/types", router.AccessAuthorized, notifyType.Retrieve).Name("消息类型列表").Build()
			r.Post("/notify/types", router.AccessAuthorized, notifyType.Create).Name("创建消息类型").Build()
			r.Put("/notify/types/:id", router.AccessAuthorized, notifyType.Update).Name("更新消息类型").Build()
			r.Delete("/notify/types/:id", router.AccessAuthorized, notifyType.Delete).Name("删除消息类型").Build()
			menuType = r.GetMenu()
		}

		{
			r := router.NewRouteInfoBuilder(appName, template, gv1, router.MenuOption{ComponentName: "NotifyTemplate", Path: "/system/notify/template/index", Icon: "ep:document"})
			r.Get("/notify/templates", router.AccessAuthorized, template.Retrieve).Name("消息模板列表").Build()
			r.Post("/notify/templates", router.AccessAuthorized, template.Create).Name("创建消息模板").Build()
			r.Put("/notify/templates/:id", router.AccessAuthorized, template.Update).Name("更新消息模板").Build()
			r.Delete("/notify/templates/:id", router.AccessAuthorized, template.Delete).Name("删除消息模板").Build()
			r.Post("/notify/templates/preview", router.AccessAuthorized, template.Preview).Name("预览消息模板").Build()
			menuTemplate = r.GetMenu()
		}

		{
			r := router.NewRouteInfoBuilder(appName, delivery, gv1, router.MenuOption{ComponentName: "NotifyDelivery", Path: "/system/notify/delivery/index", Icon: "ep:promotion"})
			r.Get("/notify/deliveries", router.AccessAuthorized, delivery.Retrieve).Name("发送记录列表").Build()
			r.Get("/notify/deliveries/:messageId", router.AccessAuthorized, delivery.Detail).Name("发送记录明细").Build()
			r.Post("/notify/deliveries/:id/retry", router.AccessAuthorized, delivery.Retry).Name("重试发送").Build()
			menuDelivery = r.GetMenu()
		}

		// 3. SSE 实时推流公开端点（一次性票据鉴权）
		pub := engine.Group("/api/v1")
		pub.GET("/notify/stream", message.Stream)

		menuRoot = &router.Menu{
			Path: "/system/notify", Name: "NotifyCenter",
			Meta:     router.Meta{Title: "通知中心", Icon: "ep:bell", ShowParent: true, ShowLink: true},
			MenuType: router.MenuTypeDir,
			Children: []*router.Menu{menuInbox, menuSend, menuSchedule, menuMessage, menuType, menuTemplate, menuDelivery, menuAnnouncement},
		}
	})
	owl.PanicIf(err)
}
