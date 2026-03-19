package v1

import (
	"bit-labs.cn/owl-admin/app/service"
	"bit-labs.cn/owl/provider/router"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

var _ router.Handler = (*DictHandle)(nil)
var _ router.CrudHandler = (*DictHandle)(nil)

type DictHandle struct {
	dictSvc *service.DictService
}

func (i *DictHandle) ModuleName() (string, string) {
	return "dict", "字典管理"
}

func NewDictHandle(dictService *service.DictService) *DictHandle {
	return &DictHandle{
		dictSvc: dictService,
	}
}

// @Summary		创建字典
// @Description	创建新的字典
// @Tags			字典管理
// @Accept			json
// @Produce		json
// @Param			request	body		service.CreateDictReq	true	"字典创建请求"
// @Success		200		{object}	router.Resp				"操作成功"
// @Failure		400		{object}	router.Resp				"参数错误"
// @Failure		500		{object}	router.Resp				"服务器内部错误"
// @Router			/api/v1/dict [POST]
func (i *DictHandle) Create(ctx *gin.Context) {
	var req service.CreateDictReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}

	err := i.dictSvc.CreateDict(ctx.Request.Context(), &req)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// @Summary		删除字典
// @Description	根据字典ID删除字典
// @Tags			字典管理
// @Produce		json
// @Param			id	path		string		true	"字典ID"
// @Success		200	{object}	router.Resp	"操作成功"
// @Failure		500	{object}	router.Resp	"服务器内部错误"
// @Router			/api/v1/dict/{id} [DELETE]
func (i *DictHandle) Delete(ctx *gin.Context) {

	id := ctx.Param("id")
	err := i.dictSvc.DeleteDict(ctx.Request.Context(), id)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

func (i *DictHandle) Detail(ctx *gin.Context) {

}

// @Summary		更新字典
// @Description	根据字典ID更新字典信息
// @Tags			字典管理
// @Accept			json
// @Produce		json
// @Param			id		path		int						true	"字典ID"
// @Param			request	body		service.UpdateDictReq	true	"字典更新请求"
// @Success		200		{object}	router.Resp				"操作成功"
// @Failure		400		{object}	router.Resp				"参数错误"
// @Failure		500		{object}	router.Resp				"服务器内部错误"
// @Router			/api/v1/dict/{id} [PUT]
func (i *DictHandle) Update(ctx *gin.Context) {
	var req service.UpdateDictReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	id := cast.ToUint(ctx.Param("id"))
	req.ID = id

	err := i.dictSvc.UpdateDict(ctx.Request.Context(), &req)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// @Summary		获取字典列表
// @Description	分页获取字典列表
// @Tags			字典管理
// @Produce		json
// @Param			page		query		int				false	"页码"
// @Param			pageSize	query		int				false	"每页数量"
// @Param			nameLike	query		string			false	"名称模糊搜索"
// @Param			statusIn	query		string			false	"状态 in 查询"
// @Param			type		query		string			false	"类型"
// @Success		200			{object}	router.PageResp	"操作成功"
// @Failure		500			{object}	router.Resp		"服务器内部错误"
// @Router			/api/v1/dict [GET]
func (i *DictHandle) Retrieve(ctx *gin.Context) {
	var req service.RetrieveDictReq
	if err := ctx.ShouldBind(&req); err != nil {
		router.BadRequest(ctx, "参数绑定失败")
		return
	}

	_, list, err := i.dictSvc.RetrieveDicts(ctx.Request.Context(), &req)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, list)
}

// @Summary		创建字典项
// @Description	为指定字典创建新的字典项
// @Tags			字典管理
// @Accept			json
// @Produce		json
// @Param			id		path		int							true	"字典ID"
// @Param			request	body		service.CreateDictItemReq	true	"字典项创建请求"
// @Success		200		{object}	router.Resp					"操作成功"
// @Failure		400		{object}	router.Resp					"参数错误"
// @Failure		500		{object}	router.Resp					"服务器内部错误"
// @Router			/api/v1/dict/{id}/item [POST]
func (i *DictHandle) CreateItem(ctx *gin.Context) {
	var req service.CreateDictItemReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	req.DictID = cast.ToUint(ctx.Param("id"))
	err := i.dictSvc.CreateItem(ctx.Request.Context(), &req)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// @Summary		更新字典项
// @Description	更新指定字典项信息
// @Tags			字典管理
// @Accept			json
// @Produce		json
// @Param			id		path		int							true	"字典ID"
// @Param			itemID	path		int							true	"字典项ID"
// @Param			request	body		service.UpdateDictItemReq	true	"字典项更新请求"
// @Success		200		{object}	router.Resp					"操作成功"
// @Failure		400		{object}	router.Resp					"参数错误"
// @Failure		500		{object}	router.Resp					"服务器内部错误"
// @Router			/api/v1/dict/{id}/item/{itemID} [PUT]
func (i *DictHandle) UpdateItem(ctx *gin.Context) {
	var req service.UpdateDictItemReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	req.DictID = cast.ToUint(ctx.Param("id"))
	req.ID = cast.ToUint(ctx.Param("itemID"))
	err := i.dictSvc.UpdateItem(ctx.Request.Context(), &req)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// @Summary		获取字典项列表
// @Description	获取指定字典的所有字典项
// @Tags			字典管理
// @Produce		json
// @Param			id	path		string		true	"字典ID"
// @Success		200	{object}	router.Resp	"操作成功"
// @Failure		500	{object}	router.Resp	"服务器内部错误"
// @Router			/api/v1/dict/{id}/item [GET]
func (i *DictHandle) RetrieveItems(ctx *gin.Context) {
	dictID := ctx.Param("id")
	_, list, err := i.dictSvc.RetrieveItems(ctx.Request.Context(), dictID)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, list)
}

// @Summary		按类型获取字典项
// @Description	按字典类型获取启用的字典项列表（仅需登录）
// @Tags			字典管理
// @Produce		json
// @Param			type	path		string		true	"字典类型"
// @Success		200		{object}	router.Resp	"操作成功"
// @Failure		500		{object}	router.Resp	"服务器内部错误"
// @Router			/api/v1/dict/types/{type}/items [GET]
func (i *DictHandle) GetItemsByType(ctx *gin.Context) {
	dictType := ctx.Param("type")
	list, err := i.dictSvc.GetDictByType(ctx.Request.Context(), dictType)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, list)
}

// @Summary		删除字典项
// @Description	删除指定字典的指定字典项
// @Tags			字典管理
// @Produce		json
// @Param			id		path		string		true	"字典ID"
// @Param			itemID	path		string		true	"字典项ID"
// @Success		200		{object}	router.Resp	"操作成功"
// @Failure		500		{object}	router.Resp	"服务器内部错误"
// @Router			/api/v1/dict/{id}/item/{itemID} [DELETE]
func (i *DictHandle) DeleteItem(ctx *gin.Context) {
	dictID := ctx.Param("id")
	itemID := ctx.Param("itemID")
	err := i.dictSvc.DeleteItems(ctx.Request.Context(), dictID, itemID)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}
