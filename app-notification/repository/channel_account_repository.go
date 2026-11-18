package repository

import (
	"context"

	"bit-labs.cn/owl-admin/app-notification/model"
	"bit-labs.cn/owl/contract"
	"bit-labs.cn/owl/provider/db"
	"gorm.io/gorm"
)

// ChannelAccountRepositoryInterface 读写渠道账号。
type ChannelAccountRepositoryInterface interface {
	Save(row *model.ChannelAccount) error
	Detail(id uint) (*model.ChannelAccount, error)
	Delete(id uint) error
	Retrieve(page, pageSize int, channel string) (int64, []model.ChannelAccount, error)
	FindDefault(channel string) (*model.ChannelAccount, error)
	contract.WithContext[ChannelAccountRepositoryInterface]
}

type ChannelAccountRepository struct {
	db *gorm.DB
	db.BaseRepository[model.ChannelAccount]
}

func NewChannelAccountRepository(d *gorm.DB) ChannelAccountRepositoryInterface {
	return &ChannelAccountRepository{db: d, BaseRepository: db.NewBaseRepository[model.ChannelAccount](d)}
}

func (r *ChannelAccountRepository) WithContext(ctx context.Context) ChannelAccountRepositoryInterface {
	next := r.db.WithContext(ctx)
	return &ChannelAccountRepository{db: next, BaseRepository: db.NewBaseRepository[model.ChannelAccount](next)}
}

func (r *ChannelAccountRepository) Save(row *model.ChannelAccount) error { return r.db.Save(row).Error }

func (r *ChannelAccountRepository) Detail(id uint) (*model.ChannelAccount, error) {
	var row model.ChannelAccount
	err := r.db.Where("id = ?", id).First(&row).Error
	return &row, err
}

func (r *ChannelAccountRepository) Delete(id uint) error {
	return r.db.Delete(&model.ChannelAccount{}, id).Error
}

func (r *ChannelAccountRepository) Retrieve(page, pageSize int, channel string) (int64, []model.ChannelAccount, error) {
	tx := r.db.Model(&model.ChannelAccount{})
	if channel != "" {
		tx = tx.Where("channel = ?", channel)
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
	var list []model.ChannelAccount
	err := tx.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return total, list, err
}

func (r *ChannelAccountRepository) FindDefault(channel string) (*model.ChannelAccount, error) {
	var row model.ChannelAccount
	err := r.db.Where("channel = ? AND status = ? AND is_default = ?", channel, model.StatusEnabled, model.StatusEnabled).First(&row).Error
	if err == gorm.ErrRecordNotFound {
		err = r.db.Where("channel = ? AND status = ?", channel, model.StatusEnabled).Order("id asc").First(&row).Error
	}
	return &row, err
}
