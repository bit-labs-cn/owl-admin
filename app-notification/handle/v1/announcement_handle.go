package v1

import (
	"bit-labs.cn/owl-admin/app-notification/service"
	"bit-labs.cn/owl/provider/router"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// AnnouncementHandle 处理公告接口。
type AnnouncementHandle struct{ svc *service.AnnouncementService }

func NewAnnouncementHandle(svc *service.AnnouncementService) *AnnouncementHandle {
	return &AnnouncementHandle{svc: svc}
}
func (h *AnnouncementHandle) ModuleName() (string, string) { return "announcement", "站点公告" }

func (h *AnnouncementHandle) Create(ctx *gin.Context) {
	var req service.CreateAnnouncementReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	if err := h.svc.Create(ctx.Request.Context(), &req); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

func (h *AnnouncementHandle) Update(ctx *gin.Context) {
	var req service.CreateAnnouncementReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	if err := h.svc.Update(ctx.Request.Context(), cast.ToUint(ctx.Param("id")), &req); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

func (h *AnnouncementHandle) Delete(ctx *gin.Context) {
	if err := h.svc.Delete(ctx.Request.Context(), cast.ToUint(ctx.Param("id"))); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

func (h *AnnouncementHandle) Detail(ctx *gin.Context) {
	row, err := h.svc.Detail(ctx.Request.Context(), cast.ToUint(ctx.Param("id")))
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, row)
}

func (h *AnnouncementHandle) Retrieve(ctx *gin.Context) {
	var req service.RetrieveAnnouncementReq
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

func (h *AnnouncementHandle) Mine(ctx *gin.Context) {
	rows, err := h.svc.Mine(ctx.Request.Context())
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, rows)
}

func (h *AnnouncementHandle) Dismiss(ctx *gin.Context) {
	if err := h.svc.Dismiss(ctx.Request.Context(), cast.ToUint(ctx.Param("id"))); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

func (h *AnnouncementHandle) Publish(ctx *gin.Context) {
	if err := h.svc.Publish(ctx.Request.Context(), cast.ToUint(ctx.Param("id")), true); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

func (h *AnnouncementHandle) Offline(ctx *gin.Context) {
	if err := h.svc.Publish(ctx.Request.Context(), cast.ToUint(ctx.Param("id")), false); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}
