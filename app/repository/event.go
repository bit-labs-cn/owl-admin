package repository

import (
	"context"
	"time"

	"bit-labs.cn/owl/contract"
	"bit-labs.cn/owl/provider/db"
	eventpkg "bit-labs.cn/owl/provider/event"
	"gorm.io/gorm"
)

// EventStatusCount 发件箱 / 收件箱按状态聚合的计数。
type EventStatusCount struct {
	Status int8  `gorm:"column:status"`
	Count  int64 `gorm:"column:count"`
}

// EventTypeCount 事件类型聚合计数。
type EventTypeCount struct {
	EventType string `gorm:"column:event_type"`
	Count     int64  `gorm:"column:count"`
}

// EventSourceCount 事件来源聚合计数。
type EventSourceCount struct {
	Source string `gorm:"column:source"`
	Count  int64  `gorm:"column:count"`
}

// EventRepositoryInterface 事件发件箱 / 收件箱的查询与运维仓储。
//
// 表由框架自动迁移（owl_outbox_events / owl_inbox_events），这里只读和做必要的运维动作，
// 不改表结构、不参与投递：投递由 Relay 负责。
type EventRepositoryInterface interface {
	// RetrieveOutbox 分页查询发件箱
	RetrieveOutbox(page, pageSize int, fn func(tx *gorm.DB)) (count int64, list []eventpkg.OutboxEvent, err error)
	// FindOutbox 查询单个发件箱事件
	FindOutbox(id uint64) (*eventpkg.OutboxEvent, error)
	// RetryOutbox 把未投递成功的事件重置为待发送，返回是否有行被更新
	RetryOutbox(id uint64) (bool, error)
	// RetrieveInbox 分页查询收件箱
	RetrieveInbox(page, pageSize int, fn func(tx *gorm.DB)) (count int64, list []eventpkg.InboxEvent, err error)
	// FindInbox 查询单个收件箱事件
	FindInbox(id uint64) (*eventpkg.InboxEvent, error)

	// CountOutboxByStatus 发件箱各状态计数
	CountOutboxByStatus() (map[int8]int64, error)
	// CountInboxByStatus 收件箱各状态计数
	CountInboxByStatus() (map[int8]int64, error)
	// TopEventTypes 指定时间之后发布量最高的事件类型
	TopEventTypes(since time.Time, limit int) ([]EventTypeCount, error)
	// TopSources 指定时间之后发布量最高的事件来源
	TopSources(since time.Time, limit int) ([]EventSourceCount, error)
	// OutboxCreatedSince 指定时间之后创建的发件箱事件时间，用于按天分桶
	OutboxCreatedSince(since time.Time, limit int) ([]time.Time, error)
	// InboxSucceededSince 指定时间之后消费成功的收件箱事件时间，用于按天分桶
	InboxSucceededSince(since time.Time, limit int) ([]time.Time, error)
	// RecentOutboxFailures 最近的投递失败事件
	RecentOutboxFailures(limit int) ([]eventpkg.OutboxEvent, error)
	// PendingOutboxCountByType 按事件类型统计待发送条数
	PendingOutboxCountByType() (map[string]int64, error)
	// OldestPendingOutbox 最早的待发送事件，没有时返回 nil
	OldestPendingOutbox() (*eventpkg.OutboxEvent, error)

	// DeleteOutboxSentBefore 删除指定时间之前已投递成功的发件箱记录
	DeleteOutboxSentBefore(t time.Time) (int64, error)
	// DeleteInboxSuccessBefore 删除指定时间之前消费成功的收件箱记录
	DeleteInboxSuccessBefore(t time.Time) (int64, error)

	contract.WithContext[EventRepositoryInterface]
}

var _ EventRepositoryInterface = (*EventRepository)(nil)

type EventRepository struct {
	db  *gorm.DB
	ctx context.Context
}

func NewEventRepository(d *gorm.DB) EventRepositoryInterface {
	return &EventRepository{db: d}
}

func (i *EventRepository) WithContext(ctx context.Context) EventRepositoryInterface {
	// 返回副本而不是原地改 db：仓储在容器里是单例，原地改会让并发请求互相覆盖上下文
	cloned := *i
	cloned.db = i.db.WithContext(ctx)
	cloned.ctx = ctx
	return &cloned
}

func (i *EventRepository) RetrieveOutbox(page, pageSize int, fn func(tx *gorm.DB)) (count int64, list []eventpkg.OutboxEvent, err error) {
	tx := i.db.Model(&eventpkg.OutboxEvent{})
	if fn != nil {
		fn(tx)
	}
	if err = tx.Count(&count).Error; err != nil {
		return
	}
	err = tx.Scopes(db.Paginate(page, pageSize)).Order("id desc").Find(&list).Error
	return
}

func (i *EventRepository) FindOutbox(id uint64) (*eventpkg.OutboxEvent, error) {
	row := new(eventpkg.OutboxEvent)
	if err := i.db.Where("id = ?", id).First(row).Error; err != nil {
		return nil, err
	}
	return row, nil
}

func (i *EventRepository) RetryOutbox(id uint64) (bool, error) {
	// 已投递成功的行不再改动，避免把一条已经生效的事件重新发一遍
	res := i.db.Model(&eventpkg.OutboxEvent{}).
		Where("id = ? AND status <> ?", id, eventpkg.OutboxStatusSent).
		Updates(map[string]any{
			"status":        eventpkg.OutboxStatusPending,
			"retry_count":   0,
			"next_retry_at": nil,
			"error_msg":     "",
			"updated_at":    time.Now(),
		})
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

func (i *EventRepository) RetrieveInbox(page, pageSize int, fn func(tx *gorm.DB)) (count int64, list []eventpkg.InboxEvent, err error) {
	tx := i.db.Model(&eventpkg.InboxEvent{})
	if fn != nil {
		fn(tx)
	}
	if err = tx.Count(&count).Error; err != nil {
		return
	}
	err = tx.Scopes(db.Paginate(page, pageSize)).Order("id desc").Find(&list).Error
	return
}

func (i *EventRepository) FindInbox(id uint64) (*eventpkg.InboxEvent, error) {
	row := new(eventpkg.InboxEvent)
	if err := i.db.Where("id = ?", id).First(row).Error; err != nil {
		return nil, err
	}
	return row, nil
}

func (i *EventRepository) CountOutboxByStatus() (map[int8]int64, error) {
	var rows []EventStatusCount
	err := i.db.Model(&eventpkg.OutboxEvent{}).
		Select("status, count(*) as count").
		Group("status").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[int8]int64, len(rows))
	for _, row := range rows {
		out[row.Status] = row.Count
	}
	return out, nil
}

func (i *EventRepository) CountInboxByStatus() (map[int8]int64, error) {
	var rows []EventStatusCount
	err := i.db.Model(&eventpkg.InboxEvent{}).
		Select("status, count(*) as count").
		Group("status").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[int8]int64, len(rows))
	for _, row := range rows {
		out[row.Status] = row.Count
	}
	return out, nil
}

func (i *EventRepository) TopEventTypes(since time.Time, limit int) ([]EventTypeCount, error) {
	if limit <= 0 {
		limit = 10
	}
	var rows []EventTypeCount
	err := i.db.Model(&eventpkg.OutboxEvent{}).
		Select("event_type, count(*) as count").
		Where("created_at >= ?", since).
		Group("event_type").
		Order("count desc").
		Limit(limit).
		Find(&rows).Error
	return rows, err
}

func (i *EventRepository) TopSources(since time.Time, limit int) ([]EventSourceCount, error) {
	if limit <= 0 {
		limit = 10
	}
	var rows []EventSourceCount
	err := i.db.Model(&eventpkg.OutboxEvent{}).
		Select("source, count(*) as count").
		Where("created_at >= ?", since).
		Group("source").
		Order("count desc").
		Limit(limit).
		Find(&rows).Error
	return rows, err
}

// OutboxCreatedSince 只取时间列，按天分桶交给调用方在 Go 里做。
// 用 SQL 的 DATE() / date_trunc 会引入数据库方言，和各数据库无关的约定冲突。
func (i *EventRepository) OutboxCreatedSince(since time.Time, limit int) ([]time.Time, error) {
	var rows []eventpkg.OutboxEvent
	tx := i.db.Model(&eventpkg.OutboxEvent{}).
		Select("created_at").
		Where("created_at >= ?", since).
		// 倒序取最近的数据：超出上限时丢掉最早的部分，不要让趋势缺了最近几天
		Order("created_at desc")
	if limit > 0 {
		tx = tx.Limit(limit)
	}
	if err := tx.Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]time.Time, len(rows))
	for idx, row := range rows {
		out[idx] = row.CreatedAt
	}
	return out, nil
}

func (i *EventRepository) InboxSucceededSince(since time.Time, limit int) ([]time.Time, error) {
	var rows []eventpkg.InboxEvent
	tx := i.db.Model(&eventpkg.InboxEvent{}).
		Select("processed_at").
		Where("processed_at >= ? AND status = ?", since, eventpkg.InboxStatusSuccess).
		Order("processed_at desc")
	if limit > 0 {
		tx = tx.Limit(limit)
	}
	if err := tx.Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]time.Time, len(rows))
	for idx, row := range rows {
		out[idx] = row.ProcessedAt
	}
	return out, nil
}

func (i *EventRepository) RecentOutboxFailures(limit int) ([]eventpkg.OutboxEvent, error) {
	if limit <= 0 {
		limit = 10
	}
	var rows []eventpkg.OutboxEvent
	err := i.db.
		Where("status = ?", eventpkg.OutboxStatusFailed).
		Order("id desc").
		Limit(limit).
		Find(&rows).Error
	return rows, err
}

func (i *EventRepository) OldestPendingOutbox() (*eventpkg.OutboxEvent, error) {
	row := new(eventpkg.OutboxEvent)
	err := i.db.
		Where("status = ?", eventpkg.OutboxStatusPending).
		Order("id asc").
		First(row).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return row, nil
}

// PendingOutboxCountByType 按事件类型统计待发送条数，用于回答「这个主题的真实积压是多少」。
// 与 Redis 的 lag 是两个维度：这里算的是 Relay 还没投出去的量，lag 是消费端还没读到的量。
func (i *EventRepository) PendingOutboxCountByType() (map[string]int64, error) {
	var rows []EventTypeCount
	err := i.db.Model(&eventpkg.OutboxEvent{}).
		Select("event_type, count(*) as count").
		Where("status = ?", eventpkg.OutboxStatusPending).
		Group("event_type").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[string]int64, len(rows))
	for _, row := range rows {
		out[row.EventType] = row.Count
	}
	return out, nil
}

func (i *EventRepository) DeleteOutboxSentBefore(t time.Time) (int64, error) {
	res := i.db.
		Where("status = ? AND created_at < ?", eventpkg.OutboxStatusSent, t).
		Delete(&eventpkg.OutboxEvent{})
	return res.RowsAffected, res.Error
}

func (i *EventRepository) DeleteInboxSuccessBefore(t time.Time) (int64, error) {
	res := i.db.
		Where("status = ? AND processed_at < ?", eventpkg.InboxStatusSuccess, t).
		Delete(&eventpkg.InboxEvent{})
	return res.RowsAffected, res.Error
}
