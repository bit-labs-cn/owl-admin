package service

import (
	"context"
	"sync"
	"time"
)

// Worker 轻量后台调度协程，负责到点的定时发送、超期待办标记与历史清理。
type Worker struct {
	svc    *NotificationService
	once   sync.Once
	cancel context.CancelFunc
}

func NewWorker(svc *NotificationService) *Worker {
	return &Worker{svc: svc}
}

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
	// 每 10 秒扫一次到点的定时发送任务，保证计划时间与实际发送时间的偏差在这一量级
	scheduleTick := time.NewTicker(10 * time.Second)
	// 每 30 秒扫描一次超期待办
	tick := time.NewTicker(30 * time.Second)
	// 每 24 小时执行一次历史归档清理
	day := time.NewTicker(24 * time.Hour)
	defer scheduleTick.Stop()
	defer tick.Stop()
	defer day.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-scheduleTick.C:
			w.dispatchSchedules(ctx)
		case <-tick.C:
			w.expireTodos(ctx)
		case <-day.C:
			w.cleanup(ctx)
		}
	}
}

// dispatchSchedules 领取到点的定时任务并发送。多实例部署时靠抢占式更新保证只发一次。
func (w *Worker) dispatchSchedules(ctx context.Context) {
	w.svc.DispatchDueSchedules(ctx, 20)
}

func (w *Worker) expireTodos(ctx context.Context) {
	if err := w.svc.messages.ExpireOverdueTodos(ctx, time.Now()); err == nil {
		w.svc.hub.Publish(ctx, PushEvent{
			PrincipalType: 1,
			PrincipalID:   0,
			Event:         "todo",
		})
	}
}

func (w *Worker) cleanup(ctx context.Context) {
	// 清理 180 天前的历史归档记录
	cutoff := time.Now().AddDate(0, 0, -180)
	_ = w.svc.messages.DeleteBefore(ctx, cutoff)
}
