package oauth2

import (
	"context"
	"net/http"
	"net/http/httptest"
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

func TestOwlFactory(t *testing.T) {
	p := NewCatalog(nil).get("owl")
	op, ok := p.(*oauth2Provider)
	if !ok || op.Name() != "owl" {
		t.Fatalf("owl provider = %T %+v", p, p)
	}
	if strings.Join(op.cfg.Scopes, " ") != "openid profile email phone" {
		t.Fatalf("scopes = %v", op.cfg.Scopes)
	}
	if op.fields.id[0] != "user_id" || op.fields.phone[0] != "mobile" {
		t.Fatalf("fields = %+v", op.fields)
	}
}

func TestOwlAuthCodeURL(t *testing.T) {
	authURL := "http://sso.local/api/v1/oauth/authorize"
	p := newOAuth2Provider(
		"owl",
		xoauth2.Endpoint{
			AuthURL:   authURL,
			TokenURL:  "http://sso.local/api/v1/oauth/token",
			AuthStyle: xoauth2.AuthStyleInParams,
		},
		[]string{"openid", "profile", "email", "phone"},
		"http://sso.local/api/v1/oauth/userinfo",
		identityFields{},
	)
	p.enable = true
	p.cfg.ClientID = "owl-admin"
	p.cfg.ClientSecret = "secret"
	p.cfg.RedirectURL = "http://admin.local/oauth2/owl/callback"
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
	if q.Get("client_id") != "owl-admin" || q.Get("redirect_uri") != p.cfg.RedirectURL || q.Get("response_type") != "code" || q.Get("state") != "room 1" || q.Get("scope") != "openid profile email phone" {
		t.Fatalf("query = %s", u.RawQuery)
	}
}

func TestOwlUserInfo(t *testing.T) {
	id := pickIdentity(map[string]any{
		"user_id":      "U1",
		"username":     "zhangsan",
		"nickname":     "张三",
		"mobile":       "13800000000",
		"email":        "a@b.c",
		"client_id":    "owl-admin",
		"access_token": "tok",
	}, identityFields{
		username: []string{"username"},
		nickName: []string{"nickname"},
		email:    []string{"email"},
		phone:    []string{"mobile"},
		id:       []string{"user_id"},
	})
	if id.Username != "zhangsan" || id.NickName != "张三" || id.Phone != "13800000000" || id.Email != "a@b.c" || id.SourceID != "U1" {
		t.Fatalf("identity = %+v", id)
	}
	if !strings.Contains(id.Raw, `"username":"zhangsan"`) || strings.Contains(id.Raw, "access_token") || strings.Contains(id.Raw, "tok") {
		t.Fatalf("owl raw = %s", id.Raw)
	}
}

func TestOwlIdentityExchange(t *testing.T) {
	var userinfoHits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			if err := r.ParseForm(); err != nil {
				t.Error(err)
				http.Error(w, "bad form", http.StatusBadRequest)
				return
			}
			if r.Form.Get("grant_type") != "authorization_code" || r.Form.Get("code") != "once" || r.Form.Get("client_id") != "owl-admin" {
				t.Errorf("form = %v", r.Form)
			}
			w.Header().Set("Content-Type", "application/json")
			if _, err := w.Write([]byte(`{"access_token":"sso-access","token_type":"Bearer","expires_in":7200}`)); err != nil {
				t.Error(err)
			}
		case "/userinfo":
			userinfoHits++
			if r.Header.Get("Authorization") != "Bearer sso-access" {
				t.Errorf("authorization = %s", r.Header.Get("Authorization"))
			}
			w.Header().Set("Content-Type", "application/json")
			if _, err := w.Write([]byte(`{"user_id":"U1","username":"zhangsan","nickname":"张三","mobile":"13800000000","email":"a@b.c","client_id":"owl-admin"}`)); err != nil {
				t.Error(err)
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	p := newOAuth2Provider(
		"owl",
		xoauth2.Endpoint{
			AuthURL:   srv.URL + "/authorize",
			TokenURL:  srv.URL + "/token",
			AuthStyle: xoauth2.AuthStyleInParams,
		},
		[]string{"openid", "profile", "email", "phone"},
		srv.URL+"/userinfo",
		identityFields{
			username: []string{"username"},
			nickName: []string{"nickname"},
			email:    []string{"email"},
			phone:    []string{"mobile"},
			id:       []string{"user_id"},
		},
	)
	p.cfg.ClientID = "owl-admin"
	p.cfg.ClientSecret = "secret"
	p.cfg.RedirectURL = "http://admin.local/oauth2/owl/callback"

	id, err := p.Identity(context.Background(), "once")
	if err != nil {
		t.Fatal(err)
	}
	if userinfoHits != 1 || id.Username != "zhangsan" || id.NickName != "张三" || id.SourceID != "U1" || id.Email != "a@b.c" || id.Phone != "13800000000" {
		t.Fatalf("hits=%d identity=%+v", userinfoHits, id)
	}
}

func TestUserInfoFromTokenURL(t *testing.T) {
	got := userInfoFromTokenURL("http://127.0.0.1:8080/api/v1/oauth/token")
	if got != "http://127.0.0.1:8080/api/v1/oauth/userinfo" {
		t.Fatalf("userinfo = %s", got)
	}
	if userInfoFromTokenURL("https://api.github.com/login/oauth/access_token") != "" {
		t.Fatal("non-token suffix should stay empty")
	}
}

func TestScopesFrom(t *testing.T) {
	got := scopesFrom("openid, profile email")
	if strings.Join(got, " ") != "openid profile email" {
		t.Fatalf("scopes = %v", got)
	}
	if scopesFrom(nil) != nil {
		t.Fatal("empty scopes should stay nil")
	}
}
