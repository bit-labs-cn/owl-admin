package service

import (
	"context"

	"bit-labs.cn/owl-admin/app/model"
	"bit-labs.cn/owl-admin/app/repository"
	errContract "bit-labs.cn/owl/contract/errors"
	"bit-labs.cn/owl/provider/db"
	"bit-labs.cn/owl/provider/redis"
	"bit-labs.cn/owl/provider/router"
	validatorv10 "github.com/go-playground/validator/v10"
	"github.com/jinzhu/copier"
	"github.com/spf13/cast"
	"gorm.io/gorm"
)

const (
	CodeUserGroupExists   = "USER_GROUP_EXISTS"
	CodeUserGroupNotFound = "USER_GROUP_NOT_FOUND"
)

func UserGroupExists() *errContract.BizError {
	return errContract.NewBizError(CodeUserGroupExists, "用户组已存在")
}

func UserGroupNotFound() *errContract.BizError {
	return errContract.NewBizError(CodeUserGroupNotFound, "用户组不存在")
}

type CreateUserGroupReq struct {
	Name    string   `json:"name" validate:"required,min=2,max=32" label:"用户组名称"`          // 用户组名称
	Code    string   `json:"code" validate:"required,alphanum,min=2,max=32" label:"用户组编码"` // 用户组编码
	Status  int      `json:"status" validate:"omitempty,oneof=1 2" label:"状态"`              // 状态(1启用,2禁用)
	Remark  string   `json:"remark" validate:"omitempty,max=255" label:"备注"`                // 备注
	UserIDs []string `json:"userIDs"`                                                        // 关联用户ID列表
}

type UpdateUserGroupReq struct {
	ID uint `json:"id,string" validate:"required" label:"用户组ID"` // 用户组ID
	CreateUserGroupReq
}

type RetrieveUserGroupReq struct {
	router.PageReq
	NameLike string `json:"name" form:"name" validate:"omitempty,max=32" label:"用户组名称"`          // 名称模糊搜索
	CodeLike string `json:"code" form:"code" validate:"omitempty,alphanum,max=32" label:"用户组编码"` // 编码模糊搜索
	Status   uint8  `json:"status" form:"status" validate:"omitempty,oneof=1 2" label:"状态"`       // 状态(1启用,2禁用)
}

type AssignUsersToGroupReq struct {
	GroupID uint     `json:"groupID,string" validate:"required" label:"用户组ID"` // 用户组ID
	UserIDs []string `json:"userIDs" validate:"required" label:"用户ID列表"`       // 用户ID列表
}


type UserGroupService struct {
	db.BaseRepository[model.UserGroup]
	repo     repository.UserGroupRepositoryInterface
	userRepo repository.UserRepositoryInterface
	locker   redis.LockerFactory
	validate *validatorv10.Validate
}

func NewUserGroupService(
	repo repository.UserGroupRepositoryInterface,
	userRepo repository.UserRepositoryInterface,
	tx *gorm.DB,
	locker redis.LockerFactory,
	validate *validatorv10.Validate,
) *UserGroupService {
	return &UserGroupService{
		BaseRepository: db.NewBaseRepository[model.UserGroup](tx),
		repo:           repo,
		userRepo:       userRepo,
		locker:         locker,
		validate:       validate,
	}
}

func (i *UserGroupService) CreateUserGroup(ctx context.Context, req *CreateUserGroupReq) error {
	if err := i.validate.Struct(req); err != nil {
		return err
	}

	l := i.locker.New()
	if err := l.Lock("user-group:create"); err != nil {
		return err
	}
	defer l.Unlock()

	if _, exists := i.repo.WithContext(ctx).Unique(0, req.Name, req.Code); exists {
		return UserGroupExists()
	}

	var group model.UserGroup
	if err := copier.Copy(&group, req); err != nil {
		return err
	}
	group.Enable()
	if req.UserIDs != nil {
		group.Users = db.GetModelsByIDs[model.User](req.UserIDs)
	}
	return i.repo.WithContext(ctx).Save(&group)
}

func (i *UserGroupService) UpdateUserGroup(ctx context.Context, req *UpdateUserGroupReq) error {
	if err := i.validate.Struct(req); err != nil {
		return err
	}

	l := i.locker.New()
	if err := l.Lock("user-group:update:" + cast.ToString(req.ID)); err != nil {
		return err
	}
	defer l.Unlock()

	if _, exists := i.repo.WithContext(ctx).Unique(req.ID, req.Name, req.Code); exists {
		return UserGroupExists()
	}

	group, err := i.repo.WithContext(ctx).Detail(req.ID)
	if err != nil {
		return err
	}

	if err = copier.Copy(group, req); err != nil {
		return err
	}
	if req.UserIDs != nil {
		group.Users = db.GetModelsByIDs[model.User](req.UserIDs)
	}
	return i.repo.WithContext(ctx).Save(group)
}

func (i *UserGroupService) DeleteUserGroup(ctx context.Context, id uint) error {
	l := i.locker.New()
	if err := l.Lock("user-group:delete:" + cast.ToString(id)); err != nil {
		return err
	}
	defer l.Unlock()

	return i.BaseRepository.Delete(id)
}

func (i *UserGroupService) ChangeStatus(ctx context.Context, req *db.ChangeStatus) error {
	if err := i.validate.Struct(req); err != nil {
		return err
	}

	l := i.locker.New()
	if err := l.Lock("user-group:status:" + cast.ToString(req.ID)); err != nil {
		return err
	}
	defer l.Unlock()

	return i.BaseRepository.ChangeStatus(req)
}

func (i *UserGroupService) RetrieveUserGroups(ctx context.Context, req *RetrieveUserGroupReq) (count int64, list []model.UserGroup, err error) {
	if err := i.validate.Struct(req); err != nil {
		return 0, nil, err
	}

	return i.repo.WithContext(ctx).Retrieve(req.Page, req.PageSize, func(tx *gorm.DB) {
		db.AppendWhereFromStruct(tx, req)
		tx.Order("created_at desc")
	})
}

func (i *UserGroupService) Options(ctx context.Context) (list []repository.UserGroupItem, err error) {
	return i.repo.WithContext(ctx).Options()
}

func (i *UserGroupService) GetUsersByGroupID(ctx context.Context, groupID uint, page, pageSize int) (count int64, list []model.User, err error) {
	return i.repo.WithContext(ctx).GetUsersByGroupID(groupID, page, pageSize)
}

func (i *UserGroupService) AssignUsersToGroup(ctx context.Context, req *AssignUsersToGroupReq) error {
	if err := i.validate.Struct(req); err != nil {
		return err
	}

	l := i.locker.New()
	if err := l.Lock("user-group:assign-users:" + cast.ToString(req.GroupID)); err != nil {
		return err
	}
	defer l.Unlock()

	group, err := i.repo.WithContext(ctx).Detail(req.GroupID)
	if err != nil {
		return err
	}

	group.SetUsers(db.GetModelsByIDs[model.User](req.UserIDs))
	return i.repo.WithContext(ctx).Save(group)
}

func (i *UserGroupService) GetUserIDsByGroupID(ctx context.Context, groupID uint) ([]string, error) {
	return i.repo.WithContext(ctx).GetUserIDsByGroupID(groupID)
}
