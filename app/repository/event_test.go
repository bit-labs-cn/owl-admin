package repository

import (
	"fmt"
	"testing"
	"time"

	eventpkg "bit-labs.cn/owl/provider/event"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupEventRepoDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	gdb, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := gdb.AutoMigrate(&eventpkg.OutboxEvent{}, &eventpkg.InboxEvent{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return gdb
}

func mustCreateOutbox(t *testing.T, gdb *gorm.DB, row *eventpkg.OutboxEvent) *eventpkg.OutboxEvent {
	t.Helper()
	if row.CreatedAt.IsZero() {
		row.CreatedAt = time.Now()
	}
	if row.UpdatedAt.IsZero() {
		row.UpdatedAt = row.CreatedAt
	}
	if err := gdb.Create(row).Error; err != nil {
		t.Fatalf("create outbox: %v", err)
	}
	return row
}

func TestRetryOutboxSkipsSentRows(t *testing.T) {
	gdb := setupEventRepoDB(t)
	repo := NewEventRepository(gdb)

	pending := mustCreateOutbox(t, gdb, &eventpkg.OutboxEvent{
		EventID: "e-pending", EventType: "a.b", Payload: "{}",
		Status: eventpkg.OutboxStatusPending,
	})
	failed := mustCreateOutbox(t, gdb, &eventpkg.OutboxEvent{
		EventID: "e-failed", EventType: "a.b", Payload: "{}",
		Status: eventpkg.OutboxStatusFailed, RetryCount: 5, ErrorMsg: "boom",
	})
	sent := mustCreateOutbox(t, gdb, &eventpkg.OutboxEvent{
		EventID: "e-sent", EventType: "a.b", Payload: "{}",
		Status: eventpkg.OutboxStatusSent,
	})

	ok, err := repo.RetryOutbox(failed.ID)
	if err != nil {
		t.Fatalf("retry failed row: %v", err)
	}
	if !ok {
		t.Fatal("失败事件重投应返回已更新")
	}

	ok, err = repo.RetryOutbox(sent.ID)
	if err != nil {
		t.Fatalf("retry sent row: %v", err)
	}
	if ok {
		t.Fatal("已投递成功的事件不应被重投")
	}

	got, err := repo.FindOutbox(failed.ID)
	if err != nil {
		t.Fatalf("find failed row: %v", err)
	}
	if got.Status != eventpkg.OutboxStatusPending || got.RetryCount != 0 || got.ErrorMsg != "" || got.NextRetryAt != nil {
		t.Fatalf("重投后状态不正确: status=%d retry=%d err=%q next=%v",
			got.Status, got.RetryCount, got.ErrorMsg, got.NextRetryAt)
	}

	stillPending, err := repo.FindOutbox(pending.ID)
	if err != nil {
		t.Fatalf("find pending row: %v", err)
	}
	if stillPending.Status != eventpkg.OutboxStatusPending {
		t.Fatalf("待发送事件状态被改动: %d", stillPending.Status)
	}
}

func TestCleanupOnlyRemovesExpiredFinishedRows(t *testing.T) {
	gdb := setupEventRepoDB(t)
	repo := NewEventRepository(gdb)

	old := time.Now().AddDate(0, 0, -30)
	recent := time.Now().Add(-time.Hour)

	mustCreateOutbox(t, gdb, &eventpkg.OutboxEvent{
		EventID: "old-sent", EventType: "a.b", Payload: "{}",
		Status: eventpkg.OutboxStatusSent, CreatedAt: old,
	})
	mustCreateOutbox(t, gdb, &eventpkg.OutboxEvent{
		EventID: "old-failed", EventType: "a.b", Payload: "{}",
		Status: eventpkg.OutboxStatusFailed, CreatedAt: old,
	})
	mustCreateOutbox(t, gdb, &eventpkg.OutboxEvent{
		EventID: "recent-sent", EventType: "a.b", Payload: "{}",
		Status: eventpkg.OutboxStatusSent, CreatedAt: recent,
	})

	inboxRows := []eventpkg.InboxEvent{
		{EventID: "old-sent", ConsumerGroup: "g1", Status: eventpkg.InboxStatusSuccess, ProcessedAt: old, CreatedAt: old},
		{EventID: "old-failed", ConsumerGroup: "g1", Status: eventpkg.InboxStatusFailed, ProcessedAt: old, CreatedAt: old},
		{EventID: "recent-sent", ConsumerGroup: "g1", Status: eventpkg.InboxStatusSuccess, ProcessedAt: recent, CreatedAt: recent},
	}
	for idx := range inboxRows {
		if err := gdb.Create(&inboxRows[idx]).Error; err != nil {
			t.Fatalf("create inbox: %v", err)
		}
	}

	cutoff := time.Now().AddDate(0, 0, -7)

	deletedOutbox, err := repo.DeleteOutboxSentBefore(cutoff)
	if err != nil {
		t.Fatalf("delete outbox: %v", err)
	}
	if deletedOutbox != 1 {
		t.Fatalf("发件箱应只删 1 条过期已发送，实际 %d", deletedOutbox)
	}

	deletedInbox, err := repo.DeleteInboxSuccessBefore(cutoff)
	if err != nil {
		t.Fatalf("delete inbox: %v", err)
	}
	if deletedInbox != 1 {
		t.Fatalf("收件箱应只删 1 条过期成功记录，实际 %d", deletedInbox)
	}

	var outboxCount int64
	if err := gdb.Model(&eventpkg.OutboxEvent{}).Count(&outboxCount).Error; err != nil {
		t.Fatal(err)
	}
	if outboxCount != 2 {
		t.Fatalf("发件箱剩余 %d 条，期望 2", outboxCount)
	}

	var inboxCount int64
	if err := gdb.Model(&eventpkg.InboxEvent{}).Count(&inboxCount).Error; err != nil {
		t.Fatal(err)
	}
	if inboxCount != 2 {
		t.Fatalf("收件箱剩余 %d 条，期望 2", inboxCount)
	}
}

func TestRetrieveOutboxFiltersByStatusAndType(t *testing.T) {
	gdb := setupEventRepoDB(t)
	repo := NewEventRepository(gdb)

	mustCreateOutbox(t, gdb, &eventpkg.OutboxEvent{EventID: "e1", EventType: "order.created", Payload: "{}", Status: eventpkg.OutboxStatusPending})
	mustCreateOutbox(t, gdb, &eventpkg.OutboxEvent{EventID: "e2", EventType: "order.created", Payload: "{}", Status: eventpkg.OutboxStatusSent})
	mustCreateOutbox(t, gdb, &eventpkg.OutboxEvent{EventID: "e3", EventType: "user.login", Payload: "{}", Status: eventpkg.OutboxStatusPending})

	count, list, err := repo.RetrieveOutbox(1, 20, func(tx *gorm.DB) {
		tx.Where("event_type = ? AND status = ?", "order.created", eventpkg.OutboxStatusPending)
	})
	if err != nil {
		t.Fatalf("retrieve: %v", err)
	}
	if count != 1 || len(list) != 1 || list[0].EventID != "e1" {
		t.Fatalf("筛选结果不正确: count=%d list=%+v", count, list)
	}
}

func TestTrendSourcesOnlyReturnRequestedColumns(t *testing.T) {
	gdb := setupEventRepoDB(t)
	repo := NewEventRepository(gdb)

	base := time.Now().Add(-48 * time.Hour)
	mustCreateOutbox(t, gdb, &eventpkg.OutboxEvent{EventID: "t1", EventType: "a.b", Payload: "{}", Status: eventpkg.OutboxStatusSent, CreatedAt: base})
	mustCreateOutbox(t, gdb, &eventpkg.OutboxEvent{EventID: "t2", EventType: "a.b", Payload: "{}", Status: eventpkg.OutboxStatusSent, CreatedAt: base.Add(time.Hour)})

	times, err := repo.OutboxCreatedSince(time.Now().AddDate(0, 0, -7), 20000)
	if err != nil {
		t.Fatalf("outbox since: %v", err)
	}
	if len(times) != 2 {
		t.Fatalf("应取到 2 条时间，实际 %d", len(times))
	}

	empty, err := repo.OutboxCreatedSince(time.Now().Add(-time.Hour), 20000)
	if err != nil {
		t.Fatalf("outbox since recent: %v", err)
	}
	if len(empty) != 0 {
		t.Fatalf("近期不应有记录，实际 %d", len(empty))
	}
}

func TestPendingOutboxCountByType(t *testing.T) {
	gdb := setupEventRepoDB(t)
	repo := NewEventRepository(gdb)

	mustCreateOutbox(t, gdb, &eventpkg.OutboxEvent{EventID: "p1", EventType: "cms.article.changed", Payload: "{}", Status: eventpkg.OutboxStatusPending})
	mustCreateOutbox(t, gdb, &eventpkg.OutboxEvent{EventID: "p2", EventType: "cms.article.changed", Payload: "{}", Status: eventpkg.OutboxStatusPending})
	mustCreateOutbox(t, gdb, &eventpkg.OutboxEvent{EventID: "p3", EventType: "cms.document.changed", Payload: "{}", Status: eventpkg.OutboxStatusPending})
	// 已发送与失败的不计入积压
	mustCreateOutbox(t, gdb, &eventpkg.OutboxEvent{EventID: "s1", EventType: "cms.article.changed", Payload: "{}", Status: eventpkg.OutboxStatusSent})
	mustCreateOutbox(t, gdb, &eventpkg.OutboxEvent{EventID: "f1", EventType: "cms.article.changed", Payload: "{}", Status: eventpkg.OutboxStatusFailed})

	counts, err := repo.PendingOutboxCountByType()
	if err != nil {
		t.Fatalf("count by type: %v", err)
	}
	if counts["cms.article.changed"] != 2 {
		t.Fatalf("cms.article.changed 待发送应为 2，实际 %d", counts["cms.article.changed"])
	}
	if counts["cms.document.changed"] != 1 {
		t.Fatalf("cms.document.changed 待发送应为 1，实际 %d", counts["cms.document.changed"])
	}
	if _, ok := counts["order.created"]; ok {
		t.Fatal("不存在的类型不应出现在结果里")
	}
}

func TestOldestPendingOutbox(t *testing.T) {
	gdb := setupEventRepoDB(t)
	repo := NewEventRepository(gdb)

	none, err := repo.OldestPendingOutbox()
	if err != nil {
		t.Fatalf("oldest pending on empty: %v", err)
	}
	if none != nil {
		t.Fatalf("空表应返回 nil，实际 %+v", none)
	}

	older := time.Now().Add(-2 * time.Hour)
	mustCreateOutbox(t, gdb, &eventpkg.OutboxEvent{EventID: "p1", EventType: "a.b", Payload: "{}", Status: eventpkg.OutboxStatusPending, CreatedAt: older})
	mustCreateOutbox(t, gdb, &eventpkg.OutboxEvent{EventID: "p2", EventType: "a.b", Payload: "{}", Status: eventpkg.OutboxStatusPending})

	oldest, err := repo.OldestPendingOutbox()
	if err != nil {
		t.Fatalf("oldest pending: %v", err)
	}
	if oldest == nil || oldest.EventID != "p1" {
		t.Fatalf("最早的待发送事件应为 p1，实际 %+v", oldest)
	}
}
