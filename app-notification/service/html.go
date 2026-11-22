package service

import (
	"html"
	"regexp"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

var htmlPolicy = newHTMLPolicy()

// allowedStyleProperties 是通知正文里允许保留的 CSS 属性。
// 只放排版与配色相关的属性：position、animation 这类会影响宿主页面布局的属性一律不放开。
var allowedStyleProperties = []string{
	"color", "background-color",
	"font-family", "font-size", "font-style", "font-weight",
	"line-height", "letter-spacing", "text-align", "text-decoration", "text-indent",
	"vertical-align", "white-space",
	"margin", "margin-top", "margin-right", "margin-bottom", "margin-left",
	"padding", "padding-top", "padding-right", "padding-bottom", "padding-left",
	"border", "border-width", "border-style", "border-color", "border-radius",
	"width", "height", "max-width", "min-width",
	"display", "list-style-type", "border-collapse", "border-spacing", "table-layout",
}

// safeHrefRe 限制链接协议：只允许 http(s)、mailto、tel，以及站内相对路径与页内锚点。
// 关掉 RequireParseableURLs 才能保留 /admin/xxx 这类相对地址，因此协议必须在这里自己卡住。
var safeHrefRe = regexp.MustCompile(`(?i)^(?:https?://[^\s]*|mailto:[^\s]+|tel:[^\s]+|//[^\s]+|/[^\s]*|#\S*|\?\S+)$`)

// safeSrcRe 限制图片地址：只允许 http(s)、站内相对路径与协议相对地址。
var safeSrcRe = regexp.MustCompile(`(?i)^(?:https?://[^\s]*|//[^\s]+|/[^\s]*)$`)

// newHTMLPolicy 构造通知正文的白名单：
// 保留富文本的结构、颜色与字号（编辑器写出来的 style 属性），去掉脚本、事件属性与非 HTTP 协议。
func newHTMLPolicy() *bluemonday.Policy {
	p := bluemonday.NewPolicy()
	p.AllowElements(
		"p", "br", "strong", "b", "em", "i", "u", "s", "strike", "del", "sub", "sup",
		"ul", "ol", "li", "blockquote", "code", "pre",
		"h1", "h2", "h3", "h4", "h5", "h6",
		"span", "div", "hr", "small", "center", "font",
		"table", "thead", "tbody", "tfoot", "tr", "th", "td",
		"a", "img",
	)

	// 颜色与字号都写在 style 上，必须放行，否则邮件里只剩黑白正文。
	p.AllowAttrs("style").Globally()
	p.AllowStyles(allowedStyleProperties...).Globally()

	// 表格类属性是邮件客户端的老兼容写法（对齐、列宽、合并单元格、背景色）。
	p.AllowAttrs("align", "valign", "width", "height", "bgcolor", "border", "cellpadding", "cellspacing", "colspan", "rowspan").
		OnElements("table", "thead", "tbody", "tfoot", "tr", "th", "td")

	p.AllowAttrs("href").Matching(safeHrefRe).OnElements("a")
	p.AllowAttrs("title").OnElements("a")
	p.AllowAttrs("target").Matching(regexp.MustCompile(`^_(blank|self|parent|top)$`)).OnElements("a")
	p.AllowAttrs("rel").Matching(regexp.MustCompile(`^[a-zA-Z\s]+$`)).OnElements("a")

	p.AllowAttrs("src").Matching(safeSrcRe).OnElements("img")
	p.AllowAttrs("alt", "width", "height", "title").OnElements("img")
	p.AllowAttrs("color", "face", "size").OnElements("font")

	p.AllowURLSchemes("http", "https", "mailto", "tel")
	// 站内相对地址（/admin/xxx）要保留：投递前会按站点地址补成绝对地址，
	// 这里要求绝对地址会把编辑器里的站内链接直接删掉。
	p.RequireParseableURLs(false)
	return p
}

// SanitizeHTML 按白名单清洗通知正文。
func SanitizeHTML(raw string) string {
	return strings.TrimSpace(htmlPolicy.Sanitize(raw))
}

// richTagRe 用来判断一段正文是不是富文本：只认白名单里会出现的标签，
// 避免「3 < 5」这类纯文本被误当成 HTML 送进清洗器。
var richTagRe = regexp.MustCompile(`(?i)</?(p|br|div|span|strong|b|em|i|u|s|ul|ol|li|h[1-6]|blockquote|pre|code|table|thead|tbody|tfoot|tr|th|td|a|img|hr|font|center|small)\b`)

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

// siteURLRe 匹配 href 与 src 里的站内绝对路径（/foo），不含 //host 这种协议相对地址。
var siteURLRe = regexp.MustCompile(`(?i)(\s(?:href|src)=["'])(/[^/"'][^"']*)`)

// AbsolutizeHTML 把正文里的站内链接与站内图片补成绝对地址。
// 邮件客户端没有当前站点上下文，相对地址会点不开、图片也会裂。
func AbsolutizeHTML(origin, content string) string {
	origin = strings.TrimRight(strings.TrimSpace(origin), "/")
	if origin == "" || content == "" {
		return content
	}
	return siteURLRe.ReplaceAllString(content, "${1}"+origin+"${2}")
}

// AbsoluteURL 把站内相对地址补成绝对地址，已经是绝对地址或锚点时原样返回。
func AbsoluteURL(origin, raw string) string {
	raw = strings.TrimSpace(raw)
	origin = strings.TrimRight(strings.TrimSpace(origin), "/")
	if raw == "" || origin == "" {
		return raw
	}
	if strings.HasPrefix(raw, "/") && !strings.HasPrefix(raw, "//") {
		return origin + raw
	}
	return raw
}
