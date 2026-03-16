package v1

import (
	"time"

	"bit-labs.cn/flex-admin/app/model"
	"bit-labs.cn/flex-admin/app/service"
	"bit-labs.cn/owl/provider/db"
	"bit-labs.cn/owl/provider/router"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

var _ router.Handler = (*UserHandle)(nil)
var _ router.CrudHandler = (*UserHandle)(nil)

type UserHandle struct {
	userSvc  *service.UserService
	roleSvc  *service.RoleService
	menuRepo *router.MenuRepository
	logSvc   *service.LogService
}

func (i *UserHandle) ModuleName() (string, string) {
	return "user", "用户管理"
}

func NewUserHandle(userService *service.UserService, roleService *service.RoleService, manager *router.MenuRepository, logSvc *service.LogService) *UserHandle {
	return &UserHandle{
		userSvc:  userService,
		roleSvc:  roleService,
		menuRepo: manager,
		logSvc:   logSvc,
	}
}

// @Summary		创建用户
// @Description	创建一个新的用户账户
// @Tags			用户管理
// @Accept			json
// @Produce		json
// @Param			request	body		service.CreateUserReq	true	"用户创建请求"
// @Success		200		{object}	router.Resp				"操作成功"
// @Failure		400		{object}	router.Resp				"参数错误"
// @Failure		500		{object}	router.Resp				"服务器内部错误"
// @Router			/api/v1/users [POST]
func (i *UserHandle) Create(ctx *gin.Context) {
	req := new(service.CreateUserReq)
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.BadRequest(ctx, err.Error())
		return
	}

	err := i.userSvc.CreateUser(ctx.Request.Context(), req)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// @Summary		删除用户
// @Description	根据用户ID删除用户
// @Tags			用户管理
// @Produce		json
// @Param			id	path		int			true	"用户ID"
// @Success		200	{object}	router.Resp	"操作成功"
// @Failure		500	{object}	router.Resp	"服务器内部错误"
// @Router			/api/v1/users/{id} [DELETE]
func (i *UserHandle) Delete(ctx *gin.Context) {

	id := cast.ToUint(ctx.Param("id"))
	err := i.userSvc.DeleteUser(ctx.Request.Context(), id)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

func (i *UserHandle) Detail(ctx *gin.Context) {

}

// @Summary		更新用户
// @Description	根据用户ID更新用户信息
// @Tags			用户管理
// @Accept			json
// @Produce		json
// @Param			id		path		int						true	"用户ID"
// @Param			request	body		service.UpdateUserReq	true	"用户更新请求"
// @Success		200		{object}	router.Resp				"操作成功"
// @Failure		400		{object}	router.Resp				"参数错误"
// @Failure		500		{object}	router.Resp				"服务器内部错误"
// @Router			/api/v1/users/{id} [PUT]
func (i *UserHandle) Update(ctx *gin.Context) {

	req := new(service.UpdateUserReq)
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.BadRequest(ctx, err.Error())
		return
	}
	id := cast.ToUint(ctx.Param("id"))
	req.ID = id

	err := i.userSvc.UpdateUser(ctx.Request.Context(), req)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// @Summary		修改用户状态
// @Description	启用或禁用指定用户
// @Tags			用户管理
// @Accept			json
// @Produce		json
// @Param			id		path		int				true	"用户ID"
// @Param			request	body		db.ChangeStatus	true	"状态修改请求"
// @Success		200		{object}	router.Resp		"操作成功"
// @Failure		400		{object}	router.Resp		"参数错误"
// @Failure		500		{object}	router.Resp		"服务器内部错误"
// @Router			/api/v1/users/{id}/status [PUT]
func (i *UserHandle) ChangeStatus(ctx *gin.Context) {
	req := new(db.ChangeStatus)
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.BadRequest(ctx, err.Error())
		return
	}
	id := cast.ToUint(ctx.Param("id"))
	req.ID = id

	err := i.userSvc.ChangeUserStatus(ctx.Request.Context(), req)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// @Summary		获取用户列表
// @Description	分页获取用户列表
// @Tags			用户管理
// @Produce		json
// @Param			page		query		int				false	"页码"
// @Param			pageSize	query		int				false	"每页数量"
// @Param			keyword		query		string			false	"搜索关键词"
// @Success		200			{object}	router.PageResp	"操作成功"
// @Failure		500			{object}	router.Resp		"服务器内部错误"
// @Router			/api/v1/users [GET]
func (i *UserHandle) Retrieve(ctx *gin.Context) {
	var req service.RetrieveUserReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		router.BadRequest(ctx, "参数绑定失败")
		return
	}
	count, list, err := i.userSvc.RetrieveUsers(ctx.Request.Context(), &req)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.PageSuccess(ctx, count, req.Page, req.PageSize, list)
}

// @Summary		分配角色给用户
// @Description	为指定用户分配角色
// @Tags			用户管理
// @Accept			json
// @Produce		json
// @Param			id		path		int							true	"用户ID"
// @Param			request	body		service.AssignRoleToUser	true	"分配角色请求"
// @Success		200		{object}	router.Resp					"操作成功"
// @Failure		400		{object}	router.Resp					"参数错误"
// @Failure		500		{object}	router.Resp					"服务器内部错误"
// @Router			/api/v1/users/{id}/roles [POST]
func (i *UserHandle) AssignRolesToUser(ctx *gin.Context) {
	req := new(service.AssignRoleToUser)
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.BadRequest(ctx, err.Error())
		return
	}
	req.UserID = cast.ToUint(ctx.Param("id"))
	if err := i.userSvc.AssignRoleToUser(ctx.Request.Context(), req); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// @Summary		获取用户角色ID列表
// @Description	查询指定用户的角色ID列表
// @Tags			用户管理
// @Produce		json
// @Param			id	path		int			true	"用户ID"
// @Success		200	{object}	router.Resp	"操作成功"
// @Failure		500	{object}	router.Resp	"服务器内部错误"
// @Router			/api/v1/users/{id}/roles [GET]
func (i *UserHandle) GetRoleIdsByUserId(ctx *gin.Context) {

	userID := cast.ToUint(ctx.Param("id"))
	ids, err := i.userSvc.GetUserRoleIDs(ctx.Request.Context(), userID)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, ids)
}

// @Summary		分配菜单/权限给用户
// @Description	为指定用户分配菜单/权限（按角色关联）
// @Tags			用户管理
// @Accept			json
// @Produce		json
// @Param			id		path		int							true	"用户ID"
// @Param			request	body		service.AssignRoleToUser	true	"分配请求"
// @Success		200		{object}	router.Resp					"操作成功"
// @Failure		400		{object}	router.Resp					"参数错误"
// @Failure		500		{object}	router.Resp					"服务器内部错误"
// @Router			/api/v1/users/{id}/menus [POST]
func (i *UserHandle) AssignMenuToUser(ctx *gin.Context) {
	req := new(service.AssignRoleToUser)
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.BadRequest(ctx, err.Error())
		return
	}
	req.UserID = cast.ToUint(ctx.Param("id"))
	if err := i.userSvc.AssignRoleToUser(ctx.Request.Context(), req); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// @Summary		获取当前用户菜单
// @Description	根据用户角色返回可见菜单列表
// @Tags			用户管理
// @Produce		json
// @Success		200	{object}	router.Resp	"操作成功"
// @Router			/api/v1/users/me/menus [GET]
func (i *UserHandle) GetMyMenus(ctx *gin.Context) {

	user, _ := ctx.Get("user")
	if user.(*model.User).IsSuperAdmin {
		router.Success(ctx, i.menuRepo.GetMenuWithoutBtn())
		return
	}
	menus := i.userSvc.GetUserMenus(ctx.Request.Context(), user.(*model.User).ID)
	router.Success(ctx, menus)
}

// @Summary		获取当前用户权限
// @Description	返回当前登录用户的权限标识列表
// @Tags			用户管理
// @Produce		json
// @Success		200	{object}	router.Resp{Data=[]string}	"操作成功"
// @Router			/api/v1/users/me/permissions [GET]
func (i *UserHandle) GetMyPermissions(ctx *gin.Context) {
	user, _ := ctx.Get("user")
	u := user.(*model.User)

	permissions, err := i.userSvc.GetMyPermissions(ctx.Request.Context(), u.ID, u.IsSuperAdmin)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, permissions)
}

// @Summary		修改我的密码
// @Description	用户自行修改密码
// @Tags			用户管理
// @Accept			json
// @Produce		json
// @Param			request	body		service.ChangePasswordReq	true	"修改密码请求"
// @Success		200		{object}	router.Resp					"操作成功"
// @Failure		400		{object}	router.Resp					"参数错误"
// @Failure		500		{object}	router.Resp					"服务器内部错误"
// @Router			/api/v1/users/me/password [PUT]
func (i *UserHandle) ChangePassword(ctx *gin.Context) {
	var req service.ChangePasswordReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.BadRequest(ctx, err.Error())
		return
	}
	uVal, _ := ctx.Get("user")
	user := uVal.(*model.User)
	req.UserID = user.ID
	if err := i.userSvc.ChangeUserPassword(ctx.Request.Context(), &req); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// @Summary		重置用户密码（仅超管）
// @Description	仅超管可用，直接重置指定用户密码
// @Tags			用户管理
// @Accept			json
// @Produce		json
// @Param			id		path		int							true	"用户ID"
// @Param			request	body		service.ResetPasswordReq	true	"重置密码请求"
// @Success		200		{object}	router.Resp					"操作成功"
// @Failure		400		{object}	router.Resp					"参数错误"
// @Failure		500		{object}	router.Resp					"服务器内部错误"
// @Router			/api/v1/users/{id}/reset [PUT]
func (i *UserHandle) ResetPassword(ctx *gin.Context) {
	var req service.ResetPasswordReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.BadRequest(ctx, err.Error())
		return
	}

	req.UserID = cast.ToUint(ctx.Param("id"))
	if err := i.userSvc.ResetUserPassword(ctx.Request.Context(), &req); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// @Summary		修改用户头像
// @Description	管理员修改指定用户头像
// @Tags			用户管理
// @Accept			json
// @Produce		json
// @Param			id		path		int	true	"用户ID"
// @Param			request	body		service.ChangeAvatarReq	true	"头像请求（avatar 为图片 URL）"
// @Success		200		{object}	router.Resp				"操作成功"
// @Router			/api/v1/users/{id}/avatar [PUT]
func (i *UserHandle) ChangeAvatar(ctx *gin.Context) {
	var req service.ChangeAvatarReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.BadRequest(ctx, err.Error())
		return
	}
	req.UserID = cast.ToUint(ctx.Param("id"))
	if err := i.userSvc.ChangeUserAvatar(ctx.Request.Context(), &req); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// @Summary		用户登录
// @Description	使用用户名与密码进行登录
// @Tags			用户管理
// @Accept			json
// @Produce		json
// @Param			request	body		service.LoginReq	true	"登录请求"
// @Success		200		{object}	router.Resp			"登录成功"
// @Failure		400		{object}	router.Resp			"参数错误"
// @Failure		500		{object}	router.Resp			"服务器内部错误"
// @Router			/api/v1/users/login [POST]
func (i *UserHandle) Login(ctx *gin.Context) {
	var req service.LoginReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.BadRequest(ctx, err.Error())
		return
	}

	login, err := i.userSvc.Login(ctx.Request.Context(), &req)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	// 记录登录日志
	var uType = "user"
	if login.User != nil && login.User.IsSuperAdmin {
		uType = "super_admin"
	}
	_ = i.logSvc.CreateLoginLog(ctx.Request.Context(), &service.CreateLoginLogReq{
		UserId:    int(login.User.ID),
		UserName:  login.User.Username,
		UserType:  uType,
		LoginTime: int(time.Now().Unix()),
		Ip:        ctx.ClientIP(),
		UserAgent: ctx.GetHeader("User-Agent"),
	})
	router.Success(ctx, login)
}
func (i *UserHandle) Register(ctx *gin.Context) {

}

// @Summary		获取当前用户信息
// @Description	返回当前登录用户的基本信息
// @Tags			用户管理
// @Produce		json
// @Success		200	{object}	router.Resp{Data=model.User}	"操作成功"
// @Router			/api/v1/users/me [GET]
func (i *UserHandle) Me(ctx *gin.Context) {
	value, _ := ctx.Get("user")
	router.Success(ctx, value)
}
