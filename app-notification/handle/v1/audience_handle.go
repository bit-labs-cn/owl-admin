package v1

import (
	"strings"

	"bit-labs.cn/owl-admin/app-notification/service"
	"bit-labs.cn/owl/provider/router"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// AudienceHandle 提供受众选择和深链页面选项。这些接口只要求登录，
// 因为任何能发通知的人都必须能看懂自己选的是谁。
type AudienceHandle struct {
	resolver *service.AudienceResolver
}

func NewAudienceHandle(resolver *service.AudienceResolver) *AudienceHandle {
	return &AudienceHandle{resolver: resolver}
}

func (h *AudienceHandle) ModuleName() (string, string) { return "audience", "受众选择" }

func (h *AudienceHandle) Options(ctx *gin.Context) {
	options, err := h.resolver.Options(ctx.Request.Context())
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, options)
}

func (h *AudienceHandle) Users(ctx *gin.Context) {
	if raw := strings.TrimSpace(ctx.Query("ids")); raw != "" {
		ids := make([]uint, 0)
		for _, part := range strings.Split(raw, ",") {
			if id := cast.ToUint(strings.TrimSpace(part)); id > 0 {
				ids = append(ids, id)
			}
		}
		rows, err := h.resolver.UsersByIDs(ctx.Request.Context(), ids)
		if err != nil {
			router.Fail(ctx, err)
			return
		}
		router.Success(ctx, rows)
		return
	}
	rows, err := h.resolver.SearchUsers(ctx.Request.Context(), ctx.Query("keyword"), cast.ToInt(ctx.Query("limit")))
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, rows)
}

func (h *AudienceHandle) Links(ctx *gin.Context) {
	rows, err := h.resolver.Links(ctx.Request.Context())
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, rows)
}

func (h *AudienceHandle) SmsTemplates(ctx *gin.Context) {
	rows, err := h.resolver.SmsTemplates(ctx.Request.Context())
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, rows)
}
