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
	menuTask         *router.Menu
	menuTemplate     *router.Menu
	menuAnnouncement *router.Menu
	menuContact      *router.Menu
	menuChannel      *router.Menu
	menuDevice       *router.Menu
	menuDelivery     *router.Menu
	menuEvent        *router.Menu
	menuSetting      *router.Menu
	menuInbox        *router.Menu
	menuSubscription *router.Menu
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
		template *v1.TemplateHandle,
		task *v1.TaskHandle,
		announcement *v1.AnnouncementHandle,
		audience *v1.AudienceHandle,
		contact *v1.ContactHandle,
		channel *v1.ChannelHandle,
		device *v1.DeviceHandle,
		delivery *v1.DeliveryHandle,
		event *v1.EventHandle,
		setting *v1.SettingHandle,
		inbox *v1.InboxHandle,
		subscription *v1.SubscriptionHandle,
	) {
		gv1 := engine.Group("/api/v1", adminmw.PermissionCheck(enforcer, jwtSvc))
		gv1.Use(adminmw.OperationLog(logSvc))

		{
			// 受众选择和深链页面是「发通知时看得懂」的基础能力，只要求登录，
			// 不额外授予通知管理权限，也不会生成菜单。
			r := router.NewRouteInfoBuilder(appName, audience, gv1, router.MenuOption{ComponentName: "NotifyAudienceOptions", Path: "/system/notify/audience/index"})
			r.Get("/notify/audience-options", router.AccessAuthenticated, audience.Options).Name("受众选择项").Build()
			r.Get("/notify/audience-users", router.AccessAuthenticated, audience.Users).Name("搜索用户").Build()
			r.Get("/notify/link-options", router.AccessAuthenticated, audience.Links).Name("通知跳转页面").Build()
			r.Get("/notify/sms-templates", router.AccessAuthenticated, audience.SmsTemplates).Name("短信模板选项").Build()
		}

		{
			r := router.NewRouteInfoBuilder(appName, task, gv1, router.MenuOption{ComponentName: "NotifyTask", Path: "/system/notify/task/index", Icon: "ep:promotion"})
			r.Get("/notify/tasks", router.AccessAuthorized, task.Retrieve).Name("通知任务列表").Build()
			r.Post("/notify/tasks", router.AccessAuthorized, task.Create).Name("创建通知任务").Build()
			r.Put("/notify/tasks/:id", router.AccessAuthorized, task.Update).Name("更新通知任务").Build()
			r.Delete("/notify/tasks/:id", router.AccessAuthorized, task.Delete).Name("删除通知任务").Build()
			r.Get("/notify/tasks/:id", router.AccessAuthorized, task.Detail).Name("通知任务详情").Build()
			r.Post("/notify/tasks-estimate", router.AccessAuthorized, task.Estimate).Name("预估通知人数").Build()
			r.Post("/notify/tasks/:id/submit", router.AccessAuthorized, task.Submit).Name("发送通知任务").Build()
			r.Post("/notify/tasks/:id/approve", router.AccessAuthorized, task.Approve).Name("审核通知任务").Build()
			r.Post("/notify/tasks/:id/cancel", router.AccessAuthorized, task.Cancel).Name("取消通知任务").Build()
			menuTask = r.GetMenu()
		}
		{
			r := router.NewRouteInfoBuilder(appName, template, gv1, router.MenuOption{ComponentName: "NotifyTemplate", Path: "/system/notify/template/index", Icon: "ep:document"})
			r.Get("/notify/templates", router.AccessAuthorized, template.Retrieve).Name("通知模板列表").Build()
			r.Post("/notify/templates", router.AccessAuthorized, template.Create).Name("创建通知模板").Build()
			r.Put("/notify/templates/:id", router.AccessAuthorized, template.Update).Name("更新通知模板").Build()
			r.Delete("/notify/templates/:id", router.AccessAuthorized, template.Delete).Name("删除通知模板").Build()
			r.Get("/notify/templates/:id", router.AccessAuthorized, template.Detail).Name("通知模板详情").Build()
			r.Post("/notify/templates-preview", router.AccessAuthorized, template.Preview).Name("预览通知模板").Build()
			menuTemplate = r.GetMenu()
		}
		{
			r := router.NewRouteInfoBuilder(appName, announcement, gv1, router.MenuOption{ComponentName: "NotifyAnnouncement", Path: "/system/notify/announcement/index", Icon: "ep:bell"})
			r.Get("/notify/announcements", router.AccessAuthorized, announcement.Retrieve).Name("公告列表").Build()
			r.Post("/notify/announcements", router.AccessAuthorized, announcement.Create).Name("创建公告").Build()
			r.Put("/notify/announcements/:id", router.AccessAuthorized, announcement.Update).Name("更新公告").Build()
			r.Delete("/notify/announcements/:id", router.AccessAuthorized, announcement.Delete).Name("删除公告").Build()
			r.Get("/notify/announcements/:id", router.AccessAuthorized, announcement.Detail).Name("公告详情").Build()
			r.Post("/notify/announcements/:id/publish", router.AccessAuthorized, announcement.Publish).Name("发布公告").Build()
			r.Post("/notify/announcements/:id/offline", router.AccessAuthorized, announcement.Offline).Name("下线公告").Build()
			menuAnnouncement = r.GetMenu()
		}
		{
			r := router.NewRouteInfoBuilder(appName, contact, gv1, router.MenuOption{ComponentName: "NotifyContact", Path: "/system/notify/contact/index", Icon: "ep:user"})
			r.Get("/notify/contact-lists", router.AccessAuthorized, contact.Lists).Name("名单列表").Build()
			r.Post("/notify/contact-lists", router.AccessAuthorized, contact.SaveList).Name("创建名单").Build()
			r.Put("/notify/contact-lists/:id", router.AccessAuthorized, contact.SaveList).Name("更新名单").Build()
			r.Delete("/notify/contact-lists/:id", router.AccessAuthorized, contact.DeleteList).Name("删除名单").Build()
			r.Get("/notify/contacts", router.AccessAuthorized, contact.Contacts).Name("联系人列表").Build()
			r.Post("/notify/contacts", router.AccessAuthorized, contact.SaveContact).Name("创建联系人").Build()
			r.Put("/notify/contacts/:id", router.AccessAuthorized, contact.SaveContact).Name("更新联系人").Build()
			r.Delete("/notify/contacts/:id", router.AccessAuthorized, contact.DeleteContact).Name("删除联系人").Build()
			r.Post("/notify/contacts-import", router.AccessAuthorized, contact.Import).Name("导入联系人").Build()
			menuContact = r.GetMenu()
		}
		{
			r := router.NewRouteInfoBuilder(appName, channel, gv1, router.MenuOption{ComponentName: "NotifyChannel", Path: "/system/notify/channel/index", Icon: "ep:connection"})
			r.Get("/notify/channel-accounts", router.AccessAuthorized, channel.Retrieve).Name("渠道账号列表").Build()
			r.Post("/notify/channel-accounts", router.AccessAuthorized, channel.Save).Name("创建渠道账号").Build()
			r.Put("/notify/channel-accounts/:id", router.AccessAuthorized, channel.Save).Name("更新渠道账号").Build()
			r.Delete("/notify/channel-accounts/:id", router.AccessAuthorized, channel.Delete).Name("删除渠道账号").Build()
			r.Get("/notify/channel-accounts/:id", router.AccessAuthorized, channel.Detail).Name("渠道账号详情").Build()
			r.Post("/notify/channel-accounts/:id/test", router.AccessAuthorized, channel.Test).Name("测试渠道发送").Build()
			menuChannel = r.GetMenu()
		}
		{
			r := router.NewRouteInfoBuilder(appName, device, gv1, router.MenuOption{ComponentName: "NotifyDevice", Path: "/system/notify/device/index", Icon: "ep:cellphone"})
			r.Get("/notify/devices", router.AccessAuthorized, device.Retrieve).Name("推送设备列表").Build()
			r.Post("/notify/devices/:id/invalidate", router.AccessAuthorized, device.Invalidate).Name("失效推送设备").Build()
			menuDevice = r.GetMenu()
		}
		{
			r := router.NewRouteInfoBuilder(appName, delivery, gv1, router.MenuOption{ComponentName: "NotifyDelivery", Path: "/system/notify/delivery/index", Icon: "ep:list"})
			r.Get("/notify/deliveries", router.AccessAuthorized, delivery.Retrieve).Name("投递记录列表").Build()
			r.Get("/notify/deliveries/:id", router.AccessAuthorized, delivery.Detail).Name("投递记录详情").Build()
			r.Post("/notify/deliveries/:id/retry", router.AccessAuthorized, delivery.Retry).Name("重试投递").Build()
			menuDelivery = r.GetMenu()
		}
		{
			r := router.NewRouteInfoBuilder(appName, event, gv1, router.MenuOption{ComponentName: "NotifyEvent", Path: "/system/notify/event/index", Icon: "ep:collection"})
			r.Get("/notify/events", router.AccessAuthorized, event.Retrieve).Name("通知事件列表").Build()
			r.Post("/notify/events", router.AccessAuthorized, event.Save).Name("创建通知事件").Build()
			r.Put("/notify/events/:id", router.AccessAuthorized, event.Save).Name("更新通知事件").Build()
			r.Delete("/notify/events/:id", router.AccessAuthorized, event.Delete).Name("删除通知事件").Build()
			menuEvent = r.GetMenu()
		}
		{
			r := router.NewRouteInfoBuilder(appName, setting, gv1, router.MenuOption{ComponentName: "NotifySetting", Path: "/system/notify/setting/index", Icon: "ep:setting"})
			r.Get("/notify/settings", router.AccessAuthorized, setting.Get).Name("读取通知设置").Build()
			r.Put("/notify/settings", router.AccessAuthorized, setting.Save).Name("保存通知设置").Build()
			menuSetting = r.GetMenu()
		}
		{
			r := router.NewRouteInfoBuilder(appName, inbox, gv1, router.MenuOption{ComponentName: "NotifyInbox", Path: "/system/notify/inbox/index", Icon: "ep:message"})
			r.Get("/notify/my/counts", router.AccessAuthenticated, inbox.Counts).Name("我的未读数").Build()
			r.Get("/notify/my/inbox", router.AccessAuthenticated, inbox.Retrieve).Name("我的消息").Build()
			r.Get("/notify/my/inbox/recent", router.AccessAuthenticated, inbox.List).Name("铃铛消息").Build()
			r.Post("/notify/my/inbox/read", router.AccessAuthenticated, inbox.Read).Name("标记已读").Build()
			r.Post("/notify/my/inbox/read-all", router.AccessAuthenticated, inbox.ReadAll).Name("全部已读").Build()
			r.Delete("/notify/my/inbox/:id", router.AccessAuthenticated, inbox.Delete).Name("删除我的消息").Build()
			r.Post("/notify/my/todos/:id/done", router.AccessAuthenticated, inbox.Done).Name("完成待办").Build()
			r.Post("/notify/my/todos/:id/reopen", router.AccessAuthenticated, inbox.Reopen).Name("重开待办").Build()
			r.Post("/notify/my/stream-ticket", router.AccessAuthenticated, inbox.Ticket).Name("实时连接票据").Build()
			r.Post("/notify/my/devices", router.AccessAuthenticated, device.Register).Name("登记推送设备").Build()
			r.Delete("/notify/my/devices/:id", router.AccessAuthenticated, device.DeleteMine).Name("注销推送设备").Build()
			r.Get("/notify/my/announcements", router.AccessAuthenticated, announcement.Mine).Name("我的公告").Build()
			r.Post("/notify/my/announcements/:id/dismiss", router.AccessAuthenticated, announcement.Dismiss).Name("关闭公告").Build()
			r.Get("/notify/my/subscriptions", router.AccessAuthenticated, subscription.List).Name("我的订阅").Build()
			r.Put("/notify/my/subscriptions", router.AccessAuthenticated, subscription.Save).Name("保存我的订阅").Build()
			menuInbox = r.GetMenu()
		}
		{
			r := router.NewRouteInfoBuilder(appName, subscription, gv1, router.MenuOption{ComponentName: "NotifySubscription", Path: "/system/notify/subscription/index", Icon: "ep:switch"})
			menuSubscription = r.GetMenu()
		}

		pub := engine.Group("/api/v1")
		pub.GET("/notify/stream", inbox.Stream)
		pub.GET("/notify/unsubscribe", contact.Unsubscribe)

		// 前端子系统包暂未重建，先在侧栏隐藏整个目录，接口与权限保持不变；
		// 前端恢复后把 ShowLink 改回 true 即可。
		menuRoot = &router.Menu{
			Path: "/system/notify", Name: "NotifyCenter",
			Meta:     router.Meta{Title: "通知中心", Icon: "ep:bell", ShowParent: true, ShowLink: false},
			MenuType: router.MenuTypeDir,
			Children: []*router.Menu{menuTask, menuTemplate, menuAnnouncement, menuContact, menuChannel, menuDevice, menuDelivery, menuEvent, menuSetting, menuInbox, menuSubscription},
		}
	})
	owl.PanicIf(err)
}
