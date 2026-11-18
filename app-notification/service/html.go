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
