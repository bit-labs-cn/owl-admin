package v1

import (
	"bit-labs.cn/owl-admin/app-notification/service"
	"bit-labs.cn/owl/provider/router"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// ChannelHandle 处理渠道账号。
type ChannelHandle struct {
	svc *service.ChannelAccountService
}

func NewChannelHandle(svc *service.ChannelAccountService) *ChannelHandle {
	return &ChannelHandle{svc: svc}
}
func (h *ChannelHandle) ModuleName() (string, string) { return "channel", "渠道账号" }

func (h *ChannelHandle) Save(ctx *gin.Context) {
	var req service.SaveChannelAccountReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	if err := h.svc.Save(ctx.Request.Context(), cast.ToUint(ctx.Param("id")), &req); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

func (h *ChannelHandle) Delete(ctx *gin.Context) {
	if err := h.svc.Delete(ctx.Request.Context(), cast.ToUint(ctx.Param("id"))); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

func (h *ChannelHandle) Retrieve(ctx *gin.Context) {
	var req service.RetrieveChannelAccountReq
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

func (h *ChannelHandle) Detail(ctx *gin.Context) {
	row, err := h.svc.Detail(ctx.Request.Context(), cast.ToUint(ctx.Param("id")))
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, row)
}

func (h *ChannelHandle) Test(ctx *gin.Context) {
	var req struct {
		Address string `json:"address"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	if err := h.svc.TestSend(ctx.Request.Context(), cast.ToUint(ctx.Param("id")), req.Address); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// DeviceHandle 处理推送设备。
type DeviceHandle struct{ svc *service.DeviceService }

func NewDeviceHandle(svc *service.DeviceService) *DeviceHandle { return &DeviceHandle{svc: svc} }
func (h *DeviceHandle) ModuleName() (string, string)           { return "device", "推送设备" }

func (h *DeviceHandle) Register(ctx *gin.Context) {
	var req service.RegisterDeviceReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	if err := h.svc.Register(ctx.Request.Context(), &req); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

func (h *DeviceHandle) DeleteMine(ctx *gin.Context) {
	if err := h.svc.DeleteMine(ctx.Request.Context(), cast.ToUint(ctx.Param("id"))); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

func (h *DeviceHandle) Retrieve(ctx *gin.Context) {
	var req service.RetrieveDeviceReq
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

func (h *DeviceHandle) Invalidate(ctx *gin.Context) {
	if err := h.svc.Invalidate(ctx.Request.Context(), cast.ToUint(ctx.Param("id"))); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// DeliveryHandle 处理投递记录。
type DeliveryHandle struct{ svc *service.DeliveryService }

func NewDeliveryHandle(svc *service.DeliveryService) *DeliveryHandle {
	return &DeliveryHandle{svc: svc}
}
func (h *DeliveryHandle) ModuleName() (string, string) { return "delivery", "投递记录" }

func (h *DeliveryHandle) Retrieve(ctx *gin.Context) {
	var req service.RetrieveDeliveryReq
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

func (h *DeliveryHandle) Detail(ctx *gin.Context) {
	row, err := h.svc.Detail(ctx.Request.Context(), cast.ToUint(ctx.Param("id")))
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, row)
}

func (h *DeliveryHandle) Retry(ctx *gin.Context) {
	if err := h.svc.Retry(ctx.Request.Context(), cast.ToUint(ctx.Param("id"))); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// EventHandle 处理事件目录。
type EventHandle struct{ svc *service.EventService }

func NewEventHandle(svc *service.EventService) *EventHandle { return &EventHandle{svc: svc} }
func (h *EventHandle) ModuleName() (string, string)         { return "event", "通知事件" }

func (h *EventHandle) Save(ctx *gin.Context) {
	var req service.SaveEventReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	if err := h.svc.Save(ctx.Request.Context(), cast.ToUint(ctx.Param("id")), &req); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

func (h *EventHandle) Delete(ctx *gin.Context) {
	if err := h.svc.Delete(ctx.Request.Context(), cast.ToUint(ctx.Param("id"))); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

func (h *EventHandle) Retrieve(ctx *gin.Context) {
	var req service.RetrieveEventReq
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

// SettingHandle 处理通知设置。
type SettingHandle struct{ svc *service.SettingService }

func NewSettingHandle(svc *service.SettingService) *SettingHandle { return &SettingHandle{svc: svc} }
func (h *SettingHandle) ModuleName() (string, string)             { return "setting", "通知设置" }

func (h *SettingHandle) Get(ctx *gin.Context) {
	row, err := h.svc.Get(ctx.Request.Context())
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, row)
}

func (h *SettingHandle) Save(ctx *gin.Context) {
	var req service.SettingValues
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	if err := h.svc.Save(ctx.Request.Context(), req); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}

// SubscriptionHandle 处理我的订阅。
type SubscriptionHandle struct{ svc *service.SubscriptionService }

func NewSubscriptionHandle(svc *service.SubscriptionService) *SubscriptionHandle {
	return &SubscriptionHandle{svc: svc}
}
func (h *SubscriptionHandle) ModuleName() (string, string) { return "subscription", "我的订阅" }

func (h *SubscriptionHandle) List(ctx *gin.Context) {
	rows, err := h.svc.List(ctx.Request.Context())
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, rows)
}

func (h *SubscriptionHandle) Save(ctx *gin.Context) {
	var req struct {
		Items []service.SubscriptionItem `json:"items"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	if err := h.svc.Save(ctx.Request.Context(), req.Items); err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, nil)
}
