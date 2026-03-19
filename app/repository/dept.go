package repository

import (
	"context"

	"bit-labs.cn/owl-admin/app/model"
	"bit-labs.cn/owl/contract"
	"bit-labs.cn/owl/provider/db"
	"gorm.io/gorm"
)

type DeptRepositoryInterface interface {
	contract.Repository[model.Dept]
	// Unique 用于唯一性判断：当存在与 (parent_id, name) 匹配的其他记录时返回 true。
	// id > 0 时会排除自身（id != ?），用于 update 场景。
	Unique(id uint, parentID int, name string) bool
	contract.WithContext[DeptRepositoryInterface]
}

var _ DeptRepositoryInterface = (*DeptRepository)(nil)

type DeptRepository struct {
	db  *gorm.DB
	ctx context.Context
	db.BaseRepository[model.Dept]
}

func NewDeptRepository(d *gorm.DB) DeptRepositoryInterface {
	return &DeptRepository{
		db:             d,
		BaseRepository: db.NewBaseRepository[model.Dept](d),
	}
}

func (i *DeptRepository) WithContext(ctx context.Context) DeptRepositoryInterface {
	i.db = i.db.WithContext(ctx)
	i.ctx = ctx
	return i
}
func (i *DeptRepository) Create(data *model.Dept) error {
	return i.BaseRepository.Save(data)
}

func (i *DeptRepository) Update(data *model.Dept) error {
	return i.Create(data)
}

func (i *DeptRepository) Unique(id uint, parentID int, name string) bool {
	_, exists := i.BaseRepository.Unique(id, func(db *gorm.DB) {
		db.Where("parent_id = ? and name = ?", parentID, name)
	})
	return exists
}

func (i *DeptRepository) Delete(id uint) error {
	return i.BaseRepository.Delete(id)
}

func (i *DeptRepository) Detail(id any) (*model.Dept, error) {
	return i.BaseRepository.Detail(id)
}

func (i *DeptRepository) Retrieve(page, pageSize int, fn func(db *gorm.DB)) (count int64, list []model.Dept, err error) {
	return i.BaseRepository.Retrieve(page, pageSize, fn)
}
