package oauth2

import (
	_ "embed"

	"bit-labs.cn/owl"
	middleware2 "bit-labs.cn/owl-admin/app/middleware"
	"bit-labs.cn/owl-admin/app/provider/jwt"
	"bit-labs.cn/owl-admin/app/service"
	"bit-labs.cn/owl/contract/foundation"
	"bit-labs.cn/owl/provider/conf"
	"bit-labs.cn/owl/provider/router"
	"github.com/casbin/casbin/v3"
	"github.com/gin-gonic/gin"
)

// Oauth2ServiceProvider 登记第三方登录，并在首次启动时生成 oauth.yaml。
type Oauth2ServiceProvider struct {
	app foundation.Application
}

var _ foundation.ServiceProvider = (*Oauth2ServiceProvider)(nil)

func (s *Oauth2ServiceProvider) Description() string {
	return "第三方 OAuth2 登录"
}

func (s *Oauth2ServiceProvider) Register() {
	s.app.Register(func(c *conf.Configure) *Catalog {
		return NewCatalog(c)
	})
	s.app.Register(NewOauthHandle)
}

func (s *Oauth2ServiceProvider) Boot() {
	err := s.app.Invoke(func(
		oauthHandle *Handle,
		engine *gin.Engine,
		enforcer *casbin.SyncedEnforcer,
		jwtSvc *jwt.JWTService,
		logService *service.LogService,
	) {
		g := engine.Group("/oauth2", middleware2.PermissionCheck(enforcer, jwtSvc))
		g.Use(middleware2.OperationLog(logService))

		r := router.NewRouteInfoBuilder("admin", oauthHandle, g, router.MenuOption{})
		r.Get("/providers", router.AccessPublic, oauthHandle.Providers).Name("第三方登录提供方").WithoutOperateLog().Build()
		r.Get("/:provider/login", router.AccessPublic, oauthHandle.Login).Name("第三方登录").Build()
		r.Get("/:provider/callback", router.AccessPublic, oauthHandle.Callback).Name("第三方登录回调").Build()
	})
	owl.PanicIf(err)
}

//go:embed oauth2.yaml
var oauthYAML string

func (s *Oauth2ServiceProvider) Conf() map[string]string {
	return map[string]string{
		"oauth2.yaml": oauthYAML,
	}
}
