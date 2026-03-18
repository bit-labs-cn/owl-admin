package service

import (
	"context"

	"bit-labs.cn/owl-admin/app/model"
	"bit-labs.cn/owl-admin/app/repository"
	"bit-labs.cn/owl/provider/db"
	"bit-labs.cn/owl/provider/router"
	validatorv10 "github.com/go-playground/validator/v10"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

type LogService struct {
	logRepo  repository.LogRepositoryInterface // 日志仓储接口
	validate *validatorv10.Validate
}

func NewLogService(repo repository.LogRepositoryInterface, validate *validatorv10.Validate) *LogService {
	return &LogService{logRepo: repo, validate: validate}
}

// 记录登录日志

type CreateOperationLogReq struct {
	UserId    int    `json:"userId" validate:"required,gte=1"`           // 用户编号
	UserName  string `json:"userName" validate:"required,max=64"`        // 用户名称
	UserType  string `json:"userType" validate:"required,max=32"`        // 用户类型（user/super_admin）
	Method    string `json:"method" validate:"required,max=16"`          // 请求方法
	Path      string `json:"path" validate:"required,max=255"`           // 请求路径
	ApiName   string `json:"apiName" validate:"omitempty,max=64"`        // 接口中文名称
	Status    int    `json:"status" validate:"required,gte=100,lte=599"` // 响应状态码（HTTP）
	CostMs    int    `json:"costMs" validate:"omitempty,gte=0"`          // 耗时毫秒
	Ip        string `json:"ip" validate:"omitempty,max=64"`             // 客户端 IP
	UserAgent string `json:"userAgent" validate:"omitempty,max=255"`     // 客户端 UA
	ReqBody   string `json:"reqBody" validate:"omitempty"`               // 请求体（文本）
}

type CreateLoginLogReq struct {
	UserId    int    `json:"userId" validate:"required,gte=1"`       // 用户编号
	UserName  string `json:"userName" validate:"required,max=64"`    // 用户名称
	UserType  string `json:"userType" validate:"required,max=32"`    // 用户类型（user/super_admin）
	LoginTime int    `json:"loginTime" validate:"required,gte=1"`    // 登录时间（Unix 秒）
	Ip        string `json:"ip" validate:"omitempty,max=64"`         // 客户端 IP
	UserAgent string `json:"userAgent" validate:"omitempty,max=255"` // 客户端 UA
}

func (i *LogService) CreateLoginLog(ctx context.Context, req *CreateLoginLogReq) error {
	if err := i.validate.Struct(req); err != nil {
		return err
	}

	var log model.LoginLog
	err := copier.Copy(&log, req)
	if err != nil {
		return err
	}

	return i.logRepo.WithContext(ctx).SaveLoginLog(&log)
}

func (i *LogService) CreateOperationLog(ctx context.Context, req *CreateOperationLogReq) error {
	if err := i.validate.Struct(req); err != nil {
		return err
	}

	var log model.OperationLog
	err := copier.Copy(&log, req)
	if err != nil {
		return err
	}

	return i.logRepo.WithContext(ctx).SaveOperationLog(&log)
}

type RetrieveLoginLogsReq struct {
	router.PageReq
	UserNameLike string `json:"userName"` // 用户名模糊
	Ip           string `json:"ip"`       // IP 地址
	UserType     string `json:"userType"` // 用户类型（user/super_admin）
}

func (i *LogService) RetrieveLoginLogs(ctx context.Context, req *RetrieveLoginLogsReq) (count int64, list []model.LoginLog, err error) {
	if err := i.validate.Struct(req); err != nil {
		return 0, nil, err
	}

	return i.logRepo.WithContext(ctx).RetrieveLoginLogs(req.Page, req.PageSize, func(tx *gorm.DB) {
		db.AppendWhereFromStruct(tx, req)
		tx.Order("login_time desc")
	})
}

type RetrieveOperationLogsReq struct {
	router.PageReq
	UserNameLike     string `json:"userName"`  // 用户名
	PathLike         string `json:"path"`      // 请求路径（模糊）
	Method           string `json:"method"`    // 请求方法
	Status           *int   `json:"status"`    // 状态码
	CreatedAtBetween string `json:"createdAt"` // 创建时间区间查询
}

func (i *LogService) RetrieveOperationLogs(ctx context.Context, req *RetrieveOperationLogsReq) (count int64, list []model.OperationLog, err error) {
	if err := i.validate.Struct(req); err != nil {
		return 0, nil, err
	}
	return i.logRepo.WithContext(ctx).RetrieveOperationLogs(req.Page, req.PageSize, func(tx *gorm.DB) {
		db.AppendWhereFromStruct(tx, req)
		tx.Order("created_at desc")
	})
}
