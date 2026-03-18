package repository

import (
	"context"

	"bit-labs.cn/owl-admin/app/model"
	"bit-labs.cn/owl/contract"
	"bit-labs.cn/owl/provider/db"
	"gorm.io/gorm"
)

// LogRepositoryInterface 日志仓储接口
type LogRepositoryInterface interface {
	// SaveLoginLog 保存登录日志
	SaveLoginLog(log *model.LoginLog) error
	// SaveOperationLog 保存操作日志
	SaveOperationLog(log *model.OperationLog) error
	// 分页查询登录日志
	RetrieveLoginLogs(page, pageSize int, fn func(tx *gorm.DB)) (count int64, list []model.LoginLog, err error)
	// 分页查询操作日志
	RetrieveOperationLogs(page, pageSize int, fn func(tx *gorm.DB)) (count int64, list []model.OperationLog, err error)
	contract.WithContext[LogRepositoryInterface]
}

var _ LogRepositoryInterface = (*LogRepository)(nil)

type LogRepository struct {
	db        *gorm.DB
	ctx       context.Context
	loginBase db.BaseRepository[model.LoginLog]
	opernBase db.BaseRepository[model.OperationLog]
}

func NewLogRepository(d *gorm.DB) LogRepositoryInterface {
	return &LogRepository{
		db:        d,
		loginBase: db.NewBaseRepository[model.LoginLog](d),
		opernBase: db.NewBaseRepository[model.OperationLog](d),
	}
}

func (i *LogRepository) WithContext(ctx context.Context) LogRepositoryInterface {
	i.db = i.db.WithContext(ctx)
	i.ctx = ctx
	return i
}

func (i *LogRepository) SaveLoginLog(log *model.LoginLog) error {
	return i.db.Create(log).Error
}

func (i *LogRepository) SaveOperationLog(log *model.OperationLog) error {
	return i.db.Create(log).Error
}

func (i *LogRepository) RetrieveLoginLogs(page, pageSize int, fn func(tx *gorm.DB)) (count int64, list []model.LoginLog, err error) {
	tx := i.db.Model(&model.LoginLog{})
	if fn != nil {
		fn(tx)
	}
	err = tx.Count(&count).Error
	if err != nil {
		return
	}
	err = tx.Scopes(db.Paginate(page, pageSize)).Order("created_at desc").Find(&list).Error
	return
}

func (i *LogRepository) RetrieveOperationLogs(page, pageSize int, fn func(tx *gorm.DB)) (count int64, list []model.OperationLog, err error) {
	tx := i.db.Model(&model.OperationLog{})
	if fn != nil {
		fn(tx)
	}
	err = tx.Count(&count).Error
	if err != nil {
		return
	}
	err = tx.Scopes(db.Paginate(page, pageSize)).Order("created_at desc").Find(&list).Error
	return
}
