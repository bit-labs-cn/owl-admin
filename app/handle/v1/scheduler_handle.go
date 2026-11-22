package v1

import (
	"bit-labs.cn/owl-admin/app/service"
	"bit-labs.cn/owl/provider/router"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

var _ router.Handler = (*SchedulerHandle)(nil)

// SchedulerHandle 定时任务管理端接口。
type SchedulerHandle struct {
	svc *service.SchedulerService
}

func NewSchedulerHandle(svc *service.SchedulerService) *SchedulerHandle {
	return &SchedulerHandle{svc: svc}
}

func (i *SchedulerHandle) ModuleName() (string, string) {
	return "scheduler", "定时任务"
}

// @Summary		定时任务列表
// @Description	分页查询内置与自定义定时任务，返回排期与最近执行结果
// @Tags			定时任务
// @Produce		json
// @Param			page		query		int				false	"页码"
// @Param			pageSize	query		int				false	"每页数量"
// @Param			name		query		string			false	"任务名称（模糊）"
// @Param			code		query		string			false	"任务编码（模糊）"
// @Param			ownerApp	query		string			false	"所属子系统"
// @Param			source		query		int				false	"来源：1内置、2自定义"
// @Param			status		query		string			false	"状态，逗号分隔：1启用、2停用、3已下线"
// @Success		200			{object}	router.PageResp	"操作成功"
// @Failure		500			{object}	router.Resp		"服务器内部错误"
// @Router			/api/v1/scheduler/jobs [GET]
func (i *SchedulerHandle) RetrieveJobs(ctx *gin.Context) {
	var req service.RetrieveJobReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		router.BadRequest(ctx, "参数绑定失败")
		return
	}
	count, list, err := i.svc.RetrieveJobs(ctx.Request.Context(), &req)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.PageSuccess(ctx, int(count), req.Page, req.PageSize, list)
}

// @Summary		定时任务详情
// @Description	返回任务完整配置，包含脚本内容；仅超管可读
// @Tags			定时任务
// @Produce		json
// @Param			id	path		int				true	"任务ID"
// @Success		200	{object}	router.Resp		"操作成功"
// @Failure		500	{object}	router.Resp		"服务器内部错误"
// @Router			/api/v1/scheduler/jobs/{id} [GET]
func (i *SchedulerHandle) JobDetail(ctx *gin.Context) {
	detail, err := i.svc.JobDetail(ctx.Request.Context(), cast.ToUint(ctx.Param("id")))
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, detail)
}

// @Summary		新建自定义定时任务
// @Description	创建一个 HTTP 或 Shell 定时任务；仅超管可操作
// @Tags			定时任务
// @Accept			json
// @Produce		json
// @Param			request	body		service.CreateJobReq	true	"任务配置"
// @Success		200		{object}	router.Resp				"操作成功"
// @Failure		400		{object}	router.Resp				"参数错误"
// @Failure		500		{object}	router.Resp				"服务器内部错误"
// @Router			/api/v1/scheduler/jobs [POST]
func (i *SchedulerHandle) CreateJob(ctx *gin.Context) {
	var req service.CreateJobReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	job, err := i.svc.CreateJob(ctx.Request.Context(), &req)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, job)
}

// @Summary		修改自定义定时任务
// @Description	修改自定义任务的 Cron、脚本与重试策略；内置任务不允许修改
// @Tags			定时任务
// @Accept			json
// @Produce		json
// @Param			id		path		int						true	"任务ID"
// @Param			request	body		service.UpdateJobReq	true	"任务配置"
// @Success		200		{object}	router.Resp				"操作成功"
// @Failure		400		{object}	router.Resp				"参数错误"
// @Failure		500		{object}	router.Resp				"服务器内部错误"
// @Router			/api/v1/scheduler/jobs/{id} [PUT]
func (i *SchedulerHandle) UpdateJob(ctx *gin.Context) {
	var req service.UpdateJobReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	req.ID = cast.ToUint(ctx.Param("id"))
	job, err := i.svc.UpdateJob(ctx.Request.Context(), &req)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, job)
}

// @Summary		删除自定义定时任务
// @Description	删除自定义任务，历史执行记录保留；内置任务不允许删除
// @Tags			定时任务
// @Produce		json
// @Param			id	path		int				true	"任务ID"
// @Success		200	{object}	router.Resp		"操作成功"
// @Failure		500	{object}	router.Resp		"服务器内部错误"
// @Router			/api/v1/scheduler/jobs/{id} [DELETE]
func (i *SchedulerHandle) DeleteJob(ctx *gin.Context) {
	if err := i.svc.DeleteJob(ctx.Request.Context(), cast.ToUint(ctx.Param("id"))); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// @Summary		启停定时任务
// @Description	启用或停用任务；内置任务允许启停，停用后不再自动排期
// @Tags			定时任务
// @Accept			json
// @Produce		json
// @Param			id		path		int						true	"任务ID"
// @Param			request	body		service.ChangeStatusReq	true	"状态"
// @Success		200		{object}	router.Resp				"操作成功"
// @Failure		400		{object}	router.Resp				"参数错误"
// @Failure		500		{object}	router.Resp				"服务器内部错误"
// @Router			/api/v1/scheduler/jobs/{id}/status [PUT]
func (i *SchedulerHandle) ChangeStatus(ctx *gin.Context) {
	var req service.ChangeStatusReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	if err := i.svc.ChangeStatus(ctx.Request.Context(), cast.ToUint(ctx.Param("id")), &req); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// @Summary		手动触发定时任务
// @Description	立即生成一条待执行记录，由调度引擎在秒级内执行
// @Tags			定时任务
// @Accept			json
// @Produce		json
// @Param			id		path		int						true	"任务ID"
// @Param			request	body		service.TriggerJobReq	false	"执行参数"
// @Success		200		{object}	router.Resp				"操作成功"
// @Failure		400		{object}	router.Resp				"参数错误"
// @Failure		500		{object}	router.Resp				"服务器内部错误"
// @Router			/api/v1/scheduler/jobs/{id}/trigger [POST]
func (i *SchedulerHandle) TriggerJob(ctx *gin.Context) {
	var req service.TriggerJobReq
	if ctx.Request.ContentLength > 0 {
		if err := ctx.ShouldBindJSON(&req); err != nil {
			router.Fail(ctx, err)
			return
		}
	}
	run, err := i.svc.TriggerJob(ctx.Request.Context(), cast.ToUint(ctx.Param("id")), &req)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.SuccessWithMsg(ctx, "已提交执行，稍后刷新执行记录查看结果", run)
}

// @Summary		可调用方法目录
// @Description	返回子系统登记的可调用方法及参数说明，供自定义任务编排；接口仅超管可访问
// @Tags			定时任务
// @Produce		json
// @Success		200	{object}	router.Resp	"操作成功"
// @Failure		500	{object}	router.Resp	"服务器内部错误"
// @Router			/api/v1/scheduler/actions [GET]
func (i *SchedulerHandle) RetrieveActions(ctx *gin.Context) {
	router.Success(ctx, i.svc.RetrieveActions())
}

// @Summary		可调用方法的动态候选
// @Description	按动作与当前已填参数返回下拉候选，例如先选开关再列它的命令
// @Tags			定时任务
// @Accept			json
// @Produce		json
// @Param			code	path		string						true	"动作编码"
// @Param			request	body		service.ActionOptionsReq	true	"已填参数"
// @Success		200		{object}	router.Resp					"操作成功"
// @Failure		400		{object}	router.Resp					"参数错误"
// @Failure		500		{object}	router.Resp					"服务器内部错误"
// @Router			/api/v1/scheduler/actions/{code}/options [POST]
func (i *SchedulerHandle) ActionOptions(ctx *gin.Context) {
	var req service.ActionOptionsReq
	if ctx.Request.ContentLength > 0 {
		if err := ctx.ShouldBindJSON(&req); err != nil {
			router.Fail(ctx, err)
			return
		}
	}
	resp, err := i.svc.ActionOptions(ctx.Request.Context(), ctx.Param("code"), &req)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, resp)
}

// @Summary		执行记录列表
// @Description	分页查询定时任务的执行记录
// @Tags			定时任务
// @Produce		json
// @Param			page		query		int				false	"页码"
// @Param			pageSize	query		int				false	"每页数量"
// @Param			jobId		query		int				false	"任务ID"
// @Param			jobCode		query		string			false	"任务编码"
// @Param			status		query		string			false	"状态，逗号分隔"
// @Param			triggerType	query		int				false	"触发方式：1定时、2手动、3重试"
// @Param			plannedAt	query		string			false	"计划时间区间，start,end（Unix 秒）"
// @Success		200			{object}	router.PageResp	"操作成功"
// @Failure		500			{object}	router.Resp		"服务器内部错误"
// @Router			/api/v1/scheduler/runs [GET]
func (i *SchedulerHandle) RetrieveRuns(ctx *gin.Context) {
	var req service.RetrieveRunReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		router.BadRequest(ctx, "参数绑定失败")
		return
	}
	count, list, err := i.svc.RetrieveRuns(ctx.Request.Context(), &req)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.PageSuccess(ctx, int(count), req.Page, req.PageSize, list)
}

// @Summary		执行记录详情
// @Description	返回执行输出、错误与堆栈
// @Tags			定时任务
// @Produce		json
// @Param			id	path		int				true	"执行记录ID"
// @Success		200	{object}	router.Resp		"操作成功"
// @Failure		500	{object}	router.Resp		"服务器内部错误"
// @Router			/api/v1/scheduler/runs/{id} [GET]
func (i *SchedulerHandle) RunDetail(ctx *gin.Context) {
	detail, err := i.svc.RunDetail(ctx.Request.Context(), cast.ToUint(ctx.Param("id")))
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, detail)
}

// @Summary		重试执行记录
// @Description	对失败、超时或已取消的记录发起一次新的执行
// @Tags			定时任务
// @Produce		json
// @Param			id	path		int				true	"执行记录ID"
// @Success		200	{object}	router.Resp		"操作成功"
// @Failure		500	{object}	router.Resp		"服务器内部错误"
// @Router			/api/v1/scheduler/runs/{id}/retry [POST]
func (i *SchedulerHandle) RetryRun(ctx *gin.Context) {
	run, err := i.svc.RetryRun(ctx.Request.Context(), cast.ToUint(ctx.Param("id")))
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.SuccessWithMsg(ctx, "已提交重试", run)
}

// @Summary		取消执行记录
// @Description	取消一条待执行记录；本实例正在执行的记录会中断任务
// @Tags			定时任务
// @Produce		json
// @Param			id	path		int				true	"执行记录ID"
// @Success		200	{object}	router.Resp		"操作成功"
// @Failure		500	{object}	router.Resp		"服务器内部错误"
// @Router			/api/v1/scheduler/runs/{id}/cancel [POST]
func (i *SchedulerHandle) CancelRun(ctx *gin.Context) {
	if err := i.svc.CancelRun(ctx.Request.Context(), cast.ToUint(ctx.Param("id"))); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// @Summary		清理执行记录
// @Description	删除指定天数之前的终态执行记录，保留未完成与失败记录之外的历史
// @Tags			定时任务
// @Accept			json
// @Produce		json
// @Param			request	body		service.CleanupRunsReq	true	"清理请求"
// @Success		200		{object}	router.Resp				"操作成功"
// @Failure		400		{object}	router.Resp				"参数错误"
// @Failure		500		{object}	router.Resp				"服务器内部错误"
// @Router			/api/v1/scheduler/runs/cleanup [POST]
func (i *SchedulerHandle) CleanupRuns(ctx *gin.Context) {
	var req service.CleanupRunsReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	count, err := i.svc.CleanupRuns(ctx.Request.Context(), &req)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.SuccessWithMsg(ctx, "清理完成", gin.H{"deleted": count})
}

// @Summary		预览 Cron 下次执行时间
// @Description	按给定 Cron 与时区返回接下来的执行时间，便于保存前确认
// @Tags			定时任务
// @Accept			json
// @Produce		json
// @Param			request	body		service.CronPreviewReq	true	"预览请求"
// @Success		200		{object}	router.Resp				"操作成功"
// @Failure		400		{object}	router.Resp				"参数错误"
// @Failure		500		{object}	router.Resp				"服务器内部错误"
// @Router			/api/v1/scheduler/cron/preview [POST]
func (i *SchedulerHandle) PreviewCron(ctx *gin.Context) {
	var req service.CronPreviewReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	resp, err := i.svc.PreviewCron(&req)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, resp)
}

// @Summary		定时任务汇总
// @Description	任务数量、执行中与今日成功失败数量，以及最近失败记录
// @Tags			定时任务
// @Produce		json
// @Success		200	{object}	router.Resp	"操作成功"
// @Failure		500	{object}	router.Resp	"服务器内部错误"
// @Router			/api/v1/scheduler/stats [GET]
func (i *SchedulerHandle) Stats(ctx *gin.Context) {
	stats, err := i.svc.Stats(ctx.Request.Context())
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, stats)
}
