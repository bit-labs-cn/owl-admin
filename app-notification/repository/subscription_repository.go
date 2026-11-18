package repository

import (
	"context"

	"bit-labs.cn/owl-admin/app-notification/model"
	"bit-labs.cn/owl/contract"
	"bit-labs.cn/owl/provider/db"
	"gorm.io/gorm"
)

// SubscriptionRepositoryInterface 读写订阅偏好。
type SubscriptionRepositoryInterface interface {
	List(principalType int, principalID uint) ([]model.Subscription, error)
	Upsert(row *model.Subscription) error
	Allows(principalType int, principalID uint, category int, channel string) (bool, error)
	contract.WithContext[SubscriptionRepositoryInterface]
}

type SubscriptionRepository struct {
	db *gorm.DB
	db.BaseRepository[model.Subscription]
}

func NewSubscriptionRepository(d *gorm.DB) SubscriptionRepositoryInterface {
	return &SubscriptionRepository{db: d, BaseRepository: db.NewBaseRepository[model.Subscription](d)}
}

func (r *SubscriptionRepository) WithContext(ctx context.Context) SubscriptionRepositoryInterface {
	next := r.db.WithContext(ctx)
	return &SubscriptionRepository{db: next, BaseRepository: db.NewBaseRepository[model.Subscription](next)}
}

func (r *SubscriptionRepository) List(principalType int, principalID uint) ([]model.Subscription, error) {
	var list []model.Subscription
	err := r.db.Where("principal_type = ? AND principal_id = ?", principalType, principalID).Find(&list).Error
	return list, err
}

func (r *SubscriptionRepository) Upsert(row *model.Subscription) error {
	var old model.Subscription
	err := r.db.Where("principal_type = ? AND principal_id = ? AND category = ? AND channel = ?", row.PrincipalType, row.PrincipalID, row.Category, row.Channel).First(&old).Error
	if err == gorm.ErrRecordNotFound {
		return r.db.Create(row).Error
	}
	if err != nil {
		return err
	}
	old.Enabled = row.Enabled
	return r.db.Save(&old).Error
}

func (r *SubscriptionRepository) Allows(principalType int, principalID uint, category int, channel string) (bool, error) {
	var row model.Subscription
	err := r.db.Where("principal_type = ? AND principal_id = ? AND category = ? AND channel = ?", principalType, principalID, category, channel).First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return row.Enabled == model.StatusEnabled, nil
}
