package v1

import (
	"bit-labs.cn/owl-admin/app/service"
	"bit-labs.cn/owl/provider/db"
	"bit-labs.cn/owl/provider/router"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

type PositionHandle struct {
	svc *service.PositionService
}

func NewPositionHandle(svc *service.PositionService) *PositionHandle {
	return &PositionHandle{svc: svc}
}

func (i *PositionHandle) ModuleName() (string, string) { return "position", "岗位管理" }

// @Summary		创建岗位
// @Description	创建新的岗位
// @Tags			岗位管理
// @Accept			json
// @Produce		json
// @Param			request	body		service.CreatePositionReq	true	"岗位创建请求"
// @Success		200		{object}	router.Resp					"操作成功"
// @Failure		400		{object}	router.Resp					"参数错误"
// @Failure		500		{object}	router.Resp					"服务器内部错误"
// @Router			/api/v1/position [POST]
func (i *PositionHandle) Create(ctx *gin.Context) {
	var req service.CreatePositionReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.BadRequest(ctx, "参数绑定失败")
		return
	}
	if err := i.svc.CreatePosition(ctx.Request.Context(), &req); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// @Summary		更新岗位
// @Description	根据岗位ID更新岗位信息
// @Tags			岗位管理
// @Accept			json
// @Produce		json
// @Param			id		path		int							true	"岗位ID"
// @Param			request	body		service.UpdatePositionReq	true	"岗位更新请求"
// @Success		200		{object}	router.Resp					"操作成功"
// @Failure		400		{object}	router.Resp					"参数错误"
// @Failure		500		{object}	router.Resp					"服务器内部错误"
// @Router			/api/v1/position/{id} [PUT]
func (i *PositionHandle) Update(ctx *gin.Context) {
	var req service.UpdatePositionReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.BadRequest(ctx, "参数绑定失败")
		return
	}
	req.ID = cast.ToUint(ctx.Param("id"))
	if err := i.svc.UpdatePosition(ctx.Request.Context(), &req); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// @Summary		删除岗位
// @Description	根据岗位ID删除岗位
// @Tags			岗位管理
// @Produce		json
// @Param			id	path		int			true	"岗位ID"
// @Success		200	{object}	router.Resp	"操作成功"
// @Failure		500	{object}	router.Resp	"服务器内部错误"
// @Router			/api/v1/position/{id} [DELETE]
func (i *PositionHandle) Delete(ctx *gin.Context) {
	id := cast.ToUint(ctx.Param("id"))
	if err := i.svc.DeletePosition(ctx.Request.Context(), id); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// @Summary		获取岗位列表
// @Description	分页获取岗位列表
// @Tags			岗位管理
// @Produce		json
// @Param			page		query		int				false	"页码"
// @Param			pageSize	query		int				false	"每页数量"
// @Param			nameLike	query		string			false	"名称模糊搜索"
// @Param			status		query		int				false	"状态"
// @Success		200			{object}	router.PageResp	"操作成功"
// @Failure		500			{object}	router.Resp		"服务器内部错误"
// @Router			/api/v1/position [GET]
func (i *PositionHandle) Retrieve(ctx *gin.Context) {
	var req service.RetrievePositionReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		router.BadRequest(ctx, "参数绑定失败")
		return
	}
	count, list, err := i.svc.RetrievePositions(ctx.Request.Context(), &req)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.PageSuccess(ctx, int(count), req.Page, req.PageSize, list)
}

// @Summary		修改岗位状态
// @Description	启用或禁用指定岗位
// @Tags			岗位管理
// @Accept			json
// @Produce		json
// @Param			id		path		int				true	"岗位ID"
// @Param			request	body		db.ChangeStatus	true	"状态修改请求"
// @Success		200		{object}	router.Resp		"操作成功"
// @Failure		400		{object}	router.Resp		"参数错误"
// @Failure		500		{object}	router.Resp		"服务器内部错误"
// @Router			/api/v1/position/{id}/status [PUT]
func (i *PositionHandle) ChangeStatus(ctx *gin.Context) {
	var req db.ChangeStatus
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.BadRequest(ctx, "参数绑定失败")
		return
	}
	req.ID = cast.ToUint(ctx.Param("id"))
	if err := i.svc.ChangeStatus(ctx.Request.Context(), &req); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// @Summary		获取岗位选项
// @Description	获取所有岗位的简单信息用于选择
// @Tags			岗位管理
// @Produce		json
// @Success		200	{object}	router.Resp	"操作成功"
// @Failure		500	{object}	router.Resp	"服务器内部错误"
// @Router			/api/v1/position-options [GET]
func (i *PositionHandle) Options(ctx *gin.Context) {
	list, err := i.svc.Options(ctx.Request.Context())
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, list)
}
