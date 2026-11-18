package v1

import (
	"bit-labs.cn/owl-admin/app-notification/contract"
	"bit-labs.cn/owl-admin/app-notification/service"
	"bit-labs.cn/owl/provider/router"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// TaskHandle 处理通知任务接口。
type TaskHandle struct{ svc *service.TaskService }

func NewTaskHandle(svc *service.TaskService) *TaskHandle { return &TaskHandle{svc: svc} }
func (h *TaskHandle) ModuleName() (string, string)       { return "task", "通知任务" }

func (h *TaskHandle) Create(ctx *gin.Context) {
	var req service.CreateTaskReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	req.Origin = requestOrigin(ctx)
	row, err := h.svc.Create(ctx.Request.Context(), &req)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, row)
}

func (h *TaskHandle) Update(ctx *gin.Context) {
	var req service.CreateTaskReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	req.Origin = requestOrigin(ctx)
	if err := h.svc.Update(ctx.Request.Context(), cast.ToUint(ctx.Param("id")), &req); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

func (h *TaskHandle) Delete(ctx *gin.Context) {
	if err := h.svc.Delete(ctx.Request.Context(), cast.ToUint(ctx.Param("id"))); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

func (h *TaskHandle) Detail(ctx *gin.Context) {
	row, err := h.svc.Detail(ctx.Request.Context(), cast.ToUint(ctx.Param("id")))
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, row)
}

func (h *TaskHandle) Retrieve(ctx *gin.Context) {
	var req service.RetrieveTaskReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	total, list, err := h.svc.Retrieve(ctx.Request.Context(), &req)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.PageSuccess(ctx, int(total), req.Page, req.PageSize, list)
}

func (h *TaskHandle) Estimate(ctx *gin.Context) {
	var req contract.Audience
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	est, err := h.svc.Estimate(ctx.Request.Context(), req)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, est)
}

func (h *TaskHandle) Submit(ctx *gin.Context) {
	if err := h.svc.Submit(ctx.Request.Context(), cast.ToUint(ctx.Param("id"))); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

func (h *TaskHandle) Approve(ctx *gin.Context) {
	if err := h.svc.Approve(ctx.Request.Context(), cast.ToUint(ctx.Param("id"))); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

func (h *TaskHandle) Cancel(ctx *gin.Context) {
	if err := h.svc.Cancel(ctx.Request.Context(), cast.ToUint(ctx.Param("id"))); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}
