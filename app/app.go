package admin

import (
	"bit-labs.cn/owl"
	"bit-labs.cn/owl-admin/app/cmd"
	"bit-labs.cn/owl-admin/app/database"
	"bit-labs.cn/owl-admin/app/database/seeder"
	"bit-labs.cn/owl-admin/app/listener"
	"bit-labs.cn/owl-admin/app/provider/jwt"
	"bit-labs.cn/owl-admin/app/route"
	"bit-labs.cn/owl/contract/foundation"
	"bit-labs.cn/owl/provider/captcha"
	"bit-labs.cn/owl/provider/db"
	"bit-labs.cn/owl/provider/permission"
	"bit-labs.cn/owl/provider/redis"
	"bit-labs.cn/owl/provider/router"
	"bit-labs.cn/owl/provider/socketio"
	"bit-labs.cn/owl/provider/storage"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"bit-labs.cn/owl-admin/app/handle/oauth"
	v1 "bit-labs.cn/owl-admin/app/handle/v1"
	"bit-labs.cn/owl-admin/app/repository"
	"bit-labs.cn/owl-admin/app/service"
)

var _ owl.SubApp = (*SubAppAdmin)(nil)

type SubAppAdmin struct {
	app foundation.Application
}

func (i *SubAppAdmin) Name() string {
	return "admin"
}

func (i *SubAppAdmin) Bootstrap() {
	i.app.Invoke(func(gdb *gorm.DB) {
		migDB := gdb.Session(&gorm.Session{Logger: gdb.Config.Logger.LogMode(logger.Error)})
		go database.Migrate(migDB)
		go seeder.InitAllDictData(migDB)
		listener.Init(i.app)
	})
}

func (i *SubAppAdmin) ServiceProviders() []foundation.ServiceProvider {
	return []foundation.ServiceProvider{
		&permission.GuardProvider{},
		&router.RouterServiceProvider{},
		&db.DBServiceProvider{},
		&jwt.JwtServiceProvider{},
		&redis.RedisServiceProvider{},
		&socketio.SocketIOServiceProvider{},
		&captcha.CaptchaServiceProvider{},
		&storage.StorageServiceProvider{},
	}
}
func (i *SubAppAdmin) Menu() []*router.Menu {
	return route.InitMenu()
}

func (i *SubAppAdmin) Commands() []*cobra.Command {
	return []*cobra.Command{
		cmd.Version,
		cmd.GenPwd,
	}
}

func (i *SubAppAdmin) RegisterRouters() {
	route.InitWeb(i.app)
	route.InitApi(i.app, i.Name())
}

func (i *SubAppAdmin) Binds() []any {
	return []any{
		oauth.NewOauthHandle,
		v1.NewApiHandle,
		v1.NewAppUpgradeHandle,
		storage.NewFileHandle,
		v1.NewDeptHandle,
		v1.NewDictHandle,
		v1.NewMenuHandle,
		v1.NewRoleHandle,
		v1.NewPositionHandle,
		v1.NewAreaHandle,
		v1.NewLogHandle,
		v1.NewUserHandle,
		v1.NewUserGroupHandle,
		service.NewDeptService,
		service.NewDictService,
		service.NewRoleService,
		service.NewLogService,
		service.NewUserService,
		service.NewAreaService,
		service.NewAppVersionService,
		repository.NewLogRepository,
		repository.NewDeptRepository,
		repository.NewDictRepository,
		repository.NewRoleRepository,
		repository.NewPositionRepository,
		repository.NewUserRepository,
		repository.NewAreaRepository,
		repository.NewAppVersionRepository,
		service.NewPositionService,
		service.NewUserGroupService,
		repository.NewUserGroupRepository,
	}
}
