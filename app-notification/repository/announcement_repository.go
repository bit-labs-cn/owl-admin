package repository

import (
	"context"
	"time"

	"bit-labs.cn/owl-admin/app-notification/model"
	"bit-labs.cn/owl/contract"
	"bit-labs.cn/owl/provider/db"
	"gorm.io/gorm"
)

// AnnouncementRepositoryInterface 读写公告。
type AnnouncementRepositoryInterface interface {
	Create(row *model.Announcement) error
	Save(row *model.Announcement) error
	Detail(id uint) (*model.Announcement, error)
	Delete(id uint) error
	Retrieve(page, pageSize int, title string, status int) (int64, []model.Announcement, error)
	ListPublished(now time.Time) ([]model.Announcement, error)
	Dismiss(principalType int, principalID, announcementID uint) error
	DismissedIDs(principalType int, principalID uint) ([]uint, error)
	contract.WithContext[AnnouncementRepositoryInterface]
}

type AnnouncementRepository struct {
	db *gorm.DB
	db.BaseRepository[model.Announcement]
}

func NewAnnouncementRepository(d *gorm.DB) AnnouncementRepositoryInterface {
	return &AnnouncementRepository{db: d, BaseRepository: db.NewBaseRepository[model.Announcement](d)}
}

func (r *AnnouncementRepository) WithContext(ctx context.Context) AnnouncementRepositoryInterface {
	next := r.db.WithContext(ctx)
	return &AnnouncementRepository{db: next, BaseRepository: db.NewBaseRepository[model.Announcement](next)}
}

func (r *AnnouncementRepository) Create(row *model.Announcement) error { return r.db.Create(row).Error }
func (r *AnnouncementRepository) Save(row *model.Announcement) error   { return r.db.Save(row).Error }

func (r *AnnouncementRepository) Detail(id uint) (*model.Announcement, error) {
	var row model.Announcement
	err := r.db.Where("id = ?", id).First(&row).Error
	return &row, err
}

func (r *AnnouncementRepository) Delete(id uint) error {
	return r.db.Delete(&model.Announcement{}, id).Error
}

func (r *AnnouncementRepository) Retrieve(page, pageSize int, title string, status int) (int64, []model.Announcement, error) {
	tx := r.db.Model(&model.Announcement{})
	if title != "" {
		tx = tx.Where("title LIKE ?", "%"+title+"%")
	}
	if status > 0 {
		tx = tx.Where("status = ?", status)
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return 0, nil, err
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	var list []model.Announcement
	err := tx.Order("sort asc, id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return total, list, err
}

func (r *AnnouncementRepository) ListPublished(now time.Time) ([]model.Announcement, error) {
	var list []model.Announcement
	err := r.db.Where("status = ?", model.AnnouncementStatusPublished).
		Where("(start_at IS NULL OR start_at <= ?)", now).
		Where("(end_at IS NULL OR end_at >= ?)", now).
		Order("sort asc, id desc").Find(&list).Error
	return list, err
}

func (r *AnnouncementRepository) Dismiss(principalType int, principalID, announcementID uint) error {
	now := time.Now()
	var row model.AnnouncementState
	err := r.db.Where("principal_type = ? AND principal_id = ? AND announcement_id = ?", principalType, principalID, announcementID).First(&row).Error
	if err == gorm.ErrRecordNotFound {
		row = model.AnnouncementState{PrincipalType: principalType, PrincipalID: principalID, AnnouncementID: announcementID, DismissedAt: &now, LastShownAt: &now}
		return r.db.Create(&row).Error
	}
	if err != nil {
		return err
	}
	row.DismissedAt = &now
	return r.db.Save(&row).Error
}

func (r *AnnouncementRepository) DismissedIDs(principalType int, principalID uint) ([]uint, error) {
	var ids []uint
	err := r.db.Model(&model.AnnouncementState{}).
		Where("principal_type = ? AND principal_id = ? AND dismissed_at IS NOT NULL", principalType, principalID).
		Pluck("announcement_id", &ids).Error
	return ids, err
}
