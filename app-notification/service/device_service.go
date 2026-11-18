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

// RegisterDeviceReq 是 App 登记推送令牌的入参。
type RegisterDeviceReq struct {
	Platform   string `json:"platform" validate:"required,max=16" label:"平台"`
	Provider   string `json:"provider" validate:"required,oneof=apns fcm" label:"服务商"`
	Token      string `json:"token" validate:"required,max=512" label:"令牌"`
	AppID      string `json:"appId" validate:"omitempty,max=128" label:"应用标识"`
	DeviceName string `json:"deviceName" validate:"omitempty,max=128" label:"设备名称"`
}

// RetrieveDeviceReq 是设备列表查询。
type RetrieveDeviceReq struct {
	router.PageReq
	PrincipalID uint `json:"principalId,string" form:"principalId" validate:"omitempty" label:"用户"`
	Status      int  `json:"status" form:"status" validate:"omitempty,min=1,max=2" label:"状态"`
}

// DeviceService 管理推送设备。
type DeviceService struct {
	repo     repository.DeviceRepositoryInterface
	db       *gorm.DB
	validate *validatorv10.Validate
}

func NewDeviceService(repo repository.DeviceRepositoryInterface, db *gorm.DB, validate *validatorv10.Validate) *DeviceService {
	return &DeviceService{repo: repo, db: db, validate: validate}
}

func (s *DeviceService) Register(ctx context.Context, req *RegisterDeviceReq) error {
	if err := s.validate.Struct(req); err != nil {
		return err
	}
	p, err := currentPrincipal(ctx)
	if err != nil {
		return err
	}
	now := time.Now()
	row, err := s.repo.WithContext(ctx).FindByToken(req.Token)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		row = &model.Device{Token: req.Token}
	} else if err != nil {
		return err
	}
	row.PrincipalType = p.Type
	row.PrincipalID = p.ID
	row.Platform = req.Platform
	row.Provider = req.Provider
	row.AppID = req.AppID
	row.DeviceName = req.DeviceName
	row.Status = model.DeviceStatusActive
	row.LastActiveAt = &now
	return s.repo.WithContext(ctx).Save(row)
}

func (s *DeviceService) DeleteMine(ctx context.Context, id uint) error {
	p, err := currentPrincipal(ctx)
	if err != nil {
		return err
	}
	return s.repo.WithContext(ctx).DeleteOwned(p.Type, p.ID, id)
}

func (s *DeviceService) Retrieve(ctx context.Context, req *RetrieveDeviceReq) (int64, []model.Device, error) {
	if err := s.validate.Struct(req); err != nil {
		return 0, nil, err
	}
	total, rows, err := s.repo.WithContext(ctx).Retrieve(req.Page, req.PageSize, req.PrincipalID, req.Status)
	if err != nil {
		return 0, nil, err
	}
	ids := make([]uint, 0, len(rows))
	for _, row := range rows {
		if row.PrincipalType == model.PrincipalTypeAdmin && row.PrincipalID > 0 {
			ids = append(ids, row.PrincipalID)
		}
	}
	names := s.principalNames(ctx, ids)
	for i := range rows {
		rows[i].PrincipalName = names[rows[i].PrincipalID]
		rows[i].Token = maskAddress(rows[i].Token)
	}
	return total, rows, nil
}

func (s *DeviceService) principalNames(ctx context.Context, ids []uint) map[uint]string {
	out := map[uint]string{}
	if len(ids) == 0 || s.db == nil {
		return out
	}
	var rows []struct {
		ID       uint
		Username string
		Nickname string
	}
	_ = s.db.WithContext(ctx).Table("admin_user").Select("id, username, nickname").Where("id IN ?", ids).Scan(&rows).Error
	for _, row := range rows {
		name := row.Nickname
		if name == "" {
			name = row.Username
		}
		out[row.ID] = name
	}
	return out
}

func (s *DeviceService) Invalidate(ctx context.Context, id uint) error {
	row, err := s.repo.WithContext(ctx).Detail(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return NotFound("设备不存在")
	}
	if err != nil {
		return err
	}
	row.Status = model.DeviceStatusInvalid
	return s.repo.WithContext(ctx).Save(row)
}
