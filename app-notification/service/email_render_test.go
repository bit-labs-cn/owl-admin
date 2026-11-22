package service

import (
	"strings"
	"testing"
)

func TestContentHTMLKeepsInlineStyle(t *testing.T) {
	got := ContentHTML(`<p style="color:#dc2626;font-size:16px;">库存不足</p>`)
	if !strings.Contains(got, "color: #dc2626") {
		t.Fatalf("颜色被清洗掉了: %s", got)
	}
	if !strings.Contains(got, "font-size: 16px") {
		t.Fatalf("字号被清洗掉了: %s", got)
	}
}

func TestContentHTMLDropsPositionStyle(t *testing.T) {
	got := ContentHTML(`<div style="position:fixed;top:0;color:#111111;">浮层</div>`)
	if strings.Contains(got, "position") {
		t.Fatalf("定位属性应被拒绝: %s", got)
	}
	if !strings.Contains(got, "color: #111111") {
		t.Fatalf("同一 style 里的合法属性不该被一起丢掉: %s", got)
	}
}

func TestContentHTMLKeepsRelativeLink(t *testing.T) {
	got := ContentHTML(`<p>看 <a href="/admin/bug/detail?id=1">缺陷详情</a></p>`)
	if !strings.Contains(got, `href="/admin/bug/detail?id=1"`) {
		t.Fatalf("站内相对链接被删了: %s", got)
	}
}

func TestContentHTMLKeepsTableLayoutAttributes(t *testing.T) {
	got := ContentHTML(`<table border="1" cellpadding="8"><tr><td colspan="2" bgcolor="#fef3c7">合计</td></tr></table>`)
	if !strings.Contains(got, "colspan=\"2\"") || !strings.Contains(got, `bgcolor="#fef3c7"`) {
		t.Fatalf("邮件表格属性丢失: %s", got)
	}
}

func TestAbsolutizeHTML(t *testing.T) {
	content := `<p><a href="/admin/run/1">详情</a><img src="/storage/a.png"><a href="https://x.test/b">外链</a></p>`
	got := AbsolutizeHTML("https://admin.example.com/", content)
	if !strings.Contains(got, `href="https://admin.example.com/admin/run/1"`) {
		t.Fatalf("站内链接没补绝对地址: %s", got)
	}
	if !strings.Contains(got, `src="https://admin.example.com/storage/a.png"`) {
		t.Fatalf("站内图片没补绝对地址: %s", got)
	}
	if !strings.Contains(got, `href="https://x.test/b"`) {
		t.Fatalf("外链不该被改写: %s", got)
	}
}

func TestAbsolutizeHTMLWithoutOrigin(t *testing.T) {
	content := `<a href="/admin/run/1">详情</a>`
	if got := AbsolutizeHTML("", content); got != content {
		t.Fatalf("没有站点地址时应原样返回, got %q", got)
	}
}

func TestAbsoluteURL(t *testing.T) {
	tests := []struct {
		name   string
		origin string
		raw    string
		want   string
	}{
		{name: "站内路径", origin: "https://admin.example.com", raw: "/admin/run/1", want: "https://admin.example.com/admin/run/1"},
		{name: "末尾带斜杠", origin: "https://admin.example.com/", raw: "/admin/run/1", want: "https://admin.example.com/admin/run/1"},
		{name: "已是绝对地址", origin: "https://admin.example.com", raw: "https://x.test/a", want: "https://x.test/a"},
		{name: "协议相对地址", origin: "https://admin.example.com", raw: "//cdn.test/a.png", want: "//cdn.test/a.png"},
		{name: "空地址", origin: "https://admin.example.com", raw: "", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AbsoluteURL(tt.origin, tt.raw); got != tt.want {
				t.Errorf("AbsoluteURL(%q, %q) = %q; want %q", tt.origin, tt.raw, got, tt.want)
			}
		})
	}
}

func TestEmailDocument(t *testing.T) {
	got := EmailDocument("构建失败：owl-sqm", `<p style="color:#dc2626;">编译错误</p>`, "https://admin.example.com/#/forge/run/detail?id=7")
	for _, want := range []string{
		"<!DOCTYPE html>",
		"<title>构建失败：owl-sqm</title>",
		`color:#dc2626`,
		`href="https://admin.example.com/#/forge/run/detail?id=7"`,
		"查看详情",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("邮件文档缺少 %q", want)
		}
	}
}

func TestEmailDocumentEscapesTitle(t *testing.T) {
	got := EmailDocument(`<b>标题</b>`, "", "")
	if strings.Contains(got, "<b>标题</b>") {
		t.Fatalf("标题应被转义: %s", got)
	}
	if strings.Contains(got, "查看详情") {
		t.Fatalf("没有链接时不该有按钮")
	}
}
