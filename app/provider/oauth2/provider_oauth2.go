package oauth2

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"bit-labs.cn/owl/provider/conf"
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
	}
	_ = configure.GetConfig("oauth."+p.name, &creds)
	p.enable = creds.Enable
	p.displayName = strings.TrimSpace(creds.Name)
	p.cfg.ClientID = strings.TrimSpace(creds.ClientID)
	p.cfg.ClientSecret = strings.TrimSpace(creds.ClientSecret)
	p.cfg.RedirectURL = strings.TrimSpace(creds.RedirectURL)
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
	}
}
