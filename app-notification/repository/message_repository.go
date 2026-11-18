package repository

import (
	"context"
	"errors"
	"time"

	"bit-labs.cn/owl-admin/app-notification/model"
	"bit-labs.cn/owl/contract"
	"gorm.io/gorm"
)

// UserMessageDTO 是给终端用户展示的一条消息/待办视图
type UserMessageDTO struct {
	ID         uint       `json:"id,string"`        // 流水记录 ID
	MessageID  uint       `json:"messageId,string"` // 消息本体 ID
	Kind       int        `json:"kind"`             // 1通知 2待办
	TypeCode   string     `json:"typeCode"`         // 消息类型编码
	TypeName   string     `json:"typeName"`         // 消息类型名称
	Title      string     `json:"title"`            // 标题
	Content    string     `json:"content"`          // 正文
	Link       string     `json:"link"`             // 跳转深链
	BizKey     string     `json:"bizKey"`           // 业务主键
	DueAt      *time.Time `json:"dueAt,omitempty"`  // 待办截止时间
	ReadStatus int        `json:"readStatus"`       // 0未读 1已读
	ReadAt     *time.Time `json:"readAt,omitempty"` // 阅读时间
	TodoStatus int        `json:"todoStatus"`       // 0无 1待处理 2已完成 3已过期
	DoneAt     *time.Time `json:"doneAt,omitempty"` // 完成时间
	CreatedAt  time.Time  `json:"createdAt"`        // 接收时间
}

type MessageRepositoryInterface interface {
	CreateDirect(ctx context.Context, tx *gorm.DB, msg *model.NotifyMessage, userIDs []uint) error
	FindByBizKey(ctx context.Context, tx *gorm.DB, bizKey string) (*model.NotifyMessage, error)
	RetrieveUserMessages(ctx context.Context, userID uint, kind int, readStatus *int, typeCode string, page, pageSize int) (int64, []UserMessageDTO, error)
	UnreadCounts(ctx context.Context, userID uint) (unreadNotice int64, pendingTodo int64, err error)
	// MarkAsRead 按接收流水 ID 标记已读，recipientIDs 为空时标记该用户全部未读。
	MarkAsRead(ctx context.Context, userID uint, recipientIDs []uint) error
	// MarkAllAsRead 标记该用户全部未读；kind > 0 时只标记该种类的消息。
	MarkAllAsRead(ctx context.Context, userID uint, kind int) error
	CompleteByBizKey(ctx context.Context, bizKey string) error
	CompleteByID(ctx context.Context, userID, recipientID uint) error
	ReopenByID(ctx context.Context, userID, recipientID uint) error
	DeleteUserMessage(ctx context.Context, userID, recipientID uint) error
	ExpireOverdueTodos(ctx context.Context, now time.Time) error
	DeleteBefore(ctx context.Context, t time.Time) error
	contract.WithContext[MessageRepositoryInterface]
}

type MessageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) MessageRepositoryInterface {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) WithContext(ctx context.Context) MessageRepositoryInterface {
	return &MessageRepository{db: r.db.WithContext(ctx)}
}

// CreateDirect 支持传入已有事务 tx，以保障分布式双写一致性
func (r *MessageRepository) CreateDirect(ctx context.Context, tx *gorm.DB, msg *model.NotifyMessage, userIDs []uint) error {
	runner := r.db.WithContext(ctx)
	if tx != nil {
		runner = tx.WithContext(ctx)
	}

	if err := runner.Create(msg).Error; err != nil {
		return err
	}

	if len(userIDs) == 0 {
		return nil
	}

	initialTodoStatus := model.TodoStatusNone
	if msg.Kind == model.KindTodo {
		initialTodoStatus = model.TodoStatusPending
	}

	recipients := make([]model.NotifyRecipient, 0, len(userIDs))
	for _, uid := range userIDs {
		recipients = append(recipients, model.NotifyRecipient{
			MessageID:  msg.ID,
			UserID:     uid,
			ReadStatus: model.ReadStatusUnread,
			TodoStatus: initialTodoStatus,
		})
	}

	return runner.Create(&recipients).Error
}

// FindByBizKey 按业务键查找已存在的消息。tx 非空时在该事务内查询。
func (r *MessageRepository) FindByBizKey(ctx context.Context, tx *gorm.DB, bizKey string) (*model.NotifyMessage, error) {
	runner := r.db.WithContext(ctx)
	if tx != nil {
		runner = tx.WithContext(ctx)
	}
	var row model.NotifyMessage
	err := runner.Where("biz_key = ?", bizKey).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *MessageRepository) RetrieveUserMessages(ctx context.Context, userID uint, kind int, readStatus *int, typeCode string, page, pageSize int) (int64, []UserMessageDTO, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	base := r.db.WithContext(ctx).Table("notification_recipient r").
		Joins("INNER JOIN notification_message m ON r.message_id = m.id").
		Joins("LEFT JOIN notification_type t ON t.code = m.type_code AND t.deleted_at IS NULL").
		Where("r.user_id = ?", userID)

	if kind > 0 {
		base = base.Where("m.kind = ?", kind)
	}
	if typeCode != "" {
		base = base.Where("m.type_code = ?", typeCode)
	}
	if readStatus != nil {
		base = base.Where("r.read_status = ?", *readStatus)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return 0, nil, err
	}

	var list []UserMessageDTO
	err := base.Select("r.id, r.message_id, m.kind, m.type_code, COALESCE(t.name, '') as type_name, m.title, m.content, m.link, m.biz_key, m.due_at, r.read_status, r.read_at, r.todo_status, r.done_at, r.created_at").
		Order("r.id desc").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&list).Error

	return total, list, err
}

func (r *MessageRepository) UnreadCounts(ctx context.Context, userID uint) (unreadNotice int64, pendingTodo int64, err error) {
	err = r.db.WithContext(ctx).Table("notification_recipient r").
		Joins("INNER JOIN notification_message m ON r.message_id = m.id").
		Where("r.user_id = ? AND r.read_status = ? AND m.kind = ?", userID, model.ReadStatusUnread, model.KindNotice).
		Count(&unreadNotice).Error
	if err != nil {
		return 0, 0, err
	}

	err = r.db.WithContext(ctx).Table("notification_recipient").
		Where("user_id = ? AND todo_status = ?", userID, model.TodoStatusPending).
		Count(&pendingTodo).Error
	return unreadNotice, pendingTodo, err
}

func (r *MessageRepository) MarkAsRead(ctx context.Context, userID uint, recipientIDs []uint) error {
	now := time.Now()
	tx := r.db.WithContext(ctx).Model(&model.NotifyRecipient{}).
		Where("user_id = ? AND read_status = ?", userID, model.ReadStatusUnread)
	if len(recipientIDs) > 0 {
		tx = tx.Where("id IN ?", recipientIDs)
	}
	return tx.Updates(map[string]any{
		"read_status": model.ReadStatusRead,
		"read_at":     &now,
	}).Error
}

func (r *MessageRepository) MarkAllAsRead(ctx context.Context, userID uint, kind int) error {
	now := time.Now()
	tx := r.db.WithContext(ctx).Model(&model.NotifyRecipient{}).
		Where("user_id = ? AND read_status = ?", userID, model.ReadStatusUnread)
	if kind > 0 {
		tx = tx.Where("message_id IN (?)", r.db.WithContext(ctx).
			Model(&model.NotifyMessage{}).
			Select("id").
			Where("kind = ? AND deleted_at IS NULL", kind))
	}
	return tx.Updates(map[string]any{
		"read_status": model.ReadStatusRead,
		"read_at":     &now,
	}).Error
}

func (r *MessageRepository) CompleteByBizKey(ctx context.Context, bizKey string) error {
	now := time.Now()
	var messageIDs []uint
	if err := r.db.WithContext(ctx).Model(&model.NotifyMessage{}).
		Where("biz_key = ? AND kind = ?", bizKey, model.KindTodo).
		Pluck("id", &messageIDs).Error; err != nil {
		return err
	}
	if len(messageIDs) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&model.NotifyRecipient{}).
		Where("message_id IN ? AND todo_status = ?", messageIDs, model.TodoStatusPending).
		Updates(map[string]any{
			"todo_status": model.TodoStatusDone,
			"done_at":     &now,
			"read_status": model.ReadStatusRead,
			"read_at":     &now,
		}).Error
}

func (r *MessageRepository) CompleteByID(ctx context.Context, userID, recipientID uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&model.NotifyRecipient{}).
		Where("id = ? AND user_id = ?", recipientID, userID).
		Updates(map[string]any{
			"todo_status": model.TodoStatusDone,
			"done_at":     &now,
			"read_status": model.ReadStatusRead,
			"read_at":     &now,
		}).Error
}

func (r *MessageRepository) ReopenByID(ctx context.Context, userID, recipientID uint) error {
	return r.db.WithContext(ctx).Model(&model.NotifyRecipient{}).
		Where("id = ? AND user_id = ?", recipientID, userID).
		Updates(map[string]any{
			"todo_status": model.TodoStatusPending,
			"done_at":     nil,
		}).Error
}

func (r *MessageRepository) DeleteUserMessage(ctx context.Context, userID, recipientID uint) error {
	return r.db.WithContext(ctx).Where("id = ? AND user_id = ?", recipientID, userID).Delete(&model.NotifyRecipient{}).Error
}

func (r *MessageRepository) ExpireOverdueTodos(ctx context.Context, now time.Time) error {
	var overdueMsgIDs []uint
	if err := r.db.WithContext(ctx).Model(&model.NotifyMessage{}).
		Where("kind = ? AND due_at IS NOT NULL AND due_at <= ?", model.KindTodo, now).
		Pluck("id", &overdueMsgIDs).Error; err != nil {
		return err
	}
	if len(overdueMsgIDs) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&model.NotifyRecipient{}).
		Where("message_id IN ? AND todo_status = ?", overdueMsgIDs, model.TodoStatusPending).
		Update("todo_status", model.TodoStatusExpired).Error
}

func (r *MessageRepository) DeleteBefore(ctx context.Context, t time.Time) error {
	var oldIDs []uint
	if err := r.db.WithContext(ctx).Model(&model.NotifyMessage{}).
		Where("created_at < ?", t).
		Pluck("id", &oldIDs).Error; err != nil {
		return err
	}
	if len(oldIDs) == 0 {
		return nil
	}
	_ = r.db.WithContext(ctx).Where("message_id IN ?", oldIDs).Delete(&model.NotifyRecipient{}).Error
	_ = r.db.WithContext(ctx).Where("message_id IN ?", oldIDs).Delete(&model.NotifyDelivery{}).Error
	return r.db.WithContext(ctx).Where("id IN ?", oldIDs).Delete(&model.NotifyMessage{}).Error
}
