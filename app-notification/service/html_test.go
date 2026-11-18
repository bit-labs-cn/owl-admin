package service

import "testing"

func TestSanitizeHTMLDropsScript(t *testing.T) {
	got := SanitizeHTML(`<p>待巡检</p><script>alert(1)</script><a href="javascript:alert(1)">x</a>`)
	if got == "" || contains(got, "script") || contains(got, "javascript:") {
		t.Fatalf("消毒结果不符合预期: %s", got)
	}
}

func TestHTMLToText(t *testing.T) {
	got := HTMLToText("<p>去 <strong>3 号机房</strong></p>")
	if got != "去 3 号机房" {
		t.Fatalf("got %q", got)
	}
}

func TestContentHTMLKeepsRichText(t *testing.T) {
	got := ContentHTML(`<p>停电通知</p><ul><li>3 号机房</li></ul><img src="https://example.com/a.png">`)
	if !contains(got, "<ul>") || !contains(got, `src="https://example.com/a.png"`) {
		t.Fatalf("富文本被破坏: %s", got)
	}
	if !contains(got, "<p>停电通知</p>") {
		t.Fatalf("段落丢失: %s", got)
	}
}

func TestContentHTMLSanitizesScript(t *testing.T) {
	got := ContentHTML(`<p>正文</p><script>alert(1)</script><p onclick="x()">点我</p>`)
	if contains(got, "script") || contains(got, "onclick") {
		t.Fatalf("危险标签没被清掉: %s", got)
	}
}

func TestContentHTMLEscapesPlainText(t *testing.T) {
	got := ContentHTML("第一行\n第二行 3 < 5")
	if contains(got, "<p>") {
		t.Fatalf("纯文本不该被当成富文本: %s", got)
	}
	if !contains(got, "&lt; 5") {
		t.Fatalf("尖括号应转义: %s", got)
	}
	if !contains(got, "<br>") {
		t.Fatalf("换行应转成 br: %s", got)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || (len(s) > 0 && (indexOf(s, sub) >= 0)))
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
