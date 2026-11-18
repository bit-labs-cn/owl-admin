package repository

import (
	"context"
	"time"

	"bit-labs.cn/owl-admin/app-notification/model"
	"bit-labs.cn/owl/contract"
	"bit-labs.cn/owl/provider/db"
	"gorm.io/gorm"
)

// TaskRepositoryInterface 读写通知任务。
type TaskRepositoryInterface interface {
	Create(row *model.Task) error
	Save(row *model.Task) error
	Detail(id uint) (*model.Task, error)
	FindByIdempotency(key string) (*model.Task, error)
	Delete(id uint) error
	Retrieve(page, pageSize int, title string, kind, status int) (int64, []model.Task, error)
	Titles(ids []uint) (map[uint]string, error)
	ListDue(now time.Time, limit int) ([]model.Task, error)
	contract.WithContext[TaskRepositoryInterface]
}

type TaskRepository struct {
	db *gorm.DB
	db.BaseRepository[model.Task]
}

func NewTaskRepository(d *gorm.DB) TaskRepositoryInterface {
	return &TaskRepository{db: d, BaseRepository: db.NewBaseRepository[model.Task](d)}
}

func (r *TaskRepository) WithContext(ctx context.Context) TaskRepositoryInterface {
	next := r.db.WithContext(ctx)
	return &TaskRepository{db: next, BaseRepository: db.NewBaseRepository[model.Task](next)}
}

func (r *TaskRepository) Create(row *model.Task) error { return r.db.Create(row).Error }
func (r *TaskRepository) Save(row *model.Task) error   { return r.db.Save(row).Error }

func (r *TaskRepository) Detail(id uint) (*model.Task, error) {
	var row model.Task
	err := r.db.Where("id = ?", id).First(&row).Error
	return &row, err
}

func (r *TaskRepository) FindByIdempotency(key string) (*model.Task, error) {
	var row model.Task
	err := r.db.Where("idempotency_key = ?", key).First(&row).Error
	return &row, err
}

func (r *TaskRepository) Delete(id uint) error {
	return r.db.Delete(&model.Task{}, id).Error
}

func (r *TaskRepository) Retrieve(page, pageSize int, title string, kind, status int) (int64, []model.Task, error) {
	tx := r.db.Model(&model.Task{})
	if title != "" {
		tx = tx.Where("title LIKE ?", "%"+title+"%")
	}
	if kind > 0 {
		tx = tx.Where("kind = ?", kind)
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
	var list []model.Task
	err := tx.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return total, list, err
}

func (r *TaskRepository) ListDue(now time.Time, limit int) ([]model.Task, error) {
	var list []model.Task
	err := r.db.Where("status = ? AND schedule_at IS NOT NULL AND schedule_at <= ?", model.TaskStatusPending, now).
		Order("schedule_at asc").Limit(limit).Find(&list).Error
	return list, err
}

func (r *TaskRepository) Titles(ids []uint) (map[uint]string, error) {
	out := make(map[uint]string, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	var rows []struct {
		ID    uint
		Title string
	}
	if err := r.db.Model(&model.Task{}).Select("id, title").Where("id IN ?", ids).Scan(&rows).Error; err != nil {
		return out, err
	}
	for _, row := range rows {
		out[row.ID] = row.Title
	}
	return out, nil
}
