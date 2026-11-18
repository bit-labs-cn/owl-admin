package repository

import (
	"context"
	"time"

	"bit-labs.cn/owl-admin/app-notification/model"
	"bit-labs.cn/owl/contract"
	"bit-labs.cn/owl/provider/db"
	"gorm.io/gorm"
)

// DeliveryRepositoryInterface 读写投递记录。
type DeliveryRepositoryInterface interface {
	Create(row *model.Delivery) error
	Save(row *model.Delivery) error
	Detail(id uint) (*model.Delivery, error)
	Retrieve(page, pageSize int, taskID uint, channel string, status int) (int64, []model.Delivery, error)
	ClaimDue(now time.Time, limit int) ([]model.Delivery, error)
	CountToday(recipientType int, recipientID uint, channel string, since time.Time) (int64, error)
	DeleteBefore(t time.Time) error
	contract.WithContext[DeliveryRepositoryInterface]
}

type DeliveryRepository struct {
	db *gorm.DB
	db.BaseRepository[model.Delivery]
}

func NewDeliveryRepository(d *gorm.DB) DeliveryRepositoryInterface {
	return &DeliveryRepository{db: d, BaseRepository: db.NewBaseRepository[model.Delivery](d)}
}

func (r *DeliveryRepository) WithContext(ctx context.Context) DeliveryRepositoryInterface {
	next := r.db.WithContext(ctx)
	return &DeliveryRepository{db: next, BaseRepository: db.NewBaseRepository[model.Delivery](next)}
}

func (r *DeliveryRepository) Create(row *model.Delivery) error { return r.db.Create(row).Error }
func (r *DeliveryRepository) Save(row *model.Delivery) error   { return r.db.Save(row).Error }

func (r *DeliveryRepository) Detail(id uint) (*model.Delivery, error) {
	var row model.Delivery
	err := r.db.Where("id = ?", id).First(&row).Error
	return &row, err
}

func (r *DeliveryRepository) Retrieve(page, pageSize int, taskID uint, channel string, status int) (int64, []model.Delivery, error) {
	tx := r.db.Model(&model.Delivery{})
	if taskID > 0 {
		tx = tx.Where("task_id = ?", taskID)
	}
	if channel != "" {
		tx = tx.Where("channel = ?", channel)
	}
	if status > 0 {
		tx = tx.Where("status = ?", status)
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return 0, nil, err
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	var list []model.Delivery
	err := tx.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return total, list, err
}

func (r *DeliveryRepository) ClaimDue(now time.Time, limit int) ([]model.Delivery, error) {
	var list []model.Delivery
	err := r.db.Where("status = ? AND (next_retry_at IS NULL OR next_retry_at <= ?)", model.DeliveryStatusPending, now).
		Order("id asc").Limit(limit).Find(&list).Error
	return list, err
}

func (r *DeliveryRepository) CountToday(recipientType int, recipientID uint, channel string, since time.Time) (int64, error) {
	var total int64
	err := r.db.Model(&model.Delivery{}).
		Where("recipient_type = ? AND recipient_id = ? AND channel = ? AND created_at >= ? AND status <> ?",
			recipientType, recipientID, channel, since, model.DeliveryStatusSkipped).
		Count(&total).Error
	return total, err
}

func (r *DeliveryRepository) DeleteBefore(t time.Time) error {
	return r.db.Where("created_at < ?", t).Delete(&model.Delivery{}).Error
}
