package v1

import (
	"net"
	"net/url"
	"strconv"
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
	host = ensurePublicHostPort(host, firstForwarded(ctx.GetHeader("X-Forwarded-Port")), scheme)
	return scheme + "://" + host
}

func ensurePublicHostPort(host, forwardedPort, scheme string) string {
	host = strings.TrimSpace(host)
	if host == "" {
		return host
	}
	if _, _, err := net.SplitHostPort(host); err == nil {
		return host
	}
	port := parseForwardedPort(forwardedPort)
	if port == "" || isDefaultPublicPort(scheme, port) {
		return host
	}
	hostname := host
	if strings.HasPrefix(hostname, "[") && strings.HasSuffix(hostname, "]") {
		hostname = hostname[1 : len(hostname)-1]
	}
	return net.JoinHostPort(hostname, port)
}

func parseForwardedPort(value string) string {
	n, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || n < 1 || n > 65535 {
		return ""
	}
	return strconv.Itoa(n)
}

func isDefaultPublicPort(scheme, port string) bool {
	switch strings.ToLower(scheme) {
	case "https":
		return port == "443"
	default:
		return port == "80"
	}
}

func hostNameOnly(host string) string {
	h, _, err := net.SplitHostPort(host)
	if err != nil {
		return strings.Trim(host, "[]")
	}
	return h
}

func hostPortOnly(host string) string {
	_, p, err := net.SplitHostPort(host)
	if err != nil {
		return ""
	}
	return p
}

// restoreMissingPublicPort 已是绝对地址、但 Host 丢了非默认端口时，用当前请求 origin 补上。
func restoreMissingPublicPort(origin, fileURL string) string {
	parsed, err := url.Parse(fileURL)
	if err != nil || parsed.Host == "" {
		return fileURL
	}
	cur, err := url.Parse(origin)
	if err != nil || cur.Host == "" {
		return fileURL
	}
	if hostNameOnly(parsed.Host) != hostNameOnly(cur.Host) {
		return fileURL
	}
	if hostPortOnly(parsed.Host) != "" || hostPortOnly(cur.Host) == "" {
		return fileURL
	}
	parsed.Scheme = cur.Scheme
	parsed.Host = cur.Host
	return parsed.String()
}

// absolutePublicURL 将存储相对路径拼成客户端可下载的绝对地址。
// 已含 :// 的对象存储 / CDN 地址原样返回；同主机缺端口时按当前请求补上。
func absolutePublicURL(ctx *gin.Context, fileURL string) string {
	fileURL = strings.TrimSpace(fileURL)
	if fileURL == "" {
		return fileURL
	}
	origin := strings.TrimRight(requestPublicOrigin(ctx), "/")
	if strings.Contains(fileURL, "://") {
		return restoreMissingPublicPort(origin, fileURL)
	}
	if strings.HasPrefix(fileURL, "/") {
		return origin + fileURL
	}
	return origin + "/" + fileURL
}
