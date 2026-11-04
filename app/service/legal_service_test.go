package service

import (
	"context"
	"fmt"
	"testing"

	"bit-labs.cn/owl-admin/app/model"
	"bit-labs.cn/owl-admin/app/repository"
	"bit-labs.cn/owl/provider/router"
	"bit-labs.cn/owl/utils"
	"github.com/glebarez/sqlite"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupLegalDB(t *testing.T) *gorm.DB {
	t.Helper()
	if err := utils.InitSnowFlakeWorker(1, 1); err != nil {
		t.Fatalf("init snowflake: %v", err)
	}
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	gdb, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Silent),
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := gdb.AutoMigrate(
		&model.User{},
		&model.LegalDocument{},
		&model.LegalVersion{},
		&model.LegalScene{},
		&model.LegalSceneItem{},
		&model.LegalEvent{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return gdb
}

func newLegalService(t *testing.T, gdb *gorm.DB) *LegalService {
	t.Helper()
	return NewLegalService(repository.NewLegalRepository(gdb), testLockerFactory{}, validator.New())
}

func mustCreateDoc(t *testing.T, svc *LegalService, key, name, kind string) {
	t.Helper()
	if err := svc.CreateDocument(context.Background(), &CreateLegalDocumentReq{Key: key, Name: name, Kind: kind}); err != nil {
		t.Fatalf("create doc %s: %v", key, err)
	}
}

func mustDocID(t *testing.T, gdb *gorm.DB, key string) uint {
	t.Helper()
	var doc model.LegalDocument
	if err := gdb.Where("key = ?", key).First(&doc).Error; err != nil {
		t.Fatalf("doc %s: %v", key, err)
	}
	return doc.ID
}

func mustDraft(t *testing.T, svc *LegalService, docID uint, version, title, content string) uint {
	t.Helper()
	if err := svc.CreateVersion(context.Background(), &CreateLegalVersionReq{
		DocumentID: docID,
		Version:    version,
		Title:      title,
		Content:    content,
	}); err != nil {
		t.Fatalf("draft %s: %v", version, err)
	}
	found, err := svc.repo.WithContext(context.Background()).FindVersion(docID, version)
	if err != nil || found == nil {
		t.Fatalf("find draft %s: %v", version, err)
	}
	return found.ID
}

func TestLegalPublishFreezeAndReaccept(t *testing.T) {
	gdb := setupLegalDB(t)
	svc := newLegalService(t, gdb)
	ctx := context.Background()
	mustCreateDoc(t, svc, "terms", "使用条款", model.LegalKindContract)
	docID := mustDocID(t, gdb, "terms")
	v1 := mustDraft(t, svc, docID, "1.0.0", "使用条款", "第一版正文")
	if err := svc.PublishVersion(ctx, &PublishLegalVersionReq{ID: v1, ChangeClass: model.LegalChangeMaterial}); err != nil {
		t.Fatalf("publish v1: %v", err)
	}
	if err := svc.UpdateVersion(ctx, &UpdateLegalVersionReq{ID: v1, Version: "1.0.0", Title: "改", Content: "改过"}); err == nil {
		t.Fatal("published version was edited")
	}

	body, err := svc.PublicDocument(ctx, "terms")
	if err != nil || body.Content != "第一版正文" || body.Version != "1.0.0" || body.ContentSHA256 == "" {
		t.Fatalf("public body = %+v, %v", body, err)
	}

	actor := LegalActor{UserID: 7, Username: "ada", Nickname: "艾达"}
	if _, err := svc.RecordAcceptances(ctx, actor, &RecordAcceptanceReq{
		SceneKey: "register",
		Items:    []RecordAcceptanceItem{{DocumentKey: "terms", Version: "1.0.0", Action: model.LegalActionAccept}},
	}); err == nil {
		t.Fatal("accept without scene succeeded")
	}
	if err := svc.CreateScene(ctx, &CreateLegalSceneReq{
		Key:  "register",
		Name: "注册",
		Kind: model.LegalSceneConsent,
		Items: []LegalSceneItemReq{{
			DocumentKey: "terms", Required: true, GroupKey: "account", Wording: "我已阅读并同意", Sort: 1,
		}},
	}); err != nil {
		t.Fatalf("scene: %v", err)
	}
	first, err := svc.RecordAcceptances(ctx, actor, &RecordAcceptanceReq{
		SceneKey: "register",
		Items:    []RecordAcceptanceItem{{DocumentKey: "terms", Version: "1.0.0", Action: model.LegalActionAccept}},
	})
	if err != nil || len(first) != 1 {
		t.Fatalf("accept: %v %+v", err, first)
	}
	again, err := svc.RecordAcceptances(ctx, actor, &RecordAcceptanceReq{
		SceneKey: "register",
		Items:    []RecordAcceptanceItem{{DocumentKey: "terms", Version: "1.0.0", Action: model.LegalActionAccept}},
	})
	if err != nil || len(again) != 1 {
		t.Fatalf("idempotent accept: %v %+v", err, again)
	}
	var n int64
	if err := gdb.Model(&model.LegalEvent{}).Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("events = %d", n)
	}
	if _, err := svc.RecordAcceptances(ctx, actor, &RecordAcceptanceReq{
		SceneKey: "register",
		Items:    []RecordAcceptanceItem{{DocumentKey: "terms", Version: "1.0.0", Action: model.LegalActionWithdraw}},
	}); err == nil {
		t.Fatal("contract withdraw succeeded")
	}

	status, err := svc.Status(ctx, actor.UserID)
	if err != nil || len(status.Pending) != 0 {
		t.Fatalf("status after accept: %+v %v", status, err)
	}

	v2 := mustDraft(t, svc, docID, "1.1.0", "使用条款", "错别字")
	if err := svc.PublishVersion(ctx, &PublishLegalVersionReq{ID: v2, ChangeClass: model.LegalChangeEditorial}); err != nil {
		t.Fatalf("publish editorial: %v", err)
	}
	var published int64
	if err := gdb.Model(&model.LegalVersion{}).Where("status = ?", model.LegalVersionPublished).Count(&published).Error; err != nil {
		t.Fatal(err)
	}
	if published != 1 {
		t.Fatalf("published = %d", published)
	}
	var archived int64
	if err := gdb.Model(&model.LegalVersion{}).Where("status = ? AND version = ?", model.LegalVersionArchived, "1.0.0").Count(&archived).Error; err != nil {
		t.Fatal(err)
	}
	if archived != 1 {
		t.Fatalf("archived v1 = %d", archived)
	}
	status, err = svc.Status(ctx, actor.UserID)
	if err != nil || len(status.Pending) != 0 {
		t.Fatalf("editorial should not block: %+v %v", status, err)
	}
	old, err := svc.PublicDocumentVersion(ctx, "terms", "1.0.0")
	if err != nil || old.Content != "第一版正文" {
		t.Fatalf("history: %+v %v", old, err)
	}

	v3 := mustDraft(t, svc, docID, "2.0.0", "使用条款", "实质修改")
	if err := svc.PublishVersion(ctx, &PublishLegalVersionReq{ID: v3, ChangeClass: model.LegalChangeMaterial}); err != nil {
		t.Fatalf("publish material: %v", err)
	}
	status, err = svc.Status(ctx, actor.UserID)
	if err != nil || len(status.Pending) != 1 || status.Pending[0].Version != "2.0.0" {
		t.Fatalf("material pending: %+v %v", status, err)
	}
	if err := svc.SceneRequirementsMet(ctx, "register", []LegalVersionRef{{DocumentKey: "terms", Version: "1.0.0"}}); err == nil {
		t.Fatal("old version satisfied current requirement")
	}
	if err := svc.SceneRequirementsMet(ctx, "register", []LegalVersionRef{{DocumentKey: "terms", Version: "2.0.0"}}); err != nil {
		t.Fatalf("current version: %v", err)
	}
	if err := svc.DeleteDocument(ctx, docID); err == nil {
		t.Fatal("deleted document that has events")
	}
}

func TestLegalReaderAndNoticeWithdraw(t *testing.T) {
	gdb := setupLegalDB(t)
	svc := newLegalService(t, gdb)
	ctx := context.Background()
	mustCreateDoc(t, svc, "terms", "使用条款", model.LegalKindContract)
	mustCreateDoc(t, svc, "privacy", "隐私政策", model.LegalKindNotice)
	mustCreateDoc(t, svc, "inactive-account", "非活跃账号处理规范", model.LegalKindContract)
	for _, key := range []string{"terms", "privacy", "inactive-account"} {
		id := mustDraft(t, svc, mustDocID(t, gdb, key), "1.0.0", key, key+" 正文")
		if err := svc.PublishVersion(ctx, &PublishLegalVersionReq{ID: id, ChangeClass: model.LegalChangeMaterial}); err != nil {
			t.Fatalf("publish %s: %v", key, err)
		}
	}
	if err := svc.CreateScene(ctx, &CreateLegalSceneReq{
		Key:  "site",
		Name: "网站条款",
		Kind: model.LegalSceneReader,
		Items: []LegalSceneItemReq{
			{DocumentKey: "terms", Sort: 1},
			{DocumentKey: "privacy", Sort: 2},
		},
	}); err != nil {
		t.Fatalf("reader: %v", err)
	}
	view, err := svc.PublicScene(ctx, "site")
	if err != nil {
		t.Fatal(err)
	}
	if len(view.Documents) != 2 || view.Documents[0].Key != "terms" || view.Documents[1].Key != "privacy" {
		t.Fatalf("reader docs = %+v", view.Documents)
	}
	if len(view.Groups) != 0 {
		t.Fatalf("reader groups = %+v", view.Groups)
	}
	if err := svc.SceneRequirementsMet(ctx, "site", nil); err != nil {
		t.Fatalf("reader requirements: %v", err)
	}

	if err := svc.CreateScene(ctx, &CreateLegalSceneReq{
		Key:  "register",
		Name: "注册",
		Kind: model.LegalSceneConsent,
		Items: []LegalSceneItemReq{
			{DocumentKey: "terms", Required: true, GroupKey: "account", Wording: "我已阅读并同意", Sort: 1},
			{DocumentKey: "inactive-account", Required: true, GroupKey: "account", Wording: "我已阅读并同意", Sort: 2},
			{DocumentKey: "privacy", Required: true, GroupKey: "privacy", Wording: "我已阅读", Sort: 3},
		},
	}); err != nil {
		t.Fatalf("consent: %v", err)
	}
	consent, err := svc.PublicScene(ctx, "register")
	if err != nil || len(consent.Groups) != 2 || consent.Groups[0].Documents[0].Key != "terms" {
		t.Fatalf("consent = %+v %v", consent, err)
	}
	if consent.Groups[0].Documents[1].Key != "inactive-account" || consent.Groups[1].Documents[0].Key != "privacy" {
		t.Fatalf("groups = %+v", consent.Groups)
	}

	actor := LegalActor{UserID: 9, Username: "bao", Nickname: "宝"}
	if _, err := svc.RecordAcceptances(ctx, actor, &RecordAcceptanceReq{
		SceneKey: "register",
		Items: []RecordAcceptanceItem{
			{DocumentKey: "privacy", Version: "1.0.0", Action: model.LegalActionAccept},
		},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.RecordAcceptances(ctx, actor, &RecordAcceptanceReq{
		SceneKey: "register",
		Items: []RecordAcceptanceItem{
			{DocumentKey: "privacy", Version: "1.0.0", Action: model.LegalActionWithdraw},
		},
	}); err != nil {
		t.Fatalf("withdraw notice: %v", err)
	}
	status, err := svc.Status(ctx, actor.UserID)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, item := range status.Pending {
		seen[item.Key] = true
	}
	if !seen["privacy"] || !seen["terms"] || !seen["inactive-account"] {
		t.Fatalf("pending = %+v", status.Pending)
	}

	user := model.User{Username: "bao", Nickname: "宝", Status: model.UserStatusEnabled}
	user.ID = 9
	if err := gdb.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	_, events, err := svc.RetrieveEvents(ctx, &RetrieveLegalEventReq{PageReq: router.PageReq{Page: 1, PageSize: 10}, DocumentKey: "privacy"})
	if err != nil || len(events) != 2 {
		t.Fatalf("events = %+v %v", events, err)
	}
	if events[0].UserLabel != "宝" || !events[0].UserPresent || events[0].DocumentName != "隐私政策" || events[0].ID == "" || events[0].ID == "0" {
		t.Fatalf("event row = %+v", events[0])
	}
}
