package repository

import (
	"context"

	"bit-labs.cn/owl-admin/app/model"
	"bit-labs.cn/owl/contract"
	"bit-labs.cn/owl/provider/db"
	"gorm.io/gorm"
)

type DictRepositoryInterface interface {
	Save(dict *model.Dict) error
	DeleteDict(ids ...string) error
	Unique(id uint, name string, Type string) (*model.Dict, bool)
	Detail(id any) (*model.Dict, error)
	Retrieve(page, pageSize int, fn func(db *gorm.DB)) (count int64, list []model.Dict, err error)
	DetailByType(dictType string) (*model.Dict, error)

	CreateItem(item *model.DictItem) error
	DeleteItem(itemIds ...string) error
	UpdateItem(item *model.DictItem) error
	RetrieveItems(dictID any) (count int64, list []model.DictItem, err error)
	RetrieveItemsByType(dictType string) (count int64, list []model.DictItem, err error)
	contract.WithContext[DictRepositoryInterface]
}

var _ DictRepositoryInterface = (*DictRepository)(nil)

type DictRepository struct {
	db  *gorm.DB
	ctx context.Context
	db.BaseRepository[model.Dict]
}

func NewDictRepository(d *gorm.DB) DictRepositoryInterface {
	return &DictRepository{
		db:             d,
		BaseRepository: db.NewBaseRepository[model.Dict](d),
	}
}

func (i *DictRepository) WithContext(ctx context.Context) DictRepositoryInterface {
	i.db = i.db.WithContext(ctx)
	i.ctx = ctx
	return i
}
func (i *DictRepository) Save(dict *model.Dict) error {
	err := i.db.Where("id", dict.ID).Save(&dict).Error
	return err
}

func (i *DictRepository) Detail(id any) (*model.Dict, error) {
	var m model.Dict
	err := i.db.Where("id = ?", id).First(&m).Error
	return &m, err
}

func (i *DictRepository) DetailByType(dictType string) (*model.Dict, error) {
	var m model.Dict
	err := i.db.Where("type = ?", dictType).First(&m).Error
	return &m, err
}

func (i *DictRepository) DeleteDict(ids ...string) error {
	return i.db.Model(&model.Dict{}).Where("id in ?", ids).Delete(nil).Error
}

func (i *DictRepository) CreateItem(item *model.DictItem) error {
	return i.db.Create(item).Error
}

func (i *DictRepository) DeleteItem(itemIds ...string) error {
	return i.db.Delete(&model.DictItem{}, itemIds).Error
}
func (i *DictRepository) UpdateItem(item *model.DictItem) error {
	return i.db.Updates(item).Error
}
func (i *DictRepository) RetrieveItems(dictID any) (count int64, list []model.DictItem, err error) {
	tx := i.db.Model(&model.DictItem{}).Where("dict_id = ?", dictID)
	if err := tx.Count(&count).Error; err != nil {
		return 0, nil, err
	}
	if err := tx.Order("created_at desc").Find(&list).Error; err != nil {
		return 0, nil, err
	}
	return count, list, nil
}

func (i *DictRepository) RetrieveItemsByType(dictType string) (count int64, list []model.DictItem, err error) {
	tx := i.db.Model(&model.DictItem{}).Where("dict_type = ?", dictType).Where("status = ?", 1)
	if err := tx.Count(&count).Error; err != nil {
		return 0, nil, err
	}
	if err := tx.Order("created_at desc").Find(&list).Error; err != nil {
		return 0, nil, err
	}
	return count, list, nil
}
func (i *DictRepository) Unique(id uint, name string, Type string) (*model.Dict, bool) {
	return i.BaseRepository.Unique(id, func(db *gorm.DB) {
		db.Where("name = ? or type = ?", name, Type)
	})
}
