package service

import (
	"context"
	"errors"

	"bit-labs.cn/owl-admin/app/model"
	"bit-labs.cn/owl-admin/app/repository"
	errContract "bit-labs.cn/owl/contract/errors"
	"gorm.io/gorm"
)

const CodeAppVersionNotFound = "APP_VERSION_NOT_FOUND"

func AppVersionNotFound() *errContract.BizError {
	return errContract.NewBizError(CodeAppVersionNotFound, "暂无可用版本")
}

type AppVersionService struct {
	repo repository.AppVersionRepositoryInterface
}

func NewAppVersionService(repo repository.AppVersionRepositoryInterface) *AppVersionService {
	return &AppVersionService{repo: repo}
}

func (i *AppVersionService) Latest(ctx context.Context, apkType *int32) (*model.AppVersion, error) {
	v, err := i.repo.WithContext(ctx).Latest(apkType)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, AppVersionNotFound()
	}
	return v, err
}
