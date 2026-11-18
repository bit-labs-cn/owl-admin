package v1

import (
	"net"
	"strings"

	"github.com/gin-gonic/gin"
)

func requestOrigin(ctx *gin.Context) string {
	scheme := "http"
	if ctx.Request.TLS != nil {
		scheme = "https"
	}
	if proto := firstHeader(ctx, "X-Forwarded-Proto"); proto != "" {
		scheme = proto
	}
	host := ctx.Request.Host
	if forwarded := firstHeader(ctx, "X-Forwarded-Host"); forwarded != "" {
		host = forwarded
	}
	return scheme + "://" + host
}

func firstHeader(ctx *gin.Context, key string) string {
	value := strings.TrimSpace(ctx.GetHeader(key))
	if i := strings.IndexByte(value, ','); i >= 0 {
		value = strings.TrimSpace(value[:i])
	}
	if key == "X-Forwarded-Host" {
		if _, _, err := net.SplitHostPort(value); err != nil && value != "" {
			return value
		}
	}
	return value
}
