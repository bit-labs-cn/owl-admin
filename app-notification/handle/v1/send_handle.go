package v1

import (
	"strings"
	"time"

	"bit-labs.cn/owl-admin/app-notification/contract"
	"bit-labs.cn/owl-admin/app-notification/repository"
	"bit-labs.cn/owl-admin/app-notification/service"
	"bit-labs.cn/owl/provider/router"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// SendHandle 处理管理端人工发送通知与待办。
type SendHandle struct {
	svc *service.NotificationService
}

func NewSendHandle(svc *service.NotificationService) *SendHandle {
	return &SendHandle{svc: svc}
}

func (h *SendHandle) ModuleName() (string, string) { return "notify-send", "发送通知" }

func toIDs(values []any) []uint {
	out := make([]uint, 0, len(values))
	for _, value := range values {
		if id := cast.ToUint(value); id > 0 {
			out = append(out, id)
		}
	}
	return out
}

// Audience 返回角色、部门树和用户组选项。
func (h *SendHandle) Audience(ctx *gin.Context) {
	options, err := h.svc.AudienceOptions(ctx.Request.Context())
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, options)
}

// Recipients 搜索收件人，或按 ID 回显已选收件人。
func (h *SendHandle) Recipients(ctx *gin.Context) {
	keyword := ctx.Query("keyword")
	userIDs := make([]uint, 0)
	for _, raw := range ctx.QueryArray("ids") {
		for _, part := range strings.Split(raw, ",") {
			if id := cast.ToUint(strings.TrimSpace(part)); id > 0 {
				userIDs = append(userIDs, id)
			}
		}
	}
	list, err := h.svc.RecipientOptions(ctx.Request.Context(), keyword, userIDs)
	if err != nil {
		router.Fail(ctx, err)
		return
	}

	router.Success(ctx, list)
}

// EstimateAudience 预估当前受众条件下的接收人数，供发送页实时显示。
func (h *SendHandle) EstimateAudience(ctx *gin.Context) {
	var req struct {
		UserIDs          []any                        `json:"userIds"`
		RoleIDs          []any                        `json:"roleIds"`
		DeptIDs          []any                        `json:"deptIds"`
		GroupIDs         []any                        `json:"groupIds"`
		IncludeDeptChild bool                         `json:"includeDeptChild"`
		External         []contract.ExternalRecipient `json:"externals"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, service.Invalid("请求参数无效"))
		return
	}
	estimate, err := h.svc.EstimateAudience(ctx.Request.Context(), repository.AudienceSelection{
		UserIDs:      toIDs(req.UserIDs),
		RoleIDs:      toIDs(req.RoleIDs),
		DeptIDs:      toIDs(req.DeptIDs),
		GroupIDs:     toIDs(req.GroupIDs),
		IncludeChild: req.IncludeDeptChild,
	}, req.External)
	if err != nil {
		router.Fail(ctx, err)
		return
	}

	router.Success(ctx, estimate)
}

// Send 人工发送一条通知或待办。
func (h *SendHandle) Send(ctx *gin.Context) {
	var req struct {
		Kind             int                          `json:"kind"`
		TypeCode         string                       `json:"typeCode"`
		Title            string                       `json:"title"`
		Content          string                       `json:"content"`
		Link             string                       `json:"link"`
		Channels         []string                     `json:"channels"`
		UserIDs          []any                        `json:"userIds"`
		RoleIDs          []any                        `json:"roleIds"`
		DeptIDs          []any                        `json:"deptIds"`
		GroupIDs         []any                        `json:"groupIds"`
		IncludeDeptChild bool                         `json:"includeDeptChild"`
		External         []contract.ExternalRecipient `json:"externals"`
		DueAt            string                       `json:"dueAt"`
		PlanAt           string                       `json:"planAt"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, service.Invalid("请求参数无效"))
		return
	}

	var dueAt *time.Time
	if req.DueAt != "" {
		parsed, err := time.Parse(time.RFC3339, req.DueAt)
		if err != nil {
			router.Fail(ctx, service.Invalid("截止时间格式无效"))
			return
		}
		dueAt = &parsed
	}

	var planAt *time.Time
	if req.PlanAt != "" {
		parsed, err := time.Parse(time.RFC3339, req.PlanAt)
		if err != nil {
			router.Fail(ctx, service.Invalid("定时发送时间格式无效"))
			return
		}
		planAt = &parsed
	}

	scheduled, err := h.svc.ManualSend(ctx.Request.Context(), service.ManualSendInput{
		Kind:               req.Kind,
		TypeCode:           req.TypeCode,
		Title:              req.Title,
		Content:            req.Content,
		Link:               req.Link,
		Channels:           req.Channels,
		UserIDs:            toIDs(req.UserIDs),
		RoleIDs:            toIDs(req.RoleIDs),
		DeptIDs:            toIDs(req.DeptIDs),
		GroupIDs:           toIDs(req.GroupIDs),
		IncludeDeptChild:   req.IncludeDeptChild,
		ExternalRecipients: req.External,
		DueAt:              dueAt,
		PlanAt:             planAt,
		SenderID:           getUserID(ctx),
	})
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	if scheduled != nil {
		router.Success(ctx, gin.H{
			"scheduled": true,
			"id":        scheduled.ID,
			"planAt":    scheduled.PlanAt,
		})
		return
	}

	router.Success(ctx, gin.H{"scheduled": false})
}
