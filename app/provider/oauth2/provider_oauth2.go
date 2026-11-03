package oauth2

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"bit-labs.cn/owl/provider/conf"
	"github.com/spf13/cast"
	xoauth2 "golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

type identityFields struct {
	username []string
	nickName []string
	email    []string
	phone    []string
	remark   []string
	id       []string
}

type oauth2Provider struct {
	name        string
	displayName string
	enable      bool
	cfg         *xoauth2.Config
	userInfoURL string
	fields      identityFields
}

func init() {
	registerFactory("github", func() Provider {
		return newOAuth2Provider(
			"github",
			github.Endpoint,
			[]string{"user:email"},
			"https://api.github.com/user",
			identityFields{
				username: []string{"login"},
				nickName: []string{"name"},
				email:    []string{"email"},
				remark:   []string{"bio"},
				id:       []string{"id"},
			},
		)
	})
	registerFactory("google", func() Provider {
		return newOAuth2Provider(
			"google",
			google.Endpoint,
			[]string{"openid", "profile", "email"},
			"https://www.googleapis.com/oauth2/v2/userinfo",
			identityFields{
				username: []string{"email", "id"},
				nickName: []string{"name"},
				email:    []string{"email"},
				id:       []string{"id"},
			},
		)
	})
	registerFactory("gitee", func() Provider {
		return newOAuth2Provider(
			"gitee",
			xoauth2.Endpoint{
				AuthURL:  "https://gitee.com/oauth/authorize",
				TokenURL: "https://gitee.com/oauth/token",
			},
			[]string{"user_info"},
			"https://gitee.com/api/v5/user",
			identityFields{
				username: []string{"login"},
				nickName: []string{"name"},
				email:    []string{"email"},
				remark:   []string{"bio"},
				id:       []string{"id"},
			},
		)
	})
	registerFactory("owl", func() Provider {
		return newOAuth2Provider(
			"owl",
			xoauth2.Endpoint{},
			[]string{"openid", "profile", "email", "phone"},
			"",
			identityFields{
				username: []string{"username"},
				nickName: []string{"nickname"},
				email:    []string{"email"},
				phone:    []string{"mobile"},
				id:       []string{"user_id"},
			},
		)
	})
}

func newOAuth2Provider(name string, endpoint xoauth2.Endpoint, scopes []string, userInfoURL string, fields identityFields) *oauth2Provider {
	return &oauth2Provider{
		name:        name,
		userInfoURL: userInfoURL,
		fields:      fields,
		cfg: &xoauth2.Config{
			Scopes:   scopes,
			Endpoint: endpoint,
		},
	}
}

func (p *oauth2Provider) Name() string { return p.name }

func (p *oauth2Provider) DisplayName() string { return p.displayName }

func (p *oauth2Provider) Load(configure *conf.Configure) {
	var creds struct {
		Enable       bool   `json:"enable"`
		Name         string `json:"name"`
		ClientID     string `json:"client-id"`
		ClientSecret string `json:"client-secret"`
		RedirectURL  string `json:"redirect-url"`
		UserInfoURL  string `json:"user-info-url"`
		Scopes       any    `json:"scopes"`
		Endpoint     struct {
			AuthURL       string `json:"auth-url"`
			TokenURL      string `json:"token-url"`
			DeviceAuthURL string `json:"device-auth-url"`
			AuthStyle     any    `json:"auth-style"`
		} `json:"endpoint"`
	}
	_ = configure.GetConfig("oauth2."+p.name, &creds)
	p.enable = creds.Enable
	p.displayName = strings.TrimSpace(creds.Name)
	p.cfg.ClientID = strings.TrimSpace(creds.ClientID)
	p.cfg.ClientSecret = strings.TrimSpace(creds.ClientSecret)
	p.cfg.RedirectURL = strings.TrimSpace(creds.RedirectURL)
	if scopes := scopesFrom(creds.Scopes); len(scopes) > 0 {
		p.cfg.Scopes = scopes
	}
	if auth := strings.TrimSpace(creds.Endpoint.AuthURL); auth != "" {
		p.cfg.Endpoint.AuthURL = auth
	}
	if token := strings.TrimSpace(creds.Endpoint.TokenURL); token != "" {
		p.cfg.Endpoint.TokenURL = token
	}
	if device := strings.TrimSpace(creds.Endpoint.DeviceAuthURL); device != "" {
		p.cfg.Endpoint.DeviceAuthURL = device
	}
	if style := strings.TrimSpace(cast.ToString(creds.Endpoint.AuthStyle)); style != "" {
		p.cfg.Endpoint.AuthStyle = authStyleFrom(creds.Endpoint.AuthStyle)
	}
	if info := strings.TrimSpace(creds.UserInfoURL); info != "" {
		p.userInfoURL = info
	} else if p.userInfoURL == "" {
		p.userInfoURL = userInfoFromTokenURL(p.cfg.Endpoint.TokenURL)
	}
}

func authStyleFrom(value any) xoauth2.AuthStyle {
	switch strings.ToLower(strings.TrimSpace(cast.ToString(value))) {
	case "1", "in-params", "params":
		return xoauth2.AuthStyleInParams
	case "2", "in-header", "header":
		return xoauth2.AuthStyleInHeader
	default:
		return xoauth2.AuthStyleAutoDetect
	}
}

func scopesFrom(value any) []string {
	switch v := value.(type) {
	case string:
		return splitScopes(v)
	case []string:
		out := make([]string, 0, len(v))
		for _, item := range v {
			out = append(out, splitScopes(item)...)
		}
		return out
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			out = append(out, splitScopes(cast.ToString(item))...)
		}
		return out
	default:
		return nil
	}
}

func splitScopes(raw string) []string {
	return strings.Fields(strings.ReplaceAll(raw, ",", " "))
}

// userInfoFromTokenURL 在未单独配置身份地址时，把以 /token 结尾的令牌地址换成同级 /userinfo。
func userInfoFromTokenURL(tokenURL string) string {
	u, err := url.Parse(strings.TrimSpace(tokenURL))
	if err != nil || u.Scheme == "" || u.Host == "" {
		return ""
	}
	path := strings.TrimSuffix(u.Path, "/")
	if !strings.HasSuffix(path, "/token") {
		return ""
	}
	u.Path = strings.TrimSuffix(path, "/token") + "/userinfo"
	u.RawQuery = ""
	u.Fragment = ""
	return u.String()
}

func (p *oauth2Provider) Enabled() bool {
	return p.enable
}

func (p *oauth2Provider) AuthCodeURL(state string) string {
	return p.cfg.AuthCodeURL(state)
}

func (p *oauth2Provider) Identity(ctx context.Context, code string) (*Identity, error) {
	token, err := p.cfg.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrExchange, err)
	}

	resp, err := p.cfg.Client(ctx, token).Get(p.userInfoURL)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUserInfo, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUserInfo, err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status %d", ErrUserInfo, resp.StatusCode)
	}

	var raw map[string]any
	if err = json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUserInfo, err)
	}
	id := pickIdentity(raw, p.fields)
	if id.Username == "" || id.SourceID == "" {
		return nil, fmt.Errorf("%w: incomplete user info", ErrUserInfo)
	}
	return id, nil
}

func pickIdentity(raw map[string]any, fields identityFields) *Identity {
	return &Identity{
		Username: firstString(raw, fields.username...),
		NickName: firstString(raw, fields.nickName...),
		Email:    firstString(raw, fields.email...),
		Phone:    firstString(raw, fields.phone...),
		Remark:   firstString(raw, fields.remark...),
		SourceID: firstString(raw, fields.id...),
		Raw:      profileRaw(raw),
	}
}
