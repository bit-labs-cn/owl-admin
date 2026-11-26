package v1

import (
	"bit-labs.cn/owl-admin/app-notification/contract"
	"bit-labs.cn/owl-admin/app-notification/model"
	"bit-labs.cn/owl-admin/app-notification/service"
	"bit-labs.cn/owl/provider/router"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// AnnouncementHandle 处理公告管理与用户端公告拉取。
type AnnouncementHandle struct {
	svc *service.NotificationService
}

func NewAnnouncementHandle(svc *service.NotificationService) *AnnouncementHandle {
	return &AnnouncementHandle{svc: svc}
}

func (h *AnnouncementHandle) ModuleName() (string, string) { return "announcement", "站点公告" }

// Mine 获取当前用户可见且未关闭的有效公告（读扩散模式）
func (h *AnnouncementHandle) Mine(ctx *gin.Context) {
	uid := getUserID(ctx)
	if uid == 0 {
		router.Fail(ctx, service.NotAllowed("未登录"))
		return
	}
	page := cast.ToInt(ctx.DefaultQuery("page", "1"))
	pageSize := cast.ToInt(ctx.DefaultQuery("pageSize", "20"))
	total, list, err := h.svc.MyAnnouncements(ctx.Request.Context(), uid, page, pageSize)
	if err != nil {
		router.Fail(ctx, err)
		return
	}

	router.PageSuccess(ctx, int(total), page, pageSize, list)
}

// Dismiss 用户主动关闭/知晓某条公告
func (h *AnnouncementHandle) Dismiss(ctx *gin.Context) {
	uid := getUserID(ctx)
	if uid == 0 {
		router.Fail(ctx, service.NotAllowed("未登录"))
		return
	}
	annID := cast.ToUint(ctx.Param("id"))
	if annID == 0 {
		router.Fail(ctx, service.Invalid("公告 ID 无效"))
		return
	}
	if err := h.svc.DismissAnnouncement(ctx.Request.Context(), annID, uid); err != nil {
		router.Fail(ctx, err)
		return
	}

	router.Success(ctx, nil)
}

// --- 管理后台接口 ---

// Retrieve 管理员分页检索公告
func (h *AnnouncementHandle) Retrieve(ctx *gin.Context) {
	page := cast.ToInt(ctx.DefaultQuery("page", "1"))
	pageSize := cast.ToInt(ctx.DefaultQuery("pageSize", "10"))
	keyword := ctx.Query("keyword")
	if keyword == "" {
		keyword = ctx.Query("title")
	}
	status := cast.ToInt(ctx.Query("status"))
	total, list, err := h.svc.AdminAnnouncements(ctx.Request.Context(), page, pageSize, keyword, status)
	if err != nil {
		router.Fail(ctx, err)
		return
	}

	router.PageSuccess(ctx, int(total), page, pageSize, list)
}

// Detail 管理员查看公告详情
func (h *AnnouncementHandle) Detail(ctx *gin.Context) {
	id := cast.ToUint(ctx.Param("id"))
	row, err := h.svc.AdminAnnouncementDetail(ctx.Request.Context(), id)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, row)
}

// ReadState 返回公告的已读汇总与每个接收人的状态。
func (h *AnnouncementHandle) ReadState(ctx *gin.Context) {
	id := cast.ToUint(ctx.Param("id"))
	if id == 0 {
		router.Fail(ctx, service.Invalid("公告 ID 无效"))
		return
	}
	view, err := h.svc.AdminAnnouncementRead(ctx.Request.Context(), id)
	if err != nil {
		router.Fail(ctx, err)
		return
	}

	router.Success(ctx, view)
}

// Create 管理员创建并广播公告
func (h *AnnouncementHandle) Create(ctx *gin.Context) {
	var req struct {
		Title        string `json:"title"`
		Content      string `json:"content"`
		Level        int    `json:"level"`
		DisplayMode  int    `json:"displayMode"`
		TargetScope  string `json:"targetScope"`
		TargetValues []uint `json:"targetValues"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	senderID := getUserID(ctx)
	err := h.svc.BroadcastAnnouncement(ctx.Request.Context(), &contract.AnnouncementMessage{
		Title:        req.Title,
		Content:      req.Content,
		Level:        req.Level,
		DisplayMode:  req.DisplayMode,
		Scope:        req.TargetScope,
		TargetValues: req.TargetValues,
		SenderID:     senderID,
	})
	if err != nil {
		router.Fail(ctx, err)
		return
	}

	router.Success(ctx, nil)
}

// Update 管理员编辑公告
func (h *AnnouncementHandle) Update(ctx *gin.Context) {
	id := cast.ToUint(ctx.Param("id"))
	if id == 0 {
		router.Fail(ctx, service.Invalid("公告 ID 无效"))
		return
	}
	var req model.NotifyAnnouncement
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	req.ID = id

	if err := h.svc.SaveAnnouncement(ctx.Request.Context(), &req); err != nil {
		router.Fail(ctx, err)
		return
	}

	router.Success(ctx, nil)
}

// Delete 管理员删除公告
func (h *AnnouncementHandle) Delete(ctx *gin.Context) {
	id := cast.ToUint(ctx.Param("id"))
	if err := h.svc.DeleteAnnouncement(ctx.Request.Context(), id); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// Publish 上线公告
func (h *AnnouncementHandle) Publish(ctx *gin.Context) {
	id := cast.ToUint(ctx.Param("id"))
	ann, err := h.svc.AdminAnnouncementDetail(ctx.Request.Context(), id)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	ann.Status = model.AnnouncementStatusPublished

	if err := h.svc.SaveAnnouncement(ctx.Request.Context(), ann); err != nil {
		router.Fail(ctx, err)
		return
	}

	router.Success(ctx, nil)
}

// Offline 下线公告
func (h *AnnouncementHandle) Offline(ctx *gin.Context) {
	id := cast.ToUint(ctx.Param("id"))
	ann, err := h.svc.AdminAnnouncementDetail(ctx.Request.Context(), id)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	ann.Status = model.AnnouncementStatusOffline

	if err := h.svc.SaveAnnouncement(ctx.Request.Context(), ann); err != nil {
		router.Fail(ctx, err)
		return
	}

	router.Success(ctx, nil)
}
