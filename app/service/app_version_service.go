package service

import (
	"context"
	"errors"

	"bit-labs.cn/owl-admin/app/model"
	"bit-labs.cn/owl-admin/app/repository"
)

var ErrNoAvailableAppVersion = errors.New("暂无可用版本")

type AppVersionService struct {
	repo repository.AppVersionRepositoryInterface
}

func NewAppVersionService(repo repository.AppVersionRepositoryInterface) *AppVersionService {
	return &AppVersionService{repo: repo}
}

func (i *AppVersionService) Latest(ctx context.Context, apkType *int32) (*model.AppVersion, error) {
	v, err := i.repo.WithContext(ctx).Latest(apkType)
	if errors.Is(err, repository.ErrAppVersionNotFound) {
		return nil, ErrNoAvailableAppVersion
	}
	return v, err
}
