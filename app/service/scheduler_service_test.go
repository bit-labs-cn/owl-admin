package service

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	errContract "bit-labs.cn/owl/contract/errors"
	schedulerContract "bit-labs.cn/owl/contract/scheduler"
	"bit-labs.cn/owl/provider/redis"
	schedulerpkg "bit-labs.cn/owl/provider/scheduler"
	"bit-labs.cn/owl/utils"
	"github.com/glebarez/sqlite"
	validatorv10 "github.com/go-playground/validator/v10"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"bit-labs.cn/owl-admin/app/repository"
)

// noopLocker 是测试用的分布式锁替身，直接放行。
type noopLocker struct{}

func (noopLocker) New() redis.Locker { return noopLock{} }

type noopLock struct{}

func (noopLock) Lock(string) error { return nil }
func (noopLock) Unlock()           {}

func newSchedulerTestService(t *testing.T) (*SchedulerService, *gorm.DB) {
	t.Helper()

	// 应用启动时会初始化雪花算法；测试里确保它可用即可。
	// 已有测试可能先用别的 worker 初始化过，这里不因参数不同而失败。
	_ = utils.InitSnowFlakeWorker(1, 3)

	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "scheduler-service.db")), &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Silent),
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("打开测试数据库失败: %v", err)
	}
	if err := db.AutoMigrate(&schedulerpkg.Job{}, &schedulerpkg.Run{}); err != nil {
		t.Fatalf("迁移测试表失败: %v", err)
	}
	if sqlDB, dbErr := db.DB(); dbErr == nil {
		t.Cleanup(func() { _ = sqlDB.Close() })
	}

	// 直接落一条内置任务，模拟子系统启动后调度引擎同步出来的结果。
	builtin := &schedulerpkg.Job{
		Code:           schedulerContract.Code("cms", "publish-due"),
		Name:           "发布到点文章",
		OwnerApp:       "cms",
		JobKey:         "publish-due",
		Source:         schedulerpkg.JobSourceBuiltin,
		Cron:           "*/5 * * * *",
		ScriptType:     int8(schedulerContract.ScriptTypeHandler),
		TimeoutSeconds: 60,
		MaxAttempts:    1,
		OverlapPolicy:  int8(schedulerContract.OverlapSkip),
		Status:         schedulerpkg.JobStatusEnabled,
	}
	if err := db.Create(builtin).Error; err != nil {
		t.Fatalf("创建内置任务失败: %v", err)
	}

	registry := schedulerContract.NewRegistry()
	if err := registry.Add("cms", schedulerContract.Job{
		Key:  "publish-due",
		Name: "发布到点文章",
		Cron: "*/5 * * * *",
		Handler: func(ctx context.Context, in schedulerContract.Input) (string, error) {
			return "ok", nil
		},
	}); err != nil {
		t.Fatalf("登记内置任务失败: %v", err)
	}
	if err := registry.AddActions("notify",
		schedulerContract.Action{
			Key:    "send-message",
			Name:   "发送通知",
			Group:  "通知中心",
			Params: []schedulerContract.Param{
				{Name: "title", Label: "标题", Type: schedulerContract.ParamString, Required: true},
			},
			Handler: func(ctx context.Context, in schedulerContract.ActionInput) (string, error) {
				return "sent:" + in.Params["title"], nil
			},
		},
		schedulerContract.Action{
			Key:    "power-on",
			Name:   "开机",
			Group:  "物理开关",
			Params: []schedulerContract.Param{
				{Name: "switch", Label: "开关", Type: schedulerContract.ParamSelect, Required: true, Dynamic: true},
			},
			Options: func(ctx context.Context, params map[string]string) (map[string][]schedulerContract.ParamOption, error) {
				return map[string][]schedulerContract.ParamOption{
					"switch": {{Value: "1", Label: "前台"}},
				}, nil
			},
			Handler: func(ctx context.Context, in schedulerContract.ActionInput) (string, error) {
				return "powered", nil
			},
		},
	); err != nil {
		t.Fatalf("登记可调用方法失败: %v", err)
	}

	manager := schedulerpkg.NewManagerWithConfig(db, registry, schedulerpkg.Config{
		Enabled:          true,
		TickSeconds:      1,
		Workers:          2,
		Timezone:         "UTC",
		RunRetentionDays: 30,
		MaxOutputBytes:   4096,
		LeaseSeconds:     30,
		ShellEnabled:     true,
	}, nil)

	svc := NewSchedulerService(repository.NewSchedulerRepository(db), noopLocker{}, validatorv10.New(), manager)
	return svc, db
}

func TestCreateJobValidatesAndComputesNextRun(t *testing.T) {
	svc, db := newSchedulerTestService(t)

	req := &CreateJobReq{
		Name:                "巡检回调",
		Cron:                "*/10 * * * *",
		Timezone:            "UTC",
		ScriptType:          int8(schedulerContract.ScriptTypeHTTP),
		ScriptConfig:        `{"url":"https://example.com/hook"}`,
		TimeoutSeconds:      30,
		MaxAttempts:         2,
		RetryBackoffSeconds: 60,
		OverlapPolicy:       int8(schedulerContract.OverlapSkip),
	}
	created, err := svc.CreateJob(context.Background(), req)
	if err != nil {
		t.Fatalf("创建自定义任务失败: %v", err)
	}
	if !strings.HasPrefix(created.Code, "custom:") {
		t.Fatalf("自定义任务编码应以 custom: 开头，实际 %s", created.Code)
	}
	if created.NextRunAt == nil {
		t.Fatal("创建后应算好下次执行时间")
	}

	stored := new(schedulerpkg.Job)
	if err := db.Where("id = ?", created.ID).First(stored).Error; err != nil {
		t.Fatalf("读取任务失败: %v", err)
	}
	if stored.Source != schedulerpkg.JobSourceCustom || stored.Status != schedulerpkg.JobStatusEnabled {
		t.Fatalf("自定义任务来源或状态不正确: %+v", stored)
	}
}

func TestCreateJobRejectsInvalidCron(t *testing.T) {
	svc, _ := newSchedulerTestService(t)

	_, err := svc.CreateJob(context.Background(), &CreateJobReq{
		Name:                "坏表达式",
		Cron:                "99 99 * * *",
		ScriptType:          int8(schedulerContract.ScriptTypeHTTP),
		ScriptConfig:        `{"url":"https://example.com/hook"}`,
		TimeoutSeconds:      30,
		MaxAttempts:         1,
		RetryBackoffSeconds: 0,
		OverlapPolicy:       int8(schedulerContract.OverlapSkip),
	})
	var bizErr *errContract.BizError
	if !errors.As(err, &bizErr) || bizErr.Code != CodeSchedulerCronInvalid {
		t.Fatalf("非法 Cron 应返回 %s，实际 %v", CodeSchedulerCronInvalid, err)
	}
}

func TestBuiltinJobIsImmutable(t *testing.T) {
	svc, db := newSchedulerTestService(t)

	var builtin schedulerpkg.Job
	if err := db.Where("source = ?", schedulerpkg.JobSourceBuiltin).First(&builtin).Error; err != nil {
		t.Fatalf("读取内置任务失败: %v", err)
	}

	_, err := svc.UpdateJob(context.Background(), &UpdateJobReq{
		ID: builtin.ID,
		CreateJobReq: CreateJobReq{
			Name:                "改个名字",
			Cron:                "*/5 * * * *",
			ScriptType:          int8(schedulerContract.ScriptTypeHTTP),
			ScriptConfig:        `{"url":"https://example.com/hook"}`,
			TimeoutSeconds:      30,
			MaxAttempts:         1,
			RetryBackoffSeconds: 0,
			OverlapPolicy:       int8(schedulerContract.OverlapSkip),
		},
	})
	var bizErr *errContract.BizError
	if !errors.As(err, &bizErr) || bizErr.Code != CodeSchedulerBuiltinImmutable {
		t.Fatalf("内置任务修改应返回 %s，实际 %v", CodeSchedulerBuiltinImmutable, err)
	}

	if err := svc.DeleteJob(context.Background(), builtin.ID); err == nil {
		t.Fatal("内置任务不应允许删除")
	}

	// 启停是允许的。
	disabled := int8(schedulerpkg.JobStatusDisabled)
	if err := svc.ChangeStatus(context.Background(), builtin.ID, &ChangeStatusReq{Status: disabled}); err != nil {
		t.Fatalf("内置任务应允许停用: %v", err)
	}
	reloaded := new(schedulerpkg.Job)
	if err := db.Where("id = ?", builtin.ID).First(reloaded).Error; err != nil {
		t.Fatalf("读取任务失败: %v", err)
	}
	if reloaded.Status != schedulerpkg.JobStatusDisabled {
		t.Fatalf("停用未生效，实际 %d", reloaded.Status)
	}
	if reloaded.NextRunAt != nil {
		t.Fatal("停用后应清空下次执行时间")
	}
}

func TestCreateActionJobStoresSummaryAndListsAllActions(t *testing.T) {
	svc, db := newSchedulerTestService(t)

	// 普通动作：参数按声明校验，摘要用候选 label。
	created, err := svc.CreateJob(context.Background(), &CreateJobReq{
		Name:                "每天发通知",
		Cron:                "0 9 * * *",
		Timezone:            "UTC",
		ScriptType:          schedulerpkg.ScriptTypeAction,
		ActionCode:          schedulerContract.ActionCode("notify", "send-message"),
		ActionParams:        map[string]string{"title": "站会提醒"},
		TimeoutSeconds:      30,
		MaxAttempts:         1,
		RetryBackoffSeconds: 60,
		OverlapPolicy:       int8(schedulerContract.OverlapSkip),
	})
	if err != nil {
		t.Fatalf("创建动作任务失败: %v", err)
	}
	if created.ActionName != "发送通知" {
		t.Fatalf("动作名称快照不正确: %q", created.ActionName)
	}
	if !strings.Contains(created.ScriptSummary, "标题：站会提醒") {
		t.Fatalf("参数摘要应带可读参数名与值，实际 %q", created.ScriptSummary)
	}

	// 必填参数缺失应被拒绝。
	_, err = svc.CreateJob(context.Background(), &CreateJobReq{
		Name:                "缺参数",
		Cron:                "0 9 * * *",
		ScriptType:          schedulerpkg.ScriptTypeAction,
		ActionCode:          schedulerContract.ActionCode("notify", "send-message"),
		ActionParams:        map[string]string{},
		TimeoutSeconds:      30,
		MaxAttempts:         1,
		RetryBackoffSeconds: 60,
		OverlapPolicy:       int8(schedulerContract.OverlapSkip),
	})
	var bizErr *errContract.BizError
	if !errors.As(err, &bizErr) || bizErr.Code != CodeSchedulerScriptInvalid {
		t.Fatalf("缺失必填参数应返回 %s，实际 %v", CodeSchedulerScriptInvalid, err)
	}

	created, err = svc.CreateJob(context.Background(), &CreateJobReq{
		Name:                "每天开机",
		Cron:                "0 8 * * *",
		Timezone:            "UTC",
		ScriptType:          schedulerpkg.ScriptTypeAction,
		ActionCode:          schedulerContract.ActionCode("notify", "power-on"),
		ActionParams:        map[string]string{"switch": "1"},
		TimeoutSeconds:      30,
		MaxAttempts:         1,
		RetryBackoffSeconds: 60,
		OverlapPolicy:       int8(schedulerContract.OverlapSkip),
	})
	if err != nil {
		t.Fatalf("创建动作任务失败: %v", err)
	}
	if !strings.Contains(created.ScriptSummary, "开关：前台") {
		t.Fatalf("动态候选的摘要应使用 label，实际 %q", created.ScriptSummary)
	}

	// 触发只看路由权限与任务状态，动作不再按角色过滤。
	run, err := svc.TriggerJob(context.Background(), created.ID, &TriggerJobReq{})
	if err != nil {
		t.Fatalf("触发动作任务失败: %v", err)
	}
	if run.Status != schedulerpkg.RunStatusPending {
		t.Fatalf("触发后应写入待执行记录，实际 %d", run.Status)
	}

	// 目录登记了什么就列出什么，不再有「仅超管」过滤。
	if len(svc.RetrieveActions()) != 2 {
		t.Fatalf("目录应列出全部动作，实际 %d", len(svc.RetrieveActions()))
	}

	// 保存的任务字段应包含动作快照。
	stored := new(schedulerpkg.Job)
	if err := db.Where("id = ?", created.ID).First(stored).Error; err != nil {
		t.Fatalf("读取任务失败: %v", err)
	}
	if stored.ScriptType != schedulerpkg.ScriptTypeAction || stored.ActionCode == "" {
		t.Fatalf("动作任务字段不完整: %+v", stored)
	}
}

func TestTriggerBuiltinJobWritesPendingRun(t *testing.T) {
	svc, db := newSchedulerTestService(t)

	var builtin schedulerpkg.Job
	if err := db.Where("source = ?", schedulerpkg.JobSourceBuiltin).First(&builtin).Error; err != nil {
		t.Fatalf("读取内置任务失败: %v", err)
	}

	run, err := svc.TriggerJob(context.Background(), builtin.ID, &TriggerJobReq{})
	if err != nil {
		t.Fatalf("手动触发失败: %v", err)
	}
	if run.Status != schedulerpkg.RunStatusPending {
		t.Fatalf("手动触发应写入待执行记录，实际 %d", run.Status)
	}

	stored := new(schedulerpkg.Run)
	if err := db.Where("id = ?", run.ID).First(stored).Error; err != nil {
		t.Fatalf("读取执行记录失败: %v", err)
	}
	if stored.JobID != builtin.ID || stored.TriggerType != int8(schedulerContract.TriggerManual) {
		t.Fatalf("执行记录字段不正确: %+v", stored)
	}

	if _, err := svc.TriggerJob(context.Background(), builtin.ID, &TriggerJobReq{Payload: "not-json"}); err == nil {
		t.Fatal("非法 JSON 参数应被拒绝")
	}
}
