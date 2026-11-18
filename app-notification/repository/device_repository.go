package repository

import (
	"context"

	"bit-labs.cn/owl-admin/app-notification/model"
	"bit-labs.cn/owl/contract"
	"bit-labs.cn/owl/provider/db"
	"gorm.io/gorm"
)

// DeviceRepositoryInterface 读写推送设备。
type DeviceRepositoryInterface interface {
	Save(row *model.Device) error
	Detail(id uint) (*model.Device, error)
	Delete(id uint) error
	DeleteOwned(principalType int, principalID, id uint) error
	FindByToken(token string) (*model.Device, error)
	ListActive(principalType int, principalID uint) ([]model.Device, error)
	Retrieve(page, pageSize int, principalID uint, status int) (int64, []model.Device, error)
	InvalidateTokens(tokens []string) error
	contract.WithContext[DeviceRepositoryInterface]
}

type DeviceRepository struct {
	db *gorm.DB
	db.BaseRepository[model.Device]
}

func NewDeviceRepository(d *gorm.DB) DeviceRepositoryInterface {
	return &DeviceRepository{db: d, BaseRepository: db.NewBaseRepository[model.Device](d)}
}

func (r *DeviceRepository) WithContext(ctx context.Context) DeviceRepositoryInterface {
	next := r.db.WithContext(ctx)
	return &DeviceRepository{db: next, BaseRepository: db.NewBaseRepository[model.Device](next)}
}

func (r *DeviceRepository) Save(row *model.Device) error { return r.db.Save(row).Error }

func (r *DeviceRepository) Detail(id uint) (*model.Device, error) {
	var row model.Device
	err := r.db.Where("id = ?", id).First(&row).Error
	return &row, err
}

func (r *DeviceRepository) Delete(id uint) error {
	return r.db.Delete(&model.Device{}, id).Error
}

func (r *DeviceRepository) DeleteOwned(principalType int, principalID, id uint) error {
	return r.db.Where("principal_type = ? AND principal_id = ? AND id = ?", principalType, principalID, id).Delete(&model.Device{}).Error
}

func (r *DeviceRepository) FindByToken(token string) (*model.Device, error) {
	var row model.Device
	err := r.db.Where("token = ?", token).First(&row).Error
	return &row, err
}

func (r *DeviceRepository) ListActive(principalType int, principalID uint) ([]model.Device, error) {
	var list []model.Device
	err := r.db.Where("principal_type = ? AND principal_id = ? AND status = ?", principalType, principalID, model.DeviceStatusActive).Find(&list).Error
	return list, err
}

func (r *DeviceRepository) Retrieve(page, pageSize int, principalID uint, status int) (int64, []model.Device, error) {
	tx := r.db.Model(&model.Device{})
	if principalID > 0 {
		tx = tx.Where("principal_id = ?", principalID)
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
	var list []model.Device
	err := tx.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return total, list, err
}

func (r *DeviceRepository) InvalidateTokens(tokens []string) error {
	if len(tokens) == 0 {
		return nil
	}
	return r.db.Model(&model.Device{}).Where("token IN ?", tokens).Update("status", model.DeviceStatusInvalid).Error
}
