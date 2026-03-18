package service

import (
	"context"

	"bit-labs.cn/owl-admin/app/model"
	"bit-labs.cn/owl-admin/app/repository"
	"bit-labs.cn/owl/provider/db"
	validatorv10 "github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type AreaService struct {
	areaRepo repository.AreaRepositoryInterface
	validate *validatorv10.Validate
}

func NewAreaService(areaRepo repository.AreaRepositoryInterface, validate *validatorv10.Validate) *AreaService {
	return &AreaService{areaRepo: areaRepo, validate: validate}
}

type RetrieveAllAreaReq struct {
	ParentID uint   `json:"parentId,string" validate:"omitempty"` // 父级区域ID
	NameLike string `json:"name" validate:"omitempty,max=64"`     // 区域名称
}

func (i *AreaService) RetrieveAll(ctx context.Context, req *RetrieveAllAreaReq) (list []model.Area, err error) {
	if err := i.validate.Struct(req); err != nil {
		return nil, err
	}

	return i.areaRepo.WithContext(ctx).ListAll(func(tx *gorm.DB) {
		db.AppendWhereFromStruct(tx, req)
	})
}
