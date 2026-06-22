package repository

import (
	"context"

	"bit-labs.cn/owl-admin/app/model"
	"bit-labs.cn/owl/contract"
	"gorm.io/gorm"
)

type TrustedDeviceRepositoryInterface interface {
	FindByUserAndDevice(userID uint, deviceID string) (*model.TrustedDevice, error)
	Upsert(device *model.TrustedDevice) error
	contract.WithContext[TrustedDeviceRepositoryInterface]
}

var _ TrustedDeviceRepositoryInterface = (*TrustedDeviceRepository)(nil)

type TrustedDeviceRepository struct {
	db  *gorm.DB
	ctx context.Context
}

func NewTrustedDeviceRepository(tx *gorm.DB) TrustedDeviceRepositoryInterface {
	return &TrustedDeviceRepository{db: tx}
}

func (r *TrustedDeviceRepository) WithContext(ctx context.Context) TrustedDeviceRepositoryInterface {
	r.db = r.db.WithContext(ctx)
	r.ctx = ctx
	return r
}

func (r *TrustedDeviceRepository) FindByUserAndDevice(userID uint, deviceID string) (*model.TrustedDevice, error) {
	var device model.TrustedDevice
	err := r.db.Where("user_id = ? AND device_id = ?", userID, deviceID).First(&device).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

func (r *TrustedDeviceRepository) Upsert(device *model.TrustedDevice) error {
	var existing model.TrustedDevice
	err := r.db.Where("user_id = ? AND device_id = ?", device.UserID, device.DeviceID).First(&existing).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return r.db.Create(device).Error
		}
		return err
	}
	device.ID = existing.ID
	return r.db.Model(&existing).Updates(map[string]interface{}{
		"last_ip":       device.LastIP,
		"last_ua":       device.LastUA,
		"verified_at":   device.VerifiedAt,
		"last_login_at": device.LastLoginAt,
	}).Error
}
