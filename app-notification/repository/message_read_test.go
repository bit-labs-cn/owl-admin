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

func setupNotifyDB(t *testing.T) *gorm.DB {
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
	if err := gdb.AutoMigrate(&model.NotifyMessage{}, &model.NotifyRecipient{}, &model.NotifyDelivery{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return gdb
}

func seedMessage(t *testing.T, gdb *gorm.DB, messageID uint, recipientID, userID uint) {
	t.Helper()
	msg := model.NotifyMessage{Kind: model.KindNotice, Title: "测试通知"}
	msg.ID = messageID
	if err := gdb.Create(&msg).Error; err != nil {
		t.Fatalf("create message: %v", err)
	}
	row := model.NotifyRecipient{
		MessageID:  messageID,
		UserID:     userID,
		ReadStatus: model.ReadStatusUnread,
		TodoStatus: model.TodoStatusNone,
	}
	row.ID = recipientID
	if err := gdb.Create(&row).Error; err != nil {
		t.Fatalf("create recipient: %v", err)
	}
}

// 列表每行给的是接收流水 ID，标记已读必须按它更新，而不是按消息 ID。
func TestMarkAsReadUsesRecipientID(t *testing.T) {
	gdb := setupNotifyDB(t)
	repo := NewMessageRepository(gdb)
	// 消息 ID 与接收流水 ID 刻意取不同值，模拟线上雪花 ID 的情况
	seedMessage(t, gdb, 900, 1001, 7)

	if err := repo.MarkAsRead(context.Background(), 7, []uint{1001}); err != nil {
		t.Fatalf("mark as read: %v", err)
	}

	var row model.NotifyRecipient
	if err := gdb.First(&row, 1001).Error; err != nil {
		t.Fatalf("load recipient: %v", err)
	}
	if row.ReadStatus != model.ReadStatusRead {
		t.Fatalf("read status = %d, want %d", row.ReadStatus, model.ReadStatusRead)
	}
	if row.ReadAt == nil {
		t.Fatal("read at should be set")
	}
}

func TestMarkAsReadSkipsOtherUsers(t *testing.T) {
	gdb := setupNotifyDB(t)
	repo := NewMessageRepository(gdb)
	seedMessage(t, gdb, 901, 1002, 8)

	if err := repo.MarkAsRead(context.Background(), 9, []uint{1002}); err != nil {
		t.Fatalf("mark as read: %v", err)
	}

	var row model.NotifyRecipient
	if err := gdb.First(&row, 1002).Error; err != nil {
		t.Fatalf("load recipient: %v", err)
	}
	if row.ReadStatus != model.ReadStatusUnread {
		t.Fatalf("read status = %d, want unread", row.ReadStatus)
	}
}

func TestMarkAllAsReadOnlyTouchesOwnUnread(t *testing.T) {
	gdb := setupNotifyDB(t)
	repo := NewMessageRepository(gdb)
	seedMessage(t, gdb, 902, 1003, 10)
	seedMessage(t, gdb, 903, 1004, 11)

	if err := repo.MarkAllAsRead(context.Background(), 10, 0); err != nil {
		t.Fatalf("mark all as read: %v", err)
	}

	var mine, other model.NotifyRecipient
	if err := gdb.First(&mine, 1003).Error; err != nil {
		t.Fatalf("load mine: %v", err)
	}
	if err := gdb.First(&other, 1004).Error; err != nil {
		t.Fatalf("load other: %v", err)
	}
	if mine.ReadStatus != model.ReadStatusRead {
		t.Fatalf("mine read status = %d", mine.ReadStatus)
	}
	if other.ReadStatus != model.ReadStatusUnread {
		t.Fatalf("other read status = %d", other.ReadStatus)
	}
}

// 全部已读带种类时，只清该种类的未读。
func TestMarkAllAsReadScopedByKind(t *testing.T) {
	gdb := setupNotifyDB(t)
	repo := NewMessageRepository(gdb)

	notice := model.NotifyMessage{Kind: model.KindNotice, Title: "通知"}
	notice.ID = 910
	if err := gdb.Create(&notice).Error; err != nil {
		t.Fatalf("create notice: %v", err)
	}
	todo := model.NotifyMessage{Kind: model.KindTodo, Title: "待办"}
	todo.ID = 911
	if err := gdb.Create(&todo).Error; err != nil {
		t.Fatalf("create todo: %v", err)
	}
	noticeRow := model.NotifyRecipient{MessageID: 910, UserID: 12, ReadStatus: model.ReadStatusUnread}
	noticeRow.ID = 1005
	todoRow := model.NotifyRecipient{MessageID: 911, UserID: 12, ReadStatus: model.ReadStatusUnread, TodoStatus: model.TodoStatusPending}
	todoRow.ID = 1006
	if err := gdb.Create(&noticeRow).Error; err != nil {
		t.Fatalf("create notice recipient: %v", err)
	}
	if err := gdb.Create(&todoRow).Error; err != nil {
		t.Fatalf("create todo recipient: %v", err)
	}

	if err := repo.MarkAllAsRead(context.Background(), 12, model.KindTodo); err != nil {
		t.Fatalf("mark todo read: %v", err)
	}

	var readNotice, readTodo model.NotifyRecipient
	if err := gdb.First(&readNotice, 1005).Error; err != nil {
		t.Fatalf("load notice: %v", err)
	}
	if err := gdb.First(&readTodo, 1006).Error; err != nil {
		t.Fatalf("load todo: %v", err)
	}
	if readNotice.ReadStatus != model.ReadStatusUnread {
		t.Fatalf("notice should stay unread, got %d", readNotice.ReadStatus)
	}
	if readTodo.ReadStatus != model.ReadStatusRead {
		t.Fatalf("todo should be read, got %d", readTodo.ReadStatus)
	}
}
