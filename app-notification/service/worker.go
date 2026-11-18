package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"bit-labs.cn/owl-admin/app-notification/contract"
	notifEvent "bit-labs.cn/owl-admin/app-notification/event"
	"bit-labs.cn/owl-admin/app-notification/model"
	contractEvent "bit-labs.cn/owl/contract/event"
	"github.com/spf13/cast"
	"gorm.io/gorm"
)

// DispatchPendingScheduled 针对过了免打扰时段或定时重试的到期投递，通过事件总线进行补发。
func (s *TaskService) DispatchPendingScheduled(ctx context.Context) error {
	rows, err := s.deliveries.WithContext(ctx).ClaimDue(time.Now(), 50)
	if err != nil {
		return err
	}
	for _, row := range rows {
		var topic string
		switch row.Channel {
		case model.ChannelEmail:
			topic = notifEvent.TopicDispatchEmail
		case model.ChannelSMS:
			topic = notifEvent.TopicDispatchSMS
		case model.ChannelApp:
			topic = notifEvent.TopicDispatchApp
		}
		if topic != "" && s.bus != nil {
			evt, err := contractEvent.NewEvent(topic, notifEvent.DispatchPayload{
				DeliveryID:    row.ID,
				TaskID:        row.TaskID,
				RecipientType: row.RecipientType,
				RecipientID:   row.RecipientID,
				Channel:       row.Channel,
				Address:       row.Address,
			}, contractEvent.WithSource("owl-notification"), contractEvent.WithSubject(fmt.Sprintf("delivery:%d", row.ID)), contractEvent.WithPriority(contractEvent.PriorityHigh))
			if err == nil {
				_ = s.bus.PublishTx(ctx, s.db, evt)
				continue
			}
		}
		// 无事件总线时降级本机执行
		_ = s.DispatchOne(ctx, row.ID, row.Channel)
	}
	return nil
}

// Retry 把失败投递重新放回队列，并通过事件总线触发重发。
func (s *TaskService) Retry(ctx context.Context, id uint) error {
	row, err := s.deliveries.WithContext(ctx).Detail(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return NotFound("投递记录不存在")
	}
	if err != nil {
		return err
	}
	if row.Status != model.DeliveryStatusFailed && row.Status != model.DeliveryStatusSkipped {
		return NotAllowed("只有失败或跳过的投递可以重试")
	}
	row.Status = model.DeliveryStatusPending
	row.ErrorCode = ""
	row.ErrorMsg = ""
	row.NextRetryAt = nil
	if err := s.deliveries.WithContext(ctx).Save(row); err != nil {
		return err
	}

	var topic string
	switch row.Channel {
	case model.ChannelEmail:
		topic = notifEvent.TopicDispatchEmail
	case model.ChannelSMS:
		topic = notifEvent.TopicDispatchSMS
	case model.ChannelApp:
		topic = notifEvent.TopicDispatchApp
	}
	if topic != "" && s.bus != nil {
		evt, err := contractEvent.NewEvent(topic, notifEvent.DispatchPayload{
			DeliveryID:    row.ID,
			TaskID:        row.TaskID,
			RecipientType: row.RecipientType,
			RecipientID:   row.RecipientID,
			Channel:       row.Channel,
			Address:       row.Address,
		}, contractEvent.WithSource("owl-notification"), contractEvent.WithSubject(fmt.Sprintf("delivery:%d", row.ID)), contractEvent.WithPriority(contractEvent.PriorityHigh))
		if err == nil {
			_ = s.bus.PublishTx(ctx, s.db, evt)
		}
	} else {
		_ = s.DispatchOne(ctx, row.ID, row.Channel)
	}

	return s.refreshStatus(ctx, row.TaskID)
}

// DispatchOne 处理单条投递记录的发送。由事件总线各渠道消费者调用。
func (s *TaskService) DispatchOne(ctx context.Context, id uint, expectedChannel string) error {
	row, err := s.deliveries.WithContext(ctx).Detail(id)
	if err != nil {
		return err
	}
	if expectedChannel != "" && row.Channel != expectedChannel {
		return nil
	}
	if row.Status == model.DeliveryStatusSent || row.Status == model.DeliveryStatusSkipped {
		return nil
	}
	task, err := s.tasks.WithContext(ctx).Detail(row.TaskID)
	if err != nil {
		s.failDelivery(ctx, row, err.Error(), false)
		return err
	}
	if task.Status == model.TaskStatusCancelled {
		row.Status = model.DeliveryStatusSkipped
		row.ErrorMsg = "任务已取消"
		_ = s.deliveries.WithContext(ctx).Save(row)
		_ = s.refreshStatus(ctx, task.ID)
		return nil
	}
	if limit := s.rateLimit(ctx); limit > 0 && row.Channel != model.ChannelInApp && task.Category != model.CategorySecurity {
		since := time.Now().Add(-24 * time.Hour)
		n, err := s.deliveries.WithContext(ctx).CountToday(row.RecipientType, row.RecipientID, row.Channel, since)
		if err == nil && n > int64(limit) {
			s.failDelivery(ctx, row, "超过每日频率限制", false)
			_ = s.refreshStatus(ctx, task.ID)
			return nil
		}
	}
	req := &contract.ChannelRequest{
		TaskID: task.ID, DeliveryID: row.ID, Title: task.Title, ContentHTML: task.ContentHTML, ContentText: task.ContentText,
		Address: row.Address, SMSTemplate: task.SMSTemplate, Kind: task.Kind,
	}
	fromJSON(task.VariableJSON, &req.Variables)
	fromJSON(task.PayloadJSON, &req.Payload)
	if task.LinkJSON != "" {
		req.Link = &contract.DeepLink{}
		fromJSON(task.LinkJSON, req.Link)
	}
	var sender contract.Channel
	switch row.Channel {
	case model.ChannelEmail:
		sender = s.email
	case model.ChannelSMS:
		sender = s.sms
	case model.ChannelApp:
		sender = s.app
		if row.RecipientType == model.RecipientTypeUser {
			devices, err := s.devices.WithContext(ctx).ListActive(model.PrincipalTypeAdmin, row.RecipientID)
			if err != nil {
				s.failDelivery(ctx, row, err.Error(), true)
				_ = s.refreshStatus(ctx, task.ID)
				return err
			}
			for _, d := range devices {
				req.DeviceTokens = append(req.DeviceTokens, contract.DeviceTarget{ID: d.ID, Provider: d.Provider, Token: d.Token, Platform: d.Platform})
			}
		}
	default:
		s.failDelivery(ctx, row, "渠道不需要异步发送", false)
		_ = s.refreshStatus(ctx, task.ID)
		return nil
	}
	if account, err := s.accounts.WithContext(ctx).FindDefault(row.Channel); err == nil {
		req.ConfigJSON = account.ConfigJSON
		req.Provider = account.Provider
		row.Provider = account.Provider
	}
	result, err := sender.Send(ctx, req)
	if result != nil && len(result.InvalidTokens) > 0 {
		_ = s.devices.WithContext(ctx).InvalidateTokens(result.InvalidTokens)
	}
	if err != nil {
		s.failDelivery(ctx, row, err.Error(), true)
		_ = s.refreshStatus(ctx, task.ID)
		return err
	}
	now := time.Now()
	row.Status = model.DeliveryStatusSent
	row.SentAt = &now
	row.ErrorMsg = ""
	row.ErrorCode = ""
	if result != nil {
		row.ProviderMsgID = result.ProviderMsgID
	}
	_ = s.deliveries.WithContext(ctx).Save(row)
	_ = s.refreshStatus(ctx, task.ID)
	return nil
}

func (s *TaskService) failDelivery(ctx context.Context, row *model.Delivery, reason string, retry bool) {
	row.RetryCount++
	row.ErrorMsg = trimErr(reason)
	if retry && row.RetryCount <= maxRetry {
		row.Status = model.DeliveryStatusPending
		next := time.Now().Add(time.Duration(row.RetryCount*row.RetryCount) * time.Minute)
		row.NextRetryAt = &next
	} else {
		row.Status = model.DeliveryStatusFailed
	}
	_ = s.deliveries.WithContext(ctx).Save(row)
}

func (s *TaskService) rateLimit(ctx context.Context) int {
	raw, ok, err := s.settings.WithContext(ctx).Get(settingRate)
	if err != nil || !ok {
		return 0
	}
	return cast.ToInt(raw)
}

// RunScheduled 发送到达时间的定时任务。
func (s *TaskService) RunScheduled(ctx context.Context) {
	rows, err := s.tasks.WithContext(ctx).ListDue(time.Now(), 20)
	if err != nil {
		return
	}
	for _, row := range rows {
		_ = s.Submit(ctx, row.ID)
	}
}

// ExpireTodos 把过了截止时间的待办标成已过期，并推一条实时事件。
func (s *TaskService) ExpireTodos(ctx context.Context) {
	rows, err := s.inbox.WithContext(ctx).ExpireDue(time.Now())
	if err != nil {
		return
	}
	seen := map[string]struct{}{}
	for _, row := range rows {
		key := cast.ToString(row.PrincipalType) + ":" + cast.ToString(row.PrincipalID)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		s.hub.Publish(ctx, PushEvent{PrincipalType: row.PrincipalType, PrincipalID: row.PrincipalID, Event: "todo"})
	}
}

// Cleanup 按保留天数清理站内信和投递记录。
func (s *TaskService) Cleanup(ctx context.Context) {
	if days := s.days(ctx, settingInboxDays, 180); days > 0 {
		_ = s.inbox.WithContext(ctx).DeleteBefore(time.Now().AddDate(0, 0, -days))
	}
	if days := s.days(ctx, settingDeliverDay, 90); days > 0 {
		_ = s.deliveries.WithContext(ctx).DeleteBefore(time.Now().AddDate(0, 0, -days))
	}
}

func (s *TaskService) days(ctx context.Context, key string, fallback int) int {
	raw, ok, err := s.settings.WithContext(ctx).Get(key)
	if err != nil || !ok || raw == "" {
		return fallback
	}
	n := cast.ToInt(raw)
	if n <= 0 {
		return fallback
	}
	return n
}

// Worker 轮询到期投递、定时任务和待办过期。
type Worker struct {
	tasks  *TaskService
	once   sync.Once
	cancel context.CancelFunc
}

func NewWorker(tasks *TaskService) *Worker { return &Worker{tasks: tasks} }

func (w *Worker) Start() {
	w.once.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		w.cancel = cancel
		go w.loop(ctx)
	})
}

func (w *Worker) Stop() {
	if w.cancel != nil {
		w.cancel()
	}
}

func (w *Worker) loop(ctx context.Context) {
	// 定时扫描降低为 10 秒（常规投递均已由 Event Bus 实时驱动，此处仅做定时任务、待办过期与免打扰延期补发）
	tick := time.NewTicker(10 * time.Second)
	hour := time.NewTicker(time.Hour)
	defer tick.Stop()
	defer hour.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			_ = w.tasks.DispatchPendingScheduled(ctx)
			w.tasks.RunScheduled(ctx)
			w.tasks.ExpireTodos(ctx)
		case <-hour.C:
			w.tasks.Cleanup(ctx)
		}
	}
}
