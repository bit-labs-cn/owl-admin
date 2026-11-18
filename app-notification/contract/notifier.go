package contract

import "context"

// Notifier 是业务子系统发送通知的入口。
type Notifier interface {
	// Send 创建并提交一条通知，返回任务 ID。
	Send(ctx context.Context, msg *Message) (string, error)
	// CompleteByBizKey 把同一业务主键下未完成的待办标成已完成。
	CompleteByBizKey(ctx context.Context, bizKey string) error
	// RegisterEvents 登记业务事件，供后台选模板和默认渠道。
	RegisterEvents(ctx context.Context, events []EventDescriptor) error
}
