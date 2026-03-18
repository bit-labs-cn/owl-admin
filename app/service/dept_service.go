package service

import (
	"context"

	"bit-labs.cn/owl-admin/app/model"
	"bit-labs.cn/owl-admin/app/repository"
	"bit-labs.cn/owl/provider/db"
	"bit-labs.cn/owl/provider/redis"
	"bit-labs.cn/owl/provider/router"
	validatorv10 "github.com/go-playground/validator/v10"
	"github.com/jinzhu/copier"
	"github.com/spf13/cast"
	"gorm.io/gorm"
)

type DeptService struct {
	deptRepo repository.DeptRepositoryInterface
	locker   redis.LockerFactory
	validate *validatorv10.Validate
}

func NewDeptService(deptRepo repository.DeptRepositoryInterface, locker redis.LockerFactory, validate *validatorv10.Validate) *DeptService {
	return &DeptService{
		deptRepo: deptRepo,
		locker:   locker,
		validate: validate,
	}
}

type CreateDeptReq struct {
	Name        string `gorm:"comment:部门名称" json:"name" validate:"required,max=64"`                                    // 部门名称
	ParentId    int    `gorm:"comment:父级部门" json:"parentId,string" validate:"omitempty"`                               // 父级部门
	Level       uint   `gorm:"comment:部门层级" json:"level" validate:"omitempty"`                                         // 部门层级
	Sort        uint   `gorm:"comment:排序" json:"sort" validate:"omitempty"`                                            // 排序
	Status      uint   `gorm:"comment:状态" json:"status" validate:"omitempty,oneof=1 2"`                                // 状态(1启用,2禁用)
	Description string `gorm:"comment:描述" json:"description" binding:"omitempty,max=255" validate:"omitempty,max=255"` // 描述
}

type UpdateDeptReq struct {
	ID uint `json:"id,string,omitempty"` // 部门ID
	CreateDeptReq
}

type RetrieveDeptReq struct {
	router.PageReq
	NameLike string `json:"name" validate:"omitempty,max=64"`      // 部门名称模糊搜索
	Status   uint   `json:"status" validate:"omitempty,oneof=1 2"` // 状态(1启用,2禁用)
}

// CreateDept 创建部门
// 就算 CreateDeptReq 直接使用了 model.Dept 作为了结构体，但是也要单独声明 CreateDeptReq 来接收参数，因为这样可扩展性更高
func (i DeptService) CreateDept(ctx context.Context, req *CreateDeptReq) error {
	if err := i.validate.Struct(req); err != nil {
		return err
	}

	l := i.locker.New()
	if err := l.Lock("dept:create"); err != nil {
		return err
	}
	defer l.Unlock()

	var dept model.Dept
	err := copier.Copy(&dept, req)
	if err != nil {
		return err
	}
	return i.deptRepo.WithContext(ctx).Create(&dept)
}

func (i DeptService) UpdateDept(ctx context.Context, req *UpdateDeptReq) error {
	if err := i.validate.Struct(req); err != nil {
		return err
	}

	l := i.locker.New()
	if err := l.Lock("dept:update:" + cast.ToString(req.ID)); err != nil {
		return err
	}
	defer l.Unlock()

	var dept model.Dept
	err := copier.Copy(&dept, req)
	if err != nil {
		return err
	}
	return i.deptRepo.WithContext(ctx).Update(&dept)
}
func (i DeptService) DeleteDept(ctx context.Context, id uint) error {

	l := i.locker.New()
	if err := l.Lock("dept:delete:" + cast.ToString(id)); err != nil {
		return err
	}
	defer l.Unlock()

	return i.deptRepo.WithContext(ctx).Delete(id)
}

func (i DeptService) RetrieveDepts(ctx context.Context, req *RetrieveDeptReq) (count int64, list []model.Dept, err error) {
	if err := i.validate.Struct(req); err != nil {
		return 0, nil, err
	}

	return i.deptRepo.WithContext(ctx).Retrieve(req.Page, req.PageSize, func(tx *gorm.DB) {
		db.AppendWhereFromStruct(tx, req)
		tx.Order("created_at desc")
	})
}
