package repository

import (
	"context"

	"bit-labs.cn/owl-admin/app/model"
	"bit-labs.cn/owl/contract"
	"bit-labs.cn/owl/provider/db"
	"gorm.io/gorm"
)

type PositionRepositoryInterface interface {
	Save(p *model.Position) error
	Detail(id any) (*model.Position, error)
	Retrieve(page, pageSize int, fn func(db *gorm.DB)) (count int64, list []model.Position, err error)
	Options() (list []PositionItem, err error)
	contract.WithContext[PositionRepositoryInterface]
}

type PositionRepository struct {
	db  *gorm.DB
	ctx context.Context
	db.BaseRepository[model.Position]
}

var _ PositionRepositoryInterface = (*PositionRepository)(nil)

func NewPositionRepository(d *gorm.DB) PositionRepositoryInterface {
	return &PositionRepository{db: d, BaseRepository: db.NewBaseRepository[model.Position](d)}
}

func (i *PositionRepository) WithContext(ctx context.Context) PositionRepositoryInterface {
	i.db = i.db.WithContext(ctx)
	i.ctx = ctx
	return i
}

func (i *PositionRepository) Save(p *model.Position) error { return i.BaseRepository.Save(p) }

func (i *PositionRepository) Detail(id any) (*model.Position, error) {
	return i.BaseRepository.Detail(id)
}

func (i *PositionRepository) Retrieve(page, pageSize int, fn func(db *gorm.DB)) (count int64, list []model.Position, err error) {
	return i.BaseRepository.Retrieve(page, pageSize, fn)
}

type PositionItem struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

func (i *PositionRepository) Options() (result []PositionItem, err error) {
	err = i.db.Model(&model.Position{}).Scan(&result).Error
	return
}
