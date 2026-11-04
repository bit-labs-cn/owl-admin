package v1

import (
	"bit-labs.cn/owl-admin/app/service"
	"bit-labs.cn/owl/provider/router"
	"github.com/gin-gonic/gin"
)

var _ router.Handler = (*SettingHandle)(nil)

// SettingHandle 提供系统配置的读取和保存。
type SettingHandle struct {
	svc *service.SettingService
}

// NewSettingHandle 创建配置接口。
func NewSettingHandle(svc *service.SettingService) *SettingHandle {
	return &SettingHandle{svc: svc}
}

// ModuleName 返回权限前缀和菜单标题。
func (i *SettingHandle) ModuleName() (string, string) {
	return "setting", "系统配置"
}

// @Summary		读取系统配置
// @Description	返回各子系统已登记的配置及当前值
// @Tags			系统配置
// @Produce		json
// @Success		200	{object}	router.Resp	"操作成功"
// @Failure		500	{object}	router.Resp	"服务器内部错误"
// @Router			/api/v1/setting [GET]
func (i *SettingHandle) Retrieve(ctx *gin.Context) {
	view, err := i.svc.View(ctx.Request.Context())
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, view)
}

// @Summary		保存系统配置
// @Description	只接受已登记的配置键
// @Tags			系统配置
// @Accept			json
// @Produce		json
// @Param			request	body		service.UpdateSettingsReq	true	"保存请求"
// @Success		200		{object}	router.Resp					"操作成功"
// @Failure		400		{object}	router.Resp					"参数错误"
// @Failure		500		{object}	router.Resp					"服务器内部错误"
// @Router			/api/v1/setting [PUT]
func (i *SettingHandle) Update(ctx *gin.Context) {
	var req service.UpdateSettingsReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	if err := i.svc.Update(ctx.Request.Context(), &req); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// @Summary		公开系统配置
// @Description	只返回登记为公开的配置项
// @Tags			系统配置
// @Produce		json
// @Success		200	{object}	router.Resp	"操作成功"
// @Failure		500	{object}	router.Resp	"服务器内部错误"
// @Router			/api/v1/setting/public [GET]
func (i *SettingHandle) Public(ctx *gin.Context) {
	view, err := i.svc.Public(ctx.Request.Context())
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, view)
}
