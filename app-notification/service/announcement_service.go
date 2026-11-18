package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"bit-labs.cn/owl-admin/app-notification/contract"
	"bit-labs.cn/owl-admin/app-notification/model"
	"bit-labs.cn/owl-admin/app-notification/repository"
	"bit-labs.cn/owl/provider/redis"
	"bit-labs.cn/owl/provider/router"
	validatorv10 "github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

// CreateAnnouncementReq 是公告入参。
type CreateAnnouncementReq struct {
	Title       string            `json:"title" validate:"required,max=255" label:"标题"`
	ContentHTML string            `json:"contentHtml" validate:"required" label:"正文"`
	Level       int               `json:"level" validate:"omitempty,min=1,max=4" label:"级别"`
	DisplayMode int               `json:"displayMode" validate:"omitempty,min=1,max=3" label:"展示方式"`
	Audience    contract.Audience `json:"audience" label:"受众"`
	StartAt     *time.Time        `json:"startAt" label:"开始时间"`
	EndAt       *time.Time        `json:"endAt" label:"结束时间"`
	Closable    int               `json:"closable" validate:"omitempty,min=1,max=2" label:"可否关闭"`
	Frequency   int               `json:"frequency" validate:"omitempty,min=1,max=3" label:"展示频率"`
	Status      int               `json:"status" validate:"omitempty,min=1,max=3" label:"状态"`
	Sort        int               `json:"sort" validate:"omitempty" label:"排序"`
}

// RetrieveAnnouncementReq 是公告列表查询。
type RetrieveAnnouncementReq struct {
	router.PageReq
	Title  string `json:"title" form:"title" validate:"omitempty,max=255" label:"标题"`
	Status int    `json:"status" form:"status" validate:"omitempty,min=1,max=3" label:"状态"`
}

// AnnouncementService 管理站点公告。
type AnnouncementService struct {
	repo     repository.AnnouncementRepositoryInterface
	hub      *Hub
	audience *AudienceResolver
	locker   redis.LockerFactory
	validate *validatorv10.Validate
}

func NewAnnouncementService(repo repository.AnnouncementRepositoryInterface, hub *Hub, audience *AudienceResolver, locker redis.LockerFactory, validate *validatorv10.Validate) *AnnouncementService {
	return &AnnouncementService{repo: repo, hub: hub, audience: audience, locker: locker, validate: validate}
}

func (s *AnnouncementService) Create(ctx context.Context, req *CreateAnnouncementReq) error {
	if err := s.validate.Struct(req); err != nil {
		return err
	}
	l := s.locker.New()
	if err := l.Lock("notify:announcement:create"); err != nil {
		return err
	}
	defer l.Unlock()
	if err := s.repo.WithContext(ctx).Create(s.fill(req)); err != nil {
		return err
	}
	if req.Status == model.AnnouncementStatusPublished {
		s.broadcast(ctx, req.Title)
	}
	return nil
}

func (s *AnnouncementService) Update(ctx context.Context, id uint, req *CreateAnnouncementReq) error {
	if err := s.validate.Struct(req); err != nil {
		return err
	}
	row, err := s.repo.WithContext(ctx).Detail(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return NotFound("公告不存在")
	}
	if err != nil {
		return err
	}
	next := s.fill(req)
	next.ID = row.ID
	next.BaseModel = row.BaseModel
	if err := s.repo.WithContext(ctx).Save(next); err != nil {
		return err
	}
	// 无论发布、下线还是改回草稿，都通知前端重新拉公告，避免用户看到过期内容。
	s.broadcast(ctx, next.Title)
	return nil
}

func (s *AnnouncementService) Delete(ctx context.Context, id uint) error {
	row, err := s.repo.WithContext(ctx).Detail(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return NotFound("公告不存在")
	}
	if err != nil {
		return err
	}
	if err := s.repo.WithContext(ctx).Delete(id); err != nil {
		return err
	}
	s.broadcast(ctx, row.Title)
	return nil
}

func (s *AnnouncementService) Detail(ctx context.Context, id uint) (*model.Announcement, error) {
	row, err := s.repo.WithContext(ctx).Detail(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, NotFound("公告不存在")
	}
	return row, err
}

func (s *AnnouncementService) Retrieve(ctx context.Context, req *RetrieveAnnouncementReq) (int64, []model.Announcement, error) {
	if err := s.validate.Struct(req); err != nil {
		return 0, nil, err
	}
	return s.repo.WithContext(ctx).Retrieve(req.Page, req.PageSize, req.Title, req.Status)
}

func (s *AnnouncementService) Mine(ctx context.Context) ([]model.Announcement, error) {
	p, err := currentPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.WithContext(ctx).ListPublished(time.Now())
	if err != nil {
		return nil, err
	}
	dismissed, err := s.repo.WithContext(ctx).DismissedIDs(p.Type, p.ID)
	if err != nil {
		return nil, err
	}
	skip := map[uint]struct{}{}
	for _, id := range dismissed {
		skip[id] = struct{}{}
	}
	out := make([]model.Announcement, 0, len(rows))
	for _, row := range rows {
		if _, ok := skip[row.ID]; ok && row.Closable == model.StatusEnabled {
			continue
		}
		var audience contract.Audience
		if raw := strings.TrimSpace(row.AudienceJSON); raw != "" {
			if err := json.Unmarshal([]byte(raw), &audience); err != nil {
				// 受众配置损坏时宁可不展示，也不要发给无关的人。
				continue
			}
		}
		matched, err := s.audience.Match(ctx, audience, p.Type, p.ID)
		if err != nil {
			return nil, err
		}
		if !matched {
			continue
		}
		out = append(out, row)
	}
	return out, nil
}

func (s *AnnouncementService) Dismiss(ctx context.Context, id uint) error {
	p, err := currentPrincipal(ctx)
	if err != nil {
		return err
	}
	return s.repo.WithContext(ctx).Dismiss(p.Type, p.ID, id)
}

func (s *AnnouncementService) Publish(ctx context.Context, id uint, online bool) error {
	row, err := s.repo.WithContext(ctx).Detail(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return NotFound("公告不存在")
	}
	if err != nil {
		return err
	}
	if online {
		row.Status = model.AnnouncementStatusPublished
	} else {
		row.Status = model.AnnouncementStatusOffline
	}
	if err := s.repo.WithContext(ctx).Save(row); err != nil {
		return err
	}
	s.broadcast(ctx, row.Title)
	return nil
}

func (s *AnnouncementService) fill(req *CreateAnnouncementReq) *model.Announcement {
	level := req.Level
	if level == 0 {
		level = model.LevelInfo
	}
	mode := req.DisplayMode
	if mode == 0 {
		mode = model.DisplayModeBanner
	}
	closable := req.Closable
	if closable == 0 {
		closable = model.StatusEnabled
	}
	freq := req.Frequency
	if freq == 0 {
		freq = model.FrequencyOnce
	}
	status := req.Status
	if status == 0 {
		status = model.AnnouncementStatusDraft
	}
	htmlBody := SanitizeHTML(req.ContentHTML)
	return &model.Announcement{
		Title: req.Title, ContentHTML: htmlBody, ContentText: HTMLToText(htmlBody), Level: level, DisplayMode: mode,
		AudienceJSON: toJSON(req.Audience), StartAt: req.StartAt, EndAt: req.EndAt, Closable: closable, Frequency: freq, Status: status, Sort: req.Sort,
	}
}

func (s *AnnouncementService) broadcast(ctx context.Context, title string) {
	s.hub.Publish(ctx, PushEvent{Event: "announcement", Payload: map[string]string{"title": title}})
}
