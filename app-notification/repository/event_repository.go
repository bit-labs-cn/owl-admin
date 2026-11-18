package repository

import (
	"context"

	"bit-labs.cn/owl-admin/app-notification/model"
	"bit-labs.cn/owl/contract"
	"bit-labs.cn/owl/provider/db"
	"gorm.io/gorm"
)

// EventRepositoryInterface 读写事件目录。
type EventRepositoryInterface interface {
	Save(row *model.Event) error
	Detail(id uint) (*model.Event, error)
	FindByCode(code string) (*model.Event, error)
	Delete(id uint) error
	Retrieve(page, pageSize int, name, source string) (int64, []model.Event, error)
	contract.WithContext[EventRepositoryInterface]
}

type EventRepository struct {
	db *gorm.DB
	db.BaseRepository[model.Event]
}

func NewEventRepository(d *gorm.DB) EventRepositoryInterface {
	return &EventRepository{db: d, BaseRepository: db.NewBaseRepository[model.Event](d)}
}

func (r *EventRepository) WithContext(ctx context.Context) EventRepositoryInterface {
	next := r.db.WithContext(ctx)
	return &EventRepository{db: next, BaseRepository: db.NewBaseRepository[model.Event](next)}
}

func (r *EventRepository) Save(row *model.Event) error { return r.db.Save(row).Error }

func (r *EventRepository) Detail(id uint) (*model.Event, error) {
	var row model.Event
	err := r.db.Where("id = ?", id).First(&row).Error
	return &row, err
}

func (r *EventRepository) FindByCode(code string) (*model.Event, error) {
	var row model.Event
	err := r.db.Where("code = ?", code).First(&row).Error
	return &row, err
}

func (r *EventRepository) Delete(id uint) error { return r.db.Delete(&model.Event{}, id).Error }

func (r *EventRepository) Retrieve(page, pageSize int, name, source string) (int64, []model.Event, error) {
	tx := r.db.Model(&model.Event{})
	if name != "" {
		tx = tx.Where("name LIKE ?", "%"+name+"%")
	}
	if source != "" {
		tx = tx.Where("source = ?", source)
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
	var list []model.Event
	err := tx.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return total, list, err
}
