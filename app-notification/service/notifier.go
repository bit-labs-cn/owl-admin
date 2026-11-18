package service

import (
	"context"

	"bit-labs.cn/owl-admin/app-notification/contract"
	"github.com/spf13/cast"
)

// Notifier 把契约接到任务服务上，供其它子系统调用。
type Notifier struct {
	tasks  *TaskService
	inbox  *InboxService
	events *EventService
}

func NewNotifier(tasks *TaskService, inbox *InboxService, events *EventService) *Notifier {
	return &Notifier{tasks: tasks, inbox: inbox, events: events}
}

func (n *Notifier) Send(ctx context.Context, msg *contract.Message) (string, error) {
	row, err := n.tasks.SendMessage(ctx, msg)
	if err != nil {
		return "", err
	}
	if row == nil {
		return "", nil
	}
	return cast.ToString(row.ID), nil
}

func (n *Notifier) CompleteByBizKey(ctx context.Context, bizKey string) error {
	return n.inbox.CompleteByBizKey(ctx, bizKey)
}

func (n *Notifier) RegisterEvents(ctx context.Context, events []contract.EventDescriptor) error {
	return n.events.Register(ctx, events)
}
