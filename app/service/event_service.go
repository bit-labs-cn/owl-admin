package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"bit-labs.cn/owl-admin/app/repository"
	errContract "bit-labs.cn/owl/contract/errors"
	"bit-labs.cn/owl/contract/log"
	eventpkg "bit-labs.cn/owl/provider/event"
	"bit-labs.cn/owl/provider/router"
	validatorv10 "github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	// workbenchDefaultRangeDays 工作台默认统计天数
	workbenchDefaultRangeDays = 7
	// workbenchMaxRangeDays 工作台最多统计天数，避免一次扫太多历史
	workbenchMaxRangeDays = 30
	// trendRowLimit 单侧趋势取数上限；超过后标记 truncated，不阻断展示
	trendRowLimit = 20000
	// recentFailureLimit 工作台「最近失败」条数
	recentFailureLimit = 10
	// activeConsumerIdleThreshold 消费者空闲多久内仍算活跃
	// 活着的消费者一直在阻塞读（2 秒一次），空闲时间只有秒级；这里放宽到 2 分钟避免误判
	activeConsumerIdleThreshold = 2 * time.Minute
)

// EventService 事件总线工作台与事件记录查询。
//
// 只读发件箱 / 收件箱与 Redis Stream 概况；唯一的写操作是「重投」和「清理历史」，
// 都不改变投递语义：重投只是把行重置为待发送，仍由框架 Relay 负责投递。
type EventService struct {
	repo     repository.EventRepositoryInterface
	bus      *eventpkg.Bus
	redis    redis.UniversalClient
	logger   log.Logger
	validate *validatorv10.Validate
}

func NewEventService(
	repo repository.EventRepositoryInterface,
	bus *eventpkg.Bus,
	redisClient redis.UniversalClient,
	logger log.Logger,
	validate *validatorv10.Validate,
) *EventService {
	return &EventService{
		repo:     repo,
		bus:      bus,
		redis:    redisClient,
		logger:   logger,
		validate: validate,
	}
}

// ---------- 状态文案 ----------

// OutboxStatusLabel 发件箱状态文案。
func OutboxStatusLabel(status int8) string {
	switch status {
	case eventpkg.OutboxStatusPending:
		return "待发送"
	case eventpkg.OutboxStatusSent:
		return "已发送"
	case eventpkg.OutboxStatusFailed:
		return "投递失败"
	default:
		return fmt.Sprintf("未知(%d)", status)
	}
}

// InboxStatusLabel 收件箱状态文案。
func InboxStatusLabel(status int8) string {
	switch status {
	case eventpkg.InboxStatusSuccess:
		return "消费成功"
	case eventpkg.InboxStatusFailed:
		return "消费失败"
	default:
		return fmt.Sprintf("未知(%d)", status)
	}
}

// ---------- 工作台 ----------

// EventStreamSnapshot Redis Stream 概况。
type EventStreamSnapshot struct {
	Key    string `json:"key"`
	Exists bool   `json:"exists"`
	Length int64  `json:"length"`
	Error  string `json:"error,omitempty"`
}

// EventConsumerGroup 消费组积压与消费者活跃情况。
type EventConsumerGroup struct {
	Name string `json:"name"`
	// TotalConsumers Redis 里累计登记过的消费者个数（含已退出进程留下的记录）
	TotalConsumers int64 `json:"totalConsumers"`
	// ActiveConsumers 最近仍在读消息的消费者个数
	ActiveConsumers int64 `json:"activeConsumers"`
	// LastActiveSeconds 最近一次读取距今秒数，没有消费者时为 -1
	LastActiveSeconds int64 `json:"lastActiveSeconds"`
	Pending           int64 `json:"pending"`
	// Lag 该组还未读取的条数。注意是「整个 stream」的未读量，包含其它主题
	Lag             int64  `json:"lag"`
	EntriesRead     int64  `json:"entriesRead"`
	LastDeliveredID string `json:"lastDeliveredId"`
	// Topic 组名去掉应用前缀后的主题
	Topic string `json:"topic"`
	// TopicPending 该主题在发件箱里还没投递成功的条数（Relay 侧积压）
	TopicPending int64 `json:"topicPending"`
	// TopicKnown 是否能从组名反推出主题（组名被截断时为 false）
	TopicKnown bool `json:"topicKnown"`
}

// EventOutboxSummary 发件箱状态计数。
type EventOutboxSummary struct {
	Total   int64 `json:"total"`
	Pending int64 `json:"pending"`
	Sent    int64 `json:"sent"`
	Failed  int64 `json:"failed"`
}

// EventInboxSummary 收件箱状态计数。
type EventInboxSummary struct {
	Total   int64 `json:"total"`
	Success int64 `json:"success"`
	Failed  int64 `json:"failed"`
}

// EventOldestPending 最老待发送事件。
type EventOldestPending struct {
	ID         uint64 `json:"id"`
	EventID    string `json:"eventId"`
	EventType  string `json:"eventType"`
	Source     string `json:"source"`
	Priority   int8   `json:"priority"`
	CreatedAt  int64  `json:"createdAt"`
	AgeSeconds int64  `json:"ageSeconds"`
}

// EventTrendPoint 按天的事件量趋势。
type EventTrendPoint struct {
	Date      string `json:"date"`
	Published int64  `json:"published"`
	Consumed  int64  `json:"consumed"`
}

// EventTypeItem 事件类型 / 来源的聚合项。
type EventTypeItem struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

// EventWorkbenchResp 工作台聚合响应。
type EventWorkbenchResp struct {
	GeneratedAt    int64                `json:"generatedAt"`
	AppName        string               `json:"appName"`
	DriverName     string               `json:"driverName"`
	RangeDays      int                  `json:"rangeDays"`
	Stream         EventStreamSnapshot  `json:"stream"`
	Outbox         EventOutboxSummary   `json:"outbox"`
	Inbox          EventInboxSummary    `json:"inbox"`
	OldestPending  *EventOldestPending  `json:"oldestPending"`
	Trend          []EventTrendPoint    `json:"trend"`
	TrendTruncated bool                 `json:"trendTruncated"`
	TopTypes       []EventTypeItem      `json:"topTypes"`
	TopSources     []EventTypeItem      `json:"topSources"`
	ConsumerGroups []EventConsumerGroup `json:"consumerGroups"`
	RecentFailures []EventOutboxRow     `json:"recentFailures"`
}

// Workbench 汇总事件总线运行情况。days 为统计天数，超出范围时收敛到 1～30。
func (i *EventService) Workbench(ctx context.Context, days int) (*EventWorkbenchResp, error) {
	if days <= 0 {
		days = workbenchDefaultRangeDays
	}
	if days > workbenchMaxRangeDays {
		days = workbenchMaxRangeDays
	}

	resp := &EventWorkbenchResp{
		GeneratedAt: time.Now().Unix(),
		AppName:     i.bus.AppName(),
		RangeDays:   days,
		DriverName:  i.bus.DriverName(),
		Trend:       []EventTrendPoint{},
		TopTypes:    []EventTypeItem{},
		TopSources:  []EventTypeItem{},
	}

	outboxCounts, err := i.repo.WithContext(ctx).CountOutboxByStatus()
	if err != nil {
		return nil, err
	}

	for status, count := range outboxCounts {
		resp.Outbox.Total += count
		switch status {
		case eventpkg.OutboxStatusPending:
			resp.Outbox.Pending = count
		case eventpkg.OutboxStatusSent:
			resp.Outbox.Sent = count
		case eventpkg.OutboxStatusFailed:
			resp.Outbox.Failed = count
		}
	}

	inboxCounts, err := i.repo.WithContext(ctx).CountInboxByStatus()
	if err != nil {
		return nil, err
	}

	for status, count := range inboxCounts {
		resp.Inbox.Total += count
		switch status {
		case eventpkg.InboxStatusSuccess:
			resp.Inbox.Success = count
		case eventpkg.InboxStatusFailed:
			resp.Inbox.Failed = count
		}
	}

	oldest, err := i.repo.WithContext(ctx).OldestPendingOutbox()
	if err != nil {
		return nil, err
	}
	if oldest != nil {
		resp.OldestPending = &EventOldestPending{
			ID:         oldest.ID,
			EventID:    oldest.EventID,
			EventType:  oldest.EventType,
			Source:     oldest.Source,
			Priority:   oldest.Priority,
			CreatedAt:  oldest.CreatedAt.Unix(),
			AgeSeconds: int64(time.Since(oldest.CreatedAt).Seconds()),
		}
	}

	since := trendSince(days)

	published, err := i.repo.WithContext(ctx).OutboxCreatedSince(since, trendRowLimit+1)
	if err != nil {
		return nil, err
	}
	consumed, err := i.repo.WithContext(ctx).InboxSucceededSince(since, trendRowLimit+1)
	if err != nil {
		return nil, err
	}
	resp.Trend, resp.TrendTruncated = buildTrend(days, published, consumed)

	types, err := i.repo.WithContext(ctx).TopEventTypes(since, 10)
	if err != nil {
		return nil, err
	}
	for _, item := range types {
		name := item.EventType
		if strings.TrimSpace(name) == "" {
			name = "未命名事件"
		}
		resp.TopTypes = append(resp.TopTypes, EventTypeItem{Name: name, Count: item.Count})
	}

	sources, err := i.repo.WithContext(ctx).TopSources(since, 10)
	if err != nil {
		return nil, err
	}
	for _, item := range sources {
		name := item.Source
		if strings.TrimSpace(name) == "" {
			name = "未标注来源"
		}
		resp.TopSources = append(resp.TopSources, EventTypeItem{Name: name, Count: item.Count})
	}

	resp.Stream, resp.ConsumerGroups = i.loadStream(ctx)

	failures, err := i.repo.WithContext(ctx).RecentOutboxFailures(recentFailureLimit)
	if err != nil {
		return nil, err
	}
	for idx := range failures {
		resp.RecentFailures = append(resp.RecentFailures, toOutboxRow(&failures[idx]))
	}

	return resp, nil
}

// trendSince 趋势起点：本地时区「days 天前」的零点。
func trendSince(days int) time.Time {
	now := time.Now()
	loc := now.Location()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	return start.AddDate(0, 0, -(days - 1))
}

// buildTrend 把事件时间按本地日期分桶。日期截断放在 Go 里做，避免写数据库方言函数。
func buildTrend(days int, published, consumed []time.Time) ([]EventTrendPoint, bool) {
	truncated := len(published) > trendRowLimit || len(consumed) > trendRowLimit

	start := trendSince(days)
	loc := start.Location()
	points := make([]EventTrendPoint, days)
	index := make(map[string]int, days)
	for offset := 0; offset < days; offset++ {
		day := start.AddDate(0, 0, offset)
		key := day.Format("2006-01-02")
		index[key] = offset
		points[offset] = EventTrendPoint{Date: key}
	}

	for _, at := range published {
		if idx, ok := index[at.In(loc).Format("2006-01-02")]; ok {
			points[idx].Published++
		}
	}
	for _, at := range consumed {
		if idx, ok := index[at.In(loc).Format("2006-01-02")]; ok {
			points[idx].Consumed++
		}
	}

	return points, truncated
}

// loadStream 读取 Redis Stream 概况与消费组积压。Redis 不可用时只填 error，不影响其它数据。
// stream key 取自当前驱动（每个应用一条），不再假设固定的全局 key。
func (i *EventService) loadStream(ctx context.Context) (EventStreamSnapshot, []EventConsumerGroup) {
	key := i.bus.Target()
	snapshot := EventStreamSnapshot{Key: key}
	groups := []EventConsumerGroup{}

	if i.bus.DriverName() != "redis" {
		snapshot.Error = "当前驱动不是 Redis，没有 Stream 可查"
		return snapshot, groups
	}
	if key == "" {
		snapshot.Error = "当前 Stream key 为空"
		return snapshot, groups
	}
	if i.redis == nil {
		snapshot.Error = "未配置 Redis 客户端"
		return snapshot, groups
	}

	probeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	exists, err := i.redis.Exists(probeCtx, key).Result()
	if err != nil {
		snapshot.Error = err.Error()
		return snapshot, groups
	}
	if exists == 0 {
		return snapshot, groups
	}
	snapshot.Exists = true

	length, err := i.redis.XLen(probeCtx, key).Result()
	if err != nil {
		snapshot.Error = err.Error()
	}
	snapshot.Length = length

	info, err := i.redis.XInfoGroups(probeCtx, key).Result()
	if err != nil {
		if snapshot.Error == "" {
			snapshot.Error = err.Error()
		}
		return snapshot, groups
	}
	for _, group := range info {
		row := EventConsumerGroup{
			Name:              group.Name,
			TotalConsumers:    group.Consumers,
			LastActiveSeconds: -1,
			Pending:           group.Pending,
			Lag:               group.Lag,
			EntriesRead:       group.EntriesRead,
			LastDeliveredID:   group.LastDeliveredID,
		}
		if consumers, err := i.redis.XInfoConsumers(probeCtx, key, group.Name).Result(); err == nil {
			row.TotalConsumers = int64(len(consumers))
			for _, consumer := range consumers {
				idle := int64(consumer.Idle.Seconds())
				if consumer.Idle <= activeConsumerIdleThreshold {
					row.ActiveConsumers++
				}
				if row.LastActiveSeconds < 0 || idle < row.LastActiveSeconds {
					row.LastActiveSeconds = idle
				}
			}
		}
		groups = append(groups, row)
	}
	i.annotateTopicBacklog(ctx, groups)
	return snapshot, groups
}

// annotateTopicBacklog 用发件箱待发送量补上「本主题真实积压」。
//
// 消费组名是「组前缀.主题」，前缀取 app-name，所以去掉前缀就能还原主题。
// 组名超长时 Bus 会截断成「前缀.摘要」，这种反推不出主题，标记 TopicKnown=false 让前端显示占位。
func (i *EventService) annotateTopicBacklog(ctx context.Context, groups []EventConsumerGroup) {
	if len(groups) == 0 {
		return
	}

	counts, err := i.repo.WithContext(ctx).PendingOutboxCountByType()
	if err != nil {
		// 只影响这一列，不阻断工作台其它数据
		if i.logger != nil {
			i.logger.Warning("load pending outbox count by type failed: ", err)
		}
		return
	}

	prefix := i.bus.AppName() + "."
	for idx := range groups {
		name := groups[idx].Name
		topic := strings.TrimPrefix(name, prefix)
		if topic == name || topic == "" {
			continue
		}
		groups[idx].Topic = topic
		groups[idx].TopicKnown = true
		// 查不到就是这条主题当前没有待发送的记录，积压为 0 是准确值
		groups[idx].TopicPending = counts[topic]
	}
}

// ---------- 发件箱 ----------

// RetrieveOutboxReq 发件箱查询条件。
type RetrieveOutboxReq struct {
	router.PageReq
	EventID          string `json:"eventId" form:"eventId" validate:"omitempty,max=64" label:"事件ID"`
	EventType        string `json:"eventType" form:"eventType" validate:"omitempty,max=128" label:"事件类型"`
	Source           string `json:"source" form:"source" validate:"omitempty,max=64" label:"来源"`
	Subject          string `json:"subject" form:"subject" validate:"omitempty,max=128" label:"业务实体"`
	Status           *int8  `json:"status" form:"status" validate:"omitempty,gte=0,lte=2" label:"状态"`
	CreatedAtBetween string `json:"createdAt" form:"createdAt" validate:"omitempty" label:"创建时间区间"`
}

// EventOutboxRow 发件箱列表行。
type EventOutboxRow struct {
	ID          uint64 `json:"id"`
	EventID     string `json:"eventId"`
	EventType   string `json:"eventType"`
	Source      string `json:"source"`
	Subject     string `json:"subject"`
	Priority    int8   `json:"priority"`
	Status      int8   `json:"status"`
	StatusLabel string `json:"statusLabel"`
	RetryCount  int    `json:"retryCount"`
	CreatedAt   int64  `json:"createdAt"`
	UpdatedAt   int64  `json:"updatedAt"`
	PublishedAt *int64 `json:"publishedAt,omitempty"`
	NextRetryAt *int64 `json:"nextRetryAt,omitempty"`
	ErrorMsg    string `json:"errorMsg,omitempty"`
}

// EventOutboxDetail 发件箱详情，附带事件载荷与元数据。
type EventOutboxDetail struct {
	EventOutboxRow
	Payload string `json:"payload"`
	Headers string `json:"headers"`
}

// RetrieveOutbox 分页查询发件箱。
func (i *EventService) RetrieveOutbox(ctx context.Context, req *RetrieveOutboxReq) (count int64, list []EventOutboxRow, err error) {
	if err := i.validate.Struct(req); err != nil {
		return 0, nil, err
	}

	start, end, err := parseTimeRange(req.CreatedAtBetween)
	if err != nil {
		return 0, nil, err
	}

	count, rows, err := i.repo.WithContext(ctx).RetrieveOutbox(req.Page, req.PageSize, func(tx *gorm.DB) {
		if req.EventID != "" {
			tx.Where("event_id = ?", req.EventID)
		}
		if req.EventType != "" {
			tx.Where("event_type like ?", "%"+req.EventType+"%")
		}
		if req.Source != "" {
			tx.Where("source = ?", req.Source)
		}
		if req.Subject != "" {
			tx.Where("subject like ?", "%"+req.Subject+"%")
		}
		if req.Status != nil {
			tx.Where("status = ?", *req.Status)
		}
		if start != nil {
			tx.Where("created_at >= ?", *start)
		}
		if end != nil {
			tx.Where("created_at <= ?", *end)
		}
	})
	if err != nil {
		return 0, nil, err
	}

	list = make([]EventOutboxRow, 0, len(rows))
	for idx := range rows {
		list = append(list, toOutboxRow(&rows[idx]))
	}
	return count, list, nil
}

// OutboxDetail 查询发件箱事件详情。
func (i *EventService) OutboxDetail(ctx context.Context, id uint64) (*EventOutboxDetail, error) {
	row, err := i.repo.WithContext(ctx).FindOutbox(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errContract.NewBizError("event.outbox.not_found", "发件箱事件不存在")
		}
		return nil, err
	}
	detail := &EventOutboxDetail{
		EventOutboxRow: toOutboxRow(row),
		Payload:        row.Payload,
		Headers:        row.Headers,
	}
	return detail, nil
}

// RetryOutbox 把未投递成功的事件重置为待发送，由 Relay 在下一个轮询周期内重新投递。
func (i *EventService) RetryOutbox(ctx context.Context, id uint64) (bool, error) {
	retried, err := i.repo.WithContext(ctx).RetryOutbox(id)
	if err != nil {
		return false, err
	}
	if retried {
		return true, nil
	}

	// 没有行被更新：要么记录不存在，要么已经投递成功，分别给出可读原因
	row, err := i.repo.WithContext(ctx).FindOutbox(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, errContract.NewBizError("event.outbox.not_found", "发件箱事件不存在")
		}
		return false, err
	}
	if row.Status == eventpkg.OutboxStatusSent {
		return false, errContract.NewBizError("event.outbox.already_sent", "该事件已投递成功，无需重投")
	}
	return false, nil
}

// ---------- 收件箱 ----------

// RetrieveInboxReq 收件箱查询条件。
type RetrieveInboxReq struct {
	router.PageReq
	EventID            string `json:"eventId" form:"eventId" validate:"omitempty,max=64" label:"事件ID"`
	ConsumerGroup      string `json:"consumerGroup" form:"consumerGroup" validate:"omitempty,max=64" label:"消费组"`
	Status             *int8  `json:"status" form:"status" validate:"omitempty,gte=1,lte=2" label:"状态"`
	ProcessedAtBetween string `json:"processedAt" form:"processedAt" validate:"omitempty" label:"处理时间区间"`
}

// EventInboxRow 收件箱列表行。
type EventInboxRow struct {
	ID            uint64 `json:"id"`
	EventID       string `json:"eventId"`
	ConsumerGroup string `json:"consumerGroup"`
	Status        int8   `json:"status"`
	StatusLabel   string `json:"statusLabel"`
	ErrorMsg      string `json:"errorMsg,omitempty"`
	ProcessedAt   int64  `json:"processedAt"`
	CreatedAt     int64  `json:"createdAt"`
}

// RetrieveInbox 分页查询收件箱。
func (i *EventService) RetrieveInbox(ctx context.Context, req *RetrieveInboxReq) (count int64, list []EventInboxRow, err error) {
	if err := i.validate.Struct(req); err != nil {
		return 0, nil, err
	}

	start, end, err := parseTimeRange(req.ProcessedAtBetween)
	if err != nil {
		return 0, nil, err
	}

	count, rows, err := i.repo.WithContext(ctx).RetrieveInbox(req.Page, req.PageSize, func(tx *gorm.DB) {
		if req.EventID != "" {
			tx.Where("event_id = ?", req.EventID)
		}
		if req.ConsumerGroup != "" {
			tx.Where("consumer_group like ?", "%"+req.ConsumerGroup+"%")
		}
		if req.Status != nil {
			tx.Where("status = ?", *req.Status)
		}
		if start != nil {
			tx.Where("processed_at >= ?", *start)
		}
		if end != nil {
			tx.Where("processed_at <= ?", *end)
		}
	})
	if err != nil {
		return 0, nil, err
	}

	list = make([]EventInboxRow, 0, len(rows))
	for idx := range rows {
		list = append(list, toInboxRow(&rows[idx]))
	}
	return count, list, nil
}

// InboxDetail 查询收件箱事件详情。
func (i *EventService) InboxDetail(ctx context.Context, id uint64) (*EventInboxRow, error) {
	row, err := i.repo.WithContext(ctx).FindInbox(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errContract.NewBizError("event.inbox.not_found", "收件箱流水不存在")
		}
		return nil, err
	}
	detail := toInboxRow(row)
	return &detail, nil
}

// ---------- 清理 ----------

// CleanupEventsReq 清理历史请求。
type CleanupEventsReq struct {
	OlderThanDays int      `json:"olderThanDays" validate:"required,gte=1,lte=365" label:"保留天数"`
	Targets       []string `json:"targets" validate:"required,min=1,dive,oneof=outbox inbox" label:"清理目标"`
}

// CleanupEventsResp 清理结果。
type CleanupEventsResp struct {
	Cutoff        int64 `json:"cutoff"`
	OutboxDeleted int64 `json:"outboxDeleted"`
	InboxDeleted  int64 `json:"inboxDeleted"`
}

// Cleanup 删除指定天数之前已投递成功 / 消费成功的历史记录。
//
// 只删已经结束的记录：待发送、投递失败、消费失败的流水保留，方便继续排查和重投。
func (i *EventService) Cleanup(ctx context.Context, req *CleanupEventsReq) (*CleanupEventsResp, error) {
	if err := i.validate.Struct(req); err != nil {
		return nil, err
	}

	cutoff := time.Now().AddDate(0, 0, -req.OlderThanDays)
	resp := &CleanupEventsResp{Cutoff: cutoff.Unix()}

	for _, target := range req.Targets {
		switch target {
		case "outbox":
			deleted, err := i.repo.WithContext(ctx).DeleteOutboxSentBefore(cutoff)
			if err != nil {
				return nil, err
			}
			resp.OutboxDeleted = deleted
		case "inbox":
			deleted, err := i.repo.WithContext(ctx).DeleteInboxSuccessBefore(cutoff)
			if err != nil {
				return nil, err
			}
			resp.InboxDeleted = deleted
		}
	}

	return resp, nil
}

// ---------- 消费者清理 ----------

// CleanupConsumersReq 清理僵尸消费者请求。
type CleanupConsumersReq struct {
	IdleDays int `json:"idleDays" validate:"required,gte=1,lte=365" label:"空闲天数"`
}

// ConsumerCleanupItem 单个消费组的清理结果。
type ConsumerCleanupItem struct {
	Group          string `json:"group"`
	Removed        int64  `json:"removed"`
	SkippedPending int64  `json:"skippedPending"`
}

// CleanupConsumersResp 清理僵尸消费者结果。
type CleanupConsumersResp struct {
	Groups  []ConsumerCleanupItem `json:"groups"`
	Removed int64                 `json:"removed"`
	Skipped int64                 `json:"skipped"`
}

// CleanupConsumers 删除各消费组中长期没有读取、且没有滞留消息的消费者记录。
//
// Redis 不会自动回收消费者：进程每次重启都会以新的 pid 生成新的消费者名，旧名字会一直留在
// 组里，把「消费者数」堆成几十个。这里只清 pending = 0 且空闲超过 idleDays 的记录；
// 还有未确认消息的消费者一律跳过，避免把待重试的 PEL 弄丢。
func (i *EventService) CleanupConsumers(ctx context.Context, req *CleanupConsumersReq) (*CleanupConsumersResp, error) {
	if err := i.validate.Struct(req); err != nil {
		return nil, err
	}
	if i.redis == nil {
		return nil, errContract.NewBizError("event.redis.unavailable", "未配置 Redis 客户端，无法清理消费者")
	}

	key := i.bus.Target()
	if i.bus.DriverName() != "redis" || key == "" {
		return nil, errContract.NewBizError("event.redis.unavailable", "当前驱动不是 Redis，没有消费者可清理")
	}
	probeCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	exists, err := i.redis.Exists(probeCtx, key).Result()
	if err != nil {
		return nil, errContract.WrapBizError("event.redis.failed", "读取事件 Stream 失败", err)
	}
	if exists == 0 {
		return &CleanupConsumersResp{Groups: []ConsumerCleanupItem{}}, nil
	}

	groups, err := i.redis.XInfoGroups(probeCtx, key).Result()
	if err != nil {
		return nil, errContract.WrapBizError("event.redis.failed", "读取消费组失败", err)
	}

	idleThreshold := time.Duration(req.IdleDays) * 24 * time.Hour
	resp := &CleanupConsumersResp{Groups: make([]ConsumerCleanupItem, 0, len(groups))}

	for _, group := range groups {
		consumers, err := i.redis.XInfoConsumers(probeCtx, key, group.Name).Result()
		if err != nil {
			return nil, errContract.WrapBizError("event.redis.failed", "读取消费者列表失败", err)
		}

		item := ConsumerCleanupItem{Group: group.Name}
		for _, consumer := range consumers {
			if consumer.Pending > 0 {
				item.SkippedPending++
				continue
			}
			if consumer.Idle < idleThreshold {
				continue
			}
			if err := i.redis.XGroupDelConsumer(probeCtx, key, group.Name, consumer.Name).Err(); err != nil {
				return nil, errContract.WrapBizError("event.redis.failed", "删除消费者失败", err)
			}
			item.Removed++
		}

		resp.Removed += item.Removed
		resp.Skipped += item.SkippedPending
		resp.Groups = append(resp.Groups, item)
	}

	return resp, nil
}

// ---------- 辅助 ----------

func toOutboxRow(row *eventpkg.OutboxEvent) EventOutboxRow {
	out := EventOutboxRow{
		ID:          row.ID,
		EventID:     row.EventID,
		EventType:   row.EventType,
		Source:      row.Source,
		Subject:     row.Subject,
		Priority:    row.Priority,
		Status:      row.Status,
		StatusLabel: OutboxStatusLabel(row.Status),
		RetryCount:  row.RetryCount,
		CreatedAt:   row.CreatedAt.Unix(),
		UpdatedAt:   row.UpdatedAt.Unix(),
		ErrorMsg:    row.ErrorMsg,
	}
	if row.PublishedAt != nil {
		published := row.PublishedAt.Unix()
		out.PublishedAt = &published
	}
	if row.NextRetryAt != nil {
		next := row.NextRetryAt.Unix()
		out.NextRetryAt = &next
	}
	return out
}

func toInboxRow(row *eventpkg.InboxEvent) EventInboxRow {
	return EventInboxRow{
		ID:            row.ID,
		EventID:       row.EventID,
		ConsumerGroup: row.ConsumerGroup,
		Status:        row.Status,
		StatusLabel:   InboxStatusLabel(row.Status),
		ErrorMsg:      row.ErrorMsg,
		ProcessedAt:   row.ProcessedAt.Unix(),
		CreatedAt:     row.CreatedAt.Unix(),
	}
}

// parseTimeRange 解析前端统一使用的 "start,end" Unix 秒区间。空串返回两个 nil。
func parseTimeRange(raw string) (*time.Time, *time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil, nil
	}
	parts := strings.Split(raw, ",")
	if len(parts) != 2 {
		return nil, nil, fmt.Errorf("时间区间格式应为 start,end")
	}
	start, err := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64)
	if err != nil {
		return nil, nil, fmt.Errorf("开始时间格式错误: %w", err)
	}
	end, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
	if err != nil {
		return nil, nil, fmt.Errorf("结束时间格式错误: %w", err)
	}
	startTime := time.Unix(start, 0)
	endTime := time.Unix(end, 0)
	return &startTime, &endTime, nil
}
