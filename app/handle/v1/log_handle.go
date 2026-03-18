package v1

import (
	"bit-labs.cn/owl-admin/app/service"
	"bit-labs.cn/owl/provider/router"
	"github.com/gin-gonic/gin"
)

var _ router.Handler = (*LogHandle)(nil)

type LogHandle struct {
	logSvc *service.LogService
}

func NewLogHandle(logSvc *service.LogService) *LogHandle {
	return &LogHandle{logSvc: logSvc}
}

func (i *LogHandle) ModuleName() (string, string) {
	return "monitor", "系统监控"
}

// @Summary		登录日志
// @Description	分页查询登录日志
// @Tags			系统监控
// @Accept			json
// @Produce		json
// @Param			request	body		service.RetrieveLoginLogsReq	true	"登录日志查询请求"
// @Success		200		{object}	router.PageResp					"操作成功"
// @Failure		500		{object}	router.Resp						"服务器内部错误"
// @Router			/api/v1/monitor/login-logs [POST]
func (i *LogHandle) LoginLogs(ctx *gin.Context) {
	var req service.RetrieveLoginLogsReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.BadRequest(ctx, err.Error())
		return
	}
	count, list, err := i.logSvc.RetrieveLoginLogs(ctx.Request.Context(), &req)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.PageSuccess(ctx, int(count), req.Page, req.PageSize, list)
}

// @Summary		操作日志
// @Description	分页查询操作日志
// @Tags			系统监控
// @Accept			json
// @Produce		json
// @Param			request	body		service.RetrieveOperationLogsReq	true	"操作日志查询请求"
// @Success		200		{object}	router.PageResp						"操作成功"
// @Failure		500		{object}	router.Resp							"服务器内部错误"
// @Router			/api/v1/monitor/operation-logs [POST]
func (i *LogHandle) OperationLogs(ctx *gin.Context) {
	var req service.RetrieveOperationLogsReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.BadRequest(ctx, err.Error())
		return
	}
	count, list, err := i.logSvc.RetrieveOperationLogs(ctx.Request.Context(), &req)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.PageSuccess(ctx, int(count), req.Page, req.PageSize, list)
}
