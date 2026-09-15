package service

import (
	"context"
	"testing"

	"bit-labs.cn/owl-admin/app/event"
	"bit-labs.cn/owl-admin/app/model"
	"github.com/asaskevich/EventBus"
	"github.com/spf13/cast"
)

func TestNormalizeBatchUserIDs(t *testing.T) {
	ids, err := normalizeBatchUserIDs([]string{"1", "1", "2", "", "0"})
	if err != nil || len(ids) != 2 || ids[0] != 1 || ids[1] != 2 {
		t.Fatalf("ids=%v err=%v", ids, err)
	}
	if _, err := normalizeBatchUserIDs(nil); err == nil {
		t.Fatal("expected empty")
	}
}

func TestBatchDeleteAndStatus(t *testing.T) {
	gdb := setupUserImportDB(t)
	svc := newImportUserService(t, gdb, nil)

	u1 := model.User{Username: "b1", Nickname: "B1", Status: 1}
	u2 := model.User{Username: "b2", Nickname: "B2", Status: 1}
	u1.SetPassword("Passw0rd")
	u2.SetPassword("Passw0rd")
	if err := gdb.Create(&u1).Error; err != nil {
		t.Fatal(err)
	}
	if err := gdb.Create(&u2).Error; err != nil {
		t.Fatal(err)
	}

	if err := svc.BatchChangeUserStatus(context.Background(), &BatchChangeUserStatusReq{
		BatchUserIDsReq: BatchUserIDsReq{IDs: []string{cast.ToString(u1.ID), cast.ToString(u2.ID)}},
		Status:          0,
	}); err != nil {
		t.Fatal(err)
	}
	var got model.User
	_ = gdb.First(&got, u1.ID)
	if got.Status != 0 {
		t.Fatalf("status=%d", got.Status)
	}

	if err := svc.BatchDeleteUsers(context.Background(), &BatchUserIDsReq{
		IDs: []string{cast.ToString(u1.ID), cast.ToString(u2.ID)},
	}); err != nil {
		t.Fatal(err)
	}
	var count int64
	_ = gdb.Model(&model.User{}).Count(&count)
	if count != 0 {
		t.Fatalf("count=%d", count)
	}
}

func TestBatchAssignRolesAndDept(t *testing.T) {
	gdb := setupUserImportDB(t)
	bus := EventBus.New()
	var published int
	bus.Subscribe(event.AssignRoleToUser, func(req *AssignRoleToUser) {
		published++
	})
	svc := newImportUserService(t, gdb, bus)

	role := model.Role{Name: "客服", Code: "cs", Status: 1}
	if err := gdb.Create(&role).Error; err != nil {
		t.Fatal(err)
	}
	dept := model.Dept{Name: "研发部", Status: 1}
	if err := gdb.Create(&dept).Error; err != nil {
		t.Fatal(err)
	}
	u1 := model.User{Username: "r1", Nickname: "R1", Status: 1}
	u2 := model.User{Username: "r2", Nickname: "R2", Status: 1}
	u1.SetPassword("Passw0rd")
	u2.SetPassword("Passw0rd")
	if err := gdb.Create(&u1).Error; err != nil {
		t.Fatal(err)
	}
	if err := gdb.Create(&u2).Error; err != nil {
		t.Fatal(err)
	}

	ids := []string{cast.ToString(u1.ID), cast.ToString(u2.ID)}
	if err := svc.BatchAssignUserRoles(context.Background(), &BatchAssignUserRolesReq{
		BatchUserIDsReq: BatchUserIDsReq{IDs: ids},
		RoleIDs:         []string{cast.ToString(role.ID)},
	}); err != nil {
		t.Fatal(err)
	}
	if published != 2 {
		t.Fatalf("published=%d", published)
	}
	loaded, err := svc.userRepo.FindById(u1.ID)
	if err != nil || len(loaded.Roles) != 1 || loaded.Roles[0].Name != "客服" {
		t.Fatalf("roles=%+v err=%v", loaded, err)
	}

	if err := svc.BatchAssignUserDept(context.Background(), &BatchAssignUserDeptReq{
		BatchUserIDsReq: BatchUserIDsReq{IDs: ids},
		DeptID:          cast.ToString(dept.ID),
	}); err != nil {
		t.Fatal(err)
	}
	loaded, err = svc.userRepo.FindById(u1.ID)
	if err != nil || len(loaded.Depts) != 1 || loaded.Depts[0].Name != "研发部" {
		t.Fatalf("depts=%+v err=%v", loaded, err)
	}

	if err := svc.BatchAssignUserDept(context.Background(), &BatchAssignUserDeptReq{
		BatchUserIDsReq: BatchUserIDsReq{IDs: ids},
		DeptID:          "",
	}); err != nil {
		t.Fatal(err)
	}
	loaded, err = svc.userRepo.FindById(u1.ID)
	if err != nil || len(loaded.Depts) != 0 {
		t.Fatalf("cleared depts=%+v err=%v", loaded.Depts, err)
	}
}
