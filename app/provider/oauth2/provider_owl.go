package oauth2

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"bit-labs.cn/owl/provider/conf"
	"github.com/spf13/cast"
	xoauth2 "golang.org/x/oauth2"
)

type owlProvider struct {
	cfg         *xoauth2.Config
	client      *http.Client
	enable      bool
	displayName string
}

func init() {
	registerFactory("owl", func() Provider {
		return &owlProvider{
			cfg:    &xoauth2.Config{},
			client: &http.Client{Timeout: 15 * time.Second},
		}
	})
}

func (p *owlProvider) Name() string { return "owl" }

func (p *owlProvider) Load(configure *conf.Configure) {
	var creds struct {
		Enable       bool   `json:"enable"`
		Name         string `json:"name"`
		ClientID     string `json:"client-id"`
		ClientSecret string `json:"client-secret"`
		RedirectURL  string `json:"redirect-url"`
		Endpoint     struct {
			AuthURL       string `json:"auth-url"`
			TokenURL      string `json:"token-url"`
			DeviceAuthURL string `json:"device-auth-url"`
			AuthStyle     any    `json:"auth-style"`
		} `json:"endpoint"`
	}
	_ = configure.GetConfig("oauth.owl", &creds)
	p.enable = creds.Enable
	p.displayName = strings.TrimSpace(creds.Name)
	p.cfg.ClientID = strings.TrimSpace(creds.ClientID)
	p.cfg.ClientSecret = strings.TrimSpace(creds.ClientSecret)
	p.cfg.RedirectURL = strings.TrimSpace(creds.RedirectURL)
	p.cfg.Endpoint = xoauth2.Endpoint{
		AuthURL:       strings.TrimSpace(creds.Endpoint.AuthURL),
		TokenURL:      strings.TrimSpace(creds.Endpoint.TokenURL),
		DeviceAuthURL: strings.TrimSpace(creds.Endpoint.DeviceAuthURL),
		AuthStyle:     authStyleFrom(creds.Endpoint.AuthStyle),
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

func (p *owlProvider) DisplayName() string { return p.displayName }

func (p *owlProvider) Enabled() bool {
	return p.enable
}

func (p *owlProvider) AuthCodeURL(state string) string {
	return p.cfg.AuthCodeURL(state)
}

func (p *owlProvider) Identity(ctx context.Context, code string) (*Identity, error) {
	payload, err := json.Marshal(map[string]string{
		"grant_type":    "authorization_code",
		"client_id":     p.cfg.ClientID,
		"client_secret": p.cfg.ClientSecret,
		"code":          code,
		"redirect_uri":  p.cfg.RedirectURL,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrExchange, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.cfg.Endpoint.TokenURL, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrExchange, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := p.client
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrExchange, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrExchange, err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status %d", ErrExchange, resp.StatusCode)
	}
	return parseOwlIdentity(body)
}

func parseOwlIdentity(body []byte) (*Identity, error) {
	var envelope struct {
		Success bool   `json:"success"`
		Msg     string `json:"msg"`
		Data    struct {
			UserID   string `json:"user_id"`
			Username string `json:"username"`
			Nickname string `json:"nickname"`
			Mobile   string `json:"mobile"`
			Email    string `json:"email"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrExchange, err)
	}
	if !envelope.Success {
		msg := strings.TrimSpace(envelope.Msg)
		if msg == "" {
			msg = "token rejected"
		}
		return nil, fmt.Errorf("%w: %s", ErrExchange, msg)
	}
	id := &Identity{
		Username: strings.TrimSpace(envelope.Data.Username),
		NickName: strings.TrimSpace(envelope.Data.Nickname),
		Email:    strings.TrimSpace(envelope.Data.Email),
		Phone:    strings.TrimSpace(envelope.Data.Mobile),
		SourceID: strings.TrimSpace(envelope.Data.UserID),
	}
	if id.Username == "" || id.SourceID == "" {
		return nil, fmt.Errorf("%w: incomplete user info", ErrUserInfo)
	}
	return id, nil
}
