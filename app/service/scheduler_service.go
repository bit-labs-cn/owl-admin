package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"bit-labs.cn/owl-admin/app/repository"
	schedulerContract "bit-labs.cn/owl/contract/scheduler"
	"bit-labs.cn/owl/provider/db"
	"bit-labs.cn/owl/provider/redis"
	"bit-labs.cn/owl/provider/router"
	schedulerpkg "bit-labs.cn/owl/provider/scheduler"
	"bit-labs.cn/owl/utils"
	validatorv10 "github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

const (
	// schedulerMaxPageSize 是列表单页最大条数，避免一次拉太多执行记录。
	schedulerMaxPageSize = 100
	// schedulerRecentFailureLimit 是工作台「最近失败」条数。
	schedulerRecentFailureLimit = 10
)

// SchedulerService 是定时任务管理端服务。
//
// 任务的排期、抢占与执行由框架调度引擎负责；这里负责管理端的查询、自定义任务维护，
// 以及把引擎返回的错误翻译成可展示的业务错误。
type SchedulerService struct {
	repo     repository.SchedulerRepositoryInterface
	locker   redis.LockerFactory
	validate *validatorv10.Validate
	manager  *schedulerpkg.Manager
}

func NewSchedulerService(
	repo repository.SchedulerRepositoryInterface,
	locker redis.LockerFactory,
	validate *validatorv10.Validate,
	manager *schedulerpkg.Manager,
) *SchedulerService {
	return &SchedulerService{
		repo:     repo,
		locker:   locker,
		validate: validate,
		manager:  manager,
	}
}

// JobDTO 是任务列表项。内置任务的脚本配置不会出现在列表里。
type JobDTO struct {
	schedulerpkg.Job
	SourceText     string `json:"sourceText"`
	StatusText     string `json:"statusText"`
	ScriptTypeText string `json:"scriptTypeText"`
	OverlapText    string `json:"overlapText"`
	LastStatusText string `json:"lastStatusText,omitempty"`
}

// JobDetailDTO 是任务详情，额外返回脚本配置（可能包含命令与请求头，仅超管可读）。
type JobDetailDTO struct {
	JobDTO
	ScriptConfig string `json:"scriptConfig"`
}

// RunDTO 是执行记录列表项。
type RunDTO struct {
	schedulerpkg.Run
	StatusText      string `json:"statusText"`
	TriggerTypeText string `json:"triggerTypeText"`
	AttemptText     string `json:"attemptText"`
}

// RetrieveJobReq 是任务列表的查询条件。
type RetrieveJobReq struct {
	router.PageReq
	NameLike string `json:"name" form:"name"`
	CodeLike string `json:"code" form:"code"`
	OwnerApp string `json:"ownerApp" form:"ownerApp"`
	Source   int8   `json:"source" form:"source"`
	StatusIn string `json:"status" form:"status"`
}

// RetrieveRunReq 是执行记录的查询条件。
type RetrieveRunReq struct {
	router.PageReq
	JobID            uint   `json:"jobId" form:"jobId"`
	JobCode          string `json:"jobCode" form:"jobCode"`
	StatusIn         string `json:"status" form:"status"`
	TriggerType      int8   `json:"triggerType" form:"triggerType"`
	PlannedAtBetween string `json:"plannedAt" form:"plannedAt"`
}

// CreateJobReq 是新建自定义任务的入参。
//
// ScriptType 为 HTTP / Shell 时用 ScriptConfig；为「调用子系统方法」时用 ActionCode + ActionParams。
type CreateJobReq struct {
	Name                string            `json:"name" validate:"required,max=64" label:"任务名称"`
	Description         string            `json:"description" validate:"max=512" label:"任务说明"`
	Cron                string            `json:"cron" validate:"required,max=64" label:"Cron 表达式"`
	Timezone            string            `json:"timezone" validate:"max=64" label:"时区"`
	ScriptType          int8              `json:"scriptType" validate:"required" label:"执行方式"`
	ScriptConfig        string            `json:"scriptConfig" validate:"max=65535" label:"脚本配置"`
	ActionCode          string            `json:"actionCode" validate:"max=160" label:"调用的方法"`
	ActionParams        map[string]string `json:"actionParams" validate:"max=32" label:"方法参数"`
	TimeoutSeconds      int               `json:"timeoutSeconds" validate:"required,min=1,max=86400" label:"超时时间"`
	MaxAttempts         int               `json:"maxAttempts" validate:"required,min=1,max=10" label:"最大尝试次数"`
	RetryBackoffSeconds int               `json:"retryBackoffSeconds" validate:"min=0,max=86400" label:"重试间隔"`
	OverlapPolicy       int8              `json:"overlapPolicy" validate:"required" label:"并发策略"`
}

// UpdateJobReq 是修改自定义任务的入参。
type UpdateJobReq struct {
	ID uint `json:"id" validate:"required" label:"任务ID"`
	CreateJobReq
}

// ChangeStatusReq 是启停任务的入参。
type ChangeStatusReq struct {
	Status int8 `json:"status" validate:"required" label:"状态"`
}

// TriggerJobReq 是手动触发的入参。
type TriggerJobReq struct {
	Payload string `json:"payload" validate:"max=8192" label:"执行参数"`
}

// CleanupRunsReq 是清理执行记录的入参。
type CleanupRunsReq struct {
	OlderThanDays int `json:"olderThanDays" validate:"required,min=1,max=3650" label:"保留天数"`
}

// CronPreviewReq 是 Cron 预览的入参。
type CronPreviewReq struct {
	Cron     string `json:"cron" validate:"required,max=64" label:"Cron 表达式"`
	Timezone string `json:"timezone" validate:"max=64" label:"时区"`
	Count    int    `json:"count" validate:"omitempty,min=1,max=20" label:"预览次数"`
}

// CronPreviewResp 是 Cron 预览结果。
type CronPreviewResp struct {
	Times []string `json:"times"`
}

// ActionDTO 是管理端可见的可调用方法。
type ActionDTO struct {
	Code        string                    `json:"code"`
	Key         string                    `json:"key"`
	Name        string                    `json:"name"`
	Description string                    `json:"description"`
	Group       string                    `json:"group"`
	AppName     string                    `json:"appName"`
	Params      []schedulerContract.Param `json:"params"`
}

// ActionOptionsReq 是拉取动态候选的入参。
type ActionOptionsReq struct {
	Params map[string]string `json:"params" validate:"max=32" label:"方法参数"`
}

// ActionOptionsResp 是按参数名分组的候选。
type ActionOptionsResp struct {
	Options map[string][]schedulerContract.ParamOption `json:"options"`
}

// SchedulerStats 是定时任务页面的汇总数据。
type SchedulerStats struct {
	JobTotal       int64    `json:"jobTotal"`
	JobEnabled     int64    `json:"jobEnabled"`
	JobDisabled    int64    `json:"jobDisabled"`
	JobRetired     int64    `json:"jobRetired"`
	Pending        int64    `json:"pending"`
	Running        int64    `json:"running"`
	SuccessToday   int64    `json:"successToday"`
	FailedToday    int64    `json:"failedToday"`
	RecentFailures []RunDTO `json:"recentFailures"`
}

func normalizeSchedulerPage(req *router.PageReq) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	if req.PageSize > schedulerMaxPageSize {
		req.PageSize = schedulerMaxPageSize
	}
}

func toJobDTO(job *schedulerpkg.Job) JobDTO {
	dto := JobDTO{
		Job:            *job,
		SourceText:     schedulerpkg.JobSourceText(job.Source),
		StatusText:     schedulerpkg.JobStatusText(job.Status),
		ScriptTypeText: schedulerpkg.ScriptTypeText(job.ScriptType),
		OverlapText:    schedulerContract.OverlapPolicy(job.OverlapPolicy).String(),
	}
	if job.LastStatus != nil {
		dto.LastStatusText = schedulerpkg.RunStatusText(*job.LastStatus)
	}
	return dto
}

func toJobDetailDTO(job *schedulerpkg.Job) JobDetailDTO {
	return JobDetailDTO{
		JobDTO:       toJobDTO(job),
		ScriptConfig: job.ScriptConfig,
	}
}

func toRunDTO(run *schedulerpkg.Run) RunDTO {
	maxAttempts := run.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 1
	}
	return RunDTO{
		Run:             *run,
		StatusText:      schedulerpkg.RunStatusText(run.Status),
		TriggerTypeText: schedulerpkg.TriggerTypeText(run.TriggerType),
		AttemptText:     fmt.Sprintf("%d/%d", run.Attempt, maxAttempts),
	}
}

// RetrieveJobs 分页查询任务。
func (s *SchedulerService) RetrieveJobs(ctx context.Context, req *RetrieveJobReq) (int64, []JobDTO, error) {
	normalizeSchedulerPage(&req.PageReq)

	count, rows, err := s.repo.WithContext(ctx).RetrieveJobs(req.Page, req.PageSize, func(tx *gorm.DB) {
		db.AppendWhereFromStruct(tx, req)
	})
	if err != nil {
		return 0, nil, err
	}
	list := make([]JobDTO, 0, len(rows))
	for i := range rows {
		list = append(list, toJobDTO(&rows[i]))
	}
	return count, list, nil
}

// JobDetail 返回任务详情，包含脚本配置。
func (s *SchedulerService) JobDetail(ctx context.Context, id uint) (*JobDetailDTO, error) {
	job, err := s.findJob(ctx, id)
	if err != nil {
		return nil, err
	}
	dto := toJobDetailDTO(job)
	return &dto, nil
}

// CreateJob 创建一条自定义任务。
func (s *SchedulerService) CreateJob(ctx context.Context, req *CreateJobReq) (*JobDTO, error) {
	if err := s.validate.Struct(req); err != nil {
		return nil, err
	}
	next, err := s.validateJobDefinition(req.Cron, req.Timezone)
	if err != nil {
		return nil, err
	}
	plan, err := s.resolveScript(ctx, req)
	if err != nil {
		return nil, err
	}

	lock := s.locker.New()
	if err := lock.Lock("scheduler:job:create"); err != nil {
		return nil, err
	}
	defer lock.Unlock()

	id, err := utils.SnowFlakeNextID()
	if err != nil {
		return nil, err
	}

	job := &schedulerpkg.Job{
		ID:                  uint(id),
		Code:                fmt.Sprintf("custom:%d", id),
		Name:                strings.TrimSpace(req.Name),
		Description:         strings.TrimSpace(req.Description),
		Source:              schedulerpkg.JobSourceCustom,
		Cron:                strings.TrimSpace(req.Cron),
		Timezone:            strings.TrimSpace(req.Timezone),
		ScriptType:          plan.ScriptType,
		ScriptConfig:        plan.ScriptConfig,
		ActionCode:          plan.ActionCode,
		ActionName:          plan.ActionName,
		ScriptSummary:       plan.ScriptSummary,
		TimeoutSeconds:      req.TimeoutSeconds,
		MaxAttempts:         req.MaxAttempts,
		RetryBackoffSeconds: req.RetryBackoffSeconds,
		OverlapPolicy:       req.OverlapPolicy,
		Status:              schedulerpkg.JobStatusEnabled,
		NextRunAt:           next,
	}
	if err := s.repo.WithContext(ctx).CreateJob(job); err != nil {
		return nil, err
	}
	dto := toJobDTO(job)
	return &dto, nil
}

// UpdateJob 修改一条自定义任务。内置任务的排期与脚本由代码决定，不允许修改。
func (s *SchedulerService) UpdateJob(ctx context.Context, req *UpdateJobReq) (*JobDTO, error) {
	if err := s.validate.Struct(req); err != nil {
		return nil, err
	}

	job, err := s.findJob(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if job.IsBuiltin() {
		return nil, SchedulerBuiltinImmutable()
	}

	next, err := s.validateJobDefinition(req.Cron, req.Timezone)
	if err != nil {
		return nil, err
	}
	plan, err := s.resolveScript(ctx, &req.CreateJobReq)
	if err != nil {
		return nil, err
	}

	lock := s.locker.New()
	if err := lock.Lock(fmt.Sprintf("scheduler:job:%d", req.ID)); err != nil {
		return nil, err
	}
	defer lock.Unlock()

	job.Name = strings.TrimSpace(req.Name)
	job.Description = strings.TrimSpace(req.Description)
	job.Cron = strings.TrimSpace(req.Cron)
	job.Timezone = strings.TrimSpace(req.Timezone)
	job.ScriptType = plan.ScriptType
	job.ScriptConfig = plan.ScriptConfig
	job.ActionCode = plan.ActionCode
	job.ActionName = plan.ActionName
	job.ScriptSummary = plan.ScriptSummary
	job.TimeoutSeconds = req.TimeoutSeconds
	job.MaxAttempts = req.MaxAttempts
	job.RetryBackoffSeconds = req.RetryBackoffSeconds
	job.OverlapPolicy = req.OverlapPolicy
	job.NextRunAt = next
	if err := s.repo.WithContext(ctx).SaveJob(job); err != nil {
		return nil, err
	}
	dto := toJobDTO(job)
	return &dto, nil
}

// DeleteJob 删除一条自定义任务。执行记录保留，作为历史快照。
func (s *SchedulerService) DeleteJob(ctx context.Context, id uint) error {
	job, err := s.findJob(ctx, id)
	if err != nil {
		return err
	}
	if job.IsBuiltin() {
		return SchedulerBuiltinImmutable()
	}
	active, err := s.repo.WithContext(ctx).CountActiveRuns(job.ID)
	if err != nil {
		return err
	}
	if active > 0 {
		return SchedulerJobRunning()
	}

	lock := s.locker.New()
	if err := lock.Lock(fmt.Sprintf("scheduler:job:%d", id)); err != nil {
		return err
	}
	defer lock.Unlock()
	return s.repo.WithContext(ctx).DeleteJob(id)
}

// ChangeStatus 启用或停用任务。内置任务同样允许启停。
func (s *SchedulerService) ChangeStatus(ctx context.Context, id uint, req *ChangeStatusReq) error {
	if err := s.validate.Struct(req); err != nil {
		return err
	}
	if req.Status != schedulerpkg.JobStatusEnabled && req.Status != schedulerpkg.JobStatusDisabled {
		return SchedulerJobNotAllowed("状态只能是启用或停用")
	}

	job, err := s.findJob(ctx, id)
	if err != nil {
		return err
	}
	if job.Status == schedulerpkg.JobStatusRetired {
		return SchedulerJobNotAllowed("任务已下线，不能修改状态")
	}

	updates := map[string]any{
		"status":     req.Status,
		"updated_at": time.Now(),
	}
	if req.Status == schedulerpkg.JobStatusEnabled {
		next, nextErr := s.manager.NextRunFor(job.Cron, job.Timezone)
		if nextErr != nil {
			return SchedulerCronInvalid(nextErr.Error())
		}
		updates["next_run_at"] = next
		updates["last_error"] = ""
	} else {
		// 停用后清掉排期，扫描循环就不会再命中它。
		updates["next_run_at"] = nil
	}
	return s.repo.WithContext(ctx).UpdateJobFields(job.ID, updates)
}

// TriggerJob 手动触发任务，写入一条待执行记录。
//
// 谁能触发由路由权限决定（需授权 + 按钮权限），这里不再按角色或动作访问级别过滤。
func (s *SchedulerService) TriggerJob(ctx context.Context, id uint, req *TriggerJobReq) (*RunDTO, error) {
	if err := s.validate.Struct(req); err != nil {
		return nil, err
	}
	payload := strings.TrimSpace(req.Payload)
	if payload != "" && !json.Valid([]byte(payload)) {
		return nil, SchedulerPayloadInvalid("执行参数必须是合法的 JSON")
	}

	job, err := s.findJob(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := s.ensureActionRegistered(job); err != nil {
		return nil, err
	}
	run, err := s.manager.Trigger(ctx, job.Code, payload)
	if err != nil {
		return nil, mapSchedulerError(err)
	}
	if run == nil {
		return nil, SchedulerJobRunning()
	}
	dto := toRunDTO(run)
	return &dto, nil
}

// ensureActionRegistered 在触发前确认任务绑定的方法仍然登记在册。
//
// 这只看方法是否存在，不看是谁在触发：子系统未部署或方法被移除时，
// 与其写入一条注定失败的记录，不如直接提示。「调用子系统方法」以外的任务不检查。
func (s *SchedulerService) ensureActionRegistered(job *schedulerpkg.Job) error {
	if job.ScriptType != schedulerpkg.ScriptTypeAction {
		return nil
	}
	if _, ok := s.manager.FindAction(job.ActionCode); !ok {
		return SchedulerScriptInvalid("任务绑定的方法已不可用：请确认所属子系统已加载")
	}
	return nil
}

// RetrieveActions 返回可调用方法目录。
//
// 接口只对超管开放（AccessSuperAdmin）；动作不再自带访问级别，
// 登记了什么就列出什么，界面按 AppName 分组直接展示。
func (s *SchedulerService) RetrieveActions() []ActionDTO {
	registrations := s.manager.Actions()
	out := make([]ActionDTO, 0, len(registrations))
	for _, item := range registrations {
		out = append(out, ActionDTO{
			Code:        item.Code,
			Key:         item.Action.Key,
			Name:        item.Action.Name,
			Description: item.Action.Description,
			Group:       item.Action.Group,
			AppName:     item.AppName,
			Params:      item.Action.Params,
		})
	}
	return out
}

// ActionOptions 拉取动作的动态参数候选。
func (s *SchedulerService) ActionOptions(ctx context.Context, code string, req *ActionOptionsReq) (*ActionOptionsResp, error) {
	if err := s.validate.Struct(req); err != nil {
		return nil, err
	}
	reg, ok := s.manager.FindAction(strings.TrimSpace(code))
	if !ok {
		return nil, SchedulerScriptInvalid("方法未注册：请确认所属子系统已加载")
	}
	options, err := s.manager.ActionOptions(ctx, reg.Code, req.Params)
	if err != nil {
		return nil, SchedulerScriptInvalid(err.Error())
	}
	return &ActionOptionsResp{Options: options}, nil
}

// RetrieveRuns 分页查询执行记录。
func (s *SchedulerService) RetrieveRuns(ctx context.Context, req *RetrieveRunReq) (int64, []RunDTO, error) {
	normalizeSchedulerPage(&req.PageReq)

	count, rows, err := s.repo.WithContext(ctx).RetrieveRuns(req.Page, req.PageSize, func(tx *gorm.DB) {
		db.AppendWhereFromStruct(tx, req)
	})
	if err != nil {
		return 0, nil, err
	}
	list := make([]RunDTO, 0, len(rows))
	for i := range rows {
		list = append(list, toRunDTO(&rows[i]))
	}
	return count, list, nil
}

// RunDetail 返回一条执行记录，包含输出与错误。
func (s *SchedulerService) RunDetail(ctx context.Context, id uint) (*RunDTO, error) {
	run, err := s.repo.WithContext(ctx).FindRun(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, SchedulerRunNotFound()
	}
	if err != nil {
		return nil, err
	}
	dto := toRunDTO(run)
	return &dto, nil
}

// RetryRun 重试一条失败、超时或已取消的记录。
func (s *SchedulerService) RetryRun(ctx context.Context, id uint) (*RunDTO, error) {
	source, err := s.repo.WithContext(ctx).FindRun(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, SchedulerRunNotFound()
	}
	if err != nil {
		return nil, err
	}
	job, err := s.findJob(ctx, source.JobID)
	if err != nil {
		return nil, err
	}
	if err := s.ensureActionRegistered(job); err != nil {
		return nil, err
	}

	run, err := s.manager.Retry(ctx, id)
	if err != nil {
		return nil, mapSchedulerError(err)
	}
	if run == nil {
		return nil, SchedulerJobRunning()
	}
	dto := toRunDTO(run)
	return &dto, nil
}

// CancelRun 取消一条待执行或本实例正在执行的记录。
func (s *SchedulerService) CancelRun(ctx context.Context, id uint) error {
	if err := s.manager.Cancel(ctx, id); err != nil {
		return mapSchedulerError(err)
	}
	return nil
}

// CleanupRuns 清理指定天数之前的终态执行记录。
func (s *SchedulerService) CleanupRuns(ctx context.Context, req *CleanupRunsReq) (int64, error) {
	if err := s.validate.Struct(req); err != nil {
		return 0, err
	}
	cutoff := time.Now().AddDate(0, 0, -req.OlderThanDays)
	return s.repo.WithContext(ctx).DeleteTerminalRunsBefore(cutoff)
}

// PreviewCron 预览接下来几次执行时间。
func (s *SchedulerService) PreviewCron(req *CronPreviewReq) (*CronPreviewResp, error) {
	if err := s.validate.Struct(req); err != nil {
		return nil, err
	}
	if err := schedulerpkg.ValidateCron(req.Cron, req.Timezone); err != nil {
		return nil, SchedulerCronInvalid(err.Error())
	}
	times, err := s.manager.PreviewCron(req.Cron, req.Timezone, req.Count)
	if err != nil {
		return nil, SchedulerCronInvalid(err.Error())
	}
	resp := &CronPreviewResp{Times: make([]string, 0, len(times))}
	for _, item := range times {
		resp.Times = append(resp.Times, item.Format("2006-01-02 15:04:05"))
	}
	return resp, nil
}

// Stats 返回任务与执行记录的汇总数据。
func (s *SchedulerService) Stats(ctx context.Context) (*SchedulerStats, error) {
	jobCounts, err := s.repo.WithContext(ctx).CountJobsByStatus()
	if err != nil {
		return nil, err
	}
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	runCounts, err := s.repo.WithContext(ctx).CountRunsByStatusSince(startOfDay)
	if err != nil {
		return nil, err
	}
	recent, err := s.repo.WithContext(ctx).RecentFailedRuns(schedulerRecentFailureLimit)
	if err != nil {
		return nil, err
	}

	stats := &SchedulerStats{
		JobEnabled:     jobCounts[schedulerpkg.JobStatusEnabled],
		JobDisabled:    jobCounts[schedulerpkg.JobStatusDisabled],
		JobRetired:     jobCounts[schedulerpkg.JobStatusRetired],
		Pending:        runCounts[schedulerpkg.RunStatusPending],
		Running:        runCounts[schedulerpkg.RunStatusRunning],
		SuccessToday:   runCounts[schedulerpkg.RunStatusSuccess],
		FailedToday:    runCounts[schedulerpkg.RunStatusFailed] + runCounts[schedulerpkg.RunStatusTimeout],
		RecentFailures: make([]RunDTO, 0, len(recent)),
	}
	for _, count := range jobCounts {
		stats.JobTotal += count
	}
	for i := range recent {
		stats.RecentFailures = append(stats.RecentFailures, toRunDTO(&recent[i]))
	}
	return stats, nil
}

// jobScript 是「执行方式」解析后的落库字段。
type jobScript struct {
	ScriptType    int8
	ScriptConfig  string
	ActionCode    string
	ActionName    string
	ScriptSummary string
}

// resolveScript 按执行方式校验脚本并生成落库字段。
//
// 调用方法时先拉一次动态候选，用它的 label 生成可读摘要，避免列表出现裸 ID。
func (s *SchedulerService) resolveScript(ctx context.Context, req *CreateJobReq) (*jobScript, error) {
	switch schedulerContract.ScriptType(req.ScriptType) {
	case schedulerContract.ScriptTypeHTTP:
		if err := schedulerpkg.ValidateHTTPConfig(req.ScriptConfig); err != nil {
			return nil, SchedulerScriptInvalid(err.Error())
		}
		return &jobScript{ScriptType: req.ScriptType, ScriptConfig: strings.TrimSpace(req.ScriptConfig)}, nil

	case schedulerContract.ScriptTypeShell:
		if !s.manager.Config().ShellEnabled {
			return nil, SchedulerScriptInvalid("Shell 任务已在服务器配置中关闭，无法创建")
		}
		if err := schedulerpkg.ValidateShellConfig(req.ScriptConfig); err != nil {
			return nil, SchedulerScriptInvalid(err.Error())
		}
		return &jobScript{ScriptType: req.ScriptType, ScriptConfig: strings.TrimSpace(req.ScriptConfig)}, nil

	case schedulerContract.ScriptType(schedulerpkg.ScriptTypeAction):
		code := strings.TrimSpace(req.ActionCode)
		if code == "" {
			return nil, SchedulerScriptInvalid("请选择要调用的方法")
		}
		reg, ok := s.manager.FindAction(code)
		if !ok {
			return nil, SchedulerScriptInvalid("方法未注册：请确认所属子系统已加载，或刷新页面后重试")
		}
		optionSets, err := s.manager.ActionOptions(ctx, code, req.ActionParams)
		if err != nil {
			return nil, SchedulerScriptInvalid(err.Error())
		}
		config, summary, err := schedulerpkg.BuildActionConfig(code, reg.Action, req.ActionParams, optionSets)
		if err != nil {
			return nil, SchedulerScriptInvalid(err.Error())
		}
		return &jobScript{
			ScriptType:    schedulerpkg.ScriptTypeAction,
			ScriptConfig:  config,
			ActionCode:    reg.Code,
			ActionName:    reg.Action.Name,
			ScriptSummary: summary,
		}, nil

	default:
		return nil, SchedulerScriptInvalid("执行方式只支持 HTTP 请求、Shell 命令或调用子系统方法")
	}
}

// findJob 读取任务并把「不存在」翻译成业务错误。
func (s *SchedulerService) findJob(ctx context.Context, id uint) (*schedulerpkg.Job, error) {
	job, err := s.repo.WithContext(ctx).FindJob(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, SchedulerJobNotFound()
	}
	if err != nil {
		return nil, err
	}
	return job, nil
}

// validateJobDefinition 校验 Cron 并算出下次执行时间。
func (s *SchedulerService) validateJobDefinition(cronExpr, timezone string) (*time.Time, error) {
	if err := schedulerpkg.ValidateCron(cronExpr, timezone); err != nil {
		return nil, SchedulerCronInvalid(err.Error())
	}
	next, err := s.manager.NextRunFor(cronExpr, timezone)
	if err != nil {
		return nil, SchedulerCronInvalid(err.Error())
	}
	return next, nil
}

// mapSchedulerError 把调度引擎的错误翻译成管理端业务错误。
func mapSchedulerError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, schedulerpkg.ErrJobNotFound):
		return SchedulerJobNotFound()
	case errors.Is(err, schedulerpkg.ErrJobDisabled):
		return SchedulerJobNotAllowed("定时任务已停用，请先启用")
	case errors.Is(err, schedulerpkg.ErrJobRetired):
		return SchedulerJobNotAllowed("定时任务已下线，不能再执行")
	case errors.Is(err, schedulerpkg.ErrJobRunning):
		return SchedulerJobNotAllowed("任务仍在执行，本次触发已跳过")
	case errors.Is(err, schedulerpkg.ErrRunNotFound):
		return SchedulerRunNotFound()
	case errors.Is(err, schedulerpkg.ErrRunNotRetryable):
		return SchedulerRunNotRetryable()
	case errors.Is(err, schedulerpkg.ErrRunNotCancelable):
		return SchedulerRunNotCancelable()
	case errors.Is(err, schedulerpkg.ErrRunRunning):
		return SchedulerRunRunning()
	case errors.Is(err, schedulerpkg.ErrSchedulerDisabled), errors.Is(err, schedulerpkg.ErrDatabaseNotReady):
		return SchedulerEngineUnavailable()
	default:
		return err
	}
}
