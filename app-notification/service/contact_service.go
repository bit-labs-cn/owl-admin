package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"bit-labs.cn/owl-admin/app-notification/model"
	"bit-labs.cn/owl-admin/app-notification/repository"
	"bit-labs.cn/owl/provider/redis"
	"bit-labs.cn/owl/provider/router"
	validatorv10 "github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

// CreateContactListReq 是名单入参。
type CreateContactListReq struct {
	Name   string `json:"name" validate:"required,max=128" label:"名称"`
	Remark string `json:"remark" validate:"omitempty,max=255" label:"备注"`
	Status int    `json:"status" validate:"omitempty,min=1,max=2" label:"状态"`
}

// CreateContactReq 是联系人入参。
type CreateContactReq struct {
	ListID uint   `json:"listId,string" validate:"required" label:"名单"`
	Name   string `json:"name" validate:"omitempty,max=128" label:"名称"`
	Email  string `json:"email" validate:"omitempty,email,max=255" label:"邮箱"`
	Phone  string `json:"phone" validate:"omitempty,max=32" label:"手机号"`
	Source string `json:"source" validate:"omitempty,max=64" label:"来源"`
}

// RetrieveContactReq 是联系人查询。
type RetrieveContactReq struct {
	router.PageReq
	ListID uint   `json:"listId,string" form:"listId" validate:"omitempty" label:"名单"`
	Name   string `json:"name" form:"name" validate:"omitempty,max=128" label:"名称"`
	Status int    `json:"status" form:"status" validate:"omitempty,min=1,max=3" label:"状态"`
}

// ContactService 管理外部名单。
type ContactService struct {
	repo     repository.ContactRepositoryInterface
	lists    repository.ContactRepositoryInterface
	locker   redis.LockerFactory
	validate *validatorv10.Validate
}

func NewContactService(repo repository.ContactRepositoryInterface, locker redis.LockerFactory, validate *validatorv10.Validate) *ContactService {
	return &ContactService{repo: repo, locker: locker, validate: validate}
}

func (s *ContactService) SaveList(ctx context.Context, id uint, req *CreateContactListReq) error {
	if err := s.validate.Struct(req); err != nil {
		return err
	}
	status := req.Status
	if status == 0 {
		status = model.StatusEnabled
	}
	row := &model.ContactList{Name: req.Name, Remark: req.Remark, Status: status}
	row.ID = id
	return s.repo.WithContext(ctx).SaveList(row)
}

func (s *ContactService) DeleteList(ctx context.Context, id uint) error {
	if _, err := s.repo.WithContext(ctx).DetailList(id); errors.Is(err, gorm.ErrRecordNotFound) {
		return NotFound("名单不存在")
	} else if err != nil {
		return err
	}
	return s.repo.WithContext(ctx).DeleteList(id)
}

func (s *ContactService) Lists(ctx context.Context, page, pageSize int, name string) (int64, []model.ContactList, error) {
	return s.repo.WithContext(ctx).RetrieveLists(page, pageSize, name)
}

func (s *ContactService) SaveContact(ctx context.Context, id uint, req *CreateContactReq) error {
	if err := s.validate.Struct(req); err != nil {
		return err
	}
	if strings.TrimSpace(req.Email) == "" && strings.TrimSpace(req.Phone) == "" {
		return Invalid("邮箱和手机号至少填一个")
	}
	if _, err := s.repo.WithContext(ctx).DetailList(req.ListID); errors.Is(err, gorm.ErrRecordNotFound) {
		return NotFound("名单不存在")
	} else if err != nil {
		return err
	}
	row := &model.Contact{ListID: req.ListID, Name: req.Name, Email: req.Email, Phone: req.Phone, Status: model.ContactStatusSubscribed, Source: req.Source}
	row.ID = id
	return s.repo.WithContext(ctx).SaveContact(row)
}

func (s *ContactService) Contacts(ctx context.Context, req *RetrieveContactReq) (int64, []model.Contact, error) {
	if err := s.validate.Struct(req); err != nil {
		return 0, nil, err
	}
	rowsTotal, rows, err := s.repo.WithContext(ctx).RetrieveContacts(req.Page, req.PageSize, req.ListID, req.Name, req.Status)
	for i := range rows {
		rows[i].Email = maskAddress(rows[i].Email)
		rows[i].Phone = maskAddress(rows[i].Phone)
	}
	return rowsTotal, rows, err
}

func (s *ContactService) DeleteContact(ctx context.Context, id uint) error {
	return s.repo.WithContext(ctx).DeleteContact(id)
}

func (s *ContactService) Unsubscribe(ctx context.Context, id uint) error {
	row, err := s.repo.WithContext(ctx).DetailContact(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return NotFound("联系人不存在")
	}
	if err != nil {
		return err
	}
	now := time.Now()
	row.Status = model.ContactStatusUnsubscribed
	row.UnsubscribedAt = &now
	return s.repo.WithContext(ctx).SaveContact(row)
}

// ImportContacts 按行导入，每行「名称,邮箱,手机号」。
func (s *ContactService) ImportContacts(ctx context.Context, listID uint, text string) (int, error) {
	if _, err := s.repo.WithContext(ctx).DetailList(listID); errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, NotFound("名单不存在")
	} else if err != nil {
		return 0, err
	}
	n := 0
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		req := &CreateContactReq{ListID: listID, Source: "import"}
		if len(parts) > 0 {
			req.Name = strings.TrimSpace(parts[0])
		}
		if len(parts) > 1 {
			req.Email = strings.TrimSpace(parts[1])
		}
		if len(parts) > 2 {
			req.Phone = strings.TrimSpace(parts[2])
		}
		if err := s.SaveContact(ctx, 0, req); err != nil {
			continue
		}
		n++
	}
	return n, nil
}
