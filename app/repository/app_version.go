package repository

import (
	"context"

	"bit-labs.cn/owl-admin/app/model"
	"bit-labs.cn/owl/contract"
	"bit-labs.cn/owl/provider/db"
	"gorm.io/gorm"
)

type AppVersionRepositoryInterface interface {
	Latest(apkType *int32) (*model.AppVersion, error)
	Retrieve(page, pageSize int, fn func(db *gorm.DB)) (count int64, list []model.AppVersion, err error)
	contract.WithContext[AppVersionRepositoryInterface]
}

var _ AppVersionRepositoryInterface = (*AppVersionRepository)(nil)

type AppVersionRepository struct {
	db  *gorm.DB
	ctx context.Context
	db.BaseRepository[model.AppVersion]
}

func NewAppVersionRepository(tx *gorm.DB) AppVersionRepositoryInterface {
	return &AppVersionRepository{
		db:             tx,
		BaseRepository: db.NewBaseRepository[model.AppVersion](tx),
	}
}

func (i *AppVersionRepository) WithContext(ctx context.Context) AppVersionRepositoryInterface {
	i.db = i.db.WithContext(ctx)
	i.ctx = ctx
	return i
}

func (i *AppVersionRepository) Latest(apkType *int32) (*model.AppVersion, error) {
	var v model.AppVersion

	tx := i.db.Model(&model.AppVersion{})

	if apkType != nil {
		tx = tx.Where("apk_type = ?", *apkType)
	}

	err := tx.Order("id desc").First(&v).Error
	return &v, err
}

func (i *AppVersionRepository) Retrieve(page, pageSize int, fn func(db *gorm.DB)) (count int64, list []model.AppVersion, err error) {
	return i.BaseRepository.Retrieve(page, pageSize, fn)
}
