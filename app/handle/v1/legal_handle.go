package v1

import (
	"bit-labs.cn/owl-admin/app/model"
	"bit-labs.cn/owl-admin/app/service"
	"bit-labs.cn/owl/provider/router"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

var _ router.Handler = (*LegalHandle)(nil)

// LegalHandle 处理条款管理、公开阅读和同意记录。
type LegalHandle struct {
	svc *service.LegalService
}

// NewLegalHandle 创建条款接口。
func NewLegalHandle(svc *service.LegalService) *LegalHandle {
	return &LegalHandle{svc: svc}
}

// ModuleName 返回模块标识和名称。
func (i *LegalHandle) ModuleName() (string, string) {
	return "legal", "条款管理"
}

func legalActor(ctx *gin.Context) service.LegalActor {
	actor := service.LegalActor{IP: ctx.ClientIP()}
	if ctx.Request != nil {
		actor.UserAgent = ctx.Request.UserAgent()
	}
	value, ok := ctx.Get("user")
	if !ok || value == nil {
		return actor
	}
	user, ok := value.(*model.User)
	if !ok || user == nil {
		return actor
	}
	actor.UserID = user.ID
	actor.Username = user.Username
	actor.Nickname = user.Nickname
	return actor
}

// @Summary		创建条款
// @Description	新建一条条款，默认启用
// @Tags			条款管理
// @Accept			json
// @Produce		json
// @Param			request	body		service.CreateLegalDocumentReq	true	"条款"
// @Success		200		{object}	router.Resp					"操作成功"
// @Router			/api/v1/legal/documents [POST]
func (i *LegalHandle) CreateDocument(ctx *gin.Context) {
	var req service.CreateLegalDocumentReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	if err := i.svc.CreateDocument(ctx.Request.Context(), &req); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// @Summary		更新条款
// @Description	修改条款名称、性质和启用状态
// @Tags			条款管理
// @Accept			json
// @Produce		json
// @Param			id		path		string							true	"条款ID"
// @Param			request	body		service.UpdateLegalDocumentReq	true	"条款"
// @Success		200		{object}	router.Resp						"操作成功"
// @Router			/api/v1/legal/documents/{id} [PUT]
func (i *LegalHandle) UpdateDocument(ctx *gin.Context) {
	var req service.UpdateLegalDocumentReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	req.ID = cast.ToUint(ctx.Param("id"))
	if err := i.svc.UpdateDocument(ctx.Request.Context(), &req); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// @Summary		删除条款
// @Description	删除没有同意记录的条款
// @Tags			条款管理
// @Produce		json
// @Param			id	path		string		true	"条款ID"
// @Success		200	{object}	router.Resp	"操作成功"
// @Router			/api/v1/legal/documents/{id} [DELETE]
func (i *LegalHandle) DeleteDocument(ctx *gin.Context) {
	if err := i.svc.DeleteDocument(ctx.Request.Context(), cast.ToUint(ctx.Param("id"))); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// @Summary		条款列表
// @Description	分页查询条款，带当前版本和所在场景
// @Tags			条款管理
// @Produce		json
// @Param			page		query		int		false	"页码"
// @Param			pageSize	query		int		false	"每页数量"
// @Param			name		query		string	false	"名称"
// @Param			kind		query		string	false	"性质"
// @Param			status		query		string	false	"状态"
// @Success		200			{object}	router.PageResp	"操作成功"
// @Router			/api/v1/legal/documents [GET]
func (i *LegalHandle) RetrieveDocuments(ctx *gin.Context) {
	var req service.RetrieveLegalDocumentReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	count, list, err := i.svc.RetrieveDocuments(ctx.Request.Context(), &req)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.PageSuccess(ctx, count, req.Page, req.PageSize, list)
}

// @Summary		条款选项
// @Description	返回全部条款，供版本和场景选择
// @Tags			条款管理
// @Produce		json
// @Success		200	{object}	router.Resp	"操作成功"
// @Router			/api/v1/legal/document-options [GET]
func (i *LegalHandle) DocumentOptions(ctx *gin.Context) {
	list, err := i.svc.DocumentOptions(ctx.Request.Context())
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, list)
}

// @Summary		创建版本草稿
// @Description	为条款新建可编辑的草稿
// @Tags			条款管理
// @Accept			json
// @Produce		json
// @Param			request	body		service.CreateLegalVersionReq	true	"草稿"
// @Success		200		{object}	router.Resp						"操作成功"
// @Router			/api/v1/legal/versions [POST]
func (i *LegalHandle) CreateVersion(ctx *gin.Context) {
	var req service.CreateLegalVersionReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	if err := i.svc.CreateVersion(ctx.Request.Context(), &req); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// @Summary		更新版本草稿
// @Description	修改草稿的版本号、标题和正文
// @Tags			条款管理
// @Accept			json
// @Produce		json
// @Param			id		path		string							true	"版本ID"
// @Param			request	body		service.UpdateLegalVersionReq	true	"草稿"
// @Success		200		{object}	router.Resp						"操作成功"
// @Router			/api/v1/legal/versions/{id} [PUT]
func (i *LegalHandle) UpdateVersion(ctx *gin.Context) {
	var req service.UpdateLegalVersionReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	req.ID = cast.ToUint(ctx.Param("id"))
	if err := i.svc.UpdateVersion(ctx.Request.Context(), &req); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// @Summary		删除版本草稿
// @Description	删除没有同意记录的草稿
// @Tags			条款管理
// @Produce		json
// @Param			id	path		string		true	"版本ID"
// @Success		200	{object}	router.Resp	"操作成功"
// @Router			/api/v1/legal/versions/{id} [DELETE]
func (i *LegalHandle) DeleteVersion(ctx *gin.Context) {
	if err := i.svc.DeleteVersion(ctx.Request.Context(), cast.ToUint(ctx.Param("id"))); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// @Summary		发布版本
// @Description	冻结草稿，并把它设为当前已发布版本
// @Tags			条款管理
// @Accept			json
// @Produce		json
// @Param			id		path		string							true	"版本ID"
// @Param			request	body		service.PublishLegalVersionReq	true	"变更级别"
// @Success		200		{object}	router.Resp						"操作成功"
// @Router			/api/v1/legal/versions/{id}/publish [POST]
func (i *LegalHandle) PublishVersion(ctx *gin.Context) {
	var req service.PublishLegalVersionReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	req.ID = cast.ToUint(ctx.Param("id"))
	if err := i.svc.PublishVersion(ctx.Request.Context(), &req); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// @Summary		版本列表
// @Description	分页查询某个条款的版本
// @Tags			条款管理
// @Produce		json
// @Param			documentId	query		string	true	"条款ID"
// @Param			page		query		int		false	"页码"
// @Param			pageSize	query		int		false	"每页数量"
// @Success		200			{object}	router.PageResp	"操作成功"
// @Router			/api/v1/legal/versions [GET]
func (i *LegalHandle) RetrieveVersions(ctx *gin.Context) {
	var req service.RetrieveLegalVersionReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	count, list, err := i.svc.RetrieveVersions(ctx.Request.Context(), &req)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.PageSuccess(ctx, count, req.Page, req.PageSize, list)
}

// @Summary		创建场景
// @Description	新建同意场景或阅读场景
// @Tags			条款管理
// @Accept			json
// @Produce		json
// @Param			request	body		service.CreateLegalSceneReq	true	"场景"
// @Success		200		{object}	router.Resp					"操作成功"
// @Router			/api/v1/legal/scenes [POST]
func (i *LegalHandle) CreateScene(ctx *gin.Context) {
	var req service.CreateLegalSceneReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	if err := i.svc.CreateScene(ctx.Request.Context(), &req); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// @Summary		更新场景
// @Description	修改场景并整组替换条目
// @Tags			条款管理
// @Accept			json
// @Produce		json
// @Param			id		path		string						true	"场景ID"
// @Param			request	body		service.UpdateLegalSceneReq	true	"场景"
// @Success		200		{object}	router.Resp					"操作成功"
// @Router			/api/v1/legal/scenes/{id} [PUT]
func (i *LegalHandle) UpdateScene(ctx *gin.Context) {
	var req service.UpdateLegalSceneReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	req.ID = cast.ToUint(ctx.Param("id"))
	if err := i.svc.UpdateScene(ctx.Request.Context(), &req); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// @Summary		删除场景
// @Description	删除场景及其条目
// @Tags			条款管理
// @Produce		json
// @Param			id	path		string		true	"场景ID"
// @Success		200	{object}	router.Resp	"操作成功"
// @Router			/api/v1/legal/scenes/{id} [DELETE]
func (i *LegalHandle) DeleteScene(ctx *gin.Context) {
	if err := i.svc.DeleteScene(ctx.Request.Context(), cast.ToUint(ctx.Param("id"))); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// @Summary		场景列表
// @Description	分页查询场景
// @Tags			条款管理
// @Produce		json
// @Param			page		query		int		false	"页码"
// @Param			pageSize	query		int		false	"每页数量"
// @Param			name		query		string	false	"名称"
// @Param			kind		query		string	false	"用途"
// @Success		200			{object}	router.PageResp	"操作成功"
// @Router			/api/v1/legal/scenes [GET]
func (i *LegalHandle) RetrieveScenes(ctx *gin.Context) {
	var req service.RetrieveLegalSceneReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	count, list, err := i.svc.RetrieveScenes(ctx.Request.Context(), &req)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.PageSuccess(ctx, count, req.Page, req.PageSize, list)
}

// @Summary		场景选项
// @Description	返回全部场景，供筛选
// @Tags			条款管理
// @Produce		json
// @Success		200	{object}	router.Resp	"操作成功"
// @Router			/api/v1/legal/scene-options [GET]
func (i *LegalHandle) SceneOptions(ctx *gin.Context) {
	list, err := i.svc.SceneOptions(ctx.Request.Context())
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, list)
}

// @Summary		同意记录
// @Description	分页查询同意、拒绝和撤回，只读
// @Tags			条款管理
// @Produce		json
// @Param			page			query		int		false	"页码"
// @Param			pageSize		query		int		false	"每页数量"
// @Param			documentKey		query		string	false	"条款标识"
// @Param			sceneKey		query		string	false	"场景标识"
// @Param			action			query		string	false	"动作"
// @Success		200				{object}	router.PageResp	"操作成功"
// @Router			/api/v1/legal/events [GET]
func (i *LegalHandle) RetrieveEvents(ctx *gin.Context) {
	var req service.RetrieveLegalEventReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	count, list, err := i.svc.RetrieveEvents(ctx.Request.Context(), &req)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.PageSuccess(ctx, count, req.Page, req.PageSize, list)
}

// @Summary		阅读场景
// @Description	手机和 PC 用同一份 JSON 获取顶栏或勾选组
// @Tags			条款管理
// @Produce		json
// @Param			key	path		string		true	"场景标识"
// @Success		200	{object}	router.Resp	"操作成功"
// @Router			/api/v1/legal/scenes/{key} [GET]
func (i *LegalHandle) PublicScene(ctx *gin.Context) {
	view, err := i.svc.PublicScene(ctx.Request.Context(), ctx.Param("key"))
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, view)
}

// @Summary		当前正文
// @Description	按条款标识返回当前已发布的 Markdown
// @Tags			条款管理
// @Produce		json
// @Param			key	path		string		true	"条款标识"
// @Success		200	{object}	router.Resp	"操作成功"
// @Router			/api/v1/legal/documents/{key} [GET]
func (i *LegalHandle) PublicDocument(ctx *gin.Context) {
	body, err := i.svc.PublicDocument(ctx.Request.Context(), ctx.Param("key"))
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, body)
}

// @Summary		历史正文
// @Description	按条款标识和版本号返回已发布或已归档正文
// @Tags			条款管理
// @Produce		json
// @Param			key		path		string		true	"条款标识"
// @Param			version	path		string		true	"版本号"
// @Success		200		{object}	router.Resp	"操作成功"
// @Router			/api/v1/legal/documents/{key}/versions/{version} [GET]
func (i *LegalHandle) PublicDocumentVersion(ctx *gin.Context) {
	body, err := i.svc.PublicDocumentVersion(ctx.Request.Context(), ctx.Param("key"), ctx.Param("version"))
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, body)
}

// @Summary		补签状态
// @Description	返回当前用户尚未接受的实质变更
// @Tags			条款管理
// @Produce		json
// @Success		200	{object}	router.Resp	"操作成功"
// @Router			/api/v1/legal/status [GET]
func (i *LegalHandle) Status(ctx *gin.Context) {
	status, err := i.svc.Status(ctx.Request.Context(), legalActor(ctx).UserID)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, status)
}

// @Summary		提交同意
// @Description	为当前用户追加同意、拒绝或撤回
// @Tags			条款管理
// @Accept			json
// @Produce		json
// @Param			request	body		service.RecordAcceptanceReq	true	"同意"
// @Success		200		{object}	router.Resp					"操作成功"
// @Router			/api/v1/legal/acceptances [POST]
func (i *LegalHandle) Accept(ctx *gin.Context) {
	var req service.RecordAcceptanceReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	list, err := i.svc.RecordAcceptances(ctx.Request.Context(), legalActor(ctx), &req)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, list)
}
