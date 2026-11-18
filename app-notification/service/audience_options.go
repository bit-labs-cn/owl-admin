package service

import (
	"context"
	"encoding/json"
	"strings"

	"bit-labs.cn/owl-admin/app-notification/contract"
	"bit-labs.cn/owl-admin/app-notification/model"
	"bit-labs.cn/owl/provider/router"
	"github.com/spf13/cast"
)

// AudienceOption 是受众选择器里的一个人、角色、分组或名单。
// ID 用字符串返回，避免雪花 ID 在浏览器里丢精度。
type AudienceOption struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Detail string `json:"detail,omitempty"`
}

// DeptOption 是部门树节点。
type DeptOption struct {
	ID       string        `json:"id"`
	Name     string        `json:"name"`
	Children []*DeptOption `json:"children,omitempty"`
}

// AudienceOptions 是一次性返回的受众选择项。
type AudienceOptions struct {
	Users  []AudienceOption `json:"users"`
	Roles  []AudienceOption `json:"roles"`
	Groups []AudienceOption `json:"groups"`
	Depts  []*DeptOption    `json:"depts"`
	Lists  []AudienceOption `json:"lists"`
}

// LinkOption 是通知可以跳转到的站内页面。
type LinkOption struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// SmsTemplateOption 是短信子系统里可用的模板。
type SmsTemplateOption struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	Content   string `json:"content"`
	Variables string `json:"variables"`
}

// Options 返回受众选择器的基础数据。用户只返回前 200 个，更多用 SearchUsers 搜索。
func (r *AudienceResolver) Options(ctx context.Context) (AudienceOptions, error) {
	out := AudienceOptions{
		Users:  []AudienceOption{},
		Roles:  []AudienceOption{},
		Groups: []AudienceOption{},
		Depts:  []*DeptOption{},
		Lists:  []AudienceOption{},
	}
	users, err := r.SearchUsers(ctx, "", 200)
	if err != nil {
		return out, err
	}
	out.Users = users

	var roles []struct {
		ID   uint
		Name string
		Code string
	}
	if err := r.db.WithContext(ctx).Table("admin_role").Select("id, name, code").Order("id asc").Scan(&roles).Error; err != nil {
		return out, err
	}
	for _, role := range roles {
		out.Roles = append(out.Roles, AudienceOption{ID: cast.ToString(role.ID), Name: role.Name, Detail: role.Code})
	}

	var groups []struct {
		ID   uint
		Name string
		Code string
	}
	if err := r.db.WithContext(ctx).Table("admin_user_group").Select("id, name, code").Where("status = ?", 1).Order("sort asc, id asc").Scan(&groups).Error; err != nil {
		return out, err
	}
	for _, group := range groups {
		out.Groups = append(out.Groups, AudienceOption{ID: cast.ToString(group.ID), Name: group.Name, Detail: group.Code})
	}

	depts, err := r.deptOptions(ctx)
	if err != nil {
		return out, err
	}
	out.Depts = depts

	var lists []struct {
		ID   uint
		Name string
	}
	if err := r.db.WithContext(ctx).Table("notify_contact_list").Select("id, name").Order("id desc").Scan(&lists).Error; err != nil {
		return out, err
	}
	for _, list := range lists {
		out.Lists = append(out.Lists, AudienceOption{ID: cast.ToString(list.ID), Name: list.Name})
	}
	return out, nil
}

// SearchUsers 按用户名、昵称、邮箱或手机号搜索启用用户。
func (r *AudienceResolver) SearchUsers(ctx context.Context, keyword string, limit int) ([]AudienceOption, error) {
	if limit < 1 || limit > 500 {
		limit = 20
	}
	tx := r.db.WithContext(ctx).Table("admin_user").
		Select("id, username, nickname, email, phone").
		Where("status = ? AND deleted_at IS NULL", adminUserStatusEnabled)
	if word := strings.TrimSpace(keyword); word != "" {
		like := "%" + word + "%"
		tx = tx.Where("username LIKE ? OR nickname LIKE ? OR email LIKE ? OR phone LIKE ?", like, like, like, like)
	}
	var rows []userOptionRow
	if err := tx.Order("id desc").Limit(limit).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return mapUserOptions(rows), nil
}

// UsersByIDs 按 ID 补全用户，供编辑旧通知时回显已选用户。
func (r *AudienceResolver) UsersByIDs(ctx context.Context, ids []uint) ([]AudienceOption, error) {
	if len(ids) == 0 {
		return []AudienceOption{}, nil
	}
	var rows []userOptionRow
	if err := r.db.WithContext(ctx).Table("admin_user").
		Select("id, username, nickname, email, phone").
		Where("id IN ? AND status = ? AND deleted_at IS NULL", ids, adminUserStatusEnabled).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	return mapUserOptions(rows), nil
}

type userOptionRow struct {
	ID       uint
	Username string
	Nickname string
	Email    string
	Phone    string
}

func mapUserOptions(rows []userOptionRow) []AudienceOption {
	out := make([]AudienceOption, 0, len(rows))
	for _, row := range rows {
		name := strings.TrimSpace(row.Nickname)
		if name == "" {
			name = row.Username
		}
		detailParts := make([]string, 0, 2)
		if row.Username != "" && row.Username != name {
			detailParts = append(detailParts, row.Username)
		}
		if row.Phone != "" {
			detailParts = append(detailParts, row.Phone)
		} else if row.Email != "" {
			detailParts = append(detailParts, row.Email)
		}
		out = append(out, AudienceOption{ID: cast.ToString(row.ID), Name: name, Detail: strings.Join(detailParts, " · ")})
	}
	return out
}

// SmsTemplates 读取短信子系统里已启用的模板。短信子系统未安装时返回空列表。
func (r *AudienceResolver) SmsTemplates(ctx context.Context) ([]SmsTemplateOption, error) {
	out := []SmsTemplateOption{}
	if !r.db.Migrator().HasTable("sms_template") {
		return out, nil
	}
	var rows []struct {
		Code      string
		Name      string
		Content   string
		Variables string
	}
	if err := r.db.WithContext(ctx).Table("sms_template").
		Select("code, name, content, variables").
		Where("status = ?", "active").
		Order("id desc").Scan(&rows).Error; err != nil {
		return out, err
	}
	for _, row := range rows {
		out = append(out, SmsTemplateOption{Code: row.Code, Name: row.Name, Content: row.Content, Variables: row.Variables})
	}
	return out, nil
}

// Links 返回可以跳转的站内页面，供通知的「点开后去哪」使用。
func (r *AudienceResolver) Links(ctx context.Context) ([]LinkOption, error) {
	var rows []struct {
		Path string
		Meta string
	}
	if err := r.db.WithContext(ctx).Table("admin_menu").
		Select("path, meta").
		Where("menu_type = ? AND path <> ''", router.MenuTypeMenu).
		Order("path asc").Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]LinkOption, 0, len(rows))
	seen := map[string]struct{}{}
	for _, row := range rows {
		path := strings.TrimSpace(row.Path)
		if path == "" || !strings.HasPrefix(path, "/") || strings.Contains(path, ":") {
			continue
		}
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		var meta router.Meta
		_ = json.Unmarshal([]byte(row.Meta), &meta)
		name := strings.TrimSpace(meta.Title)
		if name == "" {
			name = path
		}
		out = append(out, LinkOption{Name: name, Path: path})
	}
	return out, nil
}

// Match 判断某个主体是否落在受众范围内。空受众视为全体用户。
func (r *AudienceResolver) Match(ctx context.Context, audience contract.Audience, principalType int, principalID uint) (bool, error) {
	if principalType != model.PrincipalTypeAdmin || principalID == 0 {
		return false, nil
	}
	empty := len(audience.UserIDs) == 0 && len(audience.GroupIDs) == 0 && len(audience.DeptIDs) == 0 &&
		len(audience.RoleIDs) == 0 && len(audience.ContactIDs) == 0 && len(audience.ListIDs) == 0
	if audience.All || empty {
		return true, nil
	}
	for _, id := range audience.UserIDs {
		if id == principalID {
			return true, nil
		}
	}
	if len(audience.GroupIDs) > 0 {
		var count int64
		if err := r.db.WithContext(ctx).Table("admin_user_group_user").
			Where("user_group_id IN ? AND user_id = ?", audience.GroupIDs, principalID).
			Count(&count).Error; err != nil {
			return false, err
		}
		if count > 0 {
			return true, nil
		}
	}
	if len(audience.RoleIDs) > 0 {
		var count int64
		if err := r.db.WithContext(ctx).Table("admin_user_role").
			Where("role_id IN ? AND user_id = ?", audience.RoleIDs, principalID).
			Count(&count).Error; err != nil {
			return false, err
		}
		if count > 0 {
			return true, nil
		}
	}
	if len(audience.DeptIDs) > 0 {
		deptIDs, err := r.expandDepts(ctx, audience.DeptIDs)
		if err != nil {
			return false, err
		}
		if len(deptIDs) > 0 {
			var count int64
			if err := r.db.WithContext(ctx).Table("admin_user_dept").
				Where("dept_id IN ? AND user_id = ?", deptIDs, principalID).
				Count(&count).Error; err != nil {
				return false, err
			}
			if count > 0 {
				return true, nil
			}
		}
	}
	return false, nil
}

func (r *AudienceResolver) deptOptions(ctx context.Context) ([]*DeptOption, error) {
	var rows []struct {
		ID       uint
		ParentID int
		Name     string
		Sort     uint
	}
	if err := r.db.WithContext(ctx).Table("admin_dept").
		Select("id, parent_id, name, sort").
		Where("deleted_at IS NULL").
		Order("sort asc, id asc").Scan(&rows).Error; err != nil {
		return nil, err
	}
	nodes := make(map[uint]*DeptOption, len(rows))
	for _, row := range rows {
		nodes[row.ID] = &DeptOption{ID: cast.ToString(row.ID), Name: row.Name}
	}
	roots := make([]*DeptOption, 0)
	for _, row := range rows {
		node := nodes[row.ID]
		parent := nodes[uint(row.ParentID)]
		if parent != nil && uint(row.ParentID) != row.ID {
			parent.Children = append(parent.Children, node)
			continue
		}
		roots = append(roots, node)
	}
	return roots, nil
}
