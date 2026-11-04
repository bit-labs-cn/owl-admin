package v1

import (
	"bit-labs.cn/owl-admin/app/service"
	"bit-labs.cn/owl/provider/router"
	"github.com/gin-gonic/gin"
)

var _ router.Handler = (*ServerMonitorHandle)(nil)

// ServerMonitorHandle 提供主机和运行时概况。
type ServerMonitorHandle struct {
	svc *service.ServerMonitorService
}

// NewServerMonitorHandle 创建系统监控接口。
func NewServerMonitorHandle(svc *service.ServerMonitorService) *ServerMonitorHandle {
	return &ServerMonitorHandle{svc: svc}
}

// ModuleName 返回权限前缀和菜单标题。
func (i *ServerMonitorHandle) ModuleName() (string, string) {
	return "server", "系统监控"
}

// @Summary		运行概况
// @Description	返回主机、进程、数据库连接池和 Redis 概况
// @Tags			系统监控
// @Produce		json
// @Success		200	{object}	router.Resp	"操作成功"
// @Failure		500	{object}	router.Resp	"服务器内部错误"
// @Router			/api/v1/tools/server [GET]
func (i *ServerMonitorHandle) Snapshot(ctx *gin.Context) {
	router.Success(ctx, i.svc.Snapshot(ctx.Request.Context()))
}
