package service

import (
	"context"
	"errors"

	"bit-labs.cn/owl-admin/app-notification/contract"
	"bit-labs.cn/owl-admin/app-notification/model"
	"bit-labs.cn/owl-admin/app-notification/repository"
	"bit-labs.cn/owl/provider/redis"
	"bit-labs.cn/owl/provider/router"
	validatorv10 "github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

// SaveEventReq 是事件目录入参。
type SaveEventReq struct {
	Code        string   `json:"code" validate:"required,max=64" label:"编码"`
	Name        string   `json:"name" validate:"required,max=128" label:"名称"`
	Source      string   `json:"source" validate:"omitempty,max=64" label:"来源"`
	Kind        int      `json:"kind" validate:"required,min=1,max=3" label:"类型"`
	Category    int      `json:"category" validate:"required,min=1,max=4" label:"分类"`
	Channels    []string `json:"channels" validate:"omitempty" label:"默认渠道"`
	Description string   `json:"description" validate:"omitempty,max=500" label:"说明"`
	Status      int      `json:"status" validate:"omitempty,min=1,max=2" label:"状态"`
}

// RetrieveEventReq 是事件查询。
type RetrieveEventReq struct {
	router.PageReq
	Name   string `json:"name" form:"name" validate:"omitempty,max=128" label:"名称"`
	Source string `json:"source" form:"source" validate:"omitempty,max=64" label:"来源"`
}

// EventService 维护业务事件目录。
type EventService struct {
	repo     repository.EventRepositoryInterface
	locker   redis.LockerFactory
	validate *validatorv10.Validate
}

func NewEventService(repo repository.EventRepositoryInterface, locker redis.LockerFactory, validate *validatorv10.Validate) *EventService {
	return &EventService{repo: repo, locker: locker, validate: validate}
}

func (s *EventService) Save(ctx context.Context, id uint, req *SaveEventReq) error {
	if err := s.validate.Struct(req); err != nil {
		return err
	}
	l := s.locker.New()
	if err := l.Lock("notify:event:" + req.Code); err != nil {
		return err
	}
	defer l.Unlock()
	status := req.Status
	if status == 0 {
		status = model.StatusEnabled
	}
	row := &model.Event{Code: req.Code, Name: req.Name, Source: req.Source, Kind: req.Kind, Category: req.Category, ChannelsJSON: toJSON(req.Channels), Description: req.Description, Status: status}
	if id == 0 {
		if old, err := s.repo.WithContext(ctx).FindByCode(req.Code); err == nil {
			row.ID = old.ID
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
	} else {
		row.ID = id
	}
	return s.repo.WithContext(ctx).Save(row)
}

func (s *EventService) Delete(ctx context.Context, id uint) error {
	return s.repo.WithContext(ctx).Delete(id)
}

func (s *EventService) Retrieve(ctx context.Context, req *RetrieveEventReq) (int64, []model.Event, error) {
	if err := s.validate.Struct(req); err != nil {
		return 0, nil, err
	}
	return s.repo.WithContext(ctx).Retrieve(req.Page, req.PageSize, req.Name, req.Source)
}

func (s *EventService) Register(ctx context.Context, events []contract.EventDescriptor) error {
	for _, ev := range events {
		if err := s.Save(ctx, 0, &SaveEventReq{
			Code: ev.Code, Name: ev.Name, Source: ev.Source, Kind: ev.Kind, Category: ev.Category,
			Channels: ev.DefaultChannels, Description: ev.Description, Status: model.StatusEnabled,
		}); err != nil {
			return err
		}
	}
	return nil
}
