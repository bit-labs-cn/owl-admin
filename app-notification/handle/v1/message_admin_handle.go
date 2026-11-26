package v1

import (
	"bit-labs.cn/owl-admin/app-notification/repository"
	"bit-labs.cn/owl-admin/app-notification/service"
	"bit-labs.cn/owl/provider/router"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// MessageAdminHandle 处理管理端的消息记录与接收人已读情况。
type MessageAdminHandle struct {
	svc *service.NotificationService
}

func NewMessageAdminHandle(svc *service.NotificationService) *MessageAdminHandle {
	return &MessageAdminHandle{svc: svc}
}

func (h *MessageAdminHandle) ModuleName() (string, string) { return "notify-message", "消息记录" }

// Retrieve 分页列出已发消息及已读、未读和待办统计。
func (h *MessageAdminHandle) Retrieve(ctx *gin.Context) {
	page := cast.ToInt(ctx.DefaultQuery("page", "1"))
	pageSize := cast.ToInt(ctx.DefaultQuery("pageSize", "10"))
	keyword := ctx.Query("keyword")
	if keyword == "" {
		keyword = ctx.Query("title")
	}
	total, list, err := h.svc.AdminMessages(ctx.Request.Context(), repository.MessageQuery{
		Keyword:  keyword,
		TypeCode: ctx.Query("typeCode"),
		Kind:     cast.ToInt(ctx.Query("kind")),
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		router.Fail(ctx, err)
		return
	}

	router.PageSuccess(ctx, int(total), page, pageSize, list)
}

// Detail 返回消息正文与每个接收人的已读、待办状态。
func (h *MessageAdminHandle) Detail(ctx *gin.Context) {
	id := cast.ToUint(ctx.Param("id"))
	if id == 0 {
		router.Fail(ctx, service.Invalid("消息 ID 无效"))
		return
	}
	detail, err := h.svc.AdminMessageDetail(ctx.Request.Context(), id)
	if err != nil {
		router.Fail(ctx, err)
		return
	}

	router.Success(ctx, detail)
}
