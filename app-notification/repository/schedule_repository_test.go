package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"bit-labs.cn/owl-admin/app-notification/model"
	"bit-labs.cn/owl/utils"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupScheduleDB(t *testing.T) *gorm.DB {
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
	if err := gdb.AutoMigrate(&model.NotifySchedule{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return gdb
}

func TestClaimDueSchedulesOnlyClaimsDuePendingOnce(t *testing.T) {
	gdb := setupScheduleDB(t)
	repo := NewCatalogRepository(gdb)
	ctx := context.Background()
	now := time.Now()

	due := model.NotifySchedule{Title: "到点", PlanAt: now.Add(-time.Minute), Status: model.ScheduleStatusPending}
	future := model.NotifySchedule{Title: "未到点", PlanAt: now.Add(time.Hour), Status: model.ScheduleStatusPending}
	canceled := model.NotifySchedule{Title: "已取消", PlanAt: now.Add(-time.Minute), Status: model.ScheduleStatusCanceled}
	for _, row := range []*model.NotifySchedule{&due, &future, &canceled} {
		if err := gdb.Create(row).Error; err != nil {
			t.Fatalf("create schedule: %v", err)
		}
	}

	claimed, err := repo.ClaimDueSchedules(ctx, now, 10)
	if err != nil {
		t.Fatalf("claim due: %v", err)
	}
	if len(claimed) != 1 || claimed[0].Title != "到点" {
		t.Fatalf("claimed = %+v", claimed)
	}
	if claimed[0].Status != model.ScheduleStatusSending || claimed[0].Attempts != 1 {
		t.Fatalf("claimed row = %+v", claimed[0])
	}

	// 同一批任务不能被第二个实例再领一次
	again, err := repo.ClaimDueSchedules(ctx, now, 10)
	if err != nil {
		t.Fatalf("claim due again: %v", err)
	}
	if len(again) != 0 {
		t.Fatalf("second claim = %+v", again)
	}

	// 未到点的任务仍然是待发送
	stored, err := repo.FindSchedule(ctx, future.ID)
	if err != nil || stored == nil {
		t.Fatalf("find future: %v", err)
	}
	if stored.Status != model.ScheduleStatusPending || stored.Attempts != 0 {
		t.Fatalf("future row = %+v", stored)
	}
}

func TestClaimScheduleRejectsSentTask(t *testing.T) {
	gdb := setupScheduleDB(t)
	repo := NewCatalogRepository(gdb)
	ctx := context.Background()

	sent := model.NotifySchedule{Title: "已发送", PlanAt: time.Now(), Status: model.ScheduleStatusSent}
	failed := model.NotifySchedule{Title: "失败", PlanAt: time.Now(), Status: model.ScheduleStatusFailed}
	for _, row := range []*model.NotifySchedule{&sent, &failed} {
		if err := gdb.Create(row).Error; err != nil {
			t.Fatalf("create schedule: %v", err)
		}
	}

	ok, err := repo.ClaimSchedule(ctx, sent.ID, nil)
	if err != nil {
		t.Fatalf("claim sent: %v", err)
	}
	if ok {
		t.Fatal("已发送的任务不应被再次领取")
	}

	ok, err = repo.ClaimSchedule(ctx, failed.ID, nil)
	if err != nil {
		t.Fatalf("claim failed: %v", err)
	}
	if !ok {
		t.Fatal("失败的定时任务应允许重试领取")
	}
}
