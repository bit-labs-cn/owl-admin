package oauth2

import (
	"net/url"
	"strings"
	"testing"

	xoauth2 "golang.org/x/oauth2"
)

func TestFindProvider(t *testing.T) {
	cat := NewCatalog(nil)
	if cat.Find("owl") != nil {
		t.Fatal("unconfigured owl should be unsupported")
	}
	if cat.Find("github") != nil || cat.Find("google") != nil || cat.Find("gitee") != nil {
		t.Fatal("providers stay off until enable is set")
	}
	if cat.Find("missing") != nil {
		t.Fatal("unknown provider should be unsupported")
	}
	if len(cat.List()) != 0 {
		t.Fatal("disabled providers should not be listed")
	}
}

func TestPickIdentity(t *testing.T) {
	github := pickIdentity(map[string]any{
		"login":        "octocat",
		"name":         "The Octocat",
		"email":        "octo@example.com",
		"bio":          "hi",
		"id":           float64(42),
		"access_token": "secret",
	}, identityFields{
		username: []string{"login"},
		nickName: []string{"name"},
		email:    []string{"email"},
		remark:   []string{"bio"},
		id:       []string{"id"},
	})
	if github.Username != "octocat" || github.SourceID != "42" || github.Remark != "hi" {
		t.Fatalf("github identity = %+v", github)
	}
	if !strings.Contains(github.Raw, `"login":"octocat"`) || strings.Contains(github.Raw, "access_token") {
		t.Fatalf("github raw = %s", github.Raw)
	}

	google := pickIdentity(map[string]any{
		"email": "a@b.c",
		"name":  "Ann",
		"id":    "gid",
	}, identityFields{
		username: []string{"email", "id"},
		nickName: []string{"name"},
		email:    []string{"email"},
		id:       []string{"id"},
	})
	if google.Username != "a@b.c" || google.SourceID != "gid" {
		t.Fatalf("google identity = %+v", google)
	}
}

func TestAuthStyleFrom(t *testing.T) {
	if authStyleFrom("in-params") != xoauth2.AuthStyleInParams || authStyleFrom(float64(2)) != xoauth2.AuthStyleInHeader {
		t.Fatal("auth style parse mismatch")
	}
	if authStyleFrom("") != xoauth2.AuthStyleAutoDetect {
		t.Fatal("empty auth style should auto detect")
	}
}

func TestOAuth2EnableAndName(t *testing.T) {
	p := newOAuth2Provider("github", xoauth2.Endpoint{}, nil, "https://api.github.com/user", identityFields{})
	p.displayName = "GitHub"
	if p.Enabled() {
		t.Fatal("enable false should turn github off")
	}
	if p.DisplayName() != "GitHub" {
		t.Fatalf("display name = %s", p.DisplayName())
	}
	p.enable = true
	if !p.Enabled() {
		t.Fatal("enable alone should turn github on")
	}
}

func TestOwlEnableAndName(t *testing.T) {
	p := &owlProvider{
		enable:      false,
		displayName: "统一身份认证",
		cfg:         &xoauth2.Config{},
	}
	if p.Enabled() {
		t.Fatal("enable false should turn owl off")
	}
	if p.DisplayName() != "统一身份认证" {
		t.Fatalf("display name = %s", p.DisplayName())
	}
	p.enable = true
	if !p.Enabled() {
		t.Fatal("enable alone should turn owl on")
	}
}

func TestOwlAuthCodeURL(t *testing.T) {
	authURL := "http://sso.local/api/v1/oauth/authorize"
	p := &owlProvider{
		enable: true,
		cfg: &xoauth2.Config{
			ClientID:     "owl-admin",
			ClientSecret: "secret",
			RedirectURL:  "http://admin.local/oauth2/owl/callback",
			Endpoint: xoauth2.Endpoint{
				AuthURL:   authURL,
				TokenURL:  "http://sso.local/api/v1/oauth/token",
				AuthStyle: xoauth2.AuthStyleInParams,
			},
		},
	}
	if !p.Enabled() {
		t.Fatal("configured owl should be enabled")
	}
	raw := p.AuthCodeURL("room 1")
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	endpoint, err := url.Parse(authURL)
	if err != nil {
		t.Fatal(err)
	}
	if u.Scheme != endpoint.Scheme || u.Host != endpoint.Host || u.Path != endpoint.Path {
		t.Fatalf("auth url = %s", raw)
	}
	q := u.Query()
	if q.Get("client_id") != "owl-admin" || q.Get("redirect_uri") != p.cfg.RedirectURL || q.Get("response_type") != "code" || q.Get("state") != "room 1" {
		t.Fatalf("query = %s", u.RawQuery)
	}
}

func TestParseOwlIdentity(t *testing.T) {
	body := []byte(`{"success":true,"msg":"操作成功","data":{"user_id":"U1","username":"zhangsan","nickname":"张三","mobile":"13800000000","email":"a@b.c","access_token":"tok"}}`)
	id, err := parseOwlIdentity(body)
	if err != nil {
		t.Fatal(err)
	}
	if id.Username != "zhangsan" || id.NickName != "张三" || id.Phone != "13800000000" || id.Email != "a@b.c" || id.SourceID != "U1" {
		t.Fatalf("identity = %+v", id)
	}
	if !strings.Contains(id.Raw, `"username":"zhangsan"`) || strings.Contains(id.Raw, "access_token") || strings.Contains(id.Raw, "tok") {
		t.Fatalf("owl raw = %s", id.Raw)
	}

	if _, err = parseOwlIdentity([]byte(`{"success":false,"msg":"授权码无效或已过期"}`)); err == nil {
		t.Fatal("expected rejected token")
	}
	if _, err = parseOwlIdentity([]byte(`{"success":true,"data":{"username":"zhangsan"}}`)); err == nil {
		t.Fatal("expected missing user id")
	}
}
