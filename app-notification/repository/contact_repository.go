package repository

import (
	"context"

	"bit-labs.cn/owl-admin/app-notification/model"
	"bit-labs.cn/owl/contract"
	"bit-labs.cn/owl/provider/db"
	"gorm.io/gorm"
)

// ContactRepositoryInterface 读写外部名单和联系人。
type ContactRepositoryInterface interface {
	SaveList(row *model.ContactList) error
	DetailList(id uint) (*model.ContactList, error)
	DeleteList(id uint) error
	RetrieveLists(page, pageSize int, name string) (int64, []model.ContactList, error)
	SaveContact(row *model.Contact) error
	DetailContact(id uint) (*model.Contact, error)
	DeleteContact(id uint) error
	RetrieveContacts(page, pageSize int, listID uint, name string, status int) (int64, []model.Contact, error)
	ListByIDs(ids []uint) ([]model.Contact, error)
	ListByLists(listIDs []uint) ([]model.Contact, error)
	contract.WithContext[ContactRepositoryInterface]
}

type ContactRepository struct {
	db *gorm.DB
	db.BaseRepository[model.Contact]
}

func NewContactRepository(d *gorm.DB) ContactRepositoryInterface {
	return &ContactRepository{db: d, BaseRepository: db.NewBaseRepository[model.Contact](d)}
}

func (r *ContactRepository) WithContext(ctx context.Context) ContactRepositoryInterface {
	next := r.db.WithContext(ctx)
	return &ContactRepository{db: next, BaseRepository: db.NewBaseRepository[model.Contact](next)}
}

func (r *ContactRepository) SaveList(row *model.ContactList) error { return r.db.Save(row).Error }

func (r *ContactRepository) DetailList(id uint) (*model.ContactList, error) {
	var row model.ContactList
	err := r.db.Where("id = ?", id).First(&row).Error
	return &row, err
}

func (r *ContactRepository) DeleteList(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("list_id = ?", id).Delete(&model.Contact{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.ContactList{}, id).Error
	})
}

func (r *ContactRepository) RetrieveLists(page, pageSize int, name string) (int64, []model.ContactList, error) {
	tx := r.db.Model(&model.ContactList{})
	if name != "" {
		tx = tx.Where("name LIKE ?", "%"+name+"%")
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return 0, nil, err
	}
	page, pageSize = normPage(page, pageSize)
	var list []model.ContactList
	err := tx.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return total, list, err
}

func (r *ContactRepository) SaveContact(row *model.Contact) error { return r.db.Save(row).Error }

func (r *ContactRepository) DetailContact(id uint) (*model.Contact, error) {
	var row model.Contact
	err := r.db.Where("id = ?", id).First(&row).Error
	return &row, err
}

func (r *ContactRepository) DeleteContact(id uint) error {
	return r.db.Delete(&model.Contact{}, id).Error
}

func (r *ContactRepository) RetrieveContacts(page, pageSize int, listID uint, name string, status int) (int64, []model.Contact, error) {
	tx := r.db.Model(&model.Contact{})
	if listID > 0 {
		tx = tx.Where("list_id = ?", listID)
	}
	if name != "" {
		tx = tx.Where("name LIKE ?", "%"+name+"%")
	}
	if status > 0 {
		tx = tx.Where("status = ?", status)
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return 0, nil, err
	}
	page, pageSize = normPage(page, pageSize)
	var list []model.Contact
	err := tx.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return total, list, err
}

func (r *ContactRepository) ListByIDs(ids []uint) ([]model.Contact, error) {
	var list []model.Contact
	if len(ids) == 0 {
		return list, nil
	}
	err := r.db.Where("id IN ?", ids).Find(&list).Error
	return list, err
}

func (r *ContactRepository) ListByLists(listIDs []uint) ([]model.Contact, error) {
	var list []model.Contact
	if len(listIDs) == 0 {
		return list, nil
	}
	err := r.db.Where("list_id IN ?", listIDs).Find(&list).Error
	return list, err
}

func normPage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	return page, pageSize
}
