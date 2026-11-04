package v1

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestLegalRoutePatterns(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	group := engine.Group("/api/v1")
	group.GET("/legal/scenes/:key", func(ctx *gin.Context) { ctx.Status(http.StatusNoContent) })
	group.GET("/legal/documents/:key/versions/:version", func(ctx *gin.Context) { ctx.Status(http.StatusNoContent) })
	group.GET("/legal/documents/:key", func(ctx *gin.Context) { ctx.Status(http.StatusNoContent) })
	group.GET("/legal/documents", func(ctx *gin.Context) { ctx.Status(http.StatusNoContent) })
	group.GET("/legal/scenes", func(ctx *gin.Context) { ctx.Status(http.StatusNoContent) })
	group.PUT("/legal/documents/:id", func(ctx *gin.Context) { ctx.Status(http.StatusNoContent) })
	group.PUT("/legal/scenes/:id", func(ctx *gin.Context) { ctx.Status(http.StatusNoContent) })
	group.GET("/legal/versions", func(ctx *gin.Context) { ctx.Status(http.StatusNoContent) })
	group.POST("/legal/versions/:id/publish", func(ctx *gin.Context) { ctx.Status(http.StatusNoContent) })

	req := httptest.NewRequest(http.MethodGet, "/api/v1/legal/documents/terms/versions/1.0.0", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("history status = %d", rec.Code)
	}
}
