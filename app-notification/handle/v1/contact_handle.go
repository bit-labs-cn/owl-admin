package v1

import (
	"net/http"

	"bit-labs.cn/owl-admin/app-notification/service"
	"bit-labs.cn/owl/provider/router"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// ContactHandle 处理外部名单和退订。
type ContactHandle struct{ svc *service.ContactService }

func NewContactHandle(svc *service.ContactService) *ContactHandle { return &ContactHandle{svc: svc} }
func (h *ContactHandle) ModuleName() (string, string)             { return "contact", "外部名单" }

func (h *ContactHandle) SaveList(ctx *gin.Context) {
	var req service.CreateContactListReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	id := cast.ToUint(ctx.Param("id"))
	if err := h.svc.SaveList(ctx.Request.Context(), id, &req); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

func (h *ContactHandle) DeleteList(ctx *gin.Context) {
	if err := h.svc.DeleteList(ctx.Request.Context(), cast.ToUint(ctx.Param("id"))); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

func (h *ContactHandle) Lists(ctx *gin.Context) {
	total, list, err := h.svc.Lists(ctx.Request.Context(), cast.ToInt(ctx.Query("page")), cast.ToInt(ctx.Query("pageSize")), ctx.Query("name"))
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.PageSuccess(ctx, int(total), cast.ToInt(ctx.Query("page")), cast.ToInt(ctx.Query("pageSize")), list)
}

func (h *ContactHandle) SaveContact(ctx *gin.Context) {
	var req service.CreateContactReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	if err := h.svc.SaveContact(ctx.Request.Context(), cast.ToUint(ctx.Param("id")), &req); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

func (h *ContactHandle) Contacts(ctx *gin.Context) {
	var req service.RetrieveContactReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	total, list, err := h.svc.Contacts(ctx.Request.Context(), &req)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.PageSuccess(ctx, int(total), req.Page, req.PageSize, list)
}

func (h *ContactHandle) DeleteContact(ctx *gin.Context) {
	if err := h.svc.DeleteContact(ctx.Request.Context(), cast.ToUint(ctx.Param("id"))); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

func (h *ContactHandle) Import(ctx *gin.Context) {
	var req struct {
		ListID uint   `json:"listId,string"`
		Text   string `json:"text"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	n, err := h.svc.ImportContacts(ctx.Request.Context(), req.ListID, req.Text)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, gin.H{"created": n})
}

func (h *ContactHandle) Unsubscribe(ctx *gin.Context) {
	id := cast.ToUint(ctx.Query("token"))
	if err := h.svc.Unsubscribe(ctx.Request.Context(), id); err != nil {
		ctx.String(http.StatusBadRequest, "退订失败")
		return
	}
	ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte("<p>已退订，后续营销通知不会再发给你。</p>"))
}
