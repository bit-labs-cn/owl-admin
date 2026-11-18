package repository

import (
	"context"
	"time"

	"bit-labs.cn/owl-admin/app-notification/model"
)

// AnnouncementReadRecipient 是一个公告接收人的已读状态。
type AnnouncementReadRecipient struct {
	UserID      uint       `json:"userId,string"`
	UserName    string     `json:"userName"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	ReadStatus  int        `json:"readStatus"`
	ReadAt      *time.Time `json:"readAt,omitempty"`
	IsDismissed int        `json:"isDismissed"`
}

// AnnouncementReadView 是一条公告的已读汇总与接收人明细。
type AnnouncementReadView struct {
	AnnouncementID uint                        `json:"announcementId,string"`
	Title          string                      `json:"title"`
	TargetScope    string                      `json:"targetScope"`
	RecipientCount int                         `json:"recipientCount"`
	ReadCount      int                         `json:"readCount"`
	UnreadCount    int                         `json:"unreadCount"`
	DismissedCount int                         `json:"dismissedCount"`
	Recipients     []AnnouncementReadRecipient `json:"recipients"`
}

// ResolveAnnouncementAudience 把公告的受众范围解析为具体用户 ID。
// all 取全部启用用户；dept 取部门及其下级部门；role 取角色成员。
func (r *CatalogRepository) ResolveAnnouncementAudience(ctx context.Context, scope string, values []uint) ([]uint, error) {
	switch scope {
	case model.ScopeDept:
		if len(values) == 0 {
			return nil, nil
		}
		deptIDs, err := r.expandDepts(ctx, values)
		if err != nil {
			return nil, err
		}
		return r.userIDsByDept(ctx, deptIDs)
	case model.ScopeRole:
		if len(values) == 0 {
			return nil, nil
		}
		return r.userIDsByRole(ctx, values)
	default:
		var ids []uint
		err := r.db.WithContext(ctx).Table("admin_user").
			Where("deleted_at IS NULL AND status = ?", statusEnabled).
			Pluck("id", &ids).Error
		return ids, err
	}
}

// ListAnnouncementStates 返回一条公告下所有用户的状态记录。
func (r *CatalogRepository) ListAnnouncementStates(ctx context.Context, announcementID uint) ([]model.NotifyAnnouncementState, error) {
	var list []model.NotifyAnnouncementState
	err := r.db.WithContext(ctx).
		Where("announcement_id = ?", announcementID).
		Find(&list).Error
	return list, err
}
