package service

import (
	"html"
	"strings"
)

// emailFontStack 是邮件正文的字体栈：先系统 UI 字体，再中文黑体，最后兜底 sans-serif。
const emailFontStack = `-apple-system,BlinkMacSystemFont,"Segoe UI","PingFang SC","Hiragino Sans GB","Microsoft YaHei",sans-serif`

// EmailDocument 把通知正文包装成完整的 HTML 邮件文档。
//
// 站内信只需要一段 HTML 片段；邮件客户端还缺文档头、基础排版、内容外框和跳转按钮。
// content 必须是已经清洗过的 HTML，title 与 link 在这里转义。
func EmailDocument(title, content, link string) string {
	var b strings.Builder

	b.WriteString("<!DOCTYPE html>")
	b.WriteString(`<html lang="zh-CN"><head><meta charset="utf-8">`)
	b.WriteString(`<meta name="viewport" content="width=device-width,initial-scale=1">`)
	b.WriteString(`<meta name="color-scheme" content="light">`)
	b.WriteString(`<title>`)
	b.WriteString(html.EscapeString(title))
	b.WriteString(`</title></head>`)

	b.WriteString(`<body style="margin:0;padding:24px 12px;background-color:#f4f5f7;">`)
	b.WriteString(`<table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0" style="border-collapse:collapse;"><tr><td align="center">`)
	b.WriteString(`<table role="presentation" width="640" cellpadding="0" cellspacing="0" border="0" style="width:100%;max-width:640px;border-collapse:separate;background-color:#ffffff;border:1px solid #e5e7eb;border-radius:10px;">`)

	// 标题
	b.WriteString(`<tr><td style="padding:28px 28px 0 28px;">`)
	b.WriteString(`<div style="font-family:` + emailFontStack + `;font-size:18px;font-weight:600;line-height:1.5;color:#111827;">`)
	b.WriteString(html.EscapeString(title))
	b.WriteString(`</div></td></tr>`)

	// 正文
	if strings.TrimSpace(content) != "" {
		b.WriteString(`<tr><td style="padding:12px 28px 4px 28px;font-family:` + emailFontStack + `;font-size:14px;line-height:1.75;color:#374151;">`)
		b.WriteString(content)
		b.WriteString(`</td></tr>`)
	}

	// 跳转按钮
	if strings.TrimSpace(link) != "" {
		b.WriteString(`<tr><td style="padding:20px 28px 24px 28px;">`)
		b.WriteString(`<a href="` + html.EscapeString(link) + `" target="_blank" rel="noopener noreferrer" style="display:inline-block;padding:10px 20px;border-radius:6px;background-color:#2563eb;color:#ffffff;font-family:` + emailFontStack + `;font-size:14px;font-weight:600;text-decoration:none;">查看详情</a>`)
		b.WriteString(`</td></tr>`)
	}

	// 页脚
	b.WriteString(`<tr><td style="padding:16px 28px 22px 28px;border-top:1px solid #f0f1f3;font-family:` + emailFontStack + `;font-size:12px;line-height:1.6;color:#9ca3af;">`)
	b.WriteString(`本邮件由系统自动发送，请勿直接回复。`)
	b.WriteString(`</td></tr>`)

	b.WriteString(`</table></td></tr></table></body></html>`)
	return b.String()
}
