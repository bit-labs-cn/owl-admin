package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"bit-labs.cn/owl-admin/app-notification/model"
	"gorm.io/gorm"
)

// TypeQuery 是消息类型列表的筛选条件。
type TypeQuery struct {
	Keyword  string
	Status   int
	Page     int
	PageSize int
}

// TemplateQuery 是模板列表的筛选条件。
type TemplateQuery struct {
	TypeCode string
	Page     int
	PageSize int
}

// DeliveryQuery 是发送记录列表的筛选条件。MessageID 非空时只查该消息的明细。
type DeliveryQuery struct {
	Status        int
	Channel       string
	RecipientType int
	MessageID     uint
	Page          int
	PageSize      int
}

// DeliveryDTO 是管理端看到的一条发送记录。
type DeliveryDTO struct {
	ID             uint       `json:"id,string"`
	MessageID      uint       `json:"messageId,string"`
	Title          string     `json:"title"`
	UserID         uint       `json:"userId,string"`
	Channel        string     `json:"channel"`
	Status         int        `json:"status"`
	Attempts       int        `json:"attempts"`
	LastError      string     `json:"lastError"`
	RecipientType  int        `json:"recipientType"`
	RecipientName  string     `json:"recipientName"`
	RecipientEmail string     `json:"recipientEmail"`
	SentAt         *time.Time `json:"sentAt,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
}

// UserContact 是发送和展示要用的用户联系信息。
type UserContact struct {
	ID       uint   `json:"id,string"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Nickname string `json:"nickname"`
	Username string `json:"username"`
	Status   int    `json:"status"`
}

// AudienceSelection 是按角色、部门、用户组筛选接收人的条件。
type AudienceSelection struct {
	UserIDs      []uint
	RoleIDs      []uint
	DeptIDs      []uint
	GroupIDs     []uint
	IncludeChild bool
}

// AudienceOption 是角色、部门或用户组的候选项。
type AudienceOption struct {
	ID       uint             `json:"id,string"`
	Name     string           `json:"name"`
	Disabled bool             `json:"disabled"`
	Children []AudienceOption `json:"children,omitempty"`
}

// AudienceOptions 是发送页的受众选项。
type AudienceOptions struct {
	Roles  []AudienceOption `json:"roles"`
	Depts  []AudienceOption `json:"depts"`
	Groups []AudienceOption `json:"groups"`
}

// CatalogRepositoryInterface 管理类型、模板、发送记录和用户联系方式。
type CatalogRepositoryInterface interface {
	FindTypeByCode(ctx context.Context, code string) (*model.NotifyType, error)
	ListTypes(ctx context.Context, q TypeQuery) (int64, []model.NotifyType, error)
	ListEnabledTypes(ctx context.Context) ([]model.NotifyType, error)
	SaveType(ctx context.Context, row *model.NotifyType) error
	DeleteType(ctx context.Context, id uint) error
	CountTemplates(ctx context.Context, typeCode string) (int64, error)

	FindTemplate(ctx context.Context, typeCode string) (*model.NotifyTemplate, error)
	ListTemplates(ctx context.Context, q TemplateQuery) (int64, []model.NotifyTemplate, error)
	SaveTemplate(ctx context.Context, row *model.NotifyTemplate) error
	DeleteTemplate(ctx context.Context, id uint) error

	CreateDeliveries(ctx context.Context, tx *gorm.DB, rows []model.NotifyDelivery) error
	FindDelivery(ctx context.Context, id uint) (*model.NotifyDelivery, error)
	SaveDelivery(ctx context.Context, row *model.NotifyDelivery) error
	ListDeliveries(ctx context.Context, q DeliveryQuery) (int64, []DeliveryDTO, error)
	ListDeliveryBatches(ctx context.Context, q DeliveryQuery) (int64, []DeliveryBatchDTO, error)
	ListDeliveriesByMessage(ctx context.Context, messageID uint) ([]DeliveryDTO, error)

	ListContacts(ctx context.Context, userIDs []uint) ([]UserContact, error)
	SearchContacts(ctx context.Context, keyword string, userIDs []uint, limit int) ([]UserContact, error)
	AudienceOptions(ctx context.Context) (*AudienceOptions, error)
	ResolveAudience(ctx context.Context, sel AudienceSelection) ([]uint, error)

	ListMessages(ctx context.Context, q MessageQuery) (int64, []MessageSummaryDTO, error)
	FindMessageDetail(ctx context.Context, id uint) (*MessageDetailDTO, error)

	ResolveAnnouncementAudience(ctx context.Context, scope string, values []uint) ([]uint, error)
	ListAnnouncementStates(ctx context.Context, announcementID uint) ([]model.NotifyAnnouncementState, error)

	CreateSchedule(ctx context.Context, row *model.NotifySchedule) error
	FindSchedule(ctx context.Context, id uint) (*model.NotifySchedule, error)
	SaveSchedule(ctx context.Context, row *model.NotifySchedule) error
	ListSchedules(ctx context.Context, q ScheduleQuery) (int64, []ScheduleDTO, error)
	ClaimDueSchedules(ctx context.Context, now time.Time, limit int) ([]model.NotifySchedule, error)
	ClaimSchedule(ctx context.Context, id uint, planAt *time.Time) (bool, error)
}

// CatalogRepository 是类型、模板和发送记录的存储。
type CatalogRepository struct {
	db *gorm.DB
}

func NewCatalogRepository(db *gorm.DB) CatalogRepositoryInterface {
	return &CatalogRepository{db: db}
}

func (r *CatalogRepository) FindTypeByCode(ctx context.Context, code string) (*model.NotifyType, error) {
	var row model.NotifyType
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *CatalogRepository) ListTypes(ctx context.Context, q TypeQuery) (int64, []model.NotifyType, error) {
	q.Page, q.PageSize = normalizePage(q.Page, q.PageSize)
	base := r.db.WithContext(ctx).Model(&model.NotifyType{})
	if q.Keyword != "" {
		like := "%" + q.Keyword + "%"
		base = base.Where("code LIKE ? OR name LIKE ?", like, like)
	}
	if q.Status > 0 {
		base = base.Where("status = ?", q.Status)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return 0, nil, err
	}
	var list []model.NotifyType
	err := base.Order("id desc").Offset((q.Page - 1) * q.PageSize).Limit(q.PageSize).Find(&list).Error
	return total, list, err
}

func (r *CatalogRepository) ListEnabledTypes(ctx context.Context) ([]model.NotifyType, error) {
	var list []model.NotifyType
	err := r.db.WithContext(ctx).Where("status = ?", model.TypeStatusEnabled).Order("id asc").Find(&list).Error
	return list, err
}

func (r *CatalogRepository) SaveType(ctx context.Context, row *model.NotifyType) error {
	if row.ID == 0 {
		return r.db.WithContext(ctx).Create(row).Error
	}
	return r.db.WithContext(ctx).Model(&model.NotifyType{}).Where("id = ?", row.ID).Updates(map[string]any{
		"name":             row.Name,
		"description":      row.Description,
		"kind":             row.Kind,
		"default_channels": row.DefaultChannels,
		"status":           row.Status,
	}).Error
}

func (r *CatalogRepository) DeleteType(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&model.NotifyType{}, id).Error
}

func (r *CatalogRepository) CountTemplates(ctx context.Context, typeCode string) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.NotifyTemplate{}).Where("type_code = ?", typeCode).Count(&n).Error
	return n, err
}

func (r *CatalogRepository) FindTemplate(ctx context.Context, typeCode string) (*model.NotifyTemplate, error) {
	var row model.NotifyTemplate
	err := r.db.WithContext(ctx).Where("type_code = ?", typeCode).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *CatalogRepository) ListTemplates(ctx context.Context, q TemplateQuery) (int64, []model.NotifyTemplate, error) {
	q.Page, q.PageSize = normalizePage(q.Page, q.PageSize)
	base := r.db.WithContext(ctx).Model(&model.NotifyTemplate{})
	if q.TypeCode != "" {
		base = base.Where("type_code = ?", q.TypeCode)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return 0, nil, err
	}
	var list []model.NotifyTemplate
	err := base.Order("id desc").Offset((q.Page - 1) * q.PageSize).Limit(q.PageSize).Find(&list).Error
	return total, list, err
}

func (r *CatalogRepository) SaveTemplate(ctx context.Context, row *model.NotifyTemplate) error {
	if row.ID == 0 {
		return r.db.WithContext(ctx).Create(row).Error
	}
	return r.db.WithContext(ctx).Model(&model.NotifyTemplate{}).Where("id = ?", row.ID).Updates(map[string]any{
		"title": row.Title,
		"body":  row.Body,
		"link":  row.Link,
	}).Error
}

func (r *CatalogRepository) DeleteTemplate(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&model.NotifyTemplate{}, id).Error
}

func (r *CatalogRepository) CreateDeliveries(ctx context.Context, tx *gorm.DB, rows []model.NotifyDelivery) error {
	if len(rows) == 0 {
		return nil
	}
	runner := r.db.WithContext(ctx)
	if tx != nil {
		runner = tx.WithContext(ctx)
	}
	return runner.Create(&rows).Error
}

func (r *CatalogRepository) FindDelivery(ctx context.Context, id uint) (*model.NotifyDelivery, error) {
	var row model.NotifyDelivery
	err := r.db.WithContext(ctx).First(&row, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *CatalogRepository) SaveDelivery(ctx context.Context, row *model.NotifyDelivery) error {
	return r.db.WithContext(ctx).Model(&model.NotifyDelivery{}).Where("id = ?", row.ID).Updates(map[string]any{
		"status":     row.Status,
		"attempts":   row.Attempts,
		"last_error": row.LastError,
		"sent_at":    row.SentAt,
	}).Error
}

func (r *CatalogRepository) ListDeliveries(ctx context.Context, q DeliveryQuery) (int64, []DeliveryDTO, error) {
	q.Page, q.PageSize = normalizePage(q.Page, q.PageSize)
	base := r.db.WithContext(ctx).Table("notification_delivery d").
		Joins("INNER JOIN notification_message m ON m.id = d.message_id AND m.deleted_at IS NULL").
		Where("d.deleted_at IS NULL")
	if q.Status > 0 {
		base = base.Where("d.status = ?", q.Status)
	}
	if q.Channel != "" {
		base = base.Where("d.channel = ?", q.Channel)
	}
	if q.RecipientType > 0 {
		base = base.Where("d.recipient_type = ?", q.RecipientType)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return 0, nil, err
	}
	var list []DeliveryDTO
	err := base.Select("d.id, d.message_id, COALESCE(NULLIF(d.title, ''), m.title) as title, d.user_id, d.channel, d.status, d.attempts, d.last_error, d.recipient_type, d.recipient_name, d.recipient_email, d.sent_at, d.created_at").
		Order("d.id desc").
		Offset((q.Page - 1) * q.PageSize).
		Limit(q.PageSize).
		Scan(&list).Error
	return total, list, err
}

func (r *CatalogRepository) ListContacts(ctx context.Context, userIDs []uint) ([]UserContact, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}
	var list []UserContact
	err := r.db.WithContext(ctx).Table("admin_user").
		Select("id, email, phone, nickname, username, status").
		Where("id IN ? AND deleted_at IS NULL", userIDs).
		Scan(&list).Error
	return list, err
}

// SearchContacts 按关键字搜索用户，或按 ID 列表回显已选用户。传 ID 时忽略关键字。
func (r *CatalogRepository) SearchContacts(ctx context.Context, keyword string, userIDs []uint, limit int) ([]UserContact, error) {
	if limit <= 0 || limit > 200 {
		limit = 20
	}
	base := r.db.WithContext(ctx).Table("admin_user").
		Select("id, email, phone, nickname, username, status").
		Where("deleted_at IS NULL")
	if len(userIDs) > 0 {
		base = base.Where("id IN ?", userIDs)
	} else {
		keyword = strings.TrimSpace(keyword)
		if keyword != "" {
			like := "%" + keyword + "%"
			base = base.Where("username LIKE ? OR nickname LIKE ? OR email LIKE ? OR phone LIKE ?", like, like, like, like)
		}
	}
	var list []UserContact
	err := base.Order("id desc").Limit(limit).Scan(&list).Error
	return list, err
}

func normalizePage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}
