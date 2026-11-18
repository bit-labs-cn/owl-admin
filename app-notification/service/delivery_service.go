package service

import (
	"context"

	"bit-labs.cn/owl-admin/app-notification/model"
	"bit-labs.cn/owl-admin/app-notification/repository"
	"bit-labs.cn/owl/provider/router"
	validatorv10 "github.com/go-playground/validator/v10"
)

// RetrieveDeliveryReq 是投递记录查询。
type RetrieveDeliveryReq struct {
	router.PageReq
	TaskID  uint   `json:"taskId,string" form:"taskId" validate:"omitempty" label:"任务"`
	Channel string `json:"channel" form:"channel" validate:"omitempty,max=32" label:"渠道"`
	Status  int    `json:"status" form:"status" validate:"omitempty,min=1,max=5" label:"状态"`
}

// DeliveryService 查询投递记录。地址在返回前脱敏。
type DeliveryService struct {
	repo     repository.DeliveryRepositoryInterface
	tasks    *TaskService
	validate *validatorv10.Validate
}

func NewDeliveryService(repo repository.DeliveryRepositoryInterface, tasks *TaskService, validate *validatorv10.Validate) *DeliveryService {
	return &DeliveryService{repo: repo, tasks: tasks, validate: validate}
}

func (s *DeliveryService) Retrieve(ctx context.Context, req *RetrieveDeliveryReq) (int64, []model.Delivery, error) {
	if err := s.validate.Struct(req); err != nil {
		return 0, nil, err
	}
	total, rows, err := s.repo.WithContext(ctx).Retrieve(req.Page, req.PageSize, req.TaskID, req.Channel, req.Status)
	if err != nil {
		return 0, nil, err
	}
	ids := make([]uint, 0, len(rows))
	seen := map[uint]struct{}{}
	for _, row := range rows {
		if _, ok := seen[row.TaskID]; ok {
			continue
		}
		seen[row.TaskID] = struct{}{}
		ids = append(ids, row.TaskID)
	}
	titles, err := s.tasks.Titles(ctx, ids)
	if err != nil {
		return 0, nil, err
	}
	for i := range rows {
		rows[i].TaskTitle = titles[rows[i].TaskID]
		rows[i].Address = maskAddress(rows[i].Address)
	}
	return total, rows, nil
}

func (s *DeliveryService) Detail(ctx context.Context, id uint) (*model.Delivery, error) {
	row, err := s.repo.WithContext(ctx).Detail(id)
	if err != nil {
		return nil, NotFound("投递记录不存在")
	}
	row.Address = maskAddress(row.Address)
	return row, nil
}

func (s *DeliveryService) Retry(ctx context.Context, id uint) error {
	return s.tasks.Retry(ctx, id)
}
