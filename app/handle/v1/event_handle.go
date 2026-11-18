package v1

import (
	"bit-labs.cn/owl-admin/app/service"
	"bit-labs.cn/owl/provider/router"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

var _ router.Handler = (*EventHandle)(nil)

// EventHandle 事件总线工作台与事件记录接口。
type EventHandle struct {
	svc *service.EventService
}

func NewEventHandle(svc *service.EventService) *EventHandle {
	return &EventHandle{svc: svc}
}

func (i *EventHandle) ModuleName() (string, string) {
	return "event", "事件总线"
}

// @Summary		事件总线工作台
// @Description	汇总发件箱、收件箱状态计数，按天趋势，类型与来源分布，Redis Stream 消费组积压
// @Tags			事件总线
// @Produce		json
// @Param			days	query		int				false	"统计天数，1～30，默认 7"
// @Success		200		{object}	router.Resp		"操作成功"
// @Failure		500		{object}	router.Resp		"服务器内部错误"
// @Router			/api/v1/events/workbench [GET]
func (i *EventHandle) Workbench(ctx *gin.Context) {
	days := cast.ToInt(ctx.Query("days"))
	resp, err := i.svc.Workbench(ctx.Request.Context(), days)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, resp)
}

// @Summary		发件箱列表
// @Description	分页查询发件箱事件
// @Tags			事件总线
// @Produce		json
// @Param			page		query		int								false	"页码"
// @Param			pageSize	query		int								false	"每页数量"
// @Param			eventId		query		string							false	"事件ID"
// @Param			eventType	query		string							false	"事件类型（模糊）"
// @Param			source		query		string							false	"来源子系统"
// @Param			subject		query		string							false	"业务实体（模糊）"
// @Param			status		query		int								false	"状态：0-待发送、1-已发送、2-投递失败"
// @Param			createdAt	query		string							false	"创建时间区间，start,end（Unix 秒）"
// @Success		200			{object}	router.PageResp					"操作成功"
// @Failure		500			{object}	router.Resp						"服务器内部错误"
// @Router			/api/v1/events/outbox [GET]
func (i *EventHandle) RetrieveOutbox(ctx *gin.Context) {
	var req service.RetrieveOutboxReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		router.BadRequest(ctx, "参数绑定失败")
		return
	}
	count, list, err := i.svc.RetrieveOutbox(ctx.Request.Context(), &req)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.PageSuccess(ctx, int(count), req.Page, req.PageSize, list)
}

// @Summary		发件箱详情
// @Description	获取单条发件箱事件的载荷与元数据
// @Tags			事件总线
// @Produce		json
// @Param			id	path		int				true	"记录ID"
// @Success		200	{object}	router.Resp		"操作成功"
// @Failure		500	{object}	router.Resp		"服务器内部错误"
// @Router			/api/v1/events/outbox/{id} [GET]
func (i *EventHandle) OutboxDetail(ctx *gin.Context) {
	detail, err := i.svc.OutboxDetail(ctx.Request.Context(), cast.ToUint64(ctx.Param("id")))
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, detail)
}

// @Summary		重投发件箱事件
// @Description	把未投递成功的事件重置为待发送，由事件中继在下一个轮询周期内重新投递
// @Tags			事件总线
// @Produce		json
// @Param			id	path		int				true	"记录ID"
// @Success		200	{object}	router.Resp		"操作成功"
// @Failure		500	{object}	router.Resp		"服务器内部错误"
// @Router			/api/v1/events/outbox/{id}/retry [POST]
func (i *EventHandle) RetryOutbox(ctx *gin.Context) {
	retried, err := i.svc.RetryOutbox(ctx.Request.Context(), cast.ToUint64(ctx.Param("id")))
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.SuccessWithMsg(ctx, "已重置为待发送，稍后由事件中继投递", gin.H{"retried": retried})
}

// @Summary		清理事件历史
// @Description	删除指定天数之前已投递成功 / 消费成功的记录，保留未完成与失败的流水
// @Tags			事件总线
// @Accept			json
// @Produce		json
// @Param			request	body		service.CleanupEventsReq	true	"清理请求"
// @Success		200		{object}	router.Resp					"操作成功"
// @Failure		400		{object}	router.Resp					"参数错误"
// @Failure		500		{object}	router.Resp					"服务器内部错误"
// @Router			/api/v1/events/cleanup [POST]
func (i *EventHandle) Cleanup(ctx *gin.Context) {
	var req service.CleanupEventsReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	resp, err := i.svc.Cleanup(ctx.Request.Context(), &req)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, resp)
}

// @Summary		清理僵尸消费者
// @Description	删除各消费组中长期没有读取、且没有滞留消息的消费者记录；有 pending 的消费者会跳过
// @Tags			事件总线
// @Accept			json
// @Produce		json
// @Param			request	body		service.CleanupConsumersReq	true	"清理请求"
// @Success		200		{object}	router.Resp					"操作成功"
// @Failure		400		{object}	router.Resp					"参数错误"
// @Failure		500		{object}	router.Resp					"服务器内部错误"
// @Router			/api/v1/events/consumers/cleanup [POST]
func (i *EventHandle) CleanupConsumers(ctx *gin.Context) {
	var req service.CleanupConsumersReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	resp, err := i.svc.CleanupConsumers(ctx.Request.Context(), &req)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, resp)
}

// @Summary		收件箱列表
// @Description	分页查询收件箱消费流水
// @Tags			事件总线
// @Produce		json
// @Param			page			query		int				false	"页码"
// @Param			pageSize		query		int				false	"每页数量"
// @Param			eventId			query		string			false	"事件ID"
// @Param			consumerGroup	query		string			false	"消费组（模糊）"
// @Param			status			query		int				false	"状态：1-消费成功、2-消费失败"
// @Param			processedAt		query		string			false	"处理时间区间，start,end（Unix 秒）"
// @Success		200				{object}	router.PageResp	"操作成功"
// @Failure		500				{object}	router.Resp		"服务器内部错误"
// @Router			/api/v1/events/inbox [GET]
func (i *EventHandle) RetrieveInbox(ctx *gin.Context) {
	var req service.RetrieveInboxReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		router.BadRequest(ctx, "参数绑定失败")
		return
	}
	count, list, err := i.svc.RetrieveInbox(ctx.Request.Context(), &req)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.PageSuccess(ctx, int(count), req.Page, req.PageSize, list)
}

// @Summary		收件箱详情
// @Description	获取单条收件箱消费流水
// @Tags			事件总线
// @Produce		json
// @Param			id	path		int				true	"记录ID"
// @Success		200	{object}	router.Resp		"操作成功"
// @Failure		500	{object}	router.Resp		"服务器内部错误"
// @Router			/api/v1/events/inbox/{id} [GET]
func (i *EventHandle) InboxDetail(ctx *gin.Context) {
	detail, err := i.svc.InboxDetail(ctx.Request.Context(), cast.ToUint64(ctx.Param("id")))
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, detail)
}
