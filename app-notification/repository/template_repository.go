package repository

import (
	"context"

	"bit-labs.cn/owl-admin/app-notification/model"
	"bit-labs.cn/owl/contract"
	"bit-labs.cn/owl/provider/db"
	"gorm.io/gorm"
)

// TemplateRepositoryInterface 读写通知模板。
type TemplateRepositoryInterface interface {
	Create(row *model.Template) error
	Save(row *model.Template) error
	Detail(id uint) (*model.Template, error)
	FindByCode(code string) (*model.Template, error)
	Delete(id uint) error
	Retrieve(page, pageSize int, name, code string, status int) (int64, []model.Template, error)
	contract.WithContext[TemplateRepositoryInterface]
}

type TemplateRepository struct {
	db *gorm.DB
	db.BaseRepository[model.Template]
}

func NewTemplateRepository(d *gorm.DB) TemplateRepositoryInterface {
	return &TemplateRepository{db: d, BaseRepository: db.NewBaseRepository[model.Template](d)}
}

func (r *TemplateRepository) WithContext(ctx context.Context) TemplateRepositoryInterface {
	return &TemplateRepository{db: r.db.WithContext(ctx), BaseRepository: db.NewBaseRepository[model.Template](r.db.WithContext(ctx))}
}

func (r *TemplateRepository) Create(row *model.Template) error { return r.db.Create(row).Error }

func (r *TemplateRepository) Save(row *model.Template) error { return r.db.Save(row).Error }

func (r *TemplateRepository) Detail(id uint) (*model.Template, error) {
	var row model.Template
	err := r.db.Where("id = ?", id).First(&row).Error
	return &row, err
}

func (r *TemplateRepository) FindByCode(code string) (*model.Template, error) {
	var row model.Template
	err := r.db.Where("code = ?", code).First(&row).Error
	return &row, err
}

func (r *TemplateRepository) Delete(id uint) error {
	return r.db.Delete(&model.Template{}, id).Error
}

func (r *TemplateRepository) Retrieve(page, pageSize int, name, code string, status int) (int64, []model.Template, error) {
	tx := r.db.Model(&model.Template{})
	if name != "" {
		tx = tx.Where("name LIKE ?", "%"+name+"%")
	}
	if code != "" {
		tx = tx.Where("code = ?", code)
	}
	if status > 0 {
		tx = tx.Where("status = ?", status)
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return 0, nil, err
	}
	var list []model.Template
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	err := tx.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return total, list, err
}
