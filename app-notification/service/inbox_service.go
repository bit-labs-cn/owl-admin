package service

import (
	"context"
	"errors"
	"time"

	"bit-labs.cn/owl-admin/app-notification/model"
	"bit-labs.cn/owl-admin/app-notification/repository"
	"bit-labs.cn/owl/provider/router"
	validatorv10 "github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

// InboxService 面向当前登录用户的站内信和待办。
type InboxService struct {
	repo     repository.InboxRepositoryInterface
	hub      *Hub
	validate *validatorv10.Validate
}

func NewInboxService(repo repository.InboxRepositoryInterface, hub *Hub, validate *validatorv10.Validate) *InboxService {
	return &InboxService{repo: repo, hub: hub, validate: validate}
}

// RetrieveInboxReq 是我的消息查询。
type RetrieveInboxReq struct {
	router.PageReq
	Kind       int    `json:"kind" form:"kind" validate:"omitempty,min=1,max=3" label:"类型"`
	Title      string `json:"title" form:"title" validate:"omitempty,max=255" label:"标题"`
	ReadStatus int    `json:"readStatus" form:"readStatus" validate:"omitempty,min=1,max=2" label:"阅读状态"`
}

func (s *InboxService) Counts(ctx context.Context) (repository.InboxCounts, error) {
	p, err := currentPrincipal(ctx)
	if err != nil {
		return repository.InboxCounts{}, err
	}
	return s.repo.WithContext(ctx).Counts(p.Type, p.ID)
}

func (s *InboxService) List(ctx context.Context, kind, limit int) ([]model.Inbox, error) {
	p, err := currentPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	return s.repo.WithContext(ctx).List(p.Type, p.ID, kind, limit)
}

func (s *InboxService) Retrieve(ctx context.Context, req *RetrieveInboxReq) (int64, []model.Inbox, error) {
	if err := s.validate.Struct(req); err != nil {
		return 0, nil, err
	}
	p, err := currentPrincipal(ctx)
	if err != nil {
		return 0, nil, err
	}
	return s.repo.WithContext(ctx).Retrieve(p.Type, p.ID, req.Kind, req.Title, req.ReadStatus, req.Page, req.PageSize)
}

func (s *InboxService) MarkRead(ctx context.Context, ids []uint) error {
	p, err := currentPrincipal(ctx)
	if err != nil {
		return err
	}
	if err := s.repo.WithContext(ctx).MarkRead(p.Type, p.ID, ids); err != nil {
		return err
	}
	s.hub.Publish(ctx, PushEvent{PrincipalType: p.Type, PrincipalID: p.ID, Event: "read"})
	return nil
}

func (s *InboxService) MarkAllRead(ctx context.Context, kind int) error {
	p, err := currentPrincipal(ctx)
	if err != nil {
		return err
	}
	if err := s.repo.WithContext(ctx).MarkAllRead(p.Type, p.ID, kind); err != nil {
		return err
	}
	s.hub.Publish(ctx, PushEvent{PrincipalType: p.Type, PrincipalID: p.ID, Event: "read"})
	return nil
}

func (s *InboxService) Delete(ctx context.Context, id uint) error {
	p, err := currentPrincipal(ctx)
	if err != nil {
		return err
	}
	return s.repo.WithContext(ctx).Delete(p.Type, p.ID, id)
}

func (s *InboxService) Complete(ctx context.Context, id uint, done bool) error {
	p, err := currentPrincipal(ctx)
	if err != nil {
		return err
	}
	row, err := s.repo.WithContext(ctx).Detail(id)
	if errors.Is(err, gorm.ErrRecordNotFound) || (err == nil && (row.PrincipalID != p.ID || row.Kind != model.KindTodo)) {
		return NotFound("待办不存在")
	}
	if err != nil {
		return err
	}
	if err := s.repo.WithContext(ctx).Complete(p.Type, p.ID, id, done); err != nil {
		return err
	}
	s.hub.Publish(ctx, PushEvent{PrincipalType: p.Type, PrincipalID: p.ID, Event: "todo"})
	return nil
}

func (s *InboxService) CompleteByBizKey(ctx context.Context, bizKey string) error {
	if bizKey == "" {
		return Invalid("业务主键为空")
	}
	rows, err := s.repo.WithContext(ctx).CompleteByBizKey(bizKey)
	if err != nil {
		return err
	}
	seen := map[uint]struct{}{}
	for _, row := range rows {
		if _, ok := seen[row.PrincipalID]; ok {
			continue
		}
		seen[row.PrincipalID] = struct{}{}
		s.hub.Publish(ctx, PushEvent{PrincipalType: row.PrincipalType, PrincipalID: row.PrincipalID, Event: "todo"})
	}
	return nil
}

// StreamTicket 签发当前用户的 SSE 票据。
func (s *InboxService) StreamTicket(ctx context.Context) (string, error) {
	p, err := currentPrincipal(ctx)
	if err != nil {
		return "", err
	}
	return s.hub.IssueTicket(ctx, p)
}

func (s *InboxService) TouchRead(ctx context.Context, id uint) {
	_ = s.MarkRead(ctx, []uint{id})
	_ = time.Now()
}
