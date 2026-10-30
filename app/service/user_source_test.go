package service

import (
	"context"
	"testing"

	"bit-labs.cn/owl-admin/app/model"
	"bit-labs.cn/owl-admin/app/repository"
)

func TestListUserSourceOptionsOrder(t *testing.T) {
	opts := listUserSourceOptions([]string{
		"",
		model.UserSourceEmailRegister,
		model.UserSourcePhoneRegister,
		model.UserSourceGitHub,
		model.UserSourceGitHub,
		model.UserSourceGitee,
		"ldap",
	})
	want := []string{"后台添加", "手机号注册", "邮箱注册", "Gitee", "GitHub", "ldap"}
	if len(opts) != len(want) {
		t.Fatalf("options = %+v", opts)
	}
	for i := range want {
		if opts[i].Label != want[i] {
			t.Fatalf("options = %+v", opts)
		}
	}
	if opts[0].Value != model.UserSourceQueryLocal {
		t.Fatalf("admin value = %s", opts[0].Value)
	}
}

func TestListUserSourceOptionsAndFilter(t *testing.T) {
	gdb := setupUserImportDB(t)
	repo := repository.NewUserRepository(gdb)
	users := []model.User{
		{Username: "local-user", Source: model.UserSourceLocal, Status: model.UserStatusEnabled},
		{Username: "mail-user", Source: model.UserSourceEmailRegister, Status: model.UserStatusEnabled},
		{Username: "gh-user", Source: model.UserSourceGitHub, Status: model.UserStatusEnabled},
	}
	for i := range users {
		if err := gdb.Create(&users[i]).Error; err != nil {
			t.Fatalf("create user: %v", err)
		}
	}

	svc := &UserService{userRepo: repo}
	opts, err := svc.ListUserSourceOptions(context.Background())
	if err != nil {
		t.Fatalf("list sources: %v", err)
	}
	if len(opts) != 3 || opts[0].Label != "后台添加" || opts[1].Label != "邮箱注册" || opts[2].Label != "GitHub" {
		t.Fatalf("options = %+v", opts)
	}

	assertUsernames := func(query string, want ...string) {
		t.Helper()
		var names []string
		tx := gdb.Model(&model.User{})
		applyUserSourceFilter(tx, query)
		if err := tx.Order("username").Pluck("username", &names).Error; err != nil {
			t.Fatalf("query %q: %v", query, err)
		}
		if len(names) != len(want) {
			t.Fatalf("query %q names = %v, want %v", query, names, want)
		}
		for i := range want {
			if names[i] != want[i] {
				t.Fatalf("query %q names = %v, want %v", query, names, want)
			}
		}
	}
	assertUsernames("", "gh-user", "local-user", "mail-user")
	assertUsernames(model.UserSourceQueryLocal, "local-user")
	assertUsernames(model.UserSourceEmailRegister, "mail-user")
	assertUsernames(model.UserSourceGitHub, "gh-user")
}
