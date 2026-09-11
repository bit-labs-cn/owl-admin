package repository

import (
	"fmt"
	"testing"

	"bit-labs.cn/owl-admin/app/model"
	"bit-labs.cn/owl/provider/db"
	"bit-labs.cn/owl/utils"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupAssocDB(t *testing.T) *gorm.DB {
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

func TestAssignRoleDoesNotCreateEmptyRole(t *testing.T) {
	gdb := setupAssocDB(t)
	repo := NewUserRepository(gdb)

	role := model.Role{Name: "客服", Code: "customerService", Status: 1}
	if err := gdb.Create(&role).Error; err != nil {
		t.Fatalf("create role: %v", err)
	}
	user := model.User{Username: "zs", Nickname: "张三", Status: 1}
	if err := gdb.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	loaded, err := repo.FindById(user.ID)
	if err != nil {
		t.Fatalf("find user: %v", err)
	}
	loaded.SetRoles(db.GetModelsByIDs[model.Role]([]string{fmt.Sprintf("%d", role.ID)}))
	if err := repo.Save(loaded); err != nil {
		t.Fatalf("assign role: %v", err)
	}

	var roleCount, emptyCount int64
	_ = gdb.Model(&model.Role{}).Count(&roleCount)
	_ = gdb.Model(&model.Role{}).Where("name = '' OR name IS NULL").Count(&emptyCount)
	reloaded, _ := repo.FindById(user.ID)
	if roleCount != 1 || emptyCount != 0 || len(reloaded.Roles) != 1 || reloaded.Roles[0].Name != "客服" {
		t.Fatalf("want 1 named role assigned, got roles=%d empty=%d assigned=%d", roleCount, emptyCount, len(reloaded.Roles))
	}
}

func TestAssignDeptDoesNotCreateEmptyDept(t *testing.T) {
	gdb := setupAssocDB(t)
	repo := NewUserRepository(gdb)

	dept := model.Dept{Name: "A部门", Status: 1}
	if err := gdb.Create(&dept).Error; err != nil {
		t.Fatalf("create dept: %v", err)
	}
	user := model.User{Username: "zs2", Status: 1}
	if err := gdb.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	loaded, err := repo.FindById(user.ID)
	if err != nil {
		t.Fatalf("find user: %v", err)
	}
	loaded.Depts = db.GetModelsByIDs[model.Dept]([]string{fmt.Sprintf("%d", dept.ID)})
	if err := repo.Save(loaded); err != nil {
		t.Fatalf("assign dept: %v", err)
	}

	var deptCount, emptyCount int64
	_ = gdb.Model(&model.Dept{}).Count(&deptCount)
	_ = gdb.Model(&model.Dept{}).Where("name = '' OR name IS NULL").Count(&emptyCount)
	reloaded, _ := repo.FindById(user.ID)
	if deptCount != 1 || emptyCount != 0 || len(reloaded.Depts) != 1 || reloaded.Depts[0].Name != "A部门" {
		t.Fatalf("want 1 named dept assigned, got depts=%d empty=%d assigned=%d", deptCount, emptyCount, len(reloaded.Depts))
	}
}

func TestGetModelsByIDsSkipsZero(t *testing.T) {
	for _, id := range []string{"", "0"} {
		got := db.GetModelsByIDs[model.Role]([]string{id})
		if len(got) != 0 {
			t.Fatalf("id=%q should be skipped, got %+v", id, got)
		}
	}
}

func TestAssignRoleSkipsEmptyIDs(t *testing.T) {
	gdb := setupAssocDB(t)
	repo := NewUserRepository(gdb)

	role := model.Role{Name: "客服", Code: "cs", Status: 1}
	if err := gdb.Create(&role).Error; err != nil {
		t.Fatalf("create role: %v", err)
	}
	user := model.User{Username: "empty-id", Status: 1}
	if err := gdb.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	loaded, err := repo.FindById(user.ID)
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	loaded.SetRoles(db.GetModelsByIDs[model.Role]([]string{fmt.Sprintf("%d", role.ID), "", "0"}))
	if err := repo.Save(loaded); err != nil {
		t.Fatalf("save: %v", err)
	}

	var emptyCount int64
	_ = gdb.Model(&model.Role{}).Where("name = '' OR name IS NULL").Count(&emptyCount)
	reloaded, _ := repo.FindById(user.ID)
	if emptyCount != 0 || len(reloaded.Roles) != 1 {
		t.Fatalf("want 1 assigned and 0 empty, got empty=%d assigned=%d", emptyCount, len(reloaded.Roles))
	}
}

func TestReplaceJoinTableDedupesRoles(t *testing.T) {
	gdb := setupAssocDB(t)
	repo := NewUserRepository(gdb)

	role := model.Role{Name: "AA", Code: "AA", Status: 1}
	if err := gdb.Create(&role).Error; err != nil {
		t.Fatalf("create role: %v", err)
	}
	user := model.User{Username: "dup", Status: 1}
	if err := gdb.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	loaded, err := repo.FindById(user.ID)
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	loaded.SetRoles([]model.Role{role, role, role, role})
	if err := repo.Save(loaded); err != nil {
		t.Fatalf("save dup roles: %v", err)
	}

	reloaded, err := repo.FindById(user.ID)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	var joinCount int64
	_ = gdb.Table("admin_user_role").Where("user_id = ?", user.ID).Count(&joinCount)
	if joinCount != 1 || len(reloaded.Roles) != 1 {
		t.Fatalf("want 1 role/join, got assigned=%d joins=%d", len(reloaded.Roles), joinCount)
	}
}
