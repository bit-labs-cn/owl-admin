package oauth

import (
	"bit-labs.cn/owl-admin/app/provider/jwt"
	"bit-labs.cn/owl-admin/app/service"
	"bit-labs.cn/owl/provider/conf"
	"bit-labs.cn/owl/provider/router"

	"encoding/json"
	"errors"
	"io"
	"net/http"

	errContract "bit-labs.cn/owl/contract/errors"
	"bit-labs.cn/owl/contract/foundation"
	"github.com/gin-gonic/gin"
	socketio "github.com/googollee/go-socket.io"
	"github.com/jinzhu/copier"
	"github.com/spf13/cast"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

type Handle struct {
	app     foundation.Application
	userSvc *service.UserService
	sio     *socketio.Server
	jwtSvc  *jwt.JWTService
}

func (i *Handle) ModuleName() (en string, zh string) {
	return "oauth", "oauth"
}

var _ router.Handler = (*Handle)(nil)

var GitHubConfig = &oauth2.Config{
	ClientID:     "",
	ClientSecret: "",
	RedirectURL:  "",
	Scopes:       []string{"user:email"},
	Endpoint:     github.Endpoint,
}

var GoogleConfig = &oauth2.Config{
	ClientID:     "",
	ClientSecret: "",
	RedirectURL:  "",
	Scopes:       []string{"openid", "profile", "email"},
	Endpoint:     google.Endpoint,
}

var GiteeConfig = &oauth2.Config{
	ClientID:     "",
	ClientSecret: "",
	RedirectURL:  "",
	Scopes:       []string{"user_info"},
	Endpoint: oauth2.Endpoint{
		AuthURL:  "https://gitee.com/oauth/authorize",
		TokenURL: "https://gitee.com/oauth/token",
	},
}

func NewOauthHandle(
	app foundation.Application,
	configure *conf.Configure,
	userService *service.UserService,
	sio *socketio.Server,
	jwtSvc *jwt.JWTService) *Handle {
	var c struct {
		ClientID     string `json:"client-id"`
		ClientSecret string `json:"client-secret"`
		RedirectURL  string `json:"redirect-url"`
	}

	_ = configure.GetConfig("oauth.github", &c)
	_ = copier.Copy(&GitHubConfig, &c)

	_ = configure.GetConfig("oauth.gitee", &c)
	_ = copier.Copy(&GiteeConfig, &c)

	_ = configure.GetConfig("oauth.google", &c)
	_ = copier.Copy(&GoogleConfig, &c)

	return &Handle{
		app:     app,
		userSvc: userService,
		sio:     sio,
		jwtSvc:  jwtSvc,
	}
}

// @Summary		第三方登录
// @Description	通过第三方平台（GitHub、Google、Gitee）进行OAuth登录
// @Tags			OAuth认证
// @Produce		json
// @Param			provider	path	string	true	"第三方平台"	Enums(github,google,gitee)
// @Param			state		query	string	false	"状态参数"
// @Success		302			"重定向到第三方授权页面"
// @Failure		400			{object}	router.Resp	"参数错误"
// @Router			/oauth/{provider}/login [GET]
func (i *Handle) Login(c *gin.Context) {
	provider := c.Param("provider")
	conf := getOAuthConfig(provider)
	state := c.Query("state")
	if conf == nil {
		c.JSON(http.StatusBadRequest, router.Resp{Success: false, Msg: "Unsupported provider"})
		return
	}
	url := conf.AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}
func getOAuthConfig(provider string) *oauth2.Config {
	switch provider {
	case "github":
		return GitHubConfig
	case "google":
		return GoogleConfig
	case "gitee":
		return GiteeConfig
	default:
		return nil
	}
}

// @Summary		第三方授权回调
// @Description	处理第三方平台的OAuth授权回调，完成用户登录
// @Tags			OAuth认证
// @Produce		json
// @Param			provider	path		string		true	"第三方平台"	Enums(github,google,gitee)
// @Param			code		query		string		true	"授权码"
// @Param			state		query		string		false	"状态参数"
// @Success		200			{string}	string		"登录成功页面"
// @Failure		400			{object}	router.Resp	"参数错误"
// @Failure		500			{object}	router.Resp	"服务器内部错误"
// @Router			/oauth/{provider}/callback [GET]
func (i *Handle) Callback(c *gin.Context) {
	provider := c.Param("provider")
	state := c.Query("state")
	conf := getOAuthConfig(provider)
	if conf == nil {
		c.JSON(http.StatusBadRequest, router.Resp{Success: false, Msg: "Unsupported provider"})
		return
	}

	code := c.Query("code")
	token, err := conf.Exchange(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, router.Resp{Success: false, Msg: "Token exchange failed", Data: err.Error()})
		return
	}

	client := conf.Client(c.Request.Context(), token)

	// 获取用户信息
	var userInfoURL string
	switch provider {
	case "github":
		userInfoURL = "https://api.github.com/user"
	case "google":
		userInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"
	case "gitee":
		userInfoURL = "https://gitee.com/api/v5/user"
	}

	resp, err := client.Get(userInfoURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, router.Resp{Success: false, Msg: "Failed to get user info"})
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var user map[string]interface{}
	_ = json.Unmarshal(body, &user)

	// 看看用户是否在数据库中存在，如果存在则更新，不存在则写入
	createUser := &service.CreateUserReq{
		UserBatchFields: service.UserBatchFields{
			Username: cast.ToString(user["login"]),
			NickName: cast.ToString(user["name"]),
			Email:    cast.ToString(user["email"]),
			Phone:    "",
			Remark:   cast.ToString(user["bio"]),
			Status:   0,
			Sex:      3,
			Source:   provider,
			SourceID: cast.ToString(user["id"]),
		},
		Password: "",
	}
	err = i.userSvc.CreateUser(c.Request.Context(), createUser)

	if err != nil {
		var bizErr *errContract.BizError
		if errors.As(err, &bizErr) && bizErr.Code == service.CodeUserExists {
			// 用户已存在：忽略错误，继续进行第三方登录流程。
		} else {
			c.JSON(http.StatusInternalServerError, router.Resp{Success: false, Msg: err.Error()})
			return
		}
	}

	html := `
    <!DOCTYPE html>
    <html>
    <head>
        <title>登录成功</title>
        <script>
            setTimeout(function() {
                window.close();
            }, 5000);
        </script>
        <style>
            body {
                display: flex;
                justify-content: center;
                align-items: center;
                height: 100vh;
                font-family: Arial, sans-serif;
            }
            .success-box {
                text-align: center;
                padding: 20px;
                border: 1px solid #4CAF50;
                border-radius: 5px;
                background-color: #f8f9fa;
            }
        </style>
    </head>
    <body>
        <div class="success-box">
            <h2>登录成功!</h2>
            <p>窗口将在5秒后自动关闭...</p>
            <p>如果窗口没有自动关闭，请手动关闭此窗口</p>
        </div>
    </body>
    </html>
    `

	generateToken, err := i.userSvc.LoginByThirdParty(c.Request.Context(), createUser.Username, provider)
	if err != nil {
		c.JSON(http.StatusInternalServerError, router.Resp{Success: false, Msg: err.Error()})
		return
	}
	i.sio.BroadcastToNamespace("/", state, generateToken.AccessToken)
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}
