package service

import (
	"context"

	"bit-labs.cn/owl-admin/app-notification/contract"
	notifEvent "bit-labs.cn/owl-admin/app-notification/event"
	contractEvent "bit-labs.cn/owl/contract/event"
	"github.com/spf13/cast"
	"gorm.io/gorm"
)

// Notifier 把契约接到任务服务上，供其它子系统调用。
type Notifier struct {
	tasks  *TaskService
	inbox  *InboxService
	events *EventService
	bus    contractEvent.EventBus
	db     *gorm.DB
}

func NewNotifier(tasks *TaskService, inbox *InboxService, events *EventService, bus contractEvent.EventBus, db *gorm.DB) *Notifier {
	return &Notifier{tasks: tasks, inbox: inbox, events: events, bus: bus, db: db}
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
	if n.bus != nil && n.db != nil {
		evt, err := contractEvent.NewEvent(notifEvent.TopicTodoComplete, notifEvent.TodoCompletePayload{
			BizKey: bizKey,
		}, contractEvent.WithSource("owl-notification"), contractEvent.WithSubject("bizKey:"+bizKey), contractEvent.WithPriority(contractEvent.PriorityHigh))
		if err == nil {
			_ = n.bus.PublishTx(ctx, n.db, evt)
		}
	}
	return n.inbox.CompleteByBizKey(ctx, bizKey)
}

func (n *Notifier) RegisterEvents(ctx context.Context, events []contract.EventDescriptor) error {
	return n.events.Register(ctx, events)
}
