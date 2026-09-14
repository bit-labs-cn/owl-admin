package v1

import (
	"strconv"

	"bit-labs.cn/owl-admin/app/service"
	"bit-labs.cn/owl/provider/db"
	"bit-labs.cn/owl/provider/router"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

var _ router.Handler = (*AppVersionHandle)(nil)

type AppVersionHandle struct {
	svc *service.AppVersionService
}

func NewAppVersionHandle(svc *service.AppVersionService) *AppVersionHandle {
	return &AppVersionHandle{svc: svc}
}

func (i *AppVersionHandle) ModuleName() (string, string) { return "app-version", "APP升级" }

// @Summary		获取最新版本
// @Description	获取可用于升级的最新版本信息
// @Tags			APP升级
// @Produce		json
// @Param			apkType	query		int	false	"安装包类型：1-ios、2-android"
// @Success		200		{object}	router.Resp	"操作成功"
// @Router			/api/v1/app/upgrade [GET]
func (i *AppVersionHandle) Upgrade(ctx *gin.Context) {
	var apkType *int32
	if v := ctx.Query("apkType"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			router.BadRequest(ctx, "参数绑定失败")
			return
		}
		t := int32(n)
		apkType = &t
	}

	latest, err := i.svc.Latest(ctx.Request.Context(), apkType)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	if latest != nil && latest.ApkURL != nil {
		abs := absolutePublicURL(ctx, *latest.ApkURL)
		latest.ApkURL = &abs
	}
	router.Success(ctx, latest)
}

// @Summary		上传APP安装包
// @Description	上传 apk / ipa / aab，返回可访问地址
// @Tags			APP升级
// @Accept			multipart/form-data
// @Produce		json
// @Param			file	formData	file		true	"安装包文件"
// @Param			apkType	formData	int			false	"安装包类型：1-ios、2-android"
// @Success		200		{object}	router.Resp	"操作成功"
// @Failure		400		{object}	router.Resp	"参数错误"
// @Failure		500		{object}	router.Resp	"服务器内部错误"
// @Router			/api/v1/app-versions/upload [POST]
func (i *AppVersionHandle) Upload(ctx *gin.Context) {
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		router.BadRequest(ctx, "请上传安装包")
		return
	}
	f, err := fileHeader.Open()
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	defer f.Close()

	result, err := i.svc.UploadPackage(
		ctx.Request.Context(),
		fileHeader.Filename,
		fileHeader.Size,
		cast.ToInt32(ctx.PostForm("apkType")),
		f,
	)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	result.URL = absolutePublicURL(ctx, result.URL)
	router.Success(ctx, result)
}

// @Summary		创建APP版本
// @Description	创建新的APP升级版本
// @Tags			APP升级
// @Accept			json
// @Produce		json
// @Param			request	body		service.CreateAppVersionReq	true	"版本创建请求"
// @Success		200		{object}	router.Resp					"操作成功"
// @Failure		400		{object}	router.Resp					"参数错误"
// @Failure		500		{object}	router.Resp					"服务器内部错误"
// @Router			/api/v1/app-versions [POST]
func (i *AppVersionHandle) Create(ctx *gin.Context) {
	var req service.CreateAppVersionReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	if err := i.svc.Create(ctx.Request.Context(), &req); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// @Summary		更新APP版本
// @Description	根据ID更新APP升级版本
// @Tags			APP升级
// @Accept			json
// @Produce		json
// @Param			id		path		int							true	"版本ID"
// @Param			request	body		service.UpdateAppVersionReq	true	"版本更新请求"
// @Success		200		{object}	router.Resp					"操作成功"
// @Failure		400		{object}	router.Resp					"参数错误"
// @Failure		500		{object}	router.Resp					"服务器内部错误"
// @Router			/api/v1/app-versions/{id} [PUT]
func (i *AppVersionHandle) Update(ctx *gin.Context) {
	var req service.UpdateAppVersionReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	req.ID = cast.ToInt64(ctx.Param("id"))
	if err := i.svc.Update(ctx.Request.Context(), &req); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// @Summary		删除APP版本
// @Description	根据ID删除APP升级版本
// @Tags			APP升级
// @Produce		json
// @Param			id	path		int			true	"版本ID"
// @Success		200	{object}	router.Resp	"操作成功"
// @Failure		500	{object}	router.Resp	"服务器内部错误"
// @Router			/api/v1/app-versions/{id} [DELETE]
func (i *AppVersionHandle) Delete(ctx *gin.Context) {
	id := cast.ToInt64(ctx.Param("id"))
	if err := i.svc.Delete(ctx.Request.Context(), id); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// @Summary		APP版本列表
// @Description	分页获取APP升级版本
// @Tags			APP升级
// @Produce		json
// @Param			page		query		int				false	"页码"
// @Param			pageSize	query		int				false	"每页数量"
// @Param			version		query		string			false	"版本号"
// @Param			apkType		query		int				false	"安装包类型：1-ios、2-android"
// @Param			status		query		int				false	"状态"
// @Success		200			{object}	router.PageResp	"操作成功"
// @Failure		500			{object}	router.Resp		"服务器内部错误"
// @Router			/api/v1/app-versions [GET]
func (i *AppVersionHandle) Retrieve(ctx *gin.Context) {
	var req service.RetrieveAppVersionReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		router.BadRequest(ctx, "参数绑定失败")
		return
	}
	count, list, err := i.svc.Retrieve(ctx.Request.Context(), &req)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.PageSuccess(ctx, int(count), req.Page, req.PageSize, list)
}

// @Summary		修改APP版本状态
// @Description	启用或停用指定版本
// @Tags			APP升级
// @Accept			json
// @Produce		json
// @Param			id		path		int				true	"版本ID"
// @Param			request	body		db.ChangeStatus	true	"状态修改请求"
// @Success		200		{object}	router.Resp		"操作成功"
// @Failure		400		{object}	router.Resp		"参数错误"
// @Failure		500		{object}	router.Resp		"服务器内部错误"
// @Router			/api/v1/app-versions/{id}/status [PUT]
func (i *AppVersionHandle) ChangeStatus(ctx *gin.Context) {
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
