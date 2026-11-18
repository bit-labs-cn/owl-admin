package v1

import (
	"bit-labs.cn/owl-admin/app-notification/service"
	"bit-labs.cn/owl/provider/router"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// TemplateHandle 处理通知模板接口。
type TemplateHandle struct{ svc *service.TemplateService }

func NewTemplateHandle(svc *service.TemplateService) *TemplateHandle {
	return &TemplateHandle{svc: svc}
}
func (h *TemplateHandle) ModuleName() (string, string) { return "template", "通知模板" }

func (h *TemplateHandle) Create(ctx *gin.Context) {
	var req service.CreateTemplateReq
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

func (h *TemplateHandle) Update(ctx *gin.Context) {
	var req service.CreateTemplateReq
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

func (h *TemplateHandle) Delete(ctx *gin.Context) {
	if err := h.svc.Delete(ctx.Request.Context(), cast.ToUint(ctx.Param("id"))); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

func (h *TemplateHandle) Detail(ctx *gin.Context) {
	row, err := h.svc.Detail(ctx.Request.Context(), cast.ToUint(ctx.Param("id")))
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, row)
}

func (h *TemplateHandle) Retrieve(ctx *gin.Context) {
	var req service.RetrieveTemplateReq
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

func (h *TemplateHandle) Preview(ctx *gin.Context) {
	var req service.CreateTemplateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, h.svc.Preview(&req))
}
