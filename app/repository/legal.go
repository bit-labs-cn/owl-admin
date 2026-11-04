package repository

import (
	"context"
	"errors"

	"bit-labs.cn/owl-admin/app/model"
	"bit-labs.cn/owl/contract"
	"bit-labs.cn/owl/provider/db"
	"gorm.io/gorm"
)

// LegalRepositoryInterface 读写条款、版本、场景和同意事件。
type LegalRepositoryInterface interface {
	Transaction(fn func(tx LegalRepositoryInterface) error) error

	SaveDocument(doc *model.LegalDocument) error
	FindDocumentByID(id uint) (*model.LegalDocument, error)
	FindDocumentByKey(key string) (*model.LegalDocument, error)
	DeleteDocument(id uint) error
	ListDocuments(page, pageSize int, name, kind, status string) (int64, []model.LegalDocument, error)
	ListDocumentsByStatus(status string) ([]model.LegalDocument, error)
	ListDocumentOptions() ([]model.LegalDocument, error)

	SaveVersion(ver *model.LegalVersion) error
	FindVersionByID(id uint) (*model.LegalVersion, error)
	FindVersion(documentID uint, version string) (*model.LegalVersion, error)
	FindPublishedVersion(documentID uint) (*model.LegalVersion, error)
	ListVersions(documentID uint, page, pageSize int) (int64, []model.LegalVersion, error)
	ListPublishedVersions(documentIDs []uint) ([]model.LegalVersion, error)
	ArchivePublished(documentID uint, exceptID uint) error
	DeleteVersion(id uint) error
	DeleteVersionsByDocument(documentID uint) error

	SaveScene(scene *model.LegalScene) error
	FindSceneByID(id uint) (*model.LegalScene, error)
	FindSceneByKey(key string) (*model.LegalScene, error)
	DeleteScene(id uint) error
	ListScenes(page, pageSize int, name, kind string) (int64, []model.LegalScene, error)
	ListSceneOptions() ([]model.LegalScene, error)
	ReplaceSceneItems(sceneID uint, items []model.LegalSceneItem) error
	ListSceneItems(sceneID uint) ([]model.LegalSceneItem, error)
	ListItemsBySceneIDs(sceneIDs []uint) ([]model.LegalSceneItem, error)
	ListItemsByDocumentIDs(documentIDs []uint) ([]model.LegalSceneItem, error)
	DeleteItemsByDocument(documentID uint) error

	CountEventsByDocument(documentKey string) (int64, error)
	CountEventsByVersion(versionID uint) (int64, error)
	FindEvent(userID, versionID uint, action string) (*model.LegalEvent, error)
	LatestEvent(userID uint, documentKey string) (*model.LegalEvent, error)
	InsertEvent(event *model.LegalEvent) error
	ListEvents(page, pageSize int, documentKey, sceneKey, action string) (int64, []LegalEventRow, error)

	contract.WithContext[LegalRepositoryInterface]
}

// LegalEventRow 是同意记录和管理端展示用的用户、场景名称。
type LegalEventRow struct {
	model.LegalEvent
	LiveUserID   uint   `gorm:"column:live_user_id"`
	LiveNickname string `gorm:"column:live_nickname"`
	LiveUsername string `gorm:"column:live_username"`
	SceneName    string `gorm:"column:scene_name"`
}

var _ LegalRepositoryInterface = (*LegalRepository)(nil)

// LegalRepository 访问条款相关的表。
type LegalRepository struct {
	db  *gorm.DB
	ctx context.Context
}

// NewLegalRepository 返回条款仓储。
func NewLegalRepository(d *gorm.DB) LegalRepositoryInterface {
	return &LegalRepository{db: d}
}

// WithContext 在这次请求的上下文里查询。
func (i *LegalRepository) WithContext(ctx context.Context) LegalRepositoryInterface {
	return &LegalRepository{db: i.db.WithContext(ctx), ctx: ctx}
}

// Transaction 在一个事务里执行。
func (i *LegalRepository) Transaction(fn func(tx LegalRepositoryInterface) error) error {
	return i.db.Transaction(func(tx *gorm.DB) error {
		return fn(&LegalRepository{db: tx, ctx: i.ctx})
	})
}

func (i *LegalRepository) SaveDocument(doc *model.LegalDocument) error {
	return i.db.Save(doc).Error
}

func (i *LegalRepository) FindDocumentByID(id uint) (*model.LegalDocument, error) {
	var doc model.LegalDocument
	err := i.db.Where("id = ?", id).First(&doc).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func (i *LegalRepository) FindDocumentByKey(key string) (*model.LegalDocument, error) {
	var doc model.LegalDocument
	err := i.db.Where("key = ?", key).First(&doc).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func (i *LegalRepository) DeleteDocument(id uint) error {
	return i.db.Unscoped().Where("id = ?", id).Delete(&model.LegalDocument{}).Error
}

func (i *LegalRepository) ListDocuments(page, pageSize int, name, kind, status string) (int64, []model.LegalDocument, error) {
	tx := i.db.Model(&model.LegalDocument{})
	if name != "" {
		tx = tx.Where("name LIKE ?", "%"+name+"%")
	}
	if kind != "" {
		tx = tx.Where("kind = ?", kind)
	}
	if status != "" {
		tx = tx.Where("status = ?", status)
	}
	var count int64
	if err := tx.Count(&count).Error; err != nil {
		return 0, nil, err
	}
	var list []model.LegalDocument
	err := tx.Order("id desc").Scopes(db.Paginate(page, pageSize)).Find(&list).Error
	if list == nil {
		list = []model.LegalDocument{}
	}
	return count, list, err
}

func (i *LegalRepository) ListDocumentsByStatus(status string) ([]model.LegalDocument, error) {
	tx := i.db.Model(&model.LegalDocument{})
	if status != "" {
		tx = tx.Where("status = ?", status)
	}
	var list []model.LegalDocument
	err := tx.Order("id asc").Find(&list).Error
	if list == nil {
		list = []model.LegalDocument{}
	}
	return list, err
}

func (i *LegalRepository) ListDocumentOptions() ([]model.LegalDocument, error) {
	var list []model.LegalDocument
	err := i.db.Order("name asc").Find(&list).Error
	if list == nil {
		list = []model.LegalDocument{}
	}
	return list, err
}

func (i *LegalRepository) SaveVersion(ver *model.LegalVersion) error {
	return i.db.Save(ver).Error
}

func (i *LegalRepository) FindVersionByID(id uint) (*model.LegalVersion, error) {
	var ver model.LegalVersion
	err := i.db.Where("id = ?", id).First(&ver).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &ver, nil
}

func (i *LegalRepository) FindVersion(documentID uint, version string) (*model.LegalVersion, error) {
	var ver model.LegalVersion
	err := i.db.Where("document_id = ? AND version = ?", documentID, version).First(&ver).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &ver, nil
}

func (i *LegalRepository) FindPublishedVersion(documentID uint) (*model.LegalVersion, error) {
	var ver model.LegalVersion
	err := i.db.Where("document_id = ? AND status = ?", documentID, model.LegalVersionPublished).First(&ver).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &ver, nil
}

func (i *LegalRepository) ListVersions(documentID uint, page, pageSize int) (int64, []model.LegalVersion, error) {
	tx := i.db.Model(&model.LegalVersion{}).Where("document_id = ?", documentID)
	var count int64
	if err := tx.Count(&count).Error; err != nil {
		return 0, nil, err
	}
	var list []model.LegalVersion
	err := tx.Order("id desc").Scopes(db.Paginate(page, pageSize)).Find(&list).Error
	if list == nil {
		list = []model.LegalVersion{}
	}
	return count, list, err
}

func (i *LegalRepository) ListPublishedVersions(documentIDs []uint) ([]model.LegalVersion, error) {
	if len(documentIDs) == 0 {
		return []model.LegalVersion{}, nil
	}
	var list []model.LegalVersion
	err := i.db.Where("document_id IN ? AND status = ?", documentIDs, model.LegalVersionPublished).Find(&list).Error
	if list == nil {
		list = []model.LegalVersion{}
	}
	return list, err
}

func (i *LegalRepository) ArchivePublished(documentID uint, exceptID uint) error {
	tx := i.db.Model(&model.LegalVersion{}).
		Where("document_id = ? AND status = ?", documentID, model.LegalVersionPublished)
	if exceptID > 0 {
		tx = tx.Where("id <> ?", exceptID)
	}
	return tx.Update("status", model.LegalVersionArchived).Error
}

func (i *LegalRepository) DeleteVersion(id uint) error {
	return i.db.Unscoped().Where("id = ?", id).Delete(&model.LegalVersion{}).Error
}

func (i *LegalRepository) DeleteVersionsByDocument(documentID uint) error {
	return i.db.Unscoped().Where("document_id = ?", documentID).Delete(&model.LegalVersion{}).Error
}

func (i *LegalRepository) SaveScene(scene *model.LegalScene) error {
	return i.db.Save(scene).Error
}

func (i *LegalRepository) FindSceneByID(id uint) (*model.LegalScene, error) {
	var scene model.LegalScene
	err := i.db.Where("id = ?", id).First(&scene).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &scene, nil
}

func (i *LegalRepository) FindSceneByKey(key string) (*model.LegalScene, error) {
	var scene model.LegalScene
	err := i.db.Where("key = ?", key).First(&scene).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &scene, nil
}

func (i *LegalRepository) DeleteScene(id uint) error {
	if err := i.db.Unscoped().Where("scene_id = ?", id).Delete(&model.LegalSceneItem{}).Error; err != nil {
		return err
	}
	return i.db.Unscoped().Where("id = ?", id).Delete(&model.LegalScene{}).Error
}

func (i *LegalRepository) ListScenes(page, pageSize int, name, kind string) (int64, []model.LegalScene, error) {
	tx := i.db.Model(&model.LegalScene{})
	if name != "" {
		tx = tx.Where("name LIKE ?", "%"+name+"%")
	}
	if kind != "" {
		tx = tx.Where("kind = ?", kind)
	}
	var count int64
	if err := tx.Count(&count).Error; err != nil {
		return 0, nil, err
	}
	var list []model.LegalScene
	err := tx.Order("id desc").Scopes(db.Paginate(page, pageSize)).Find(&list).Error
	if list == nil {
		list = []model.LegalScene{}
	}
	return count, list, err
}

func (i *LegalRepository) ListSceneOptions() ([]model.LegalScene, error) {
	var list []model.LegalScene
	err := i.db.Order("name asc").Find(&list).Error
	if list == nil {
		list = []model.LegalScene{}
	}
	return list, err
}

func (i *LegalRepository) ReplaceSceneItems(sceneID uint, items []model.LegalSceneItem) error {
	if err := i.db.Unscoped().Where("scene_id = ?", sceneID).Delete(&model.LegalSceneItem{}).Error; err != nil {
		return err
	}
	if len(items) == 0 {
		return nil
	}
	return i.db.Create(&items).Error
}

func (i *LegalRepository) ListSceneItems(sceneID uint) ([]model.LegalSceneItem, error) {
	var list []model.LegalSceneItem
	err := i.db.Where("scene_id = ?", sceneID).Order("sort asc").Order("id asc").Find(&list).Error
	if list == nil {
		list = []model.LegalSceneItem{}
	}
	return list, err
}

func (i *LegalRepository) ListItemsBySceneIDs(sceneIDs []uint) ([]model.LegalSceneItem, error) {
	if len(sceneIDs) == 0 {
		return []model.LegalSceneItem{}, nil
	}
	var list []model.LegalSceneItem
	err := i.db.Where("scene_id IN ?", sceneIDs).Order("sort asc").Order("id asc").Find(&list).Error
	if list == nil {
		list = []model.LegalSceneItem{}
	}
	return list, err
}

func (i *LegalRepository) ListItemsByDocumentIDs(documentIDs []uint) ([]model.LegalSceneItem, error) {
	if len(documentIDs) == 0 {
		return []model.LegalSceneItem{}, nil
	}
	var list []model.LegalSceneItem
	err := i.db.Where("document_id IN ?", documentIDs).Order("sort asc").Find(&list).Error
	if list == nil {
		list = []model.LegalSceneItem{}
	}
	return list, err
}

func (i *LegalRepository) DeleteItemsByDocument(documentID uint) error {
	return i.db.Unscoped().Where("document_id = ?", documentID).Delete(&model.LegalSceneItem{}).Error
}

func (i *LegalRepository) CountEventsByDocument(documentKey string) (int64, error) {
	var count int64
	err := i.db.Model(&model.LegalEvent{}).Where("document_key = ?", documentKey).Count(&count).Error
	return count, err
}

func (i *LegalRepository) CountEventsByVersion(versionID uint) (int64, error) {
	var count int64
	err := i.db.Model(&model.LegalEvent{}).Where("version_id = ?", versionID).Count(&count).Error
	return count, err
}

func (i *LegalRepository) FindEvent(userID, versionID uint, action string) (*model.LegalEvent, error) {
	var event model.LegalEvent
	err := i.db.Where("user_id = ? AND version_id = ? AND action = ?", userID, versionID, action).First(&event).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (i *LegalRepository) LatestEvent(userID uint, documentKey string) (*model.LegalEvent, error) {
	var event model.LegalEvent
	err := i.db.Where("user_id = ? AND document_key = ?", userID, documentKey).Order("id desc").First(&event).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (i *LegalRepository) InsertEvent(event *model.LegalEvent) error {
	return i.db.Create(event).Error
}

func (i *LegalRepository) ListEvents(page, pageSize int, documentKey, sceneKey, action string) (int64, []LegalEventRow, error) {
	tx := i.db.Table("admin_legal_event AS e").
		Select("e.*, u.id AS live_user_id, u.nickname AS live_nickname, u.username AS live_username, s.name AS scene_name").
		Joins("LEFT JOIN admin_user AS u ON u.id = e.user_id AND u.deleted_at IS NULL").
		Joins("LEFT JOIN admin_legal_scene AS s ON s.key = e.scene_key AND s.deleted_at IS NULL").
		Where("e.deleted_at IS NULL")
	if documentKey != "" {
		tx = tx.Where("e.document_key = ?", documentKey)
	}
	if sceneKey != "" {
		tx = tx.Where("e.scene_key = ?", sceneKey)
	}
	if action != "" {
		tx = tx.Where("e.action = ?", action)
	}
	var count int64
	if err := tx.Count(&count).Error; err != nil {
		return 0, nil, err
	}
	var list []LegalEventRow
	err := tx.Order("e.id desc").Scopes(db.Paginate(page, pageSize)).Find(&list).Error
	if list == nil {
		list = []LegalEventRow{}
	}
	return count, list, err
}
