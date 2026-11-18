package service

import (
	"context"

	"bit-labs.cn/owl-admin/app-notification/model"
	"bit-labs.cn/owl-admin/app-notification/repository"
)

// SubscriptionItem 是一条订阅开关。
type SubscriptionItem struct {
	Category int    `json:"category" validate:"required,min=1,max=4" label:"分类"`
	Channel  string `json:"channel" validate:"required,max=32" label:"渠道"`
	Enabled  int    `json:"enabled" validate:"required,min=1,max=2" label:"开关"`
}

// SubscriptionService 管理当前用户的渠道偏好。安全类不能关闭。
type SubscriptionService struct {
	repo repository.SubscriptionRepositoryInterface
}

func NewSubscriptionService(repo repository.SubscriptionRepositoryInterface) *SubscriptionService {
	return &SubscriptionService{repo: repo}
}

func (s *SubscriptionService) List(ctx context.Context) ([]model.Subscription, error) {
	p, err := currentPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	return s.repo.WithContext(ctx).List(p.Type, p.ID)
}

func (s *SubscriptionService) Save(ctx context.Context, items []SubscriptionItem) error {
	p, err := currentPrincipal(ctx)
	if err != nil {
		return err
	}
	for _, item := range items {
		if item.Category == model.CategorySecurity && item.Enabled != model.StatusEnabled {
			return NotAllowed("安全类通知不能关闭")
		}
		row := &model.Subscription{PrincipalType: p.Type, PrincipalID: p.ID, Category: item.Category, Channel: item.Channel, Enabled: item.Enabled}
		if err := s.repo.WithContext(ctx).Upsert(row); err != nil {
			return err
		}
	}
	return nil
}
