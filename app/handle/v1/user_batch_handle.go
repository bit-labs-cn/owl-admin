package v1

import (
	"bit-labs.cn/owl-admin/app/service"
	"bit-labs.cn/owl/provider/router"
	"github.com/gin-gonic/gin"
)

// @Summary		批量删除用户
// @Description	按 ID 列表批量删除用户
// @Tags			用户管理
// @Accept			json
// @Produce		json
// @Param			request	body		service.BatchUserIDsReq	true	"用户ID列表"
// @Success		200		{object}	router.Resp				"操作成功"
// @Failure		400		{object}	router.Resp				"参数错误"
// @Router			/api/v1/users/batch-delete [POST]
func (i *UserHandle) BatchDelete(ctx *gin.Context) {
	var req service.BatchUserIDsReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	if err := i.userSvc.BatchDeleteUsers(ctx.Request.Context(), &req); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// @Summary		批量修改用户状态
// @Description	批量启用或停用用户
// @Tags			用户管理
// @Accept			json
// @Produce		json
// @Param			request	body		service.BatchChangeUserStatusReq	true	"批量状态请求"
// @Success		200		{object}	router.Resp							"操作成功"
// @Failure		400		{object}	router.Resp							"参数错误"
// @Router			/api/v1/users/batch-status [PUT]
func (i *UserHandle) BatchChangeStatus(ctx *gin.Context) {
	var req service.BatchChangeUserStatusReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	if err := i.userSvc.BatchChangeUserStatus(ctx.Request.Context(), &req); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// @Summary		批量分配用户角色
// @Description	覆盖所选用户的角色列表
// @Tags			用户管理
// @Accept			json
// @Produce		json
// @Param			request	body		service.BatchAssignUserRolesReq	true	"批量角色请求"
// @Success		200		{object}	router.Resp						"操作成功"
// @Failure		400		{object}	router.Resp						"参数错误"
// @Router			/api/v1/users/batch-roles [PUT]
func (i *UserHandle) BatchAssignRoles(ctx *gin.Context) {
	var req service.BatchAssignUserRolesReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	if err := i.userSvc.BatchAssignUserRoles(ctx.Request.Context(), &req); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// @Summary		批量设置用户部门
// @Description	覆盖所选用户的归属部门；parentId 为空表示清空
// @Tags			用户管理
// @Accept			json
// @Produce		json
// @Param			request	body		service.BatchAssignUserDeptReq	true	"批量部门请求"
// @Success		200		{object}	router.Resp						"操作成功"
// @Failure		400		{object}	router.Resp						"参数错误"
// @Router			/api/v1/users/batch-dept [PUT]
func (i *UserHandle) BatchAssignDept(ctx *gin.Context) {
	var req service.BatchAssignUserDeptReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	if err := i.userSvc.BatchAssignUserDept(ctx.Request.Context(), &req); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}
