package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"bit-labs.cn/owl-admin/app/model"
	"bit-labs.cn/owl-admin/app/repository"
	errContract "bit-labs.cn/owl/contract/errors"
	"bit-labs.cn/owl/provider/db"
	"bit-labs.cn/owl/provider/redis"
	"bit-labs.cn/owl/provider/router"
	filestorage "bit-labs.cn/owl/provider/storage"
	validatorv10 "github.com/go-playground/validator/v10"
	"github.com/jinzhu/copier"
	"github.com/spf13/cast"
	"gorm.io/gorm"
)

const (
	CodeAppVersionNotFound     = "APP_VERSION_NOT_FOUND"
	CodeAppVersionExists       = "APP_VERSION_EXISTS"
	CodeAppPackageInvalid      = "APP_PACKAGE_INVALID"
	CodeAppPackageTooLarge     = "APP_PACKAGE_TOO_LARGE"
	CodeAppPackageTypeMismatch = "APP_PACKAGE_TYPE_MISMATCH"
	CodeAppVersionFormat       = "APP_VERSION_FORMAT"

	maxAppPackageSize = 500 * 1024 * 1024
)

var (
	buildVersionRe = regexp.MustCompile(`^[1-9]\d*$`)
	semverNameRe   = regexp.MustCompile(`^\d+\.\d+\.\d+$`)
)

func AppVersionNotFound() *errContract.BizError {
	return errContract.NewBizError(CodeAppVersionNotFound, "暂无可用版本")
}

func AppVersionExists() *errContract.BizError {
	return errContract.NewBizError(CodeAppVersionExists, "该平台已存在相同版本号")
}

func AppPackageInvalid() *errContract.BizError {
	return errContract.NewBizError(CodeAppPackageInvalid, "仅支持 apk / ipa / aab 安装包")
}

func AppPackageTooLarge() *errContract.BizError {
	return errContract.NewBizError(CodeAppPackageTooLarge, "安装包不能超过 500MB")
}

func AppPackageTypeMismatch() *errContract.BizError {
	return errContract.NewBizError(CodeAppPackageTypeMismatch, "安装包类型与文件扩展名不匹配")
}

func AppVersionFormatInvalid() *errContract.BizError {
	return errContract.NewBizError(CodeAppVersionFormat, "版本号须为整数（如 100），版本名称须为 x.y.z（如 1.0.0）")
}

func validateVersionFields(version, versionName string) error {
	if !buildVersionRe.MatchString(version) || !semverNameRe.MatchString(versionName) {
		return AppVersionFormatInvalid()
	}
	return nil
}

type CreateAppVersionReq struct {
	Version     string `json:"version" validate:"required,max=16" label:"版本号"`
	VersionName string `json:"versionName" validate:"required,max=32" label:"版本名称"`
	ApkURL      string `json:"apkUrl" validate:"required,max=512" label:"安装包"`
	ApkType     int32  `json:"apkType" validate:"required,oneof=1 2" label:"安装包类型"`
	Content     string `json:"content" validate:"omitempty,max=2000" label:"更新内容"`
	Remark      string `json:"remark" validate:"omitempty,max=255" label:"备注"`
	Status      int32  `json:"status" validate:"oneof=0 1" label:"状态"`
}

type UploadPackageResult struct {
	URL          string `json:"url"`
	OriginalName string `json:"originalName"`
}

type UpdateAppVersionReq struct {
	ID int64 `json:"id,string" validate:"required,gt=0" label:"版本ID"`
	CreateAppVersionReq
}

type RetrieveAppVersionReq struct {
	router.PageReq
	VersionLike string `json:"version" form:"version" validate:"omitempty,max=64" label:"版本号"`
	ApkType     *int32 `json:"apkType" form:"apkType" validate:"omitempty,oneof=1 2" label:"安装包类型"`
	Status      *int32 `json:"status" form:"status" validate:"omitempty,oneof=0 1" label:"状态"`
}

type AppVersionService struct {
	db.BaseRepository[model.AppVersion]
	repo     repository.AppVersionRepositoryInterface
	locker   redis.LockerFactory
	validate *validatorv10.Validate
	storage  *filestorage.StorageManager
}

func NewAppVersionService(
	repo repository.AppVersionRepositoryInterface,
	tx *gorm.DB,
	locker redis.LockerFactory,
	validate *validatorv10.Validate,
	storage *filestorage.StorageManager,
) *AppVersionService {
	return &AppVersionService{
		BaseRepository: db.NewBaseRepository[model.AppVersion](tx),
		repo:           repo,
		locker:         locker,
		validate:       validate,
		storage:        storage,
	}
}

func (i *AppVersionService) UploadPackage(ctx context.Context, filename string, size int64, apkType int32, reader io.Reader) (*UploadPackageResult, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".apk", ".ipa", ".aab":
	default:
		return nil, AppPackageInvalid()
	}
	if size > maxAppPackageSize {
		return nil, AppPackageTooLarge()
	}
	if apkType == 1 && ext != ".ipa" {
		return nil, AppPackageTypeMismatch()
	}
	if apkType == 2 && ext != ".apk" && ext != ".aab" {
		return nil, AppPackageTypeMismatch()
	}

	objectPath := fmt.Sprintf("app-versions/%s/%d%s", time.Now().Format("2006/01/02"), time.Now().UnixNano(), ext)
	info, err := i.storage.Put(ctx, objectPath, reader, size)
	if err != nil {
		return nil, err
	}
	return &UploadPackageResult{URL: info.URL, OriginalName: filename}, nil
}

func (i *AppVersionService) Latest(ctx context.Context, apkType *int32) (*model.AppVersion, error) {
	v, err := i.repo.WithContext(ctx).Latest(apkType)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, AppVersionNotFound()
	}
	return v, err
}

func (i *AppVersionService) Create(ctx context.Context, req *CreateAppVersionReq) error {
	if err := i.validate.Struct(req); err != nil {
		return err
	}
	if err := validateVersionFields(req.Version, req.VersionName); err != nil {
		return err
	}

	l := i.locker.New()
	if err := l.Lock("app-version:create"); err != nil {
		return err
	}
	defer l.Unlock()

	exists, err := i.repo.WithContext(ctx).ExistsByVersionAndType(0, req.Version, req.ApkType)
	if err != nil {
		return err
	}
	if exists {
		return AppVersionExists()
	}

	var m model.AppVersion
	if err := copier.Copy(&m, req); err != nil {
		return err
	}
	now := time.Now()
	m.CreateTime = &now
	m.UpdateTime = &now
	return i.repo.WithContext(ctx).Save(&m)
}

func (i *AppVersionService) Update(ctx context.Context, req *UpdateAppVersionReq) error {
	if err := i.validate.Struct(req); err != nil {
		return err
	}
	if err := validateVersionFields(req.Version, req.VersionName); err != nil {
		return err
	}

	l := i.locker.New()
	if err := l.Lock("app-version:update:" + cast.ToString(req.ID)); err != nil {
		return err
	}
	defer l.Unlock()

	m, err := i.repo.WithContext(ctx).Detail(req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return AppVersionNotFound()
		}
		return err
	}

	exists, err := i.repo.WithContext(ctx).ExistsByVersionAndType(req.ID, req.Version, req.ApkType)
	if err != nil {
		return err
	}
	if exists {
		return AppVersionExists()
	}

	if err = copier.Copy(m, req); err != nil {
		return err
	}
	now := time.Now()
	m.UpdateTime = &now
	return i.repo.WithContext(ctx).Save(m)
}

func (i *AppVersionService) Delete(ctx context.Context, id int64) error {
	l := i.locker.New()
	if err := l.Lock("app-version:delete:" + cast.ToString(id)); err != nil {
		return err
	}
	defer l.Unlock()

	return i.repo.WithContext(ctx).Delete(id)
}

func (i *AppVersionService) ChangeStatus(ctx context.Context, req *db.ChangeStatus) error {
	if err := i.validate.Struct(req); err != nil {
		return err
	}

	l := i.locker.New()
	if err := l.Lock("app-version:status:" + cast.ToString(req.ID)); err != nil {
		return err
	}
	defer l.Unlock()

	return i.BaseRepository.ChangeStatus(req)
}

func (i *AppVersionService) Retrieve(ctx context.Context, req *RetrieveAppVersionReq) (count int64, list []model.AppVersion, err error) {
	if err := i.validate.Struct(req); err != nil {
		return 0, nil, err
	}

	return i.repo.WithContext(ctx).Retrieve(req.Page, req.PageSize, func(tx *gorm.DB) {
		if req.VersionLike != "" {
			tx.Where("version LIKE ?", "%"+req.VersionLike+"%")
		}
		if req.ApkType != nil {
			tx.Where("apk_type = ?", *req.ApkType)
		}
		if req.Status != nil {
			tx.Where("status = ?", *req.Status)
		}
		tx.Order("id desc")
	})
}
