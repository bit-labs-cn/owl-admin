package v1

import (
	"bit-labs.cn/owl-admin/app/service"
	"bit-labs.cn/owl/provider/db"
	"bit-labs.cn/owl/provider/router"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

var _ router.Handler = (*RoleHandle)(nil)
var _ router.CrudHandler = (*RoleHandle)(nil)

type RoleHandle struct {
	roleService *service.RoleService
	userService *service.UserService
}

func (i *RoleHandle) ModuleName() (string, string) {
	return "role", "角色管理"
}

func NewRoleHandle(roleService *service.RoleService, userService *service.UserService) *RoleHandle {
	return &RoleHandle{
		roleService: roleService,
		userService: userService,
	}
}

// @Summary		创建角色
// @Description	创建新的角色
// @Tags			角色管理
// @Accept			json
// @Produce		json
// @Param			request	body		service.CreateRoleReq	true	"角色创建请求"
// @Success		200		{object}	router.Resp				"操作成功"
// @Failure		400		{object}	router.Resp				"参数错误"
// @Failure		500		{object}	router.Resp				"服务器内部错误"
// @Router			/api/v1/role [POST]
func (i *RoleHandle) Create(ctx *gin.Context) {
	req := new(service.CreateRoleReq)
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.BadRequest(ctx, err.Error())
		return
	}

	err := i.roleService.CreateRole(ctx.Request.Context(), req)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

func (i *RoleHandle) Detail(ctx *gin.Context) {

}

// @Summary		更新角色
// @Description	根据角色ID更新角色信息
// @Tags			角色管理
// @Accept			json
// @Produce		json
// @Param			id		path		int						true	"角色ID"
// @Param			request	body		service.UpdateRoleReq	true	"角色更新请求"
// @Success		200		{object}	router.Resp				"操作成功"
// @Failure		400		{object}	router.Resp				"参数错误"
// @Failure		500		{object}	router.Resp				"服务器内部错误"
// @Router			/api/v1/role/{id} [PUT]
func (i *RoleHandle) Update(ctx *gin.Context) {
	req := new(service.UpdateRoleReq)
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.BadRequest(ctx, err.Error())
		return
	}

	id := cast.ToUint(ctx.Param("id"))
	req.ID = id

	err := i.roleService.UpdateRole(ctx.Request.Context(), req)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

func (i *RoleHandle) Delete(ctx *gin.Context) {
	id := cast.ToUint(ctx.Param("id"))
	err := i.roleService.DeleteRole(ctx.Request.Context(), id)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// @Summary		获取角色列表
// @Description	分页获取角色列表
// @Tags			角色管理
// @Produce		json
// @Param			page		query		int				false	"页码"
// @Param			pageSize	query		int				false	"每页数量"
// @Param			name		query		string			false	"名称模糊搜索"
// @Param			code		query		string			false	"编码"
// @Param			status		query		int				false	"状态"
// @Success		200			{object}	router.PageResp	"操作成功"
// @Failure		500			{object}	router.Resp		"服务器内部错误"
// @Router			/api/v1/role [GET]
func (i *RoleHandle) Retrieve(ctx *gin.Context) {
	var req service.RetrieveRoleReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		router.BadRequest(ctx, "参数绑定失败")
		return
	}

	count, list, err := i.roleService.RetrieveRoles(ctx.Request.Context(), &req)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.PageSuccess(ctx, int(count), req.Page, req.PageSize, list)
}

// @Summary		分配菜单给角色
// @Description	为指定角色分配菜单权限
// @Tags			角色管理
// @Accept			json
// @Produce		json
// @Param			id		path		int							true	"角色ID"
// @Param			request	body		service.AssignMenuToRole	true	"菜单分配请求"
// @Success		200		{object}	router.Resp					"操作成功"
// @Failure		400		{object}	router.Resp					"参数错误"
// @Failure		500		{object}	router.Resp					"服务器内部错误"
// @Router			/api/v1/role/{id}/menus [POST]
func (i *RoleHandle) AssignMenusToRole(ctx *gin.Context) {
	req := new(service.AssignMenuToRole)
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.BadRequest(ctx, err.Error())
		return
	}
	req.RoleID = cast.ToUint(ctx.Param("id"))
	err := i.roleService.AssignMenusToRole(ctx.Request.Context(), req)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// @Summary		获取角色菜单ID列表
// @Description	获取指定角色拥有的菜单ID列表
// @Tags			角色管理
// @Produce		json
// @Param			id	path		string		true	"角色ID"
// @Success		200	{object}	router.Resp	"操作成功"
// @Router			/api/v1/role/{id}/menus [GET]
func (i *RoleHandle) GetRoleMenuIDs(ctx *gin.Context) {
	id := ctx.Param("id")
	menuIds := i.roleService.GetRolesMenuIDs(ctx.Request.Context(), id)
	router.Success(ctx, menuIds)

}

// GetUsersByRoleID 获取角色下的用户列表
func (i *RoleHandle) GetUsersByRoleID(ctx *gin.Context) {
	roleID := cast.ToUint(ctx.Param("id"))

	page := cast.ToInt(ctx.DefaultQuery("page", "1"))
	pageSize := cast.ToInt(ctx.DefaultQuery("pageSize", "10"))

	count, list, err := i.userService.RetrieveUsersByRoleID(ctx.Request.Context(), roleID, page, pageSize)
	if err != nil {
		router.Fail(ctx, err)
		return
	}

	router.PageSuccess(ctx, count, page, pageSize, list)
}

// @Summary		获取角色选项
// @Description	获取所有角色的简单信息用于选择
// @Tags			角色管理
// @Produce		json
// @Success		200	{object}	router.Resp	"操作成功"
// @Failure		500	{object}	router.Resp	"服务器内部错误"
// @Router			/api/v1/role/options [GET]
func (i *RoleHandle) RoleOptions(ctx *gin.Context) {
	x, err := i.roleService.Options(ctx.Request.Context())
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, x)
}

// @Summary		修改角色状态
// @Description	启用或禁用指定角色
// @Tags			角色管理
// @Accept			json
// @Produce		json
// @Param			id		path		int				true	"角色ID"
// @Param			request	body		db.ChangeStatus	true	"状态修改请求"
// @Success		200		{object}	router.Resp		"操作成功"
// @Failure		400		{object}	router.Resp		"参数错误"
// @Failure		500		{object}	router.Resp		"服务器内部错误"
// @Router			/api/v1/role/{id}/status [PUT]
func (i *RoleHandle) ChangeStatus(ctx *gin.Context) {

	req := new(db.ChangeStatus)
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.BadRequest(ctx, err.Error())
		return
	}
	id := cast.ToUint(ctx.Param("id"))
	req.ID = id
	if err := i.roleService.ChangeStatus(ctx.Request.Context(), req); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}
