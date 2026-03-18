package v1

import (
	"bit-labs.cn/owl-admin/app/service"
	"bit-labs.cn/owl/provider/router"
	"github.com/gin-gonic/gin"
)

var _ router.Handler = (*AreaHandle)(nil)

type AreaHandle struct {
	areaSvc *service.AreaService
}

func NewAreaHandle(areaSvc *service.AreaService) *AreaHandle {
	return &AreaHandle{areaSvc: areaSvc}
}

func (i AreaHandle) ModuleName() (en string, zh string) {
	return "area", "省市区"
}

// @Summary		获取省市区列表
// @Description	查询所有省市区数据（平铺，不构建树）
// @Tags			省市区
// @Accept			json
// @Produce		json
// @Param			request	body		service.RetrieveAllAreaReq		true	"查询请求"
// @Success		200		{object}	router.Resp{data=[]model.Area}	"操作成功"
// @Failure		400		{object}	router.Resp						"参数错误"
// @Failure		500		{object}	router.Resp						"服务器内部错误"
// @Router			/api/v1/areas/all [POST]
func (i AreaHandle) RetrieveAll(ctx *gin.Context) {
	var req service.RetrieveAllAreaReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.BadRequest(ctx, "参数绑定失败")
		return
	}
	list, err := i.areaSvc.RetrieveAll(ctx.Request.Context(), &req)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, list)
}
