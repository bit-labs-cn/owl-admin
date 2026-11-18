package service

import (
	"context"

	"bit-labs.cn/owl-admin/app-notification/contract"
	"bit-labs.cn/owl-admin/app-notification/model"
	"bit-labs.cn/owl-admin/app-notification/repository"
	"gorm.io/gorm"
)

// 以下取值与 owl-admin 的用户表保持一致。通知中心按表名读取身份数据，
// 不引用 owl-admin 的 Go 包，避免两个子系统在代码层互相依赖。
const (
	adminUserStatusEnabled = 1
)

// Recipient 是解析后的一个接收人。
type Recipient struct {
	Type  int
	ID    uint
	Name  string
	Email string
	Phone string
}

// AudienceEstimate 是发送前的人数预估。
type AudienceEstimate struct {
	Users        int `json:"users"`
	Contacts     int `json:"contacts"`
	MissingEmail int `json:"missingEmail"`
	MissingPhone int `json:"missingPhone"`
}

// AudienceResolver 把用户、分组、部门、角色和外部名单解析成具体接收人。
type AudienceResolver struct {
	db       *gorm.DB
	contacts repository.ContactRepositoryInterface
}

func NewAudienceResolver(db *gorm.DB, contacts repository.ContactRepositoryInterface) *AudienceResolver {
	return &AudienceResolver{db: db, contacts: contacts}
}

func (r *AudienceResolver) Resolve(ctx context.Context, audience contract.Audience) ([]Recipient, AudienceEstimate, error) {
	users, err := r.users(ctx, audience)
	if err != nil {
		return nil, AudienceEstimate{}, err
	}
	contacts, err := r.external(ctx, audience)
	if err != nil {
		return nil, AudienceEstimate{}, err
	}
	out := make([]Recipient, 0, len(users)+len(contacts))
	est := AudienceEstimate{Users: len(users), Contacts: len(contacts)}
	for _, u := range users {
		if u.Email == "" {
			est.MissingEmail++
		}
		if u.Phone == "" {
			est.MissingPhone++
		}
		out = append(out, u)
	}
	for _, c := range contacts {
		if c.Email == "" {
			est.MissingEmail++
		}
		if c.Phone == "" {
			est.MissingPhone++
		}
		out = append(out, c)
	}
	return out, est, nil
}

func (r *AudienceResolver) users(ctx context.Context, audience contract.Audience) ([]Recipient, error) {
	ids := map[uint]struct{}{}
	if audience.All {
		var all []uint
		err := r.db.WithContext(ctx).Table("admin_user").
			Where("status = ?", adminUserStatusEnabled).
			Where("deleted_at IS NULL").
			Pluck("id", &all).Error
		if err != nil {
			return nil, err
		}
		for _, id := range all {
			ids[id] = struct{}{}
		}
	}
	for _, id := range audience.UserIDs {
		if id > 0 {
			ids[id] = struct{}{}
		}
	}
	if len(audience.GroupIDs) > 0 {
		var groupUsers []uint
		err := r.db.WithContext(ctx).Table("admin_user_group_user").Where("user_group_id IN ?", audience.GroupIDs).Pluck("user_id", &groupUsers).Error
		if err != nil {
			return nil, err
		}
		for _, id := range groupUsers {
			ids[id] = struct{}{}
		}
	}
	if len(audience.DeptIDs) > 0 {
		deptIDs, err := r.expandDepts(ctx, audience.DeptIDs)
		if err != nil {
			return nil, err
		}
		var deptUsers []uint
		if len(deptIDs) > 0 {
			err = r.db.WithContext(ctx).Table("admin_user_dept").Where("dept_id IN ?", deptIDs).Pluck("user_id", &deptUsers).Error
			if err != nil {
				return nil, err
			}
		}
		for _, id := range deptUsers {
			ids[id] = struct{}{}
		}
	}
	if len(audience.RoleIDs) > 0 {
		var roleUsers []uint
		err := r.db.WithContext(ctx).Table("admin_user_role").Where("role_id IN ?", audience.RoleIDs).Pluck("user_id", &roleUsers).Error
		if err != nil {
			return nil, err
		}
		for _, id := range roleUsers {
			ids[id] = struct{}{}
		}
	}
	if len(ids) == 0 {
		return nil, nil
	}
	idList := make([]uint, 0, len(ids))
	for id := range ids {
		idList = append(idList, id)
	}
	var users []struct {
		ID       uint
		Username string
		Nickname string
		Email    string
		Phone    string
	}
	err := r.db.WithContext(ctx).Table("admin_user").
		Select("id, username, nickname, email, phone").
		Where("id IN ? AND status = ? AND deleted_at IS NULL", idList, adminUserStatusEnabled).
		Scan(&users).Error
	if err != nil {
		return nil, err
	}
	out := make([]Recipient, 0, len(users))
	for _, u := range users {
		name := u.Nickname
		if name == "" {
			name = u.Username
		}
		out = append(out, Recipient{Type: model.RecipientTypeUser, ID: u.ID, Name: name, Email: u.Email, Phone: u.Phone})
	}
	return out, nil
}

func (r *AudienceResolver) expandDepts(ctx context.Context, roots []uint) ([]uint, error) {
	var all []struct {
		ID       uint
		ParentID int
	}
	if err := r.db.WithContext(ctx).Table("admin_dept").
		Select("id, parent_id").
		Where("deleted_at IS NULL").
		Scan(&all).Error; err != nil {
		return nil, err
	}
	children := map[int][]uint{}
	for _, d := range all {
		children[d.ParentID] = append(children[d.ParentID], d.ID)
	}
	seen := map[uint]struct{}{}
	var walk func(uint)
	walk = func(id uint) {
		if _, ok := seen[id]; ok {
			return
		}
		seen[id] = struct{}{}
		for _, child := range children[int(id)] {
			walk(child)
		}
	}
	for _, id := range roots {
		walk(id)
	}
	out := make([]uint, 0, len(seen))
	for id := range seen {
		out = append(out, id)
	}
	return out, nil
}

func (r *AudienceResolver) external(ctx context.Context, audience contract.Audience) ([]Recipient, error) {
	repo := r.contacts.WithContext(ctx)
	rows, err := repo.ListByIDs(audience.ContactIDs)
	if err != nil {
		return nil, err
	}
	more, err := repo.ListByLists(audience.ListIDs)
	if err != nil {
		return nil, err
	}
	rows = append(rows, more...)
	seen := map[uint]struct{}{}
	out := make([]Recipient, 0, len(rows))
	for _, row := range rows {
		if _, ok := seen[row.ID]; ok {
			continue
		}
		seen[row.ID] = struct{}{}
		if row.Status != model.ContactStatusSubscribed {
			continue
		}
		out = append(out, Recipient{Type: model.RecipientTypeContact, ID: row.ID, Name: row.Name, Email: row.Email, Phone: row.Phone})
	}
	return out, nil
}
