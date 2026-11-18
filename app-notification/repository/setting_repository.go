package repository

import (
	"context"

	"bit-labs.cn/owl-admin/app-notification/model"
	"bit-labs.cn/owl/contract"
	"bit-labs.cn/owl/provider/db"
	"gorm.io/gorm"
)

// SettingRepositoryInterface 读写通知运行配置。
type SettingRepositoryInterface interface {
	Get(key string) (string, bool, error)
	Set(key, value string) error
	List() ([]model.Setting, error)
	contract.WithContext[SettingRepositoryInterface]
}

type SettingRepository struct {
	db *gorm.DB
	db.BaseRepository[model.Setting]
}

func NewSettingRepository(d *gorm.DB) SettingRepositoryInterface {
	return &SettingRepository{db: d, BaseRepository: db.NewBaseRepository[model.Setting](d)}
}

func (r *SettingRepository) WithContext(ctx context.Context) SettingRepositoryInterface {
	next := r.db.WithContext(ctx)
	return &SettingRepository{db: next, BaseRepository: db.NewBaseRepository[model.Setting](next)}
}

func (r *SettingRepository) Get(key string) (string, bool, error) {
	var row model.Setting
	err := r.db.Where("key = ?", key).First(&row).Error
	if err == gorm.ErrRecordNotFound {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return row.Value, true, nil
}

func (r *SettingRepository) Set(key, value string) error {
	var row model.Setting
	err := r.db.Where("key = ?", key).First(&row).Error
	if err == gorm.ErrRecordNotFound {
		row.Key = key
		row.Value = value
		return r.db.Create(&row).Error
	}
	if err != nil {
		return err
	}
	row.Value = value
	return r.db.Save(&row).Error
}

func (r *SettingRepository) List() ([]model.Setting, error) {
	var list []model.Setting
	err := r.db.Order("key asc").Find(&list).Error
	return list, err
}
