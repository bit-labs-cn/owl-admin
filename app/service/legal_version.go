package service

import (
	"context"
	"strings"
	"time"

	"bit-labs.cn/owl-admin/app/model"
	"bit-labs.cn/owl-admin/app/repository"
	"bit-labs.cn/owl/provider/router"
	"github.com/spf13/cast"
)

// CreateLegalVersionReq 新建草稿。
type CreateLegalVersionReq struct {
	DocumentID uint   `json:"documentId,string" validate:"required" label:"条款"`
	Version    string `json:"version" validate:"required,max=32" label:"版本号"`
	Title      string `json:"title" validate:"required,max=128" label:"标题"`
	Content    string `json:"content" validate:"required,max=200000" label:"正文"`
}

// UpdateLegalVersionReq 修改草稿。
type UpdateLegalVersionReq struct {
	ID      uint   `json:"id,string"`
	Version string `json:"version" validate:"required,max=32" label:"版本号"`
	Title   string `json:"title" validate:"required,max=128" label:"标题"`
	Content string `json:"content" validate:"required,max=200000" label:"正文"`
}

// PublishLegalVersionReq 发布草稿。
type PublishLegalVersionReq struct {
	ID          uint   `json:"id,string"`
	ChangeClass string `json:"changeClass" validate:"required,oneof=editorial material" label:"变更级别"`
}

// RetrieveLegalVersionReq 分页查询某个条款的版本。
type RetrieveLegalVersionReq struct {
	router.PageReq
	DocumentID uint `json:"documentId,string" form:"documentId" validate:"required" label:"条款"`
}

// LegalVersionItem 是管理端看到的一个版本。
type LegalVersionItem struct {
	ID            string     `json:"id"`
	DocumentID    string     `json:"documentId"`
	Version       string     `json:"version"`
	Title         string     `json:"title"`
	Content       string     `json:"content"`
	ContentSHA256 string     `json:"contentSha256"`
	ChangeClass   string     `json:"changeClass"`
	ChangeLabel   string     `json:"changeLabel"`
	Status        string     `json:"status"`
	StatusLabel   string     `json:"statusLabel"`
	EffectiveAt   *time.Time `json:"effectiveAt,omitempty"`
	PublishedAt   *time.Time `json:"publishedAt,omitempty"`
	CreatedAt     *time.Time `json:"createdAt,omitempty"`
}

// CreateVersion 为条款新建草稿。
func (i *LegalService) CreateVersion(ctx context.Context, req *CreateLegalVersionReq) error {
	if err := i.validate.Struct(req); err != nil {
		return err
	}
	version, err := normalizeVersionNumber(req.Version)
	if err != nil {
		return err
	}
	if strings.TrimSpace(req.Content) == "" {
		return LegalContentEmpty()
	}
	unlock, err := i.lock("legal:document-id:" + cast.ToString(req.DocumentID))
	if err != nil {
		return err
	}
	defer unlock()

	repo := i.repo.WithContext(ctx)
	doc, err := repo.FindDocumentByID(req.DocumentID)
	if err != nil {
		return err
	}
	if doc == nil {
		return LegalDocumentNotFound()
	}
	existing, err := repo.FindVersion(doc.ID, version)
	if err != nil {
		return err
	}
	if existing != nil {
		return LegalVersionExists()
	}
	ver := &model.LegalVersion{
		DocumentID: doc.ID,
		Version:    version,
		Title:      strings.TrimSpace(req.Title),
		Content:    req.Content,
		Status:     model.LegalVersionDraft,
	}
	return repo.SaveVersion(ver)
}

// UpdateVersion 修改草稿的版本号、标题和正文。
func (i *LegalService) UpdateVersion(ctx context.Context, req *UpdateLegalVersionReq) error {
	if err := i.validate.Struct(req); err != nil {
		return err
	}
	version, err := normalizeVersionNumber(req.Version)
	if err != nil {
		return err
	}
	if strings.TrimSpace(req.Content) == "" {
		return LegalContentEmpty()
	}
	unlock, err := i.lock("legal:version:" + cast.ToString(req.ID))
	if err != nil {
		return err
	}
	defer unlock()

	repo := i.repo.WithContext(ctx)
	ver, err := repo.FindVersionByID(req.ID)
	if err != nil {
		return err
	}
	if ver == nil {
		return LegalVersionNotFound()
	}
	if ver.Status != model.LegalVersionDraft {
		return LegalVersionFrozen()
	}
	if ver.Version != version {
		other, err := repo.FindVersion(ver.DocumentID, version)
		if err != nil {
			return err
		}
		if other != nil {
			return LegalVersionExists()
		}
	}
	ver.Version = version
	ver.Title = strings.TrimSpace(req.Title)
	ver.Content = req.Content
	return repo.SaveVersion(ver)
}

// DeleteVersion 删除没有同意记录的草稿。
func (i *LegalService) DeleteVersion(ctx context.Context, id uint) error {
	unlock, err := i.lock("legal:version:" + cast.ToString(id))
	if err != nil {
		return err
	}
	defer unlock()

	repo := i.repo.WithContext(ctx)
	ver, err := repo.FindVersionByID(id)
	if err != nil {
		return err
	}
	if ver == nil {
		return LegalVersionNotFound()
	}
	if ver.Status != model.LegalVersionDraft {
		return LegalVersionNotDraft()
	}
	count, err := repo.CountEventsByVersion(ver.ID)
	if err != nil {
		return err
	}
	if count > 0 {
		return LegalDocumentHasEvents()
	}
	return repo.DeleteVersion(ver.ID)
}

// PublishVersion 冻结草稿并把它设为当前已发布版本。原先已发布的改为归档。
func (i *LegalService) PublishVersion(ctx context.Context, req *PublishLegalVersionReq) error {
	if err := i.validate.Struct(req); err != nil {
		return err
	}
	unlock, err := i.lock("legal:version:" + cast.ToString(req.ID))
	if err != nil {
		return err
	}
	defer unlock()

	repo := i.repo.WithContext(ctx)
	ver, err := repo.FindVersionByID(req.ID)
	if err != nil {
		return err
	}
	if ver == nil {
		return LegalVersionNotFound()
	}
	if ver.Status != model.LegalVersionDraft {
		return LegalVersionNotDraft()
	}
	now := time.Now()
	ver.Status = model.LegalVersionPublished
	ver.ChangeClass = req.ChangeClass
	ver.ContentSHA256 = legalHash(ver.Content)
	ver.PublishedAt = &now
	ver.EffectiveAt = &now
	return repo.Transaction(func(tx repository.LegalRepositoryInterface) error {
		if err := tx.ArchivePublished(ver.DocumentID, ver.ID); err != nil {
			return err
		}
		return tx.SaveVersion(ver)
	})
}

// RetrieveVersions 分页返回某个条款的版本，含正文。
func (i *LegalService) RetrieveVersions(ctx context.Context, req *RetrieveLegalVersionReq) (int, []LegalVersionItem, error) {
	if err := i.validate.Struct(req); err != nil {
		return 0, nil, err
	}
	repo := i.repo.WithContext(ctx)
	doc, err := repo.FindDocumentByID(req.DocumentID)
	if err != nil {
		return 0, nil, err
	}
	if doc == nil {
		return 0, nil, LegalDocumentNotFound()
	}
	count, list, err := repo.ListVersions(req.DocumentID, req.Page, req.PageSize)
	if err != nil {
		return 0, nil, err
	}
	out := make([]LegalVersionItem, 0, len(list))
	for _, ver := range list {
		out = append(out, toVersionItem(ver))
	}
	return int(count), out, nil
}

func toVersionItem(ver model.LegalVersion) LegalVersionItem {
	return LegalVersionItem{
		ID:            cast.ToString(ver.ID),
		DocumentID:    cast.ToString(ver.DocumentID),
		Version:       ver.Version,
		Title:         ver.Title,
		Content:       ver.Content,
		ContentSHA256: ver.ContentSHA256,
		ChangeClass:   ver.ChangeClass,
		ChangeLabel:   model.LegalChangeLabel(ver.ChangeClass),
		Status:        ver.Status,
		StatusLabel:   model.LegalVersionStatusLabel(ver.Status),
		EffectiveAt:   ver.EffectiveAt,
		PublishedAt:   ver.PublishedAt,
		CreatedAt:     ver.CreatedAt,
	}
}
