package repository

import (
	"context"
	"fmt"
	"time"

	"bit-labs.cn/owl-admin/app-notification/model"
)

// MessageQuery 是管理端消息记录的筛选条件。
type MessageQuery struct {
	Keyword  string
	TypeCode string
	Kind     int
	Page     int
	PageSize int
}

// MessageSummaryDTO 是一条已发消息及其接收统计。
type MessageSummaryDTO struct {
	ID             uint      `json:"id,string"`
	Kind           int       `json:"kind"`
	TypeCode       string    `json:"typeCode"`
	TypeName       string    `json:"typeName"`
	Title          string    `json:"title"`
	SenderID       uint      `json:"senderId,string"`
	SenderName     string    `json:"senderName"`
	RecipientCount int64     `json:"recipientCount"`
	ReadCount      int64     `json:"readCount"`
	UnreadCount    int64     `json:"unreadCount"`
	TodoPending    int64     `json:"todoPending"`
	TodoDone       int64     `json:"todoDone"`
	DeliveryFailed int64     `json:"deliveryFailed"`
	CreatedAt      time.Time `json:"createdAt"`
}

// MessageRecipientDTO 是消息详情里的一个接收人及其阅读、待办状态。
type MessageRecipientDTO struct {
	RecipientID uint       `json:"id,string"`
	UserID      uint       `json:"userId,string"`
	UserName    string     `json:"userName"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	ReadStatus  int        `json:"readStatus"`
	ReadAt      *time.Time `json:"readAt,omitempty"`
	TodoStatus  int        `json:"todoStatus"`
	DoneAt      *time.Time `json:"doneAt,omitempty"`
}

// MessageDetailDTO 是消息本体与接收人明细。
type MessageDetailDTO struct {
	MessageSummaryDTO
	Content    string                `json:"content"`
	Link       string                `json:"link"`
	DueAt      *time.Time            `json:"dueAt,omitempty"`
	Recipients []MessageRecipientDTO `json:"recipients"`
}

// messageSummaryFields 组装消息概要与接收统计的查询列。
// 失败数取发送记录里状态为失败的行，状态值来自模型常量，避免 SQL 里写死数字。
func messageSummaryFields() string {
	return fmt.Sprintf(`m.id, m.kind, m.type_code,
		COALESCE(t.name, '') as type_name, m.title, m.sender_id, m.created_at,
		COALESCE(u.username, '') as sender_name,
		(SELECT COUNT(*) FROM notification_recipient r WHERE r.message_id = m.id AND r.deleted_at IS NULL) as recipient_count,
		(SELECT COUNT(*) FROM notification_recipient r WHERE r.message_id = m.id AND r.deleted_at IS NULL AND r.read_status = %d) as read_count,
		(SELECT COUNT(*) FROM notification_recipient r WHERE r.message_id = m.id AND r.deleted_at IS NULL AND r.read_status = %d) as unread_count,
		(SELECT COUNT(*) FROM notification_recipient r WHERE r.message_id = m.id AND r.deleted_at IS NULL AND r.todo_status = %d) as todo_pending,
		(SELECT COUNT(*) FROM notification_recipient r WHERE r.message_id = m.id AND r.deleted_at IS NULL AND r.todo_status = %d) as todo_done,
		(SELECT COUNT(*) FROM notification_delivery d WHERE d.message_id = m.id AND d.deleted_at IS NULL AND d.status = %d) as delivery_failed`,
		model.ReadStatusRead, model.ReadStatusUnread,
		model.TodoStatusPending, model.TodoStatusDone,
		model.DeliveryStatusFailed,
	)
}

// ListMessages 按条件分页列出消息及接收统计。
func (r *CatalogRepository) ListMessages(ctx context.Context, q MessageQuery) (int64, []MessageSummaryDTO, error) {
	q.Page, q.PageSize = normalizePage(q.Page, q.PageSize)
	base := r.db.WithContext(ctx).Table("notification_message m").
		Joins("LEFT JOIN notification_type t ON t.code = m.type_code AND t.deleted_at IS NULL").
		Joins("LEFT JOIN admin_user u ON u.id = m.sender_id AND u.deleted_at IS NULL").
		Where("m.deleted_at IS NULL")
	if q.Keyword != "" {
		like := "%" + q.Keyword + "%"
		base = base.Where("m.title LIKE ? OR m.content LIKE ?", like, like)
	}
	if q.TypeCode != "" {
		base = base.Where("m.type_code = ?", q.TypeCode)
	}
	if q.Kind > 0 {
		base = base.Where("m.kind = ?", q.Kind)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return 0, nil, err
	}
	var list []MessageSummaryDTO
	err := base.Select(messageSummaryFields()).
		Order("m.id desc").
		Offset((q.Page - 1) * q.PageSize).
		Limit(q.PageSize).
		Scan(&list).Error
	return total, list, err
}

// FindMessageSummary 查询单条消息的概要。
func (r *CatalogRepository) FindMessageSummary(ctx context.Context, id uint) (*MessageSummaryDTO, error) {
	var row MessageSummaryDTO
	err := r.db.WithContext(ctx).Table("notification_message m").
		Joins("LEFT JOIN notification_type t ON t.code = m.type_code AND t.deleted_at IS NULL").
		Joins("LEFT JOIN admin_user u ON u.id = m.sender_id AND u.deleted_at IS NULL").
		Where("m.id = ? AND m.deleted_at IS NULL", id).
		Select(messageSummaryFields()).
		Scan(&row).Error
	if err != nil {
		return nil, err
	}
	if row.ID == 0 {
		return nil, nil
	}
	return &row, nil
}

// FindMessageDetail 返回消息正文与全部接收人状态。
func (r *CatalogRepository) FindMessageDetail(ctx context.Context, id uint) (*MessageDetailDTO, error) {
	summary, err := r.FindMessageSummary(ctx, id)
	if err != nil || summary == nil {
		return nil, err
	}
	var body struct {
		Content string     `gorm:"column:content"`
		Link    string     `gorm:"column:link"`
		DueAt   *time.Time `gorm:"column:due_at"`
	}
	if err := r.db.WithContext(ctx).Table("notification_message").
		Select("content, link, due_at").
		Where("id = ?", id).
		Scan(&body).Error; err != nil {
		return nil, err
	}

	var recipients []MessageRecipientDTO
	err = r.db.WithContext(ctx).Table("notification_recipient r").
		Joins("LEFT JOIN admin_user u ON u.id = r.user_id AND u.deleted_at IS NULL").
		Select(`r.id, r.user_id, COALESCE(NULLIF(u.nickname, ''), u.username, '') as user_name,
			COALESCE(u.username, '') as username, COALESCE(u.email, '') as email,
			r.read_status, r.read_at, r.todo_status, r.done_at`).
		Where("r.message_id = ? AND r.deleted_at IS NULL", id).
		Order("r.read_status asc, r.id asc").
		Scan(&recipients).Error
	if err != nil {
		return nil, err
	}

	return &MessageDetailDTO{
		MessageSummaryDTO: *summary,
		Content:           body.Content,
		Link:              body.Link,
		DueAt:             body.DueAt,
		Recipients:        recipients,
	}, nil
}
