package service

import (
	"context"
	"errors"

	"bit-labs.cn/owl-admin/app-notification/model"
	"bit-labs.cn/owl-admin/app-notification/repository"
	"bit-labs.cn/owl/provider/redis"
	"bit-labs.cn/owl/provider/router"
	validatorv10 "github.com/go-playground/validator/v10"
	"github.com/spf13/cast"
	"gorm.io/gorm"
)

// CreateTemplateReq 是模板入参。
type CreateTemplateReq struct {
	Code         string `json:"code" validate:"required,max=64" label:"编码"`
	Name         string `json:"name" validate:"required,max=128" label:"名称"`
	EventCode    string `json:"eventCode" validate:"omitempty,max=64" label:"事件编码"`
	Category     int    `json:"category" validate:"required,min=1,max=4" label:"分类"`
	Kind         int    `json:"kind" validate:"required,min=1,max=3" label:"类型"`
	Title        string `json:"title" validate:"required,max=255" label:"标题"`
	ContentHTML  string `json:"contentHtml" validate:"omitempty" label:"正文"`
	ContentText  string `json:"contentText" validate:"omitempty" label:"纯文本"`
	SMSTemplate  string `json:"smsTemplate" validate:"omitempty,max=64" label:"短信模板"`
	VariableJSON string `json:"variableJson" validate:"omitempty" label:"变量"`
	LinkJSON     string `json:"linkJson" validate:"omitempty" label:"深链"`
	ActionJSON   string `json:"actionJson" validate:"omitempty" label:"动作"`
	Status       int    `json:"status" validate:"omitempty,min=1,max=2" label:"状态"`
	Remark       string `json:"remark" validate:"omitempty,max=255" label:"备注"`
}

// RetrieveTemplateReq 是模板列表查询。
type RetrieveTemplateReq struct {
	router.PageReq
	Name   string `json:"name" form:"name" validate:"omitempty,max=128" label:"名称"`
	Code   string `json:"code" form:"code" validate:"omitempty,max=64" label:"编码"`
	Status int    `json:"status" form:"status" validate:"omitempty,min=1,max=2" label:"状态"`
}

// TemplateService 管理通知模板。
type TemplateService struct {
	repo     repository.TemplateRepositoryInterface
	locker   redis.LockerFactory
	validate *validatorv10.Validate
}

func NewTemplateService(repo repository.TemplateRepositoryInterface, locker redis.LockerFactory, validate *validatorv10.Validate) *TemplateService {
	return &TemplateService{repo: repo, locker: locker, validate: validate}
}

func (s *TemplateService) Create(ctx context.Context, req *CreateTemplateReq) error {
	if err := s.validate.Struct(req); err != nil {
		return err
	}
	l := s.locker.New()
	if err := l.Lock("notify:template:" + req.Code); err != nil {
		return err
	}
	defer l.Unlock()
	if _, err := s.repo.WithContext(ctx).FindByCode(req.Code); err == nil {
		return Conflict("模板编码已存在")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	row := s.fill(req)
	return s.repo.WithContext(ctx).Create(row)
}

func (s *TemplateService) Update(ctx context.Context, id uint, req *CreateTemplateReq) error {
	if err := s.validate.Struct(req); err != nil {
		return err
	}
	row, err := s.repo.WithContext(ctx).Detail(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return NotFound("模板不存在")
	}
	if err != nil {
		return err
	}
	next := s.fill(req)
	next.ID = row.ID
	next.BaseModel = row.BaseModel
	return s.repo.WithContext(ctx).Save(next)
}

func (s *TemplateService) Delete(ctx context.Context, id uint) error {
	if _, err := s.repo.WithContext(ctx).Detail(id); errors.Is(err, gorm.ErrRecordNotFound) {
		return NotFound("模板不存在")
	} else if err != nil {
		return err
	}
	return s.repo.WithContext(ctx).Delete(id)
}

func (s *TemplateService) Detail(ctx context.Context, id uint) (*model.Template, error) {
	row, err := s.repo.WithContext(ctx).Detail(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, NotFound("模板不存在")
	}
	return row, err
}

func (s *TemplateService) Retrieve(ctx context.Context, req *RetrieveTemplateReq) (int64, []model.Template, error) {
	if err := s.validate.Struct(req); err != nil {
		return 0, nil, err
	}
	return s.repo.WithContext(ctx).Retrieve(req.Page, req.PageSize, req.Name, req.Code, req.Status)
}

func (s *TemplateService) Preview(req *CreateTemplateReq) map[string]string {
	htmlBody := SanitizeHTML(applyVars(req.ContentHTML, nil))
	text := req.ContentText
	if text == "" {
		text = HTMLToText(htmlBody)
	}
	return map[string]string{"contentHtml": htmlBody, "contentText": text}
}

func (s *TemplateService) fill(req *CreateTemplateReq) *model.Template {
	status := req.Status
	if status == 0 {
		status = model.StatusEnabled
	}
	htmlBody := SanitizeHTML(req.ContentHTML)
	text := req.ContentText
	if text == "" {
		text = HTMLToText(htmlBody)
	}
	return &model.Template{
		Code: req.Code, Name: req.Name, EventCode: req.EventCode, Category: req.Category, Kind: req.Kind,
		Title: req.Title, ContentHTML: htmlBody, ContentText: text, SMSTemplate: req.SMSTemplate,
		VariableJSON: req.VariableJSON, LinkJSON: req.LinkJSON, ActionJSON: req.ActionJSON, Status: status, Remark: req.Remark,
	}
}

func idOf(v string) uint { return cast.ToUint(v) }
