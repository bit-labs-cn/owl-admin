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

// InboxHandle 处理当前用户的站内信、待办和实时流。
type InboxHandle struct {
	svc *service.InboxService
	hub *service.Hub
}

func NewInboxHandle(svc *service.InboxService, hub *service.Hub) *InboxHandle {
	return &InboxHandle{svc: svc, hub: hub}
}
func (h *InboxHandle) ModuleName() (string, string) { return "inbox", "我的消息" }

func (h *InboxHandle) Counts(ctx *gin.Context) {
	row, err := h.svc.Counts(ctx.Request.Context())
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, row)
}

func (h *InboxHandle) List(ctx *gin.Context) {
	kind := cast.ToInt(ctx.Query("kind"))
	rows, err := h.svc.List(ctx.Request.Context(), kind, 30)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, rows)
}

func (h *InboxHandle) Retrieve(ctx *gin.Context) {
	var req service.RetrieveInboxReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	total, list, err := h.svc.Retrieve(ctx.Request.Context(), &req)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.PageSuccess(ctx, int(total), req.Page, req.PageSize, list)
}

func (h *InboxHandle) Read(ctx *gin.Context) {
	var req struct {
		IDs []string `json:"ids"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	ids := make([]uint, 0, len(req.IDs))
	for _, id := range req.IDs {
		ids = append(ids, cast.ToUint(id))
	}
	if err := h.svc.MarkRead(ctx.Request.Context(), ids); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

func (h *InboxHandle) ReadAll(ctx *gin.Context) {
	if err := h.svc.MarkAllRead(ctx.Request.Context(), cast.ToInt(ctx.Query("kind"))); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

func (h *InboxHandle) Delete(ctx *gin.Context) {
	if err := h.svc.Delete(ctx.Request.Context(), cast.ToUint(ctx.Param("id"))); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

func (h *InboxHandle) Done(ctx *gin.Context) {
	if err := h.svc.Complete(ctx.Request.Context(), cast.ToUint(ctx.Param("id")), true); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

func (h *InboxHandle) Reopen(ctx *gin.Context) {
	if err := h.svc.Complete(ctx.Request.Context(), cast.ToUint(ctx.Param("id")), false); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

func (h *InboxHandle) Ticket(ctx *gin.Context) {
	token, err := h.svc.StreamTicket(ctx.Request.Context())
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, gin.H{"ticket": token})
}

func (h *InboxHandle) Stream(ctx *gin.Context) {
	p, err := h.hub.ConsumeTicket(ctx.Request.Context(), ctx.Query("ticket"))
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
