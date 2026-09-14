package v1

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func firstForwarded(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if i := strings.IndexByte(value, ','); i >= 0 {
		value = strings.TrimSpace(value[:i])
	}
	return value
}

func requestPublicOrigin(ctx *gin.Context) string {
	scheme := "http"
	if ctx.Request.TLS != nil {
		scheme = "https"
	}
	if proto := firstForwarded(ctx.GetHeader("X-Forwarded-Proto")); proto != "" {
		scheme = proto
	}
	host := ctx.Request.Host
	if forwarded := firstForwarded(ctx.GetHeader("X-Forwarded-Host")); forwarded != "" {
		host = forwarded
	}
	return scheme + "://" + host
}

// absolutePublicURL 将存储相对路径拼成客户端可下载的绝对地址。
// 已含 :// 的对象存储 / CDN 地址原样返回。
func absolutePublicURL(ctx *gin.Context, fileURL string) string {
	fileURL = strings.TrimSpace(fileURL)
	if fileURL == "" || strings.Contains(fileURL, "://") {
		return fileURL
	}
	origin := strings.TrimRight(requestPublicOrigin(ctx), "/")
	if strings.HasPrefix(fileURL, "/") {
		return origin + fileURL
	}
	return origin + "/" + fileURL
}
