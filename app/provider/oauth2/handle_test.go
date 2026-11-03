package oauth2

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCallbackAccessDeniedClosesWindow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &Handle{}
	r := gin.New()
	r.GET("/oauth2/:provider/callback", h.Callback)

	req := httptest.NewRequest(http.MethodGet, "/oauth2/owl/callback?error=access_denied&error_description=%E7%94%A8%E6%88%B7%E5%8F%96%E6%B6%88%E4%BA%86%E6%8E%88%E6%9D%83&state=abc", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Fatalf("content-type = %s", ct)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "window.close()") {
		t.Fatalf("body = %s", body)
	}
	if strings.Contains(body, "Token exchange") || strings.Contains(body, "access_denied") || strings.Contains(body, "用户取消了授权") {
		t.Fatalf("body = %s", body)
	}
}
