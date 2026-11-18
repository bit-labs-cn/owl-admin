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
