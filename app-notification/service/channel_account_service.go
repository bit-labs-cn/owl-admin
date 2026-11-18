package service

import (
	"context"
	"encoding/json"
	"errors"

	"bit-labs.cn/owl-admin/app-notification/contract"
	"bit-labs.cn/owl-admin/app-notification/model"
	"bit-labs.cn/owl-admin/app-notification/repository"
	"bit-labs.cn/owl/provider/router"
	validatorv10 "github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

// SaveChannelAccountReq 是渠道账号入参。
type SaveChannelAccountReq struct {
	Channel    string `json:"channel" validate:"required,max=32" label:"渠道"`
	Provider   string `json:"provider" validate:"omitempty,max=32" label:"服务商"`
	Name       string `json:"name" validate:"required,max=128" label:"名称"`
	ConfigJSON string `json:"configJson" validate:"required" label:"配置"`
	IsDefault  int    `json:"isDefault" validate:"omitempty,min=1,max=2" label:"是否默认"`
	Status     int    `json:"status" validate:"omitempty,min=1,max=2" label:"状态"`
}

// RetrieveChannelAccountReq 是渠道账号查询。
type RetrieveChannelAccountReq struct {
	router.PageReq
	Channel string `json:"channel" form:"channel" validate:"omitempty,max=32" label:"渠道"`
}

// ChannelAccountService 管理渠道账号。配置原样保存，列表接口不回传密钥正文以外的说明由前端遮罩。
type ChannelAccountService struct {
	repo     repository.ChannelAccountRepositoryInterface
	email    *EmailChannel
	sms      *SMSChannel
	validate *validatorv10.Validate
}

func NewChannelAccountService(repo repository.ChannelAccountRepositoryInterface, email *EmailChannel, sms *SMSChannel, validate *validatorv10.Validate) *ChannelAccountService {
	return &ChannelAccountService{repo: repo, email: email, sms: sms, validate: validate}
}

func (s *ChannelAccountService) Save(ctx context.Context, id uint, req *SaveChannelAccountReq) error {
	if err := s.validate.Struct(req); err != nil {
		return err
	}
	status := req.Status
	if status == 0 {
		status = model.StatusEnabled
	}
	isDefault := req.IsDefault
	if isDefault == 0 {
		isDefault = model.StatusDisabled
	}
	row := &model.ChannelAccount{Channel: req.Channel, Provider: req.Provider, Name: req.Name, ConfigJSON: req.ConfigJSON, IsDefault: isDefault, Status: status}
	if id > 0 {
		old, err := s.repo.WithContext(ctx).Detail(id)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NotFound("渠道账号不存在")
		}
		if err != nil {
			return err
		}
		row.BaseModel = old.BaseModel
	}
	row.ID = id
	return s.repo.WithContext(ctx).Save(row)
}

func (s *ChannelAccountService) Delete(ctx context.Context, id uint) error {
	if _, err := s.repo.WithContext(ctx).Detail(id); errors.Is(err, gorm.ErrRecordNotFound) {
		return NotFound("渠道账号不存在")
	} else if err != nil {
		return err
	}
	return s.repo.WithContext(ctx).Delete(id)
}

func (s *ChannelAccountService) Retrieve(ctx context.Context, req *RetrieveChannelAccountReq) (int64, []model.ChannelAccount, error) {
	if err := s.validate.Struct(req); err != nil {
		return 0, nil, err
	}
	total, rows, err := s.repo.WithContext(ctx).Retrieve(req.Page, req.PageSize, req.Channel)
	for i := range rows {
		rows[i].ConfigJSON = maskConfig(rows[i].ConfigJSON)
	}
	return total, rows, err
}

func (s *ChannelAccountService) Detail(ctx context.Context, id uint) (*model.ChannelAccount, error) {
	row, err := s.repo.WithContext(ctx).Detail(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, NotFound("渠道账号不存在")
	}
	return row, err
}

// TestSend 用该账号发一条测试。address 是邮箱、手机号或推送 token。
func (s *ChannelAccountService) TestSend(ctx context.Context, id uint, address string) error {
	row, err := s.Detail(ctx, id)
	if err != nil {
		return err
	}
	req := &contract.ChannelRequest{
		Title: "通知中心测试", ContentHTML: "<p>这是一封测试通知</p>", ContentText: "这是一条测试通知",
		Address: address, ConfigJSON: row.ConfigJSON, Provider: row.Provider, Kind: model.KindNotice,
		Variables: map[string]string{"code": "123456"},
	}
	var extra struct {
		TemplateCode string `json:"templateCode"`
	}
	_ = json.Unmarshal([]byte(row.ConfigJSON), &extra)
	req.SMSTemplate = extra.TemplateCode
	switch row.Channel {
	case model.ChannelEmail:
		_, err = s.email.Send(ctx, req)
	case model.ChannelSMS:
		_, err = s.sms.Send(ctx, req)
	case model.ChannelApp:
		return Invalid("请在 App 内登记设备后再测试推送")
	default:
		return Invalid("该渠道不支持测试发送")
	}
	return err
}

func maskConfig(raw string) string {
	if len(raw) <= 8 {
		return "已配置"
	}
	return raw[:4] + "…" + "已隐藏"
}
