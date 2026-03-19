package service

import (
	"context"
	"strings"

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

type CreateDictReq struct {
	Name   string `json:"name" validate:"required,max=32" label:"字典名"`              // 字典名（中）
	Type   string `json:"type"`                                                     // 字典名（英）
	Status uint8  `json:"status,string" validate:"required,min=1,max=2" label:"状态"` // 状态(1启用,2禁用)
	Desc   string `json:"desc"`                                                     // 描述
	Sort   uint8  `json:"sort,string" validate:"required,min=1,max=255" label:"排序"` // 排序
}

type UpdateDictReq struct {
	ID uint `json:"id,string"` // 字典ID
	CreateDictReq
}

type DictService struct {
	dictRepo repository.DictRepositoryInterface // 字典仓储接口
	locker   redis.LockerFactory                // 分布式锁工厂
	validate *validatorv10.Validate
}

func NewDictService(dictRepo repository.DictRepositoryInterface, locker redis.LockerFactory, validate *validatorv10.Validate) *DictService {
	return &DictService{
		dictRepo: dictRepo,
		locker:   locker,
		validate: validate,
	}
}
func (i DictService) CreateDict(ctx context.Context, req *CreateDictReq) error {
	if err := i.validate.Struct(req); err != nil {
		return err
	}

	l := i.locker.New()
	if err := l.Lock("dict:create"); err != nil {
		return err
	}
	defer l.Unlock()

	dict := new(model.Dict)
	err := copier.Copy(&dict, req)
	if err != nil {
		return err
	}

	return i.dictRepo.WithContext(ctx).Save(dict)
}

func (i DictService) UpdateDict(ctx context.Context, req *UpdateDictReq) error {
	if err := i.validate.Struct(req); err != nil {
		return err
	}

	l := i.locker.New()
	if err := l.Lock("dict:update:" + cast.ToString(req.ID)); err != nil {
		return err
	}
	defer l.Unlock()

	dict, err := i.dictRepo.WithContext(ctx).Detail(req.ID)
	if err != nil {
		return err
	}

	err = copier.Copy(&dict, req)
	if err != nil {
		return err
	}

	return i.dictRepo.WithContext(ctx).Save(dict)
}

type RetrieveDictReq struct {
	router.PageReq
	NameLike string `json:"name" binding:"omitempty,max=64" validate:"omitempty,max=64" label:"字典名"`  // 名称模糊搜索
	StatusIn string `json:"status" binding:"omitempty" validate:"omitempty" label:"状态"`               // 状态 in 查询
	Type     string `json:"type" binding:"omitempty,max=32" validate:"omitempty,max=32" label:"字典类型"` // 字典类型
}

func (i DictService) RetrieveDicts(ctx context.Context, req *RetrieveDictReq) (count int64, list []model.Dict, err error) {
	if err := i.validate.Struct(req); err != nil {
		return 0, nil, err
	}
	return i.dictRepo.WithContext(ctx).Retrieve(req.Page, req.PageSize, func(tx *gorm.DB) {
		db.AppendWhereFromStruct(tx, req)
		tx.Order("created_at desc")
	})
}

// 字典项创建请求
type CreateDictItemReq struct {
	Label    string `json:"label" validate:"required,max=64" label:"展示值"`                // 展示值
	Value    string `json:"value" validate:"required,max=128" label:"字典值"`               // 字典值
	Extend   string `json:"extend" validate:"omitempty,max=255" label:"扩展值"`             // 扩展值
	Status   uint8  `json:"status,string" validate:"required,min=1,max=2" label:"状态"`    // 启用状态
	Sort     uint   `json:"sort,string" validate:"omitempty,min=1,max=65535" label:"排序"` // 排序标记
	DictType string `json:"dictType" validate:"omitempty,max=64" label:"字典类型"`           // 字典类型
	DictID   uint   `json:"dictID,string" validate:"required" label:"字典ID"`              // 字典ID
}

// 字典项更新请求
type UpdateDictItemReq struct {
	ID       uint   `json:"id,string" validate:"required" label:"字典项ID"`                 // 字典项ID
	Label    string `json:"label" validate:"omitempty,max=64" label:"展示值"`               // 展示值
	Value    string `json:"value" validate:"omitempty,max=128" label:"字典值"`              // 字典值
	Extend   string `json:"extend" validate:"omitempty,max=255" label:"扩展值"`             // 扩展值
	Status   uint8  `json:"status,string" validate:"omitempty,min=1,max=2" label:"状态"`   // 启用状态
	Sort     uint   `json:"sort,string" validate:"omitempty,min=1,max=65535" label:"排序"` // 排序标记
	DictType string `json:"dictType" validate:"omitempty,max=64" label:"字典类型"`           // 字典类型
	DictID   uint   `json:"dictID,string" validate:"required" label:"字典ID"`              // 字典ID
}

func (i DictService) CreateItem(ctx context.Context, req *CreateDictItemReq) error {
	if err := i.validate.Struct(req); err != nil {
		return err
	}

	l := i.locker.New()
	if err := l.Lock("dict:item:create:" + cast.ToString(req.DictID)); err != nil {
		return err
	}
	defer l.Unlock()

	_, err := i.dictRepo.WithContext(ctx).Detail(req.DictID)
	if err != nil {
		return err
	}

	var item model.DictItem
	if err := copier.Copy(&item, req); err != nil {
		return err
	}
	return i.dictRepo.WithContext(ctx).CreateItem(&item)
}

func (i DictService) DeleteItems(ctx context.Context, dictID any, itemIds ...string) error {

	l := i.locker.New()
	if err := l.Lock("dict:item:delete:" + cast.ToString(dictID) + ":" + strings.Join(itemIds, ",")); err != nil {
		return err
	}
	defer l.Unlock()

	_, err := i.dictRepo.WithContext(ctx).Detail(dictID)
	if err != nil {
		return err
	}

	return i.dictRepo.WithContext(ctx).DeleteItem(itemIds...)
}
func (i DictService) UpdateItem(ctx context.Context, req *UpdateDictItemReq) error {
	if err := i.validate.Struct(req); err != nil {
		return err
	}

	l := i.locker.New()
	if err := l.Lock("dict:item:update:" + cast.ToString(req.DictID) + ":" + cast.ToString(req.ID)); err != nil {
		return err
	}
	defer l.Unlock()

	_, err := i.dictRepo.WithContext(ctx).Detail(req.DictID)
	if err != nil {
		return err
	}

	var item model.DictItem
	if err := copier.Copy(&item, req); err != nil {
		return err
	}

	return i.dictRepo.WithContext(ctx).UpdateItem(&item)
}

func (i DictService) RetrieveItems(ctx context.Context, dictID any) (int64, []model.DictItem, error) {

	return i.dictRepo.WithContext(ctx).RetrieveItems(dictID)
}

func (i DictService) DeleteDict(ctx context.Context, ids ...string) error {
	l := i.locker.New()
	if err := l.Lock("dict:delete:" + strings.Join(ids, ",")); err != nil {
		return err
	}
	defer l.Unlock()

	return i.dictRepo.WithContext(ctx).DeleteDict(ids...)
}

func (i DictService) GetDictByType(ctx context.Context, dictType string) ([]model.DictItem, error) {

	if _, err := i.dictRepo.WithContext(ctx).DetailByType(dictType); err != nil {
		return nil, err
	}

	_, list, err := i.dictRepo.WithContext(ctx).RetrieveItemsByType(dictType)
	if err != nil {
		return nil, err
	}

	return list, nil
}
