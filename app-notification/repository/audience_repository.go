package repository

import (
	"context"

	"bit-labs.cn/owl-admin/app/model"
)

// statusEnabled 是角色、用户组、部门与用户共用的启用取值。
const statusEnabled = 1

// AudienceOptions 返回角色、部门树和用户组选项。
func (r *CatalogRepository) AudienceOptions(ctx context.Context) (*AudienceOptions, error) {
	roles, err := r.roleOptions(ctx)
	if err != nil {
		return nil, err
	}
	depts, err := r.deptTree(ctx)
	if err != nil {
		return nil, err
	}
	groups, err := r.groupOptions(ctx)
	if err != nil {
		return nil, err
	}
	return &AudienceOptions{Roles: roles, Depts: depts, Groups: groups}, nil
}

func (r *CatalogRepository) roleOptions(ctx context.Context) ([]AudienceOption, error) {
	var rows []model.Role
	if err := r.db.WithContext(ctx).Select("id, name, status").Order("id asc").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]AudienceOption, 0, len(rows))
	for _, row := range rows {
		out = append(out, AudienceOption{ID: row.ID, Name: row.Name, Disabled: row.Status != statusEnabled})
	}
	return out, nil
}

func (r *CatalogRepository) groupOptions(ctx context.Context) ([]AudienceOption, error) {
	var rows []model.UserGroup
	if err := r.db.WithContext(ctx).Select("id, name, status").Order("sort asc, id asc").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]AudienceOption, 0, len(rows))
	for _, row := range rows {
		out = append(out, AudienceOption{ID: row.ID, Name: row.Name, Disabled: row.Status != statusEnabled})
	}
	return out, nil
}

// deptTree 按 parent_id 组装部门树。父级不在本表里的部门当作根。
func (r *CatalogRepository) deptTree(ctx context.Context) ([]AudienceOption, error) {
	var rows []model.Dept
	if err := r.db.WithContext(ctx).Select("id, name, parent_id, status").Order("sort asc, id asc").Find(&rows).Error; err != nil {
		return nil, err
	}
	exists := make(map[uint]struct{}, len(rows))
	for _, row := range rows {
		exists[row.ID] = struct{}{}
	}
	children := map[uint][]model.Dept{}
	for _, row := range rows {
		parent := uint(row.ParentId)
		if _, ok := exists[parent]; !ok {
			parent = 0
		}
		children[parent] = append(children[parent], row)
	}
	var build func(parent uint) []AudienceOption
	build = func(parent uint) []AudienceOption {
		items := children[parent]
		out := make([]AudienceOption, 0, len(items))
		for _, item := range items {
			out = append(out, AudienceOption{
				ID:       item.ID,
				Name:     item.Name,
				Disabled: item.Status != statusEnabled,
				Children: build(item.ID),
			})
		}

		return out
	}
	return build(0), nil
}

// ResolveAudience 把用户、角色、部门（可含下级）和用户组合并成去重后的启用用户 ID。
func (r *CatalogRepository) ResolveAudience(ctx context.Context, sel AudienceSelection) ([]uint, error) {
	collected := make(map[uint]struct{}, len(sel.UserIDs))
	for _, id := range sel.UserIDs {
		if id > 0 {
			collected[id] = struct{}{}
		}
	}

	if len(sel.RoleIDs) > 0 {
		ids, err := r.userIDsByRole(ctx, sel.RoleIDs)
		if err != nil {
			return nil, err
		}
		mergeIDs(collected, ids)
	}

	if len(sel.GroupIDs) > 0 {
		ids, err := r.userIDsByGroup(ctx, sel.GroupIDs)
		if err != nil {
			return nil, err
		}
		mergeIDs(collected, ids)
	}

	if len(sel.DeptIDs) > 0 {
		deptIDs := sel.DeptIDs
		if sel.IncludeChild {
			expanded, err := r.expandDepts(ctx, sel.DeptIDs)
			if err != nil {
				return nil, err
			}
			deptIDs = expanded
		}
		ids, err := r.userIDsByDept(ctx, deptIDs)
		if err != nil {
			return nil, err
		}
		mergeIDs(collected, ids)
	}

	out := make([]uint, 0, len(collected))
	for id := range collected {
		out = append(out, id)
	}
	return out, nil
}

func (r *CatalogRepository) userIDsByRole(ctx context.Context, roleIDs []uint) ([]uint, error) {
	var ids []uint
	err := r.db.WithContext(ctx).Table("admin_user_role ur").
		Joins("JOIN admin_user u ON u.id = ur.user_id AND u.deleted_at IS NULL AND u.status = ?", statusEnabled).
		Where("ur.role_id IN ?", roleIDs).
		Distinct().Pluck("ur.user_id", &ids).Error
	return ids, err
}

func (r *CatalogRepository) userIDsByGroup(ctx context.Context, groupIDs []uint) ([]uint, error) {
	var ids []uint
	err := r.db.WithContext(ctx).Table("admin_user_group_user gu").
		Joins("JOIN admin_user u ON u.id = gu.user_id AND u.deleted_at IS NULL AND u.status = ?", statusEnabled).
		Where("gu.user_group_id IN ?", groupIDs).
		Distinct().Pluck("gu.user_id", &ids).Error
	return ids, err
}

func (r *CatalogRepository) userIDsByDept(ctx context.Context, deptIDs []uint) ([]uint, error) {
	var ids []uint
	err := r.db.WithContext(ctx).Table("admin_user_dept ud").
		Joins("JOIN admin_user u ON u.id = ud.user_id AND u.deleted_at IS NULL AND u.status = ?", statusEnabled).
		Where("ud.dept_id IN ?", deptIDs).
		Distinct().Pluck("ud.user_id", &ids).Error
	return ids, err
}

// expandDepts 返回给定部门及其全部下级部门 ID。
func (r *CatalogRepository) expandDepts(ctx context.Context, roots []uint) ([]uint, error) {
	var rows []model.Dept
	if err := r.db.WithContext(ctx).Select("id, parent_id").Find(&rows).Error; err != nil {
		return nil, err
	}
	children := map[uint][]uint{}
	for _, row := range rows {
		parent := uint(row.ParentId)
		children[parent] = append(children[parent], row.ID)
	}
	seen := map[uint]struct{}{}
	queue := make([]uint, 0, len(roots))
	for _, id := range roots {
		if id > 0 {
			queue = append(queue, id)
		}
	}
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		queue = append(queue, children[id]...)
	}
	out := make([]uint, 0, len(seen))
	for id := range seen {
		out = append(out, id)
	}
	return out, nil
}

func mergeIDs(dst map[uint]struct{}, ids []uint) {
	for _, id := range ids {
		if id > 0 {
			dst[id] = struct{}{}
		}
	}
}
