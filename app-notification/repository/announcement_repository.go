package repository

import (
	"context"
	"time"

	"bit-labs.cn/owl-admin/app-notification/model"
	"bit-labs.cn/owl/contract"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UserAnnouncementDTO 包含公告本体及用户个人关闭/已读状态
type UserAnnouncementDTO struct {
	ID          uint       `json:"id,string"`
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	Level       int        `json:"level"`
	DisplayMode int        `json:"displayMode"`
	StartAt     *time.Time `json:"startAt,omitempty"`
	EndAt       *time.Time `json:"endAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	IsDismissed int        `json:"isDismissed"`
	ReadAt      *time.Time `json:"readAt,omitempty"`
}

type AnnouncementRepositoryInterface interface {
	Create(ctx context.Context, ann *model.NotifyAnnouncement) error
	Update(ctx context.Context, ann *model.NotifyAnnouncement) error
	Delete(ctx context.Context, id uint) error
	Detail(ctx context.Context, id uint) (*model.NotifyAnnouncement, error)
	RetrieveAdmin(ctx context.Context, page, pageSize int, keyword string, status int) (int64, []model.NotifyAnnouncement, error)
	RetrieveUserAnnouncements(ctx context.Context, userID uint, page, pageSize int) (int64, []UserAnnouncementDTO, error)
	Dismiss(ctx context.Context, announcementID, userID uint) error
	contract.WithContext[AnnouncementRepositoryInterface]
}

type AnnouncementRepository struct {
	db *gorm.DB
}

func NewAnnouncementRepository(db *gorm.DB) AnnouncementRepositoryInterface {
	return &AnnouncementRepository{db: db}
}

func (r *AnnouncementRepository) WithContext(ctx context.Context) AnnouncementRepositoryInterface {
	return &AnnouncementRepository{db: r.db.WithContext(ctx)}
}

func (r *AnnouncementRepository) Create(ctx context.Context, ann *model.NotifyAnnouncement) error {
	return r.db.WithContext(ctx).Create(ann).Error
}

func (r *AnnouncementRepository) Update(ctx context.Context, ann *model.NotifyAnnouncement) error {
	return r.db.WithContext(ctx).Save(ann).Error
}

func (r *AnnouncementRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.NotifyAnnouncement{}, id).Error
}

func (r *AnnouncementRepository) Detail(ctx context.Context, id uint) (*model.NotifyAnnouncement, error) {
	var row model.NotifyAnnouncement
	err := r.db.WithContext(ctx).First(&row, id).Error
	return &row, err
}

func (r *AnnouncementRepository) RetrieveAdmin(ctx context.Context, page, pageSize int, keyword string, status int) (int64, []model.NotifyAnnouncement, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	tx := r.db.WithContext(ctx).Model(&model.NotifyAnnouncement{})
	if keyword != "" {
		tx = tx.Where("title LIKE ?", "%"+keyword+"%")
	}
	if status > 0 {
		tx = tx.Where("status = ?", status)
	}

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return 0, nil, err
	}
	var list []model.NotifyAnnouncement
	err := tx.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return total, list, err
}

// RetrieveUserAnnouncements 读扩散查询：只读取生效中且未被当前用户关闭的公告
func (r *AnnouncementRepository) RetrieveUserAnnouncements(ctx context.Context, userID uint, page, pageSize int) (int64, []UserAnnouncementDTO, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	now := time.Now()
	base := r.db.WithContext(ctx).Table("notification_announcement a").
		Joins("LEFT JOIN notification_announcement_state s ON a.id = s.announcement_id AND s.user_id = ?", userID).
		Where("a.status = ?", model.AnnouncementStatusPublished).
		Where("(a.start_at IS NULL OR a.start_at <= ?)", now).
		Where("(a.end_at IS NULL OR a.end_at >= ?)", now).
		Where("(s.is_dismissed IS NULL OR s.is_dismissed = 0)")

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return 0, nil, err
	}

	var list []UserAnnouncementDTO
	err := base.Select("a.id, a.title, a.content, a.level, a.display_mode, a.start_at, a.end_at, a.created_at, COALESCE(s.is_dismissed, 0) as is_dismissed, s.read_at").
		Order("a.level desc, a.id desc").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Scan(&list).Error

	return total, list, err
}

func (r *AnnouncementRepository) Dismiss(ctx context.Context, announcementID, userID uint) error {
	now := time.Now()
	state := model.NotifyAnnouncementState{
		AnnouncementID: announcementID,
		UserID:         userID,
		IsDismissed:    1,
		ReadAt:         &now,
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "announcement_id"}, {Name: "user_id"}},
		DoUpdates: clause.Assignments(map[string]any{"is_dismissed": 1, "read_at": &now}),
	}).Create(&state).Error
}
