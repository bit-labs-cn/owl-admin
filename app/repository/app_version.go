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
	Save(v *model.AppVersion) error
	Detail(id any) (*model.AppVersion, error)
	Delete(ids ...any) error
	ExistsByVersionAndType(id int64, version string, apkType int32) (bool, error)
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

	tx := i.db.Model(&model.AppVersion{}).Where("(status IS NULL OR status = ?)", int32(1))

	if apkType != nil {
		tx = tx.Where("apk_type = ?", *apkType)
	}

	err := tx.Order("id desc").First(&v).Error
	return &v, err
}

func (i *AppVersionRepository) Retrieve(page, pageSize int, fn func(db *gorm.DB)) (count int64, list []model.AppVersion, err error) {
	return i.BaseRepository.Retrieve(page, pageSize, fn)
}

func (i *AppVersionRepository) Save(v *model.AppVersion) error {
	return i.BaseRepository.Save(v)
}

func (i *AppVersionRepository) Detail(id any) (*model.AppVersion, error) {
	return i.BaseRepository.Detail(id)
}

func (i *AppVersionRepository) Delete(ids ...any) error {
	return i.BaseRepository.Delete(ids...)
}

func (i *AppVersionRepository) ExistsByVersionAndType(id int64, version string, apkType int32) (bool, error) {
	var count int64
	tx := i.db.Model(&model.AppVersion{}).Where("version = ? AND apk_type = ?", version, apkType)
	if id > 0 {
		tx = tx.Where("id != ?", id)
	}
	err := tx.Count(&count).Error
	return count > 0, err
}
