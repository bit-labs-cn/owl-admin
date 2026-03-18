package repository

import (
	"context"

	"bit-labs.cn/owl-admin/app/model"
	"bit-labs.cn/owl/contract"
	"bit-labs.cn/owl/provider/db"
	"gorm.io/gorm"
)

type AreaRepositoryInterface interface {
	contract.Repository[model.Area]
	ListAll(fn func(db *gorm.DB)) (list []model.Area, err error)
	contract.WithContext[AreaRepositoryInterface]
}

var _ AreaRepositoryInterface = (*AreaRepository)(nil)

type AreaRepository struct {
	db  *gorm.DB
	ctx context.Context
	db.BaseRepository[model.Area]
}

func NewAreaRepository(d *gorm.DB) AreaRepositoryInterface {
	return &AreaRepository{
		db:             d,
		BaseRepository: db.NewBaseRepository[model.Area](d),
	}
}

func (i *AreaRepository) WithContext(ctx context.Context) AreaRepositoryInterface {
	i.db = i.db.WithContext(ctx)
	i.ctx = ctx
	return i
}

func (i *AreaRepository) Create(data *model.Area) error {
	return i.BaseRepository.Save(data)
}

func (i *AreaRepository) Update(data *model.Area) error {
	return i.BaseRepository.Save(data)
}

func (i *AreaRepository) Delete(id uint) error {
	return i.BaseRepository.Delete(id)
}

func (i *AreaRepository) Detail(id any) (*model.Area, error) {
	return i.BaseRepository.Detail(id)
}

func (i *AreaRepository) Retrieve(page, pageSize int, fn func(db *gorm.DB)) (count int64, list []model.Area, err error) {
	return i.BaseRepository.Retrieve(page, pageSize, fn)
}

func (i *AreaRepository) ListAll(fn func(db *gorm.DB)) (list []model.Area, err error) {
	newDB := i.db.Model(new(model.Area))
	if fn != nil {
		fn(newDB)
	}
	err = newDB.Find(&list).Error
	return
}
