package repository

import (
	"context"
	"fmt"
	"time"

	"bit-labs.cn/owl-admin/app-notification/model"
	"gorm.io/gorm"
)

// DeliveryBatchDTO 是一次发送（一条消息）的投递合计。
type DeliveryBatchDTO struct {
	MessageID    uint      `json:"messageId,string"`
	Title        string    `json:"title"`
	Kind         int       `json:"kind"`
	TypeName     string    `json:"typeName"`
	Total        int64     `json:"total"`
	SuccessCount int64     `json:"successCount"`
	FailedCount  int64     `json:"failedCount"`
	SkippedCount int64     `json:"skippedCount"`
	PendingCount int64     `json:"pendingCount"`
	ChannelCount int64     `json:"channelCount"`
	RecipientSum int64     `json:"recipientCount"`
	SenderName   string    `json:"senderName"`
	CreatedAt    time.Time `json:"createdAt"`
}

// ListDeliveryBatches 按消息聚合发送记录。筛选条件作用于明细，批次里的合计也只算命中的明细。
func (r *CatalogRepository) ListDeliveryBatches(ctx context.Context, q DeliveryQuery) (int64, []DeliveryBatchDTO, error) {
	q.Page, q.PageSize = normalizePage(q.Page, q.PageSize)
	base := func() *gorm.DB {
		tx := r.db.WithContext(ctx).Table("notification_delivery d").
			Joins("INNER JOIN notification_message m ON m.id = d.message_id AND m.deleted_at IS NULL").
			Joins("LEFT JOIN notification_type t ON t.code = m.type_code AND t.deleted_at IS NULL").
			Where("d.deleted_at IS NULL")
		if q.Status > 0 {
			tx = tx.Where("d.status = ?", q.Status)
		}
		if q.Channel != "" {
			tx = tx.Where("d.channel = ?", q.Channel)
		}
		if q.RecipientType > 0 {
			tx = tx.Where("d.recipient_type = ?", q.RecipientType)
		}
		return tx
	}

	var total int64
	if err := base().Distinct("d.message_id").Count(&total).Error; err != nil {
		return 0, nil, err
	}
	if total == 0 {
		return 0, nil, nil
	}

	var list []DeliveryBatchDTO
	err := base().
		Select(fmt.Sprintf(`d.message_id, COALESCE(NULLIF(MAX(d.title), ''), MAX(m.title)) as title, m.kind,
			COALESCE(MAX(t.name), '') as type_name, m.created_at,
			(SELECT COALESCE(NULLIF(u.nickname, ''), u.username, '') FROM admin_user u WHERE u.id = m.sender_id AND u.deleted_at IS NULL) as sender_name,
			COUNT(*) as total,
			SUM(CASE WHEN d.status = %d THEN 1 ELSE 0 END) as success_count,
			SUM(CASE WHEN d.status = %d THEN 1 ELSE 0 END) as failed_count,
			SUM(CASE WHEN d.status = %d THEN 1 ELSE 0 END) as skipped_count,
			SUM(CASE WHEN d.status = %d THEN 1 ELSE 0 END) as pending_count,
			COUNT(DISTINCT d.channel) as channel_count,
			COUNT(DISTINCT CASE WHEN d.recipient_type <> %d THEN d.user_id ELSE 0 END) as recipient_sum`,
			model.DeliveryStatusSuccess, model.DeliveryStatusFailed,
			model.DeliveryStatusSkipped, model.DeliveryStatusPending,
			model.RecipientTypeBroadcast)).
		// 必须把 m.sender_id 放进 GROUP BY：MySQL 默认开启 ONLY_FULL_GROUP_BY，
		// 否则上面取发送人的子查询会报 1055 错误，整个列表接口 500 后前端显示为空。
		Group("d.message_id, m.title, m.kind, m.sender_id, t.name, m.created_at").
		Order("MAX(d.id) desc").
		Offset((q.Page - 1) * q.PageSize).
		Limit(q.PageSize).
		Scan(&list).Error
	return total, list, err
}

// ListDeliveriesByMessage 返回一次发送下的全部渠道记录。
func (r *CatalogRepository) ListDeliveriesByMessage(ctx context.Context, messageID uint) ([]DeliveryDTO, error) {
	var list []DeliveryDTO
	err := r.db.WithContext(ctx).Table("notification_delivery d").
		Where("d.message_id = ? AND d.deleted_at IS NULL", messageID).
		Select("d.id, d.message_id, d.title, d.user_id, d.channel, d.status, d.attempts, d.last_error, d.recipient_type, d.recipient_name, d.recipient_email, d.sent_at, d.created_at").
		Order("d.channel asc, d.id asc").
		Scan(&list).Error
	return list, err
}
