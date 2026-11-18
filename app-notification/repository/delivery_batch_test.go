package repository

import (
	"context"
	"fmt"
	"testing"

	"bit-labs.cn/owl-admin/app-notification/model"
	"bit-labs.cn/owl/utils"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupDeliveryDB(t *testing.T) *gorm.DB {
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
	if err := gdb.AutoMigrate(&model.NotifyMessage{}, &model.NotifyRecipient{}, &model.NotifyDelivery{}, &model.NotifyType{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	// 批次查询取发送人名称会关联 admin_user，这里建最小结构即可
	if err := gdb.Exec(`CREATE TABLE IF NOT EXISTS admin_user (
		id integer PRIMARY KEY,
		nickname text,
		username text,
		deleted_at datetime
	)`).Error; err != nil {
		t.Fatalf("create admin_user: %v", err)
	}
	return gdb
}

func TestListDeliveryBatchesGroupsByMessage(t *testing.T) {
	gdb := setupDeliveryDB(t)
	repo := NewCatalogRepository(gdb)
	ctx := context.Background()

	msg := model.NotifyMessage{Kind: model.KindNotice, Title: "测试通知"}
	msg.ID = 500
	if err := gdb.Create(&msg).Error; err != nil {
		t.Fatalf("create message: %v", err)
	}
	rows := []model.NotifyDelivery{
		{MessageID: 500, UserID: 1, Channel: model.ChannelInApp, Title: "测试通知", Status: model.DeliveryStatusSuccess, RecipientType: model.RecipientTypeUser},
		{MessageID: 500, UserID: 1, Channel: model.ChannelEmail, Title: "测试通知", Status: model.DeliveryStatusFailed, RecipientType: model.RecipientTypeUser},
		{MessageID: 500, UserID: 2, Channel: model.ChannelInApp, Title: "测试通知", Status: model.DeliveryStatusSuccess, RecipientType: model.RecipientTypeUser},
	}
	if err := gdb.Create(&rows).Error; err != nil {
		t.Fatalf("create deliveries: %v", err)
	}

	total, list, err := repo.ListDeliveryBatches(ctx, DeliveryQuery{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list batches: %v", err)
	}
	if total != 1 {
		t.Fatalf("total = %d, want 1", total)
	}
	if len(list) != 1 {
		t.Fatalf("list = %+v", list)
	}
	batch := list[0]
	if batch.MessageID != 500 || batch.Title != "测试通知" {
		t.Fatalf("batch = %+v", batch)
	}
	if batch.Total != 3 || batch.SuccessCount != 2 || batch.FailedCount != 1 {
		t.Fatalf("counts = %+v", batch)
	}
	if batch.ChannelCount != 2 {
		t.Fatalf("channel count = %d", batch.ChannelCount)
	}
}

func TestListDeliveryBatchesFilteredByStatus(t *testing.T) {
	gdb := setupDeliveryDB(t)
	repo := NewCatalogRepository(gdb)
	ctx := context.Background()

	msg := model.NotifyMessage{Kind: model.KindNotice, Title: "失败通知"}
	msg.ID = 501
	if err := gdb.Create(&msg).Error; err != nil {
		t.Fatalf("create message: %v", err)
	}
	other := model.NotifyMessage{Kind: model.KindNotice, Title: "成功通知"}
	other.ID = 502
	if err := gdb.Create(&other).Error; err != nil {
		t.Fatalf("create message: %v", err)
	}
	rows := []model.NotifyDelivery{
		{MessageID: 501, UserID: 1, Channel: model.ChannelEmail, Title: "失败通知", Status: model.DeliveryStatusFailed, RecipientType: model.RecipientTypeUser},
		{MessageID: 502, UserID: 1, Channel: model.ChannelEmail, Title: "成功通知", Status: model.DeliveryStatusSuccess, RecipientType: model.RecipientTypeUser},
	}
	if err := gdb.Create(&rows).Error; err != nil {
		t.Fatalf("create deliveries: %v", err)
	}

	total, list, err := repo.ListDeliveryBatches(ctx, DeliveryQuery{Status: model.DeliveryStatusFailed, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list batches: %v", err)
	}
	if total != 1 || len(list) != 1 || list[0].MessageID != 501 {
		t.Fatalf("filtered total=%d list=%+v", total, list)
	}
}
