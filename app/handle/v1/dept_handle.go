package v1

import (
	"bit-labs.cn/owl-admin/app/service"
	"bit-labs.cn/owl/provider/router"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

var _ router.Handler = (*DeptHandle)(nil)
var _ router.CrudHandler = (*DeptHandle)(nil)

type DeptHandle struct {
	deptSvc *service.DeptService
}

func NewDeptHandle(deptSvc *service.DeptService) *DeptHandle {
	return &DeptHandle{deptSvc: deptSvc}
}
func (i DeptHandle) ModuleName() (en string, zh string) {
	return "dept", "部门管理"
}

// @Summary		创建部门
// @Description	创建一个新的部门
// @Tags			部门管理
// @Accept			json
// @Produce		json
// @Param			request	body		service.CreateDeptReq	true	"部门创建请求"
// @Success		200		{object}	router.Resp				"操作成功"
// @Failure		400		{object}	router.Resp				"参数错误"
// @Failure		500		{object}	router.Resp				"服务器内部错误"
// @Router			/api/v1/dept [POST]
func (i DeptHandle) Create(ctx *gin.Context) {
	var req service.CreateDeptReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.BadRequest(ctx, "参数绑定失败")
		return
	}

	err := i.deptSvc.CreateDept(ctx.Request.Context(), &req)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// @Summary		更新部门
// @Description	根据部门ID更新部门信息
// @Tags			部门管理
// @Accept			json
// @Produce		json
// @Param			id		path		int						true	"部门ID"
// @Param			request	body		service.UpdateDeptReq	true	"部门更新请求"
// @Success		200		{object}	router.Resp				"操作成功"
// @Failure		400		{object}	router.Resp				"参数错误"
// @Failure		500		{object}	router.Resp				"服务器内部错误"
// @Router			/api/v1/dept/{id} [PUT]
func (i DeptHandle) Update(ctx *gin.Context) {
	var req service.UpdateDeptReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.BadRequest(ctx, "参数绑定失败")
		return
	}

	err := i.deptSvc.UpdateDept(ctx.Request.Context(), &req)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// @Summary		删除部门
// @Description	根据部门ID删除部门
// @Tags			部门管理
// @Produce		json
// @Param			id	path		int			true	"部门ID"
// @Success		200	{object}	router.Resp	"操作成功"
// @Failure		500	{object}	router.Resp	"服务器内部错误"
// @Router			/api/v1/dept/{id} [DELETE]
func (i DeptHandle) Delete(ctx *gin.Context) {
	id := cast.ToUint(ctx.Param("id"))
	err := i.deptSvc.DeleteDept(ctx.Request.Context(), id)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// @Summary		获取部门列表
// @Description	获取所有部门
// @Tags			部门管理
// @Produce		json
// @Success		200	{object}	router.Resp{data=[]model.Dept}	"操作成功"
// @Failure		500	{object}	router.Resp						"服务器内部错误"
// @Router			/api/v1/dept [GET]
func (i DeptHandle) Retrieve(ctx *gin.Context) {
	var req service.RetrieveDeptReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		router.BadRequest(ctx, "参数绑定失败")
		return
	}
	_, list, err := i.deptSvc.RetrieveDepts(ctx.Request.Context(), &req)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, list)
}

func (i DeptHandle) Detail(ctx *gin.Context) {

}
