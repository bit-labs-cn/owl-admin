package v1

import (
	"bit-labs.cn/owl-admin/app-notification/repository"
	"bit-labs.cn/owl-admin/app-notification/service"
	"bit-labs.cn/owl/provider/router"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// ScheduleHandle 处理定时发送任务的列表、取消、立即发送与重试。
type ScheduleHandle struct {
	svc *service.NotificationService
}

func NewScheduleHandle(svc *service.NotificationService) *ScheduleHandle {
	return &ScheduleHandle{svc: svc}
}

func (h *ScheduleHandle) ModuleName() (string, string) { return "notify-schedule", "定时发送" }

// Retrieve 分页列出定时发送任务。
func (h *ScheduleHandle) Retrieve(ctx *gin.Context) {
	page := cast.ToInt(ctx.DefaultQuery("page", "1"))
	pageSize := cast.ToInt(ctx.DefaultQuery("pageSize", "10"))
	total, list, err := h.svc.AdminSchedules(ctx.Request.Context(), repository.ScheduleQuery{
		Keyword:  ctx.Query("keyword"),
		Kind:     cast.ToInt(ctx.Query("kind")),
		Status:   cast.ToInt(ctx.Query("status")),
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.PageSuccess(ctx, int(total), page, pageSize, list)
}

// Cancel 取消一条还没发出去的定时任务。
func (h *ScheduleHandle) Cancel(ctx *gin.Context) {
	id := cast.ToUint(ctx.Param("id"))
	if id == 0 {
		router.Fail(ctx, service.Invalid("定时任务 ID 无效"))
		return
	}
	if err := h.svc.CancelSchedule(ctx.Request.Context(), id); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// SendNow 跳过计划时间立刻发送。
func (h *ScheduleHandle) SendNow(ctx *gin.Context) {
	id := cast.ToUint(ctx.Param("id"))
	if id == 0 {
		router.Fail(ctx, service.Invalid("定时任务 ID 无效"))
		return
	}
	if err := h.svc.SendScheduleNow(ctx.Request.Context(), id); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// Retry 重试一条发送失败的定时任务。
func (h *ScheduleHandle) Retry(ctx *gin.Context) {
	id := cast.ToUint(ctx.Param("id"))
	if id == 0 {
		router.Fail(ctx, service.Invalid("定时任务 ID 无效"))
		return
	}
	if err := h.svc.RetrySchedule(ctx.Request.Context(), id); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}
