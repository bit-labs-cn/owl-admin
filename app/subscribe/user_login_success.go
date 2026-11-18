package subscribe

import (
	"context"

	"bit-labs.cn/owl-admin/app/event"
	"bit-labs.cn/owl-admin/app/service"
	contractEvent "bit-labs.cn/owl/contract/event"
)

// subscribeUserLoginSuccess 监听登录成功事件，写入登录日志。
func subscribeUserLoginSuccess(bus contractEvent.EventBus, logSvc *service.LogService) error {
	return bus.Subscribe(event.UserLoginSuccess, contractEvent.AsHandler(func(ctx context.Context, evt *service.LoginSuccessEvent) error {
		userType := "user"
		if evt.IsSuperAdmin {
			userType = "super_admin"
		}
		return logSvc.CreateLoginLog(ctx, &service.CreateLoginLogReq{
			UserId:    int(evt.UserID),
			UserName:  evt.Username,
			UserType:  userType,
			LoginTime: evt.LoginTime,
			Ip:        evt.IP,
			UserAgent: evt.UserAgent,
		})
	}))
}
