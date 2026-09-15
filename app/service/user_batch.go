package service

import (
	"context"

	"bit-labs.cn/owl-admin/app/event"
	"bit-labs.cn/owl-admin/app/model"
	errContract "bit-labs.cn/owl/contract/errors"
	"bit-labs.cn/owl/provider/db"
	"github.com/spf13/cast"
)

const (
	UserBatchMaxIDs = 100

	CodeUserBatchEmpty   = "USER_BATCH_EMPTY"
	CodeUserBatchTooMany = "USER_BATCH_TOO_MANY"
)

func UserBatchEmpty() *errContract.BizError {
	return errContract.NewBizError(CodeUserBatchEmpty, "请选择要操作的用户")
}

func UserBatchTooMany() *errContract.BizError {
	return errContract.NewBizError(CodeUserBatchTooMany, "单次最多操作 100 个用户")
}

// BatchUserIDsReq 批量操作用户 ID 列表
type BatchUserIDsReq struct {
	IDs []string `json:"ids" validate:"required,min=1,max=100,dive,required" label:"用户ID列表"`
}

// BatchChangeUserStatusReq 批量修改用户状态
type BatchChangeUserStatusReq struct {
	BatchUserIDsReq
	Status int `json:"status" validate:"oneof=0 1" label:"状态"`
}

// BatchAssignUserRolesReq 批量分配角色（覆盖原有角色）
type BatchAssignUserRolesReq struct {
	BatchUserIDsReq
	RoleIDs []string `json:"roleIDs" validate:"omitempty,dive,required" label:"角色ID列表"`
}

// BatchAssignUserDeptReq 批量设置归属部门；parentId 为空表示清空部门
type BatchAssignUserDeptReq struct {
	BatchUserIDsReq
	DeptID string `json:"parentId" validate:"omitempty" label:"归属部门"`
}

func normalizeBatchUserIDs(ids []string) ([]uint, error) {
	if len(ids) == 0 {
		return nil, UserBatchEmpty()
	}
	if len(ids) > UserBatchMaxIDs {
		return nil, UserBatchTooMany()
	}
	out := make([]uint, 0, len(ids))
	seen := map[uint]struct{}{}
	for _, raw := range ids {
		id := cast.ToUint(raw)
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	if len(out) == 0 {
		return nil, UserBatchEmpty()
	}
	if len(out) > UserBatchMaxIDs {
		return nil, UserBatchTooMany()
	}
	return out, nil
}

// BatchDeleteUsers 批量删除用户
func (i *UserService) BatchDeleteUsers(ctx context.Context, req *BatchUserIDsReq) error {
	if err := i.validate.Struct(req); err != nil {
		return err
	}
	ids, err := normalizeBatchUserIDs(req.IDs)
	if err != nil {
		return err
	}

	l := i.locker.New()
	if err := l.Lock("user:batch-delete"); err != nil {
		return err
	}
	defer l.Unlock()

	anyIDs := make([]any, len(ids))
	for n, id := range ids {
		anyIDs[n] = id
	}
	return i.userRepo.WithContext(ctx).Delete(anyIDs...)
}

// BatchChangeUserStatus 批量启用/停用用户
func (i *UserService) BatchChangeUserStatus(ctx context.Context, req *BatchChangeUserStatusReq) error {
	if err := i.validate.Struct(req); err != nil {
		return err
	}
	ids, err := normalizeBatchUserIDs(req.IDs)
	if err != nil {
		return err
	}

	l := i.locker.New()
	if err := l.Lock("user:batch-status"); err != nil {
		return err
	}
	defer l.Unlock()

	return i.userRepo.WithContext(ctx).UpdateStatusByIDs(ids, req.Status)
}

// BatchAssignUserRoles 批量覆盖分配角色
func (i *UserService) BatchAssignUserRoles(ctx context.Context, req *BatchAssignUserRolesReq) error {
	if err := i.validate.Struct(req); err != nil {
		return err
	}
	ids, err := normalizeBatchUserIDs(req.IDs)
	if err != nil {
		return err
	}

	l := i.locker.New()
	if err := l.Lock("user:batch-roles"); err != nil {
		return err
	}
	defer l.Unlock()

	roles := db.GetModelsByIDs[model.Role](req.RoleIDs)
	roleIDs := make([]string, 0, len(roles))
	for _, role := range roles {
		roleIDs = append(roleIDs, cast.ToString(role.ID))
	}

	users := make([]*model.User, 0, len(ids))
	for _, id := range ids {
		user, err := i.userRepo.WithContext(ctx).FindById(id)
		if err != nil {
			return err
		}
		user.SetRoles(roles)
		users = append(users, user)
	}
	if err := i.userRepo.WithContext(ctx).SaveBatch(users); err != nil {
		return err
	}
	for _, user := range users {
		i.eventBus.Publish(event.AssignRoleToUser, &AssignRoleToUser{
			UserID:  user.ID,
			RoleIDs: roleIDs,
		})
	}
	return nil
}

// BatchAssignUserDept 批量设置归属部门
func (i *UserService) BatchAssignUserDept(ctx context.Context, req *BatchAssignUserDeptReq) error {
	if err := i.validate.Struct(req); err != nil {
		return err
	}
	ids, err := normalizeBatchUserIDs(req.IDs)
	if err != nil {
		return err
	}

	l := i.locker.New()
	if err := l.Lock("user:batch-dept"); err != nil {
		return err
	}
	defer l.Unlock()

	var depts []model.Dept
	if req.DeptID != "" {
		depts = db.GetModelsByIDs[model.Dept]([]string{req.DeptID})
		if len(depts) == 0 || depts[0].ID == 0 {
			return DeptNotFound()
		}
		// 校验部门真实存在
		if _, err := i.deptRepo.WithContext(ctx).Detail(depts[0].ID); err != nil {
			return DeptNotFound()
		}
	}

	users := make([]*model.User, 0, len(ids))
	for _, id := range ids {
		user, err := i.userRepo.WithContext(ctx).FindById(id)
		if err != nil {
			return err
		}
		user.Depts = depts
		users = append(users, user)
	}
	return i.userRepo.WithContext(ctx).SaveBatch(users)
}
