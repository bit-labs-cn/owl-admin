package v1

import (
	"bit-labs.cn/owl-admin/app/service"
	"bit-labs.cn/owl/provider/db"
	"bit-labs.cn/owl/provider/router"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

var _ router.Handler = (*UserGroupHandle)(nil)

type UserGroupHandle struct {
	svc *service.UserGroupService
}

func NewUserGroupHandle(svc *service.UserGroupService) *UserGroupHandle {
	return &UserGroupHandle{svc: svc}
}

func (i *UserGroupHandle) ModuleName() (string, string) { return "userGroup", "用户组管理" }

// @Summary		创建用户组
// @Description	创建新的用户组
// @Tags			用户组管理
// @Accept			json
// @Produce		json
// @Param			request	body		service.CreateUserGroupReq	true	"用户组创建请求"
// @Success		200		{object}	router.Resp					"操作成功"
// @Failure		400		{object}	router.Resp					"参数错误"
// @Failure		500		{object}	router.Resp					"服务器内部错误"
// @Router			/api/v1/user-groups [POST]
func (i *UserGroupHandle) Create(ctx *gin.Context) {
	var req service.CreateUserGroupReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	if err := i.svc.CreateUserGroup(ctx.Request.Context(), &req); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// @Summary		更新用户组
// @Description	根据用户组ID更新用户组信息
// @Tags			用户组管理
// @Accept			json
// @Produce		json
// @Param			id		path		int							true	"用户组ID"
// @Param			request	body		service.UpdateUserGroupReq	true	"用户组更新请求"
// @Success		200		{object}	router.Resp					"操作成功"
// @Failure		400		{object}	router.Resp					"参数错误"
// @Failure		500		{object}	router.Resp					"服务器内部错误"
// @Router			/api/v1/user-groups/{id} [PUT]
func (i *UserGroupHandle) Update(ctx *gin.Context) {
	var req service.UpdateUserGroupReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	req.ID = cast.ToUint(ctx.Param("id"))
	if err := i.svc.UpdateUserGroup(ctx.Request.Context(), &req); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// @Summary		删除用户组
// @Description	根据用户组ID删除用户组
// @Tags			用户组管理
// @Produce		json
// @Param			id	path		int			true	"用户组ID"
// @Success		200	{object}	router.Resp	"操作成功"
// @Failure		500	{object}	router.Resp	"服务器内部错误"
// @Router			/api/v1/user-groups/{id} [DELETE]
func (i *UserGroupHandle) Delete(ctx *gin.Context) {
	id := cast.ToUint(ctx.Param("id"))
	if err := i.svc.DeleteUserGroup(ctx.Request.Context(), id); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// @Summary		获取用户组列表
// @Description	分页获取用户组列表
// @Tags			用户组管理
// @Produce		json
// @Param			page		query		int				false	"页码"
// @Param			pageSize	query		int				false	"每页数量"
// @Param			name		query		string			false	"名称模糊搜索"
// @Param			code		query		string			false	"编码"
// @Param			status		query		int				false	"状态"
// @Success		200			{object}	router.PageResp	"操作成功"
// @Failure		500			{object}	router.Resp		"服务器内部错误"
// @Router			/api/v1/user-groups [GET]
func (i *UserGroupHandle) Retrieve(ctx *gin.Context) {
	var req service.RetrieveUserGroupReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		router.BadRequest(ctx, "参数绑定失败")
		return
	}
	count, list, err := i.svc.RetrieveUserGroups(ctx.Request.Context(), &req)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.PageSuccess(ctx, int(count), req.Page, req.PageSize, list)
}

// @Summary		修改用户组状态
// @Description	启用或禁用指定用户组
// @Tags			用户组管理
// @Accept			json
// @Produce		json
// @Param			id		path		int				true	"用户组ID"
// @Param			request	body		db.ChangeStatus	true	"状态修改请求"
// @Success		200		{object}	router.Resp		"操作成功"
// @Failure		400		{object}	router.Resp		"参数错误"
// @Failure		500		{object}	router.Resp		"服务器内部错误"
// @Router			/api/v1/user-groups/{id}/status [PUT]
func (i *UserGroupHandle) ChangeStatus(ctx *gin.Context) {
	var req db.ChangeStatus
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	req.ID = cast.ToUint(ctx.Param("id"))
	if err := i.svc.ChangeStatus(ctx.Request.Context(), &req); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// @Summary		获取用户组选项
// @Description	获取所有启用的用户组信息用于选择
// @Tags			用户组管理
// @Produce		json
// @Success		200	{object}	router.Resp	"操作成功"
// @Failure		500	{object}	router.Resp	"服务器内部错误"
// @Router			/api/v1/user-groups-options [GET]
func (i *UserGroupHandle) Options(ctx *gin.Context) {
	list, err := i.svc.Options(ctx.Request.Context())
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, list)
}

// @Summary		获取用户组内的用户列表
// @Description	分页获取指定用户组下的用户列表
// @Tags			用户组管理
// @Produce		json
// @Param			id			path		int				true	"用户组ID"
// @Param			page		query		int				false	"页码"
// @Param			pageSize	query		int				false	"每页数量"
// @Success		200			{object}	router.PageResp	"操作成功"
// @Failure		500			{object}	router.Resp		"服务器内部错误"
// @Router			/api/v1/user-groups/{id}/users [GET]
func (i *UserGroupHandle) GetUsersByGroupID(ctx *gin.Context) {
	groupID := cast.ToUint(ctx.Param("id"))
	page := cast.ToInt(ctx.DefaultQuery("page", "1"))
	pageSize := cast.ToInt(ctx.DefaultQuery("pageSize", "10"))

	count, list, err := i.svc.GetUsersByGroupID(ctx.Request.Context(), groupID, page, pageSize)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.PageSuccess(ctx, int(count), page, pageSize, list)
}

// @Summary		获取用户组的用户ID列表
// @Description	返回指定用户组关联的所有用户ID（轻量接口，用于表单回显）
// @Tags			用户组管理
// @Produce		json
// @Param			id	path		int			true	"用户组ID"
// @Success		200	{object}	router.Resp	"操作成功"
// @Failure		500	{object}	router.Resp	"服务器内部错误"
// @Router			/api/v1/user-groups/{id}/user-ids [GET]
func (i *UserGroupHandle) GetUserIDsByGroupID(ctx *gin.Context) {
	groupID := cast.ToUint(ctx.Param("id"))
	ids, err := i.svc.GetUserIDsByGroupID(ctx.Request.Context(), groupID)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, ids)
}

// @Summary		为用户组分配用户
// @Description	为指定用户组分配用户
// @Tags			用户组管理
// @Accept			json
// @Produce		json
// @Param			id		path		int								true	"用户组ID"
// @Param			request	body		service.AssignUsersToGroupReq	true	"分配用户请求"
// @Success		200		{object}	router.Resp						"操作成功"
// @Failure		400		{object}	router.Resp						"参数错误"
// @Failure		500		{object}	router.Resp						"服务器内部错误"
// @Router			/api/v1/user-groups/{id}/users [PUT]
func (i *UserGroupHandle) AssignUsersToGroup(ctx *gin.Context) {
	var req service.AssignUsersToGroupReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	req.GroupID = cast.ToUint(ctx.Param("id"))
	if err := i.svc.AssignUsersToGroup(ctx.Request.Context(), &req); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}
