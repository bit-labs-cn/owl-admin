package v1

import (
	"crypto/tls"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func newPublicURLContext(method, target string, tlsOn bool, headers map[string]string) *gin.Context {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(method, target, nil)
	if tlsOn {
		req.TLS = &tls.ConnectionState{}
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	ctx.Request = req
	return ctx
}

func TestAbsolutePublicURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		target  string
		tlsOn   bool
		headers map[string]string
		in      string
		want    string
	}{
		{
			name:   "relative with leading slash",
			target: "http://11.14.2.11:8080/api/v1/app-versions/upload",
			in:     "/storage/app-versions/2026/09/12/xxx.apk",
			want:   "http://11.14.2.11:8080/storage/app-versions/2026/09/12/xxx.apk",
		},
		{
			name:   "relative without leading slash",
			target: "http://11.14.2.11:8080/api/v1/app/upgrade",
			in:     "storage/app-versions/xxx.apk",
			want:   "http://11.14.2.11:8080/storage/app-versions/xxx.apk",
		},
		{
			name:   "already absolute unchanged",
			target: "http://11.14.2.11:8080/api/v1/app/upgrade",
			in:     "https://cdn.example.com/app.apk",
			want:   "https://cdn.example.com/app.apk",
		},
		{
			name:   "empty unchanged",
			target: "http://11.14.2.11:8080/api/v1/app/upgrade",
			in:     "  ",
			want:   "",
		},
		{
			name:   "https from tls",
			target: "https://admin.example.com/api/v1/app-versions/upload",
			tlsOn:  true,
			in:     "/storage/a.apk",
			want:   "https://admin.example.com/storage/a.apk",
		},
		{
			name:   "forwarded proto and host",
			target: "http://127.0.0.1:8080/api/v1/app-versions/upload",
			headers: map[string]string{
				"X-Forwarded-Proto": "https, http",
				"X-Forwarded-Host":  "dl.example.com, 127.0.0.1:8080",
			},
			in:   "/storage/a.apk",
			want: "https://dl.example.com/storage/a.apk",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := newPublicURLContext("GET", tt.target, tt.tlsOn, tt.headers)
			if got := absolutePublicURL(ctx, tt.in); got != tt.want {
				t.Fatalf("absolutePublicURL() = %q, want %q", got, tt.want)
			}
		})
	}
}
