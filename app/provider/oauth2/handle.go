package oauth2

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"html/template"
	"net/http"

	"bit-labs.cn/owl-admin/app/model"
	"bit-labs.cn/owl-admin/app/provider/jwt"
	"bit-labs.cn/owl-admin/app/service"
	errContract "bit-labs.cn/owl/contract/errors"
	"bit-labs.cn/owl/contract/foundation"
	"bit-labs.cn/owl/provider/router"
	"github.com/gin-gonic/gin"
	socketio "github.com/googollee/go-socket.io"
)

type Handle struct {
	app     foundation.Application
	catalog *Catalog
	userSvc *service.UserService
	sio     *socketio.Server
	jwtSvc  *jwt.JWTService
}

func (i *Handle) ModuleName() (en string, zh string) {
	return "oauth", "oauth"
}

var _ router.Handler = (*Handle)(nil)

func NewOauthHandle(
	app foundation.Application,
	catalog *Catalog,
	userService *service.UserService,
	sio *socketio.Server,
	jwtSvc *jwt.JWTService) *Handle {
	return &Handle{
		app:     app,
		catalog: catalog,
		userSvc: userService,
		sio:     sio,
		jwtSvc:  jwtSvc,
	}
}

// @Summary		已启用的第三方登录
// @Description	返回已启用提供方的标识和显示名称，供登录页展示
// @Tags			OAuth认证
// @Produce		json
// @Success		200	{object}	router.Resp	"操作成功"
// @Router			/oauth2/providers [GET]
func (i *Handle) Providers(c *gin.Context) {
	router.Success(c, i.catalog.List())
}

// @Summary		第三方登录
// @Description	通过第三方平台（GitHub、Google、Gitee、owl-oauth）进行 OAuth 登录
// @Tags			OAuth认证
// @Produce		json
// @Param			provider	path	string	true	"第三方平台"	Enums(github,google,gitee,owl)
// @Param			state		query	string	false	"状态参数"
// @Success		302			"重定向到第三方授权页面"
// @Failure		400			{object}	router.Resp	"参数错误"
// @Router			/oauth2/{provider}/login [GET]
func (i *Handle) Login(c *gin.Context) {
	p := i.catalog.Find(c.Param("provider"))
	if p == nil {
		c.JSON(http.StatusBadRequest, router.Resp{Success: false, Msg: "Unsupported provider"})
		return
	}
	c.Redirect(http.StatusTemporaryRedirect, p.AuthCodeURL(c.Query("state")))
}

// @Summary		第三方授权回调
// @Description	处理第三方平台的 OAuth 授权回调，完成用户登录
// @Tags			OAuth认证
// @Produce		json
// @Param			provider	path		string		true	"第三方平台"	Enums(github,google,gitee,owl)
// @Param			code		query		string		true	"授权码"
// @Param			state		query		string		false	"状态参数"
// @Success		200			{string}	string		"登录成功页面"
// @Failure		400			{object}	router.Resp	"参数错误"
// @Failure		500			{object}	router.Resp	"服务器内部错误"
// @Router			/oauth2/{provider}/callback [GET]
func (i *Handle) Callback(c *gin.Context) {
	provider := c.Param("provider")
	p := i.catalog.Find(provider)
	if p == nil {
		c.JSON(http.StatusBadRequest, router.Resp{Success: false, Msg: "Unsupported provider"})
		return
	}

	id, err := p.Identity(c.Request.Context(), c.Query("code"))
	if err != nil {
		msg := "Failed to get user info"
		if errors.Is(err, ErrExchange) {
			msg = "Token exchange failed"
		}
		c.JSON(http.StatusInternalServerError, router.Resp{Success: false, Msg: msg, Data: err.Error()})
		return
	}

	password, err := randomPassword()
	if err != nil {
		c.JSON(http.StatusInternalServerError, router.Resp{Success: false, Msg: err.Error()})
		return
	}

	createUser := &service.CreateUserReq{
		UserBatchFields: service.UserBatchFields{
			Username: id.Username,
			NickName: id.NickName,
			Email:    id.Email,
			Phone:    id.Phone,
			Remark:   id.Remark,
			Status:   model.UserStatusEnabled,
			Sex:      model.UserSexUnknown,
			Source:   provider,
			SourceID: id.SourceID,
		},
		Password:  password,
		SourceRaw: id.Raw,
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

	generateToken, err := i.userSvc.LoginByThirdParty(c.Request.Context(), createUser.Username, provider, createUser.SourceID, id.Raw)
	if err != nil {
		c.JSON(http.StatusInternalServerError, router.Resp{Success: false, Msg: err.Error()})
		return
	}
	state := c.Query("state")

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, renderLoginSuccess(generateToken.AccessToken, state))
}

func randomPassword() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func renderLoginSuccess(token, state string) string {
	raw, err := json.Marshal(map[string]string{
		"source": "owl-oauth",
		"token":  token,
		"state":  state,
	})
	if err != nil {
		raw = []byte("null")
	}
	var buf bytes.Buffer
	_ = loginSuccessTmpl.Execute(&buf, struct {
		Message template.JS
	}{Message: template.JS(raw)})
	return buf.String()
}

var loginSuccessTmpl = template.Must(template.New("login-success").Parse(`<!DOCTYPE html>
<html>
<head>
    <title>登录成功</title>
    <script>
        (function () {
            var msg = {{.Message}};
            try {
                if (window.opener && !window.opener.closed) {
                    window.opener.postMessage(msg, "*");
                }
            } catch (e) {}
            setTimeout(function () { window.close(); }, 600);
        })();
    </script>
    <style>
        body {
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100vh;
            margin: 0;
            font-family: "PingFang SC", "Microsoft YaHei", sans-serif;
            background: #f4f7fb;
            color: #0f172a;
        }
        .success-box {
            text-align: center;
            padding: 28px 32px;
            border-radius: 16px;
            background: #fff;
            box-shadow: 0 12px 32px rgba(15, 23, 42, 0.08);
        }
        h2 { margin: 0 0 8px; font-size: 20px; }
        p { margin: 0; color: #64748b; font-size: 14px; }
    </style>
</head>
<body>
    <div class="success-box">
        <h2>登录成功</h2>
        <p>正在返回，窗口将自动关闭</p>
    </div>
</body>
</html>`))
