package service

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"bit-labs.cn/owl-admin/app/event"
	"bit-labs.cn/owl-admin/app/model"
	"bit-labs.cn/owl-admin/app/repository"
	"bit-labs.cn/owl/provider/db"
	owlredis "bit-labs.cn/owl/provider/redis"
	"bit-labs.cn/owl/utils"
	"github.com/asaskevich/EventBus"
	"github.com/glebarez/sqlite"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/cast"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type testLock struct{}

func (testLock) Lock(string) error { return nil }
func (testLock) Unlock()           {}

type testLockerFactory struct{}

func (testLockerFactory) New() owlredis.Locker { return testLock{} }

var _ owlredis.LockerFactory = testLockerFactory{}

func setupUserImportDB(t *testing.T) *gorm.DB {
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
	if err := gdb.AutoMigrate(&model.User{}, &model.Role{}, &model.Dept{}, &model.UserGroup{}, &model.Menu{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return gdb
}

func newImportUserService(t *testing.T, gdb *gorm.DB, bus EventBus.Bus) *UserService {
	t.Helper()
	if bus == nil {
		bus = EventBus.New()
	}
	return &UserService{
		db:             gdb,
		userRepo:       repository.NewUserRepository(gdb),
		deptRepo:       repository.NewDeptRepository(gdb),
		roleRepo:       repository.NewRoleRepository(gdb),
		eventBus:       bus,
		locker:         testLockerFactory{},
		validate:       validator.New(),
		BaseRepository: db.NewBaseRepository[model.User](gdb),
	}
}

func buildImportWorkbook(t *testing.T, headers []string, rows [][]string) []byte {
	t.Helper()
	f := excelize.NewFile()
	defer f.Close()
	sheet := f.GetSheetName(0)
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
	}
	for r, row := range rows {
		for c, v := range row {
			cell, _ := excelize.CoordinatesToCellName(c+1, r+2)
			_ = f.SetCellValue(sheet, cell, v)
		}
	}
	buf, err := f.WriteToBuffer()
	if err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestParseUserImportRowsAndEnums(t *testing.T) {
	rows := [][]string{
		{"用户账号", "用户昵称", "初始密码", "手机号", "邮箱", "性别", "状态", "归属部门", "角色", "备注"},
		{"u1", "用户一", "Passw0rd", "13800138000", "a@b.com", "女", "停用", "总公司/研发部", "客服,访客", "备注1"},
		{"", "", "", "", "", "", "", "", "", ""},
		{"u2", "用户二", "Passw0rd", "", "", "", "", "", "", ""},
	}
	drafts, err := parseUserImportRows(rows)
	if err != nil {
		t.Fatal(err)
	}
	if len(drafts) != 2 {
		t.Fatalf("want 2 drafts, got %d", len(drafts))
	}
	if drafts[0].Username != "u1" || drafts[0].DeptPath != "总公司/研发部" || drafts[0].RolesRaw != "客服,访客" {
		t.Fatalf("draft0: %+v", drafts[0])
	}
	if drafts[1].Line != 4 || drafts[1].Username != "u2" {
		t.Fatalf("draft1: %+v", drafts[1])
	}

	sex, ok := parseImportSex("女")
	if !ok || sex != 1 {
		t.Fatalf("sex female => %d", sex)
	}
	sex, ok = parseImportSex("")
	if !ok || sex != 0 {
		t.Fatalf("sex default => %d", sex)
	}
	status, ok := parseImportStatus("停用")
	if !ok || status != 0 {
		t.Fatalf("status disable => %d", status)
	}
	status, ok = parseImportStatus("")
	if !ok || status != 1 {
		t.Fatalf("status default => %d", status)
	}
	names := splitImportRoleNames("客服，访客、客服")
	if len(names) != 2 || names[0] != "客服" || names[1] != "访客" {
		t.Fatalf("roles: %v", names)
	}
}

func TestMapUserImportHeaderMissing(t *testing.T) {
	_, err := mapUserImportHeader([]string{"用户昵称", "初始密码"})
	if err == nil || !strings.Contains(err.Error(), "用户账号") {
		t.Fatalf("expected missing username, got %v", err)
	}
}

func TestBuildDeptPathMaps(t *testing.T) {
	root := model.Dept{Name: "总公司", ParentId: 0}
	root.ID = 1
	child := model.Dept{Name: "研发部", ParentId: 1}
	child.ID = 2
	paths, ambig := buildDeptPathMaps([]model.Dept{root, child})
	if len(ambig) != 0 {
		t.Fatalf("ambig: %v", ambig)
	}
	if paths["总公司"].ID != 1 || paths["总公司/研发部"].ID != 2 {
		t.Fatalf("paths: %+v", paths)
	}
}

func TestBuildUserImportTemplate(t *testing.T) {
	gdb := setupUserImportDB(t)
	svc := newImportUserService(t, gdb, nil)
	_ = gdb.Create(&model.Role{Name: "客服", Code: "cs", Status: 1}).Error
	dept := model.Dept{Name: "技术部", Status: 1}
	_ = gdb.Create(&dept).Error

	data, name, ctype, err := svc.BuildUserImportTemplate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if name != UserImportTemplateName || !strings.Contains(ctype, "spreadsheet") || len(data) < 100 {
		t.Fatalf("template meta name=%s ctype=%s size=%d", name, ctype, len(data))
	}
	rows, err := readUserImportRows(data)
	if err != nil {
		t.Fatal(err)
	}
	drafts, err := parseUserImportRows(rows)
	if err != nil {
		t.Fatal(err)
	}
	if len(drafts) < 1 {
		t.Fatalf("expected example row, got %d", len(drafts))
	}
}

func TestImportUsersPartialSuccess(t *testing.T) {
	gdb := setupUserImportDB(t)
	bus := EventBus.New()
	var published []*AssignRoleToUser
	bus.Subscribe(event.AssignRoleToUser, func(req *AssignRoleToUser) {
		published = append(published, req)
	})
	svc := newImportUserService(t, gdb, bus)

	role := model.Role{Name: "客服", Code: "cs", Status: 1}
	if err := gdb.Create(&role).Error; err != nil {
		t.Fatal(err)
	}
	root := model.Dept{Name: "总公司", Status: 1}
	if err := gdb.Create(&root).Error; err != nil {
		t.Fatal(err)
	}
	child := model.Dept{Name: "研发部", ParentId: int(root.ID), Status: 1}
	if err := gdb.Create(&child).Error; err != nil {
		t.Fatal(err)
	}
	exist := model.User{Username: "exist", Nickname: "已存在", Status: 1}
	exist.SetPassword("Passw0rd")
	if err := gdb.Create(&exist).Error; err != nil {
		t.Fatal(err)
	}

	path := "总公司/研发部"
	data := buildImportWorkbook(t, userImportHeaders, [][]string{
		{"ok1", "成功一", "Passw0rd", "13800138001", "ok1@example.com", "女", "启用", path, "客服", "r1"},
		{"exist", "重复账号", "Passw0rd", "", "", "男", "启用", "", "", ""},
		{"ok2", "成功二", "Passw0rd", "", "", "", "", "", "", ""},
		{"bad", "", "Passw0rd", "", "", "", "", "", "", ""},
		{"ok1", "文件内重复", "Passw0rd", "", "", "", "", "", "", ""},
	})

	result, err := svc.ImportUsers(context.Background(), "users.xlsx", data)
	if err != nil {
		t.Fatal(err)
	}
	if result.Total != 5 || result.Created != 2 || result.Failed != 3 {
		t.Fatalf("result=%+v", result)
	}
	if len(result.Errors) < 3 {
		t.Fatalf("errors=%+v", result.Errors)
	}

	var users []model.User
	if err := gdb.Preload("Roles").Preload("Depts").Find(&users).Error; err != nil {
		t.Fatal(err)
	}
	if len(users) != 3 {
		t.Fatalf("users count=%d", len(users))
	}

	var ok1 *model.User
	for i := range users {
		if users[i].Username == "ok1" {
			ok1 = &users[i]
		}
	}
	if ok1 == nil {
		t.Fatal("ok1 missing")
	}
	if ok1.Sex != 1 || ok1.Status != 1 || ok1.Password == "Passw0rd" || ok1.Password == "" {
		t.Fatalf("ok1 fields: sex=%d status=%d pwd=%q", ok1.Sex, ok1.Status, ok1.Password)
	}
	if !utils.BcryptCheck("Passw0rd", ok1.Password) {
		t.Fatal("password not bcrypt hashed")
	}
	if len(ok1.Depts) != 1 || ok1.Depts[0].Name != "研发部" {
		t.Fatalf("ok1 depts=%+v", ok1.Depts)
	}
	if len(ok1.Roles) != 1 || ok1.Roles[0].Name != "客服" {
		t.Fatalf("ok1 roles=%+v", ok1.Roles)
	}
	if len(published) != 1 || published[0].UserID != ok1.ID || cast.ToString(published[0].RoleIDs[0]) != cast.ToString(role.ID) {
		t.Fatalf("casbin event=%+v", published)
	}

	var emptyRole, emptyDept int64
	_ = gdb.Model(&model.Role{}).Where("name = '' OR name IS NULL").Count(&emptyRole)
	_ = gdb.Model(&model.Dept{}).Where("name = '' OR name IS NULL").Count(&emptyDept)
	if emptyRole != 0 || emptyDept != 0 {
		t.Fatalf("empty associations role=%d dept=%d", emptyRole, emptyDept)
	}
}

func TestImportUsersRejectNonXlsx(t *testing.T) {
	gdb := setupUserImportDB(t)
	svc := newImportUserService(t, gdb, nil)
	_, err := svc.ImportUsers(context.Background(), "users.csv", []byte("a,b\n1,2"))
	if err == nil || !strings.Contains(err.Error(), "xlsx") {
		t.Fatalf("want xlsx error, got %v", err)
	}
}

func TestImportUsersEmpty(t *testing.T) {
	gdb := setupUserImportDB(t)
	svc := newImportUserService(t, gdb, nil)
	data := buildImportWorkbook(t, userImportHeaders, nil)
	_, err := svc.ImportUsers(context.Background(), "users.xlsx", data)
	if err == nil {
		t.Fatal("expected empty")
	}
}

func TestBuildUserImportErrorsExcel(t *testing.T) {
	svc := &UserService{}
	data, name, ctype, err := svc.BuildUserImportErrorsExcel(&ExportUserImportErrorsReq{
		Errors: []ImportUserError{
			{Row: 2, Username: "zhangsan", Reason: "用户账号已存在"},
			{Row: 3, Username: "ls", Reason: "用户账号已存在"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if name != UserImportErrorsFileName || !strings.Contains(ctype, "spreadsheet") || len(data) < 100 {
		t.Fatalf("meta name=%s ctype=%s size=%d", name, ctype, len(data))
	}
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	rows, err := f.GetRows(f.GetSheetName(0))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 3 || rows[0][0] != "行号" || rows[1][1] != "zhangsan" {
		t.Fatalf("rows=%+v", rows)
	}
}

func TestSaveBatchTransactionalAssociations(t *testing.T) {
	gdb := setupUserImportDB(t)
	repo := repository.NewUserRepository(gdb)
	role := model.Role{Name: "AA", Code: "AA", Status: 1}
	if err := gdb.Create(&role).Error; err != nil {
		t.Fatal(err)
	}
	dept := model.Dept{Name: "D1", Status: 1}
	if err := gdb.Create(&dept).Error; err != nil {
		t.Fatal(err)
	}
	u1 := &model.User{Username: "batch1", Nickname: "B1", Status: 1}
	u1.SetPassword("Passw0rd")
	u1.SetRoles([]model.Role{role})
	u1.Depts = []model.Dept{dept}
	u2 := &model.User{Username: "batch2", Nickname: "B2", Status: 1}
	u2.SetPassword("Passw0rd")
	if err := repo.SaveBatch([]*model.User{u1, u2}); err != nil {
		t.Fatal(err)
	}
	loaded, err := repo.FindById(u1.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Roles) != 1 || len(loaded.Depts) != 1 {
		t.Fatalf("assoc roles=%d depts=%d", len(loaded.Roles), len(loaded.Depts))
	}
}
