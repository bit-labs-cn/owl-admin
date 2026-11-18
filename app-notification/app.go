package notification

import (
	"bit-labs.cn/owl"
	"bit-labs.cn/owl-admin/app-notification/contract"
	"bit-labs.cn/owl-admin/app-notification/database"
	v1 "bit-labs.cn/owl-admin/app-notification/handle/v1"
	"bit-labs.cn/owl-admin/app-notification/repository"
	"bit-labs.cn/owl-admin/app-notification/route"
	"bit-labs.cn/owl-admin/app-notification/service"
	"bit-labs.cn/owl-admin/app-notification/subscribe"
	"bit-labs.cn/owl/contract/foundation"
	"bit-labs.cn/owl/provider/mailer"
	"bit-labs.cn/owl/provider/router"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

var _ owl.SubApp = (*SubAppNotification)(nil)

// SubAppNotification 是通知中心子系统。它跟着 owl-admin 走，但作为独立子应用
// 注册，迁移、菜单和路由都走子应用契约，不占用 admin 的 ServiceProvider。
type SubAppNotification struct {
	app foundation.Application
}

func (s *SubAppNotification) Name() string { return "notify" }

// Bootstrap 只做与数据库无关的装配。投递扫描会查表，放到 AfterMigrate 启动，
// 避免表还没建好就开始查询。
func (s *SubAppNotification) Bootstrap() {
	_ = s.app.Invoke(func(
		email *service.EmailChannel,
		hub *service.Hub,
		notifier *service.Notifier,
		mail *mailer.MailerManager,
	) {
		email.SetMailer(mail)
		hub.Start()
		contract.Register(notifier)
		subscribe.Init(s.app)
	})
}

func (s *SubAppNotification) ServiceProviders() []foundation.ServiceProvider { return nil }

func (s *SubAppNotification) RegisterRouters() { route.InitAPI(s.app, s.Name()) }

func (s *SubAppNotification) RegisterMenus() []*router.Menu { return route.InitMenu() }

func (s *SubAppNotification) RegisterCommands() []*cobra.Command { return nil }

func (s *SubAppNotification) RegisterMigrate() []any { return database.Models() }

func (s *SubAppNotification) BeforeMigrate(db *gorm.DB) error {
	return database.BackfillNulls(db)
}

func (s *SubAppNotification) AfterMigrate(db *gorm.DB) error {
	return s.app.Invoke(func(worker *service.Worker) {
		worker.Start()
	})
}

// Binds 顺序：repository → service → handle。
func (s *SubAppNotification) Binds() []any {
	return []any{
		repository.NewTemplateRepository,
		repository.NewTaskRepository,
		repository.NewDeliveryRepository,
		repository.NewInboxRepository,
		repository.NewDeviceRepository,
		repository.NewAnnouncementRepository,
		repository.NewSubscriptionRepository,
		repository.NewContactRepository,
		repository.NewChannelAccountRepository,
		repository.NewEventRepository,
		repository.NewSettingRepository,
		service.NewAudienceResolver,
		service.NewHub,
		service.NewEmailChannel,
		service.NewSMSChannel,
		service.NewAppChannel,
		service.NewTemplateService,
		service.NewTaskService,
		service.NewInboxService,
		service.NewAnnouncementService,
		service.NewContactService,
		service.NewChannelAccountService,
		service.NewDeviceService,
		service.NewDeliveryService,
		service.NewEventService,
		service.NewSettingService,
		service.NewSubscriptionService,
		service.NewNotifier,
		service.NewWorker,
		v1.NewTemplateHandle,
		v1.NewTaskHandle,
		v1.NewAnnouncementHandle,
		v1.NewAudienceHandle,
		v1.NewContactHandle,
		v1.NewChannelHandle,
		v1.NewDeviceHandle,
		v1.NewDeliveryHandle,
		v1.NewEventHandle,
		v1.NewSettingHandle,
		v1.NewInboxHandle,
		v1.NewSubscriptionHandle,
	}
}
