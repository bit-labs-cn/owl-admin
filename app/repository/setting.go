package repository

import (
	"context"
	"errors"

	"bit-labs.cn/owl-admin/app/model"
	"bit-labs.cn/owl/contract"
	"gorm.io/gorm"
)

// SettingRepositoryInterface 读写配置覆盖值。
type SettingRepositoryInterface interface {
	All() ([]model.Setting, error)
	List(subsystem string) ([]model.Setting, error)
	Get(subsystem, key string) (model.Setting, bool, error)
	Apply(subsystem string, upserts map[string]string, drops []string) error
	contract.WithContext[SettingRepositoryInterface]
}

// SettingRepository 用 admin_setting 保存覆盖值。
type SettingRepository struct {
	db  *gorm.DB
	ctx context.Context
}

var _ SettingRepositoryInterface = (*SettingRepository)(nil)

// NewSettingRepository 创建配置仓储。
func NewSettingRepository(d *gorm.DB) SettingRepositoryInterface {
	return &SettingRepository{db: d}
}

// WithContext 绑定本次请求的上下文。
func (i *SettingRepository) WithContext(ctx context.Context) SettingRepositoryInterface {
	i.db = i.db.WithContext(ctx)
	i.ctx = ctx
	return i
}

// All 返回全部覆盖值。
func (i *SettingRepository) All() ([]model.Setting, error) {
	var list []model.Setting
	err := i.db.Find(&list).Error
	return list, err
}

// List 返回一个子系统的覆盖值。
func (i *SettingRepository) List(subsystem string) ([]model.Setting, error) {
	var list []model.Setting
	err := i.db.Where(&model.Setting{Subsystem: subsystem}).Find(&list).Error
	return list, err
}

// Get 读取一条覆盖值。没有记录时 found 为 false。
func (i *SettingRepository) Get(subsystem, key string) (model.Setting, bool, error) {
	var row model.Setting
	err := i.db.Where(&model.Setting{Subsystem: subsystem, Key: key}).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.Setting{}, false, nil
	}
	if err != nil {
		return model.Setting{}, false, err
	}
	return row, true, nil
}

// Apply 写入变更。与默认值相同的键从 drops 删除，其余 upsert。
func (i *SettingRepository) Apply(subsystem string, upserts map[string]string, drops []string) error {
	return i.db.Transaction(func(tx *gorm.DB) error {
		for _, key := range drops {
			err := tx.Unscoped().Where(&model.Setting{Subsystem: subsystem, Key: key}).Delete(&model.Setting{}).Error
			if err != nil {
				return err
			}
		}
		for key, value := range upserts {
			var row model.Setting
			err := tx.Where(&model.Setting{Subsystem: subsystem, Key: key}).First(&row).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				row = model.Setting{Subsystem: subsystem, Key: key, Value: value}
				if err = tx.Create(&row).Error; err != nil {
					return err
				}
				continue
			}
			if err != nil {
				return err
			}
			if err = tx.Model(&row).Update("value", value).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
