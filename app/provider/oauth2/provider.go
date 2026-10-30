package oauth2

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"sync"

	"bit-labs.cn/owl/provider/conf"
	"github.com/spf13/cast"
)

// Identity 第三方账号映射到本地用户的字段。
type Identity struct {
	Username string
	NickName string
	Email    string
	Phone    string
	Remark   string
	SourceID string
	Raw      string // 第三方返回的用户资料原文，不含令牌
}

// Provider 一个可注册的登录提供方。新增提供方实现该接口并 register，不必改登录和回调。
type Provider interface {
	Name() string
	DisplayName() string
	Load(configure *conf.Configure)
	Enabled() bool
	AuthCodeURL(state string) string
	Identity(ctx context.Context, code string) (*Identity, error)
}

var (
	ErrExchange = errors.New("token exchange failed")
	ErrUserInfo = errors.New("failed to get user info")

	factories = map[string]func() Provider{}
)

func registerFactory(name string, fn func() Provider) {
	factories[name] = fn
}

// Catalog 按提供方名称按需读取配置。未访问过的提供方不会加载。
type Catalog struct {
	configure *conf.Configure
	mu        sync.Mutex
	cache     map[string]Provider
}

func NewCatalog(configure *conf.Configure) *Catalog {
	return &Catalog{
		configure: configure,
		cache:     map[string]Provider{},
	}
}

func (c *Catalog) get(name string) Provider {
	c.mu.Lock()
	defer c.mu.Unlock()
	if p, ok := c.cache[name]; ok {
		return p
	}
	fn, ok := factories[name]
	if !ok {
		return nil
	}
	p := fn()
	if c.configure != nil {
		p.Load(c.configure)
	}
	c.cache[name] = p
	return p
}

// Find 返回 enable 为 true 的提供方。第一次访问时才读取该提供方的配置。
func (c *Catalog) Find(name string) Provider {
	p := c.get(name)
	if p == nil || !p.Enabled() {
		return nil
	}
	return p
}

// ProviderView 是登录页展示的提供方。
type ProviderView struct {
	Provider string `json:"provider"`
	Name     string `json:"name"`
}

// List 返回 enable 为 true 的提供方标识和显示名称。
func (c *Catalog) List() []ProviderView {
	names := make([]string, 0, len(factories))
	for name := range factories {
		names = append(names, name)
	}
	sort.Strings(names)
	list := make([]ProviderView, 0)
	for _, name := range names {
		p := c.get(name)
		if p == nil || !p.Enabled() {
			continue
		}
		label := strings.TrimSpace(p.DisplayName())
		if label == "" {
			label = p.Name()
		}
		list = append(list, ProviderView{Provider: p.Name(), Name: label})
	}
	return list
}

func profileRaw(raw map[string]any) string {
	if len(raw) == 0 {
		return ""
	}
	cleaned := make(map[string]any, len(raw))
	for key, value := range raw {
		switch strings.ToLower(key) {
		case "access_token", "refresh_token", "id_token", "token", "client_secret":
			continue
		default:
			cleaned[key] = value
		}
	}
	body, err := json.Marshal(cleaned)
	if err != nil {
		return ""
	}
	return string(body)
}

func firstString(raw map[string]any, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(cast.ToString(raw[key])); value != "" {
			return value
		}
	}
	return ""
}
