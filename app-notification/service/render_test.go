package service

import "testing"

func TestRenderTextMissingVar(t *testing.T) {
	_, err := renderText("单号 {{.OrderNo}}", map[string]any{})
	if err == nil {
		t.Fatal("expected missing variable error")
	}
}

func TestRenderHTMLEscapes(t *testing.T) {
	out, err := renderHTML("<p>{{.Name}}</p>", map[string]any{"Name": "<b>"})
	if err != nil {
		t.Fatal(err)
	}
	if out != "<p>&lt;b&gt;</p>" {
		t.Fatalf("html = %s", out)
	}
}

func TestApplyTextFallback(t *testing.T) {
	out, err := applyText("", "原文", nil)
	if err != nil {
		t.Fatal(err)
	}
	if out != "原文" {
		t.Fatalf("fallback = %s", out)
	}
}
