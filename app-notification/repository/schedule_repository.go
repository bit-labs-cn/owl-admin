package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"bit-labs.cn/owl-admin/app-notification/contract"
	"bit-labs.cn/owl-admin/app-notification/model"
	"gorm.io/gorm"
)

// ScheduleAudience 是定时任务保存下来的受众选择条件。
// 到点发送时才解析成具体用户，避免计划期间的组织调整被冻结。
type ScheduleAudience struct {
	UserIDs      []uint                       `json:"userIds,omitempty"`
	RoleIDs      []uint                       `json:"roleIds,omitempty"`
	DeptIDs      []uint                       `json:"deptIds,omitempty"`
	GroupIDs     []uint                       `json:"groupIds,omitempty"`
	IncludeChild bool                         `json:"includeDeptChild,omitempty"`
	Externals    []contract.ExternalRecipient `json:"externals,omitempty"`
}

// Selection 还原成受众解析条件。
func (a ScheduleAudience) Selection() AudienceSelection {
	return AudienceSelection{
		UserIDs:      a.UserIDs,
		RoleIDs:      a.RoleIDs,
		DeptIDs:      a.DeptIDs,
		GroupIDs:     a.GroupIDs,
		IncludeChild: a.IncludeChild,
	}
}

// ScheduleQuery 是定时发送列表的筛选条件。
type ScheduleQuery struct {
	Keyword  string
	Kind     int
	Status   int
	Page     int
	PageSize int
}

// ScheduleDTO 是管理端看到的一条定时发送任务。
type ScheduleDTO struct {
	ID              uint       `json:"id,string"`
	Kind            int        `json:"kind"`
	TypeCode        string     `json:"typeCode"`
	TypeName        string     `json:"typeName"`
	Title           string     `json:"title"`
	Channels        []string   `json:"channels"`
	AudienceSummary string     `json:"audienceSummary"`
	RecipientCount  int        `json:"recipientCount"`
	PlanAt          time.Time  `json:"planAt"`
	Status          int        `json:"status"`
	Attempts        int        `json:"attempts"`
	LastError       string     `json:"lastError"`
	MessageID       uint       `json:"messageId,string"`
	SenderID        uint       `json:"senderId,string"`
	SenderName      string     `json:"senderName"`
	SentAt          *time.Time `json:"sentAt,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
}

// scheduleRow 用于把 JSON 渠道列按原文取出，再转成渠道数组。
type scheduleRow struct {
	ScheduleDTO
	ChannelsText string `gorm:"column:channels_text"`
}

const scheduleSelectFields = `s.id, s.kind, s.type_code, COALESCE(t.name, '') as type_name, s.title,
	s.plan_at, s.status, s.attempts, s.last_error, s.message_id, s.sender_id, s.sent_at, s.created_at,
	COALESCE(s.audience_summary, '') as audience_summary, s.recipient_count,
	COALESCE(s.channels, '') as channels_text,
	COALESCE(NULLIF(u.nickname, ''), u.username, '') as sender_name`

// CreateSchedule 保存一条定时发送任务。
func (r *CatalogRepository) CreateSchedule(ctx context.Context, row *model.NotifySchedule) error {
	return r.db.WithContext(ctx).Create(row).Error
}

// FindSchedule 按 ID 查一条定时发送任务。
func (r *CatalogRepository) FindSchedule(ctx context.Context, id uint) (*model.NotifySchedule, error) {
	var row model.NotifySchedule
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// SaveSchedule 回写任务状态、尝试次数、错误信息与发送结果。
func (r *CatalogRepository) SaveSchedule(ctx context.Context, row *model.NotifySchedule) error {
	return r.db.WithContext(ctx).Model(&model.NotifySchedule{}).Where("id = ?", row.ID).Updates(map[string]any{
		"status":     row.Status,
		"attempts":   row.Attempts,
		"last_error": row.LastError,
		"message_id": row.MessageID,
		"plan_at":    row.PlanAt,
		"sent_at":    row.SentAt,
	}).Error
}

// ListSchedules 分页列出定时发送任务。
func (r *CatalogRepository) ListSchedules(ctx context.Context, q ScheduleQuery) (int64, []ScheduleDTO, error) {
	q.Page, q.PageSize = normalizePage(q.Page, q.PageSize)
	base := func() *gorm.DB {
		tx := r.db.WithContext(ctx).Table("notification_schedule s").
			Joins("LEFT JOIN notification_type t ON t.code = s.type_code AND t.deleted_at IS NULL").
			Joins("LEFT JOIN admin_user u ON u.id = s.sender_id AND u.deleted_at IS NULL").
			Where("s.deleted_at IS NULL")
		if q.Keyword != "" {
			like := "%" + q.Keyword + "%"
			tx = tx.Where("s.title LIKE ? OR s.content LIKE ?", like, like)
		}
		if q.Kind > 0 {
			tx = tx.Where("s.kind = ?", q.Kind)
		}
		if q.Status > 0 {
			tx = tx.Where("s.status = ?", q.Status)
		}
		return tx
	}

	var total int64
	if err := base().Count(&total).Error; err != nil {
		return 0, nil, err
	}
	if total == 0 {
		return 0, nil, nil
	}

	var rows []scheduleRow
	err := base().
		Select(scheduleSelectFields).
		Order("s.plan_at desc, s.id desc").
		Offset((q.Page - 1) * q.PageSize).
		Limit(q.PageSize).
		Scan(&rows).Error
	if err != nil {
		return 0, nil, err
	}
	list := make([]ScheduleDTO, 0, len(rows))
	for _, row := range rows {
		item := row.ScheduleDTO
		item.Channels = decodeChannelList(row.ChannelsText)
		list = append(list, item)
	}
	return total, list, nil
}

// ClaimDueSchedules 领取到点且仍在待发送的任务。
// 用「带状态条件的更新」抢占，多实例同时轮询也只有一个实例能领到同一条任务。
func (r *CatalogRepository) ClaimDueSchedules(ctx context.Context, now time.Time, limit int) ([]model.NotifySchedule, error) {
	if limit <= 0 {
		limit = 20
	}
	var due []model.NotifySchedule
	if err := r.db.WithContext(ctx).
		Where("status = ? AND plan_at <= ?", model.ScheduleStatusPending, now).
		Order("plan_at asc").
		Limit(limit).
		Find(&due).Error; err != nil {
		return nil, err
	}
	claimed := make([]model.NotifySchedule, 0, len(due))
	for _, row := range due {
		res := r.db.WithContext(ctx).Model(&model.NotifySchedule{}).
			Where("id = ? AND status = ? AND plan_at <= ?", row.ID, model.ScheduleStatusPending, now).
			Updates(map[string]any{
				"status":   model.ScheduleStatusSending,
				"attempts": gorm.Expr("attempts + 1"),
			})
		if res.Error != nil {
			return nil, res.Error
		}
		if res.RowsAffected == 0 {
			continue
		}
		row.Status = model.ScheduleStatusSending
		row.Attempts++
		claimed = append(claimed, row)
	}
	return claimed, nil
}

// ClaimSchedule 抢占指定的一条任务，允许「待发送」和「失败」两种状态。
// planAt 非空时同时改写计划时间（「立即发送」用）。返回是否抢到。
func (r *CatalogRepository) ClaimSchedule(ctx context.Context, id uint, planAt *time.Time) (bool, error) {
	updates := map[string]any{
		"status":   model.ScheduleStatusSending,
		"attempts": gorm.Expr("attempts + 1"),
	}
	if planAt != nil {
		updates["plan_at"] = *planAt
	}
	res := r.db.WithContext(ctx).Model(&model.NotifySchedule{}).
		Where("id = ? AND status IN ?", id, []int{model.ScheduleStatusPending, model.ScheduleStatusFailed}).
		Updates(updates)
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

func decodeChannelList(raw string) []string {
	if raw == "" {
		return nil
	}
	var out []string
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil
	}
	return out
}
