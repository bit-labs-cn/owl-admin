package service

import (
	"html"
	"regexp"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

var htmlPolicy = newHTMLPolicy()

func newHTMLPolicy() *bluemonday.Policy {
	p := bluemonday.NewPolicy()
	p.AllowElements("p", "br", "strong", "em", "u", "s", "ul", "ol", "li", "blockquote", "code", "pre", "h1", "h2", "h3", "h4", "span", "div", "hr", "table", "thead", "tbody", "tr", "th", "td")
	p.AllowAttrs("href").OnElements("a")
	p.AllowAttrs("src", "alt").OnElements("img")
	p.AllowURLSchemes("http", "https", "mailto")
	p.RequireParseableURLs(true)
	return p
}

// SanitizeHTML 按白名单清洗通知正文。
func SanitizeHTML(raw string) string {
	return strings.TrimSpace(htmlPolicy.Sanitize(raw))
}

// richTagRe 用来判断一段正文是不是富文本：只认白名单里会出现的标签，
// 避免「3 < 5」这类纯文本被误当成 HTML 送进清洗器。
var richTagRe = regexp.MustCompile(`(?i)</?(p|br|div|span|strong|b|em|i|u|s|ul|ol|li|h[1-6]|blockquote|pre|code|table|thead|tbody|tr|th|td|a|img|hr)\b`)

// IsHTML 判断正文是否已经是富文本。
func IsHTML(raw string) bool {
	return richTagRe.MatchString(raw)
}

// ContentHTML 把正文归一成可安全入库、可安全投递的 HTML。
// 富文本按白名单清洗；纯文本按 HTML 转义并把换行转成 <br>，保持原有段落。
func ContentHTML(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if !IsHTML(raw) {
		return strings.ReplaceAll(html.EscapeString(raw), "\n", "<br>")
	}
	return SanitizeHTML(raw)
}

var tagRe = regexp.MustCompile(`<[^>]+>`)
var spaceRe = regexp.MustCompile(`\s+`)

// HTMLToText 从 HTML 得到短信和推送用的纯文本。
func HTMLToText(raw string) string {
	s := tagRe.ReplaceAllString(raw, " ")
	s = html.UnescapeString(s)
	return strings.TrimSpace(spaceRe.ReplaceAllString(s, " "))
}

var imgSrcRe = regexp.MustCompile(`(?i)(<img\b[^>]*\bsrc=["'])(/[^"']*)`)

// AbsolutizeHTML 把正文里的站内图片地址补成绝对地址。
func AbsolutizeHTML(origin, content string) string {
	origin = strings.TrimRight(strings.TrimSpace(origin), "/")
	if origin == "" || content == "" {
		return content
	}
	return imgSrcRe.ReplaceAllString(content, "${1}"+origin+"${2}")
}
