package v1

import (
	"fmt"
	"net/http"
	"time"

	"bit-labs.cn/owl-admin/app-notification/service"
	"bit-labs.cn/owl/provider/router"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// MessageHandle 处理站内消息、待办与实时 SSE 推送。
type MessageHandle struct {
	svc *service.NotificationService
	hub *service.Hub
}

func NewMessageHandle(svc *service.NotificationService, hub *service.Hub) *MessageHandle {
	return &MessageHandle{svc: svc, hub: hub}
}

func (h *MessageHandle) ModuleName() (string, string) { return "inbox", "我的消息" }

func getUserID(ctx *gin.Context) uint {
	if v, exists := ctx.Get("user_id"); exists {
		return cast.ToUint(v)
	}
	return cast.ToUint(ctx.Request.Context().Value("user_id"))
}

// Counts 获取当前用户的未读通知与待办数
func (h *MessageHandle) Counts(ctx *gin.Context) {
	uid := getUserID(ctx)
	if uid == 0 {
		router.Fail(ctx, service.NotAllowed("未登录"))
		return
	}
	unreadNotice, pendingTodo, err := h.svc.MyCounts(ctx.Request.Context(), uid)
	if err != nil {
		router.Fail(ctx, err)
		return
	}

	router.Success(ctx, gin.H{
		"unreadNotice": unreadNotice,
		"pendingTodo":  pendingTodo,
		"notice":       unreadNotice,
		"todo":         pendingTodo,
		"message":      0,
		"total":        unreadNotice + pendingTodo,
	})
}

// Retrieve 分页查询当前用户的消息（可根据 kind/readStatus 筛选）
func (h *MessageHandle) Retrieve(ctx *gin.Context) {
	uid := getUserID(ctx)
	if uid == 0 {
		router.Fail(ctx, service.NotAllowed("未登录"))
		return
	}

	page := cast.ToInt(ctx.DefaultQuery("page", "1"))
	pageSize := cast.ToInt(ctx.DefaultQuery("pageSize", "20"))
	kind := cast.ToInt(ctx.Query("kind")) // 0: all, 1: notice, 2: todo
	var readStatus *int
	if rsStr := ctx.Query("readStatus"); rsStr != "" {
		rs := cast.ToInt(rsStr)
		readStatus = &rs
	}
	typeCode := ctx.Query("typeCode")

	total, list, err := h.svc.MyMessages(ctx.Request.Context(), uid, kind, readStatus, typeCode, page, pageSize)
	if err != nil {
		router.Fail(ctx, err)
		return
	}

	router.PageSuccess(ctx, int(total), page, pageSize, list)
}

// Read 批量标记消息已读
func (h *MessageHandle) Read(ctx *gin.Context) {
	uid := getUserID(ctx)
	if uid == 0 {
		router.Fail(ctx, service.NotAllowed("未登录"))
		return
	}

	var req struct {
		IDs []any `json:"ids"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	recipientIDs := make([]uint, 0, len(req.IDs))
	for _, id := range req.IDs {
		if u := cast.ToUint(id); u > 0 {
			recipientIDs = append(recipientIDs, u)
		}
	}
	if err := h.svc.MarkRead(ctx.Request.Context(), uid, recipientIDs); err != nil {
		router.Fail(ctx, err)
		return
	}

	router.Success(ctx, nil)
}

// ReadAll 一键全部标记已读
func (h *MessageHandle) ReadAll(ctx *gin.Context) {
	uid := getUserID(ctx)
	if uid == 0 {
		router.Fail(ctx, service.NotAllowed("未登录"))
		return
	}
	if err := h.svc.MarkAllRead(ctx.Request.Context(), uid, cast.ToInt(ctx.Query("kind"))); err != nil {
		router.Fail(ctx, err)
		return
	}

	router.Success(ctx, nil)
}

// Complete 业务/用户办结待办（根据 bizKey）
func (h *MessageHandle) Complete(ctx *gin.Context) {
	bizKey := ctx.Query("bizKey")
	if bizKey == "" {
		var req struct {
			BizKey string `json:"bizKey"`
		}
		_ = ctx.ShouldBindJSON(&req)
		bizKey = req.BizKey
	}
	if bizKey == "" {
		router.Fail(ctx, service.Invalid("缺少业务标识 bizKey"))
		return
	}
	if err := h.svc.CompleteByBizKey(ctx.Request.Context(), bizKey); err != nil {
		router.Fail(ctx, err)
		return
	}

	router.Success(ctx, nil)
}

// Done 页面点击完成待办（根据流水分发记录 id）
func (h *MessageHandle) Done(ctx *gin.Context) {
	uid := getUserID(ctx)
	if uid == 0 {
		router.Fail(ctx, service.NotAllowed("未登录"))
		return
	}
	id := cast.ToUint(ctx.Param("id"))
	if id == 0 {
		router.Fail(ctx, service.Invalid("消息记录 ID 无效"))
		return
	}
	if err := h.svc.CompleteByID(ctx.Request.Context(), uid, id); err != nil {
		router.Fail(ctx, err)
		return
	}

	router.Success(ctx, nil)
}

// Reopen 页面点击重开待办
func (h *MessageHandle) Reopen(ctx *gin.Context) {
	uid := getUserID(ctx)
	if uid == 0 {
		router.Fail(ctx, service.NotAllowed("未登录"))
		return
	}
	id := cast.ToUint(ctx.Param("id"))
	if id == 0 {
		router.Fail(ctx, service.Invalid("消息记录 ID 无效"))
		return
	}
	if err := h.svc.ReopenByID(ctx.Request.Context(), uid, id); err != nil {
		router.Fail(ctx, err)
		return
	}

	router.Success(ctx, nil)
}

// Delete 用户删除自己的消息记录
func (h *MessageHandle) Delete(ctx *gin.Context) {
	uid := getUserID(ctx)
	if uid == 0 {
		router.Fail(ctx, service.NotAllowed("未登录"))
		return
	}
	id := cast.ToUint(ctx.Param("id"))
	if id == 0 {
		router.Fail(ctx, service.Invalid("消息记录 ID 无效"))
		return
	}

	if err := h.svc.DeleteUserMessage(ctx.Request.Context(), uid, id); err != nil {
		router.Fail(ctx, err)
		return
	}

	router.Success(ctx, nil)
}

// Ticket 签发 SSE 实时推送临时连接票据
func (h *MessageHandle) Ticket(ctx *gin.Context) {
	uid := getUserID(ctx)
	if uid == 0 {
		router.Fail(ctx, service.NotAllowed("未登录"))
		return
	}
	token, err := h.hub.IssueTicket(ctx.Request.Context(), service.Principal{Type: 1, ID: uid})
	if err != nil {
		router.Fail(ctx, err)
		return
	}

	router.Success(ctx, gin.H{"ticket": token})
}

// Stream 浏览器 SSE 长连接接入
func (h *MessageHandle) Stream(ctx *gin.Context) {
	ticket := ctx.Query("ticket")
	if ticket == "" {
		router.Fail(ctx, service.TicketInvalid())
		return
	}
	p, err := h.hub.ConsumeTicket(ctx.Request.Context(), ticket)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	ch, cancel := h.hub.Subscribe(p.Type, p.ID)
	defer cancel()

	ctx.Writer.Header().Set("Content-Type", "text/event-stream")
	ctx.Writer.Header().Set("Cache-Control", "no-cache")
	ctx.Writer.Header().Set("Connection", "keep-alive")
	ctx.Writer.Header().Set("X-Accel-Buffering", "no")
	ctx.Status(http.StatusOK)
	flusher, ok := ctx.Writer.(http.Flusher)
	if !ok {
		return
	}
	fmt.Fprintf(ctx.Writer, "retry: 3000\n\n")
	flusher.Flush()
	heart := time.NewTicker(20 * time.Second)
	defer heart.Stop()

	var seq uint64
	for {
		select {
		case <-ctx.Request.Context().Done():
			return
		case <-heart.C:
			fmt.Fprintf(ctx.Writer, ": ping\n\n")
			flusher.Flush()
		case raw, open := <-ch:
			if !open {
				return
			}
			seq++
			fmt.Fprintf(ctx.Writer, "id: %d\nevent: message\ndata: %s\n\n", seq, raw)
			flusher.Flush()
		}
	}
}
