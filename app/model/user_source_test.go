package model

import "testing"

func TestUserSourceLabel(t *testing.T) {
	cases := []struct {
		source string
		label  string
		query  string
	}{
		{source: "", label: "后台添加", query: UserSourceQueryLocal},
		{source: UserSourceEmailRegister, label: "邮箱注册", query: UserSourceEmailRegister},
		{source: UserSourcePhoneRegister, label: "手机号注册", query: UserSourcePhoneRegister},
		{source: UserSourceGitHub, label: "GitHub", query: UserSourceGitHub},
		{source: UserSourceOwl, label: "统一身份认证", query: UserSourceOwl},
		{source: "ldap", label: "ldap", query: "ldap"},
	}
	for _, tc := range cases {
		if got := UserSourceLabel(tc.source); got != tc.label {
			t.Fatalf("label(%q) = %q, want %q", tc.source, got, tc.label)
		}
		if got := UserSourceQueryValue(tc.source); got != tc.query {
			t.Fatalf("query(%q) = %q, want %q", tc.source, got, tc.query)
		}
	}
}

func TestParseUserSourceQuery(t *testing.T) {
	if _, ok := ParseUserSourceQuery(""); ok {
		t.Fatal("empty query should mean no filter")
	}
	if _, ok := ParseUserSourceQuery("   "); ok {
		t.Fatal("blank query should mean no filter")
	}
	stored, ok := ParseUserSourceQuery(UserSourceQueryLocal)
	if !ok || stored != UserSourceLocal {
		t.Fatalf("local query stored=%q ok=%v", stored, ok)
	}
	stored, ok = ParseUserSourceQuery(" github ")
	if !ok || stored != "github" {
		t.Fatalf("github query stored=%q ok=%v", stored, ok)
	}
}
