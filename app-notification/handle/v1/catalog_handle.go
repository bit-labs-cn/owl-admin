package v1

import (
	"bit-labs.cn/owl-admin/app-notification/repository"
	"bit-labs.cn/owl-admin/app-notification/service"
	"bit-labs.cn/owl/provider/router"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// TypeHandle 处理消息类型管理。
type TypeHandle struct {
	svc *service.NotificationService
}

func NewTypeHandle(svc *service.NotificationService) *TypeHandle {
	return &TypeHandle{svc: svc}
}

func (h *TypeHandle) ModuleName() (string, string) { return "notify-type", "消息类型" }

func (h *TypeHandle) Retrieve(ctx *gin.Context) {
	page := cast.ToInt(ctx.DefaultQuery("page", "1"))
	pageSize := cast.ToInt(ctx.DefaultQuery("pageSize", "10"))
	keyword := ctx.Query("keyword")
	if keyword == "" {
		keyword = ctx.Query("name")
	}
	total, list, err := h.svc.AdminTypes(ctx.Request.Context(), repository.TypeQuery{
		Keyword: keyword, Status: cast.ToInt(ctx.Query("status")), Page: page, PageSize: pageSize,
	})
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.PageSuccess(ctx, int(total), page, pageSize, list)
}

func (h *TypeHandle) Options(ctx *gin.Context) {
	list, err := h.svc.TypeOptions(ctx.Request.Context())
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, list)
}

func (h *TypeHandle) Create(ctx *gin.Context) {
	var req typeBody
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, service.Invalid("请求参数无效"))
		return
	}
	if err := h.svc.SaveType(ctx.Request.Context(), req.input(0)); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

func (h *TypeHandle) Update(ctx *gin.Context) {
	id := cast.ToUint(ctx.Param("id"))
	if id == 0 {
		router.Fail(ctx, service.Invalid("类型 ID 无效"))
		return
	}
	var req typeBody
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, service.Invalid("请求参数无效"))
		return
	}
	if err := h.svc.SaveType(ctx.Request.Context(), req.input(id)); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

func (h *TypeHandle) Delete(ctx *gin.Context) {
	id := cast.ToUint(ctx.Param("id"))
	if id == 0 {
		router.Fail(ctx, service.Invalid("类型 ID 无效"))
		return
	}
	if err := h.svc.DeleteType(ctx.Request.Context(), id); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// TemplateHandle 处理消息模板管理。
type TemplateHandle struct {
	svc *service.NotificationService
}

func NewTemplateHandle(svc *service.NotificationService) *TemplateHandle {
	return &TemplateHandle{svc: svc}
}

func (h *TemplateHandle) ModuleName() (string, string) { return "notify-template", "消息模板" }

func (h *TemplateHandle) Retrieve(ctx *gin.Context) {
	page := cast.ToInt(ctx.DefaultQuery("page", "1"))
	pageSize := cast.ToInt(ctx.DefaultQuery("pageSize", "10"))
	total, list, err := h.svc.AdminTemplates(ctx.Request.Context(), repository.TemplateQuery{
		TypeCode: ctx.Query("typeCode"), Page: page, PageSize: pageSize,
	})
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.PageSuccess(ctx, int(total), page, pageSize, list)
}

func (h *TemplateHandle) Create(ctx *gin.Context) {
	var req templateBody
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, service.Invalid("请求参数无效"))
		return
	}
	if err := h.svc.SaveTemplate(ctx.Request.Context(), req.input(0)); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

func (h *TemplateHandle) Update(ctx *gin.Context) {
	id := cast.ToUint(ctx.Param("id"))
	if id == 0 {
		router.Fail(ctx, service.Invalid("模板 ID 无效"))
		return
	}
	var req templateBody
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, service.Invalid("请求参数无效"))
		return
	}
	if err := h.svc.SaveTemplate(ctx.Request.Context(), req.input(id)); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

func (h *TemplateHandle) Delete(ctx *gin.Context) {
	id := cast.ToUint(ctx.Param("id"))
	if id == 0 {
		router.Fail(ctx, service.Invalid("模板 ID 无效"))
		return
	}
	if err := h.svc.DeleteTemplate(ctx.Request.Context(), id); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

func (h *TemplateHandle) Preview(ctx *gin.Context) {
	var req struct {
		Title string         `json:"title"`
		Body  string         `json:"body"`
		Link  string         `json:"link"`
		Vars  map[string]any `json:"vars"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, service.Invalid("请求参数无效"))
		return
	}
	copy, err := h.svc.PreviewTemplate(req.Title, req.Body, req.Link, req.Vars)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, gin.H{"title": copy.Title, "body": copy.Content, "link": copy.Link})
}

// DeliveryHandle 处理发送记录。
type DeliveryHandle struct {
	svc *service.NotificationService
}

func NewDeliveryHandle(svc *service.NotificationService) *DeliveryHandle {
	return &DeliveryHandle{svc: svc}
}

func (h *DeliveryHandle) ModuleName() (string, string) { return "notify-delivery", "发送记录" }

func (h *DeliveryHandle) Retrieve(ctx *gin.Context) {
	page := cast.ToInt(ctx.DefaultQuery("page", "1"))
	pageSize := cast.ToInt(ctx.DefaultQuery("pageSize", "10"))
	total, list, err := h.svc.AdminDeliveryBatches(ctx.Request.Context(), repository.DeliveryQuery{
		Status: cast.ToInt(ctx.Query("status")), Channel: ctx.Query("channel"),
		RecipientType: cast.ToInt(ctx.Query("recipientType")),
		Page:          page, PageSize: pageSize,
	})
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.PageSuccess(ctx, int(total), page, pageSize, list)
}

// Detail 返回一次发送下全部渠道的投递明细。
func (h *DeliveryHandle) Detail(ctx *gin.Context) {
	messageID := cast.ToUint(ctx.Param("messageId"))
	if messageID == 0 {
		router.Fail(ctx, service.Invalid("消息 ID 无效"))
		return
	}
	_, list, err := h.svc.AdminDeliveries(ctx.Request.Context(), repository.DeliveryQuery{MessageID: messageID})
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, list)
}

func (h *DeliveryHandle) Retry(ctx *gin.Context) {
	id := cast.ToUint(ctx.Param("id"))
	if id == 0 {
		router.Fail(ctx, service.Invalid("发送记录 ID 无效"))
		return
	}
	if err := h.svc.RetryDelivery(ctx.Request.Context(), id); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

type typeBody struct {
	Code            string   `json:"code"`
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	Kind            int      `json:"kind"`
	DefaultChannels []string `json:"defaultChannels"`
	Status          int      `json:"status"`
}

func (b typeBody) input(id uint) service.TypeInput {
	return service.TypeInput{
		ID:              id,
		Code:            b.Code,
		Name:            b.Name,
		Description:     b.Description,
		Kind:            b.Kind,
		DefaultChannels: b.DefaultChannels,
		Status:          b.Status,
	}
}

type templateBody struct {
	TypeCode string `json:"typeCode"`
	Title    string `json:"title"`
	Body     string `json:"body"`
	Link     string `json:"link"`
}

func (b templateBody) input(id uint) service.TemplateInput {
	return service.TemplateInput{
		ID:       id,
		TypeCode: b.TypeCode,
		Title:    b.Title,
		Body:     b.Body,
		Link:     b.Link,
	}
}
