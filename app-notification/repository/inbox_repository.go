package repository

import (
	"context"
	"strings"
	"time"

	"bit-labs.cn/owl-admin/app-notification/model"
	"bit-labs.cn/owl/contract"
	"bit-labs.cn/owl/provider/db"
	"gorm.io/gorm"
)

// InboxCounts 是铃铛角标需要的数量。
type InboxCounts struct {
	Notice  int64 `json:"notice"`
	Message int64 `json:"message"`
	Todo    int64 `json:"todo"`
}

// InboxRepositoryInterface 读写站内信。
type InboxRepositoryInterface interface {
	Create(row *model.Inbox) error
	Save(row *model.Inbox) error
	Detail(id uint) (*model.Inbox, error)
	FindByPrincipalTask(principalType int, principalID, taskID uint) (*model.Inbox, error)
	List(principalType int, principalID uint, kind int, limit int) ([]model.Inbox, error)
	Retrieve(principalType int, principalID uint, kind int, keyword string, readStatus, page, pageSize int) (int64, []model.Inbox, error)
	Counts(principalType int, principalID uint) (InboxCounts, error)
	MarkRead(principalType int, principalID uint, ids []uint) error
	MarkAllRead(principalType int, principalID uint, kind int) error
	Delete(principalType int, principalID, id uint) error
	Complete(principalType int, principalID, id uint, done bool) error
	CompleteByBizKey(bizKey string) ([]model.Inbox, error)
	ExpireDue(now time.Time) ([]model.Inbox, error)
	DeleteBefore(t time.Time) error
	contract.WithContext[InboxRepositoryInterface]
}

type InboxRepository struct {
	db *gorm.DB
	db.BaseRepository[model.Inbox]
}

func NewInboxRepository(d *gorm.DB) InboxRepositoryInterface {
	return &InboxRepository{db: d, BaseRepository: db.NewBaseRepository[model.Inbox](d)}
}

func (r *InboxRepository) WithContext(ctx context.Context) InboxRepositoryInterface {
	next := r.db.WithContext(ctx)
	return &InboxRepository{db: next, BaseRepository: db.NewBaseRepository[model.Inbox](next)}
}

func (r *InboxRepository) Create(row *model.Inbox) error { return r.db.Create(row).Error }
func (r *InboxRepository) Save(row *model.Inbox) error   { return r.db.Save(row).Error }

func (r *InboxRepository) Detail(id uint) (*model.Inbox, error) {
	var row model.Inbox
	err := r.db.Where("id = ?", id).First(&row).Error
	return &row, err
}

func (r *InboxRepository) FindByPrincipalTask(principalType int, principalID, taskID uint) (*model.Inbox, error) {
	var row model.Inbox
	err := r.db.Where("principal_type = ? AND principal_id = ? AND task_id = ?", principalType, principalID, taskID).First(&row).Error
	return &row, err
}

func (r *InboxRepository) List(principalType int, principalID uint, kind int, limit int) ([]model.Inbox, error) {
	if limit < 1 {
		limit = 20
	}
	tx := r.db.Where("principal_type = ? AND principal_id = ?", principalType, principalID)
	if kind > 0 {
		tx = tx.Where("kind = ?", kind)
	}
	var list []model.Inbox
	err := tx.Order("id desc").Limit(limit).Find(&list).Error
	return list, err
}

func (r *InboxRepository) Retrieve(principalType int, principalID uint, kind int, keyword string, readStatus, page, pageSize int) (int64, []model.Inbox, error) {
	tx := r.db.Model(&model.Inbox{}).Where("principal_type = ? AND principal_id = ?", principalType, principalID)
	if kind > 0 {
		tx = tx.Where("kind = ?", kind)
	}
	if word := strings.TrimSpace(keyword); word != "" {
		tx = tx.Where("title LIKE ?", "%"+word+"%")
	}
	if readStatus > 0 {
		tx = tx.Where("read_status = ?", readStatus)
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
	var list []model.Inbox
	err := tx.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return total, list, err
}

func (r *InboxRepository) Counts(principalType int, principalID uint) (InboxCounts, error) {
	var out InboxCounts
	base := r.db.Model(&model.Inbox{}).Where("principal_type = ? AND principal_id = ? AND read_status = ?", principalType, principalID, model.ReadStatusUnread)
	if err := base.Where("kind = ?", model.KindNotice).Count(&out.Notice).Error; err != nil {
		return out, err
	}
	if err := r.db.Model(&model.Inbox{}).Where("principal_type = ? AND principal_id = ? AND read_status = ? AND kind = ?", principalType, principalID, model.ReadStatusUnread, model.KindMessage).Count(&out.Message).Error; err != nil {
		return out, err
	}
	err := r.db.Model(&model.Inbox{}).Where("principal_type = ? AND principal_id = ? AND kind = ? AND todo_status = ?", principalType, principalID, model.KindTodo, model.TodoStatusPending).Count(&out.Todo).Error
	return out, err
}

func (r *InboxRepository) MarkRead(principalType int, principalID uint, ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	now := time.Now()
	return r.db.Model(&model.Inbox{}).
		Where("principal_type = ? AND principal_id = ? AND id IN ?", principalType, principalID, ids).
		Updates(map[string]any{"read_status": model.ReadStatusRead, "read_at": now}).Error
}

func (r *InboxRepository) MarkAllRead(principalType int, principalID uint, kind int) error {
	now := time.Now()
	tx := r.db.Model(&model.Inbox{}).Where("principal_type = ? AND principal_id = ? AND read_status = ?", principalType, principalID, model.ReadStatusUnread)
	if kind > 0 {
		tx = tx.Where("kind = ?", kind)
	}
	return tx.Updates(map[string]any{"read_status": model.ReadStatusRead, "read_at": now}).Error
}

func (r *InboxRepository) Delete(principalType int, principalID, id uint) error {
	return r.db.Where("principal_type = ? AND principal_id = ? AND id = ?", principalType, principalID, id).Delete(&model.Inbox{}).Error
}

func (r *InboxRepository) Complete(principalType int, principalID, id uint, done bool) error {
	status := model.TodoStatusPending
	var doneAt any
	if done {
		status = model.TodoStatusDone
		now := time.Now()
		doneAt = now
	}
	return r.db.Model(&model.Inbox{}).
		Where("principal_type = ? AND principal_id = ? AND id = ? AND kind = ?", principalType, principalID, id, model.KindTodo).
		Updates(map[string]any{"todo_status": status, "done_at": doneAt}).Error
}

func (r *InboxRepository) CompleteByBizKey(bizKey string) ([]model.Inbox, error) {
	var list []model.Inbox
	err := r.db.Where("biz_key = ? AND kind = ? AND todo_status = ?", bizKey, model.KindTodo, model.TodoStatusPending).Find(&list).Error
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return list, nil
	}
	now := time.Now()
	err = r.db.Model(&model.Inbox{}).
		Where("biz_key = ? AND kind = ? AND todo_status = ?", bizKey, model.KindTodo, model.TodoStatusPending).
		Updates(map[string]any{"todo_status": model.TodoStatusDone, "done_at": now, "read_status": model.ReadStatusRead, "read_at": now}).Error
	return list, err
}

func (r *InboxRepository) ExpireDue(now time.Time) ([]model.Inbox, error) {
	var list []model.Inbox
	err := r.db.Where("kind = ? AND todo_status = ? AND due_at IS NOT NULL AND due_at < ?", model.KindTodo, model.TodoStatusPending, now).Find(&list).Error
	if err != nil || len(list) == 0 {
		return list, err
	}
	err = r.db.Model(&model.Inbox{}).
		Where("kind = ? AND todo_status = ? AND due_at IS NOT NULL AND due_at < ?", model.KindTodo, model.TodoStatusPending, now).
		Update("todo_status", model.TodoStatusExpired).Error
	return list, err
}

func (r *InboxRepository) DeleteBefore(t time.Time) error {
	return r.db.Where("created_at < ? AND kind <> ?", t, model.KindTodo).Delete(&model.Inbox{}).Error
}
