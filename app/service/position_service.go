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

type CreatePositionReq struct {
	Name   string `json:"name" validate:"required,min=2,max=32" label:"岗位名称"` // 岗位名称
	Remark string `json:"remark" validate:"omitempty,max=255" label:"备注"`     // 备注
	Status int    `json:"status" validate:"required,oneof=1 2" label:"状态"`    // 状态(1启用,2禁用)
}

type UpdatePositionReq struct {
	ID uint `json:"id,string" validate:"required,gt=0" label:"岗位ID"` // 岗位ID
	CreatePositionReq
}

type RetrievePositionReq struct {
	router.PageReq
	NameLike string `json:"name" validate:"omitempty,max=32" label:"岗位名称"`    // 名称模糊搜索
	Status   int    `json:"status" validate:"omitempty,oneof=1 2" label:"状态"` // 状态(1启用,2禁用)
}

type PositionService struct {
	db.BaseRepository[model.Position]
	repo     repository.PositionRepositoryInterface
	locker   redis.LockerFactory
	validate *validatorv10.Validate
}

func NewPositionService(
	repo repository.PositionRepositoryInterface,
	tx *gorm.DB,
	locker redis.LockerFactory,
	validate *validatorv10.Validate,
) *PositionService {
	return &PositionService{BaseRepository: db.NewBaseRepository[model.Position](tx), repo: repo, locker: locker, validate: validate}
}

func (i *PositionService) CreatePosition(ctx context.Context, req *CreatePositionReq) error {
	if err := i.validate.Struct(req); err != nil {
		return err
	}

	l := i.locker.New()
	if err := l.Lock("position:create"); err != nil {
		return err
	}
	defer l.Unlock()

	var m model.Position
	err := copier.Copy(&m, req)
	if err != nil {
		return err
	}

	return i.repo.WithContext(ctx).Save(&m)
}

func (i *PositionService) UpdatePosition(ctx context.Context, req *UpdatePositionReq) error {
	if err := i.validate.Struct(req); err != nil {
		return err
	}

	l := i.locker.New()
	if err := l.Lock("position:update:" + cast.ToString(req.ID)); err != nil {
		return err
	}
	defer l.Unlock()

	m, err := i.repo.WithContext(ctx).Detail(req.ID)
	if err != nil {
		return err
	}

	if err = copier.Copy(m, req); err != nil {
		return err
	}
	return i.repo.WithContext(ctx).Save(m)
}

func (i *PositionService) DeletePosition(ctx context.Context, id uint) error {

	l := i.locker.New()
	if err := l.Lock("position:delete:" + cast.ToString(id)); err != nil {
		return err
	}
	defer l.Unlock()

	return i.BaseRepository.Delete(id)
}

func (i *PositionService) ChangeStatus(ctx context.Context, req *db.ChangeStatus) error {
	if err := i.validate.Struct(req); err != nil {
		return err
	}

	l := i.locker.New()
	if err := l.Lock("position:status:" + cast.ToString(req.ID)); err != nil {
		return err
	}
	defer l.Unlock()

	return i.BaseRepository.ChangeStatus(req)
}

func (i *PositionService) RetrievePositions(ctx context.Context, req *RetrievePositionReq) (count int64, list []model.Position, err error) {
	if err := i.validate.Struct(req); err != nil {
		return 0, nil, err
	}

	return i.repo.WithContext(ctx).Retrieve(req.Page, req.PageSize, func(tx *gorm.DB) {
		db.AppendWhereFromStruct(tx, req)
		tx.Order("created_at desc")
	})
}

func (i *PositionService) Options(ctx context.Context) (list []repository.PositionItem, err error) {
	return i.repo.WithContext(ctx).Options()
}
