package notification

import (
	"context"

	"bit-labs.cn/owl"
	"bit-labs.cn/owl-admin/app-notification/contract"
	"bit-labs.cn/owl-admin/app-notification/database"
	v1 "bit-labs.cn/owl-admin/app-notification/handle/v1"
	"bit-labs.cn/owl-admin/app-notification/repository"
	"bit-labs.cn/owl-admin/app-notification/route"
	"bit-labs.cn/owl-admin/app-notification/service"
	"bit-labs.cn/owl-admin/app-notification/service/plugins"
	"bit-labs.cn/owl-admin/app-notification/subscribe"
	"bit-labs.cn/owl/contract/foundation"
	"bit-labs.cn/owl/contract/log"
	"bit-labs.cn/owl/provider/mailer"
	"bit-labs.cn/owl/provider/router"
	"bit-labs.cn/owl/provider/setting"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

var _ owl.SubApp = (*SubAppNotification)(nil)

// SubAppNotification 是通知中心子系统。
type SubAppNotification struct {
	app foundation.Application
}

func (s *SubAppNotification) Name() string { return "notify" }

func (s *SubAppNotification) Bootstrap() {
	_ = s.app.Invoke(func(
		notifSvc *service.NotificationService,
		hub *service.Hub,
		mail *mailer.MailerManager,
		logger log.Logger,
	) {
		// 1. 注册默认渠道插件（如邮件）
		if mail != nil {
			notifSvc.RegisterChannel(plugins.NewEmailPlugin(mail))
		}
		notifSvc.RegisterChannel(plugins.NewWeComPlugin(""))
		// 2. 启动 Redis SSE 实时推送 Hub
		hub.Start()
		// 3. 注册全局 contract.Notifier 单例供各业务域直接使用
		contract.Register(notifSvc)
		// 4. 初始化事件总线监听
		subscribe.Init(s.app)
	})
}

func (s *SubAppNotification) RegisterConfigs() setting.Catalog {
	return service.NotifySettingCatalog()
}

func (s *SubAppNotification) ServiceProviders() []foundation.ServiceProvider { return nil }

func (s *SubAppNotification) RegisterRouters() { route.InitAPI(s.app, s.Name()) }

func (s *SubAppNotification) RegisterMenus() []*router.Menu { return route.InitMenu() }

func (s *SubAppNotification) RegisterCommands() []*cobra.Command { return nil }

func (s *SubAppNotification) RegisterMigrate() []any { return database.Models() }

func (s *SubAppNotification) BeforeMigrate(db *gorm.DB) error {
	return nil
}

func (s *SubAppNotification) AfterMigrate(db *gorm.DB) error {
	return s.app.Invoke(func(
		notifSvc *service.NotificationService,
		worker *service.Worker,
		logger log.Logger,
	) {
		if err := notifSvc.EnsureBuiltinTypes(context.Background()); err != nil {
			logger.Error("notify: seed builtin types failed: ", err)
		}
		worker.Start()
	})
}

// Binds 注册核心仓库、服务与处理器。
func (s *SubAppNotification) Binds() []any {
	return []any{
		repository.NewMessageRepository,
		repository.NewAnnouncementRepository,
		repository.NewCatalogRepository,
		service.NewHub,
		service.NewNotificationService,
		service.NewWorker,
		v1.NewAnnouncementHandle,
		v1.NewMessageHandle,
		v1.NewTypeHandle,
		v1.NewTemplateHandle,
		v1.NewDeliveryHandle,
		v1.NewSendHandle,
		v1.NewScheduleHandle,
		v1.NewMessageAdminHandle,
	}
}
