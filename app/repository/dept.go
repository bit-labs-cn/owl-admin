package repository

import (
	"context"
	"errors"

	"bit-labs.cn/owl-admin/app/model"
	"bit-labs.cn/owl/contract"
	"bit-labs.cn/owl/provider/db"
	"gorm.io/gorm"
)

var ErrDeptExists = errors.New("部门已存在")
var ErrDeptNotExists = errors.New("部门不存在")

type DeptRepositoryInterface interface {
	contract.Repository[model.Dept]
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
	_, exists := i.BaseRepository.Unique(data.ID, func(db *gorm.DB) {
		db.Where("parent_id = ? and name = ?", data.ParentId, data.Name)
	})

	if exists {
		return ErrDeptExists
	}
	err := i.BaseRepository.Save(data)

	return err
}

func (i *DeptRepository) Update(data *model.Dept) error {
	return i.Create(data)
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
