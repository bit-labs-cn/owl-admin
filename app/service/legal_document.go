package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"

	"bit-labs.cn/owl-admin/app/model"
	"bit-labs.cn/owl-admin/app/repository"
	"bit-labs.cn/owl/provider/redis"
	"bit-labs.cn/owl/provider/router"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/cast"
)

var legalKeyPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{0,63}$`)
var legalVersionPattern = regexp.MustCompile(`^[0-9A-Za-z][0-9A-Za-z._-]{0,31}$`)

// LegalService 管理条款版本、出示场景和同意记录。
type LegalService struct {
	repo     repository.LegalRepositoryInterface
	locker   redis.LockerFactory
	validate *validator.Validate
}

// NewLegalService 创建条款服务。
func NewLegalService(repo repository.LegalRepositoryInterface, locker redis.LockerFactory, validate *validator.Validate) *LegalService {
	return &LegalService{repo: repo, locker: locker, validate: validate}
}

func (i *LegalService) lock(key string) (func(), error) {
	l := i.locker.New()
	if err := l.Lock(key); err != nil {
		return nil, err
	}
	return l.Unlock, nil
}

func normalizeLegalKey(raw string) (string, error) {
	key := strings.ToLower(strings.TrimSpace(raw))
	if !legalKeyPattern.MatchString(key) {
		return "", LegalKeyInvalid()
	}
	return key, nil
}

func normalizeVersionNumber(raw string) (string, error) {
	version := strings.TrimSpace(raw)
	if !legalVersionPattern.MatchString(version) {
		return "", LegalVersionNumberInvalid()
	}
	return version, nil
}

func legalHash(content string) string {
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}

func legalUserLabel(nickname, username string) string {
	if strings.TrimSpace(nickname) != "" {
		return nickname
	}
	if strings.TrimSpace(username) != "" {
		return username
	}
	return "未知用户"
}

// CreateLegalDocumentReq 新建条款。
type CreateLegalDocumentReq struct {
	Key  string `json:"key" validate:"required,max=64" label:"标识"`
	Name string `json:"name" validate:"required,max=128" label:"名称"`
	Kind string `json:"kind" validate:"required,oneof=contract notice optional" label:"性质"`
}

// UpdateLegalDocumentReq 修改条款。标识不能改。
type UpdateLegalDocumentReq struct {
	ID     uint   `json:"id,string"`
	Name   string `json:"name" validate:"required,max=128" label:"名称"`
	Kind   string `json:"kind" validate:"required,oneof=contract notice optional" label:"性质"`
	Status string `json:"status" validate:"required,oneof=enabled disabled" label:"状态"`
}

// RetrieveLegalDocumentReq 分页查询条款。
type RetrieveLegalDocumentReq struct {
	router.PageReq
	Name   string `json:"name" form:"name" validate:"omitempty,max=128" label:"名称"`
	Kind   string `json:"kind" form:"kind" validate:"omitempty,oneof=contract notice optional" label:"性质"`
	Status string `json:"status" form:"status" validate:"omitempty,oneof=enabled disabled" label:"状态"`
}

// LegalDocumentItem 是管理端看到的一条条款。
type LegalDocumentItem struct {
	ID             string   `json:"id"`
	Key            string   `json:"key"`
	Name           string   `json:"name"`
	Kind           string   `json:"kind"`
	KindLabel      string   `json:"kindLabel"`
	Status         string   `json:"status"`
	StatusLabel    string   `json:"statusLabel"`
	CurrentVersion string   `json:"currentVersion"`
	SceneNames     []string `json:"sceneNames"`
}

// LegalOption 是下拉选项。
type LegalOption struct {
	ID     string `json:"id,omitempty"`
	Key    string `json:"key"`
	Name   string `json:"name"`
	Kind   string `json:"kind,omitempty"`
	Status string `json:"status,omitempty"`
}

// CreateDocument 新建条款，默认启用。
func (i *LegalService) CreateDocument(ctx context.Context, req *CreateLegalDocumentReq) error {
	if err := i.validate.Struct(req); err != nil {
		return err
	}
	key, err := normalizeLegalKey(req.Key)
	if err != nil {
		return err
	}
	unlock, err := i.lock("legal:document:" + key)
	if err != nil {
		return err
	}
	defer unlock()

	repo := i.repo.WithContext(ctx)
	existing, err := repo.FindDocumentByKey(key)
	if err != nil {
		return err
	}
	if existing != nil {
		return LegalDocumentExists()
	}
	doc := &model.LegalDocument{
		Key:    key,
		Name:   strings.TrimSpace(req.Name),
		Kind:   req.Kind,
		Status: model.LegalStatusEnabled,
	}
	return repo.SaveDocument(doc)
}

// UpdateDocument 修改名称、性质和启用状态。
func (i *LegalService) UpdateDocument(ctx context.Context, req *UpdateLegalDocumentReq) error {
	if err := i.validate.Struct(req); err != nil {
		return err
	}
	unlock, err := i.lock("legal:document-id:" + cast.ToString(req.ID))
	if err != nil {
		return err
	}
	defer unlock()

	repo := i.repo.WithContext(ctx)
	doc, err := repo.FindDocumentByID(req.ID)
	if err != nil {
		return err
	}
	if doc == nil {
		return LegalDocumentNotFound()
	}
	doc.Name = strings.TrimSpace(req.Name)
	doc.Kind = req.Kind
	doc.Status = req.Status
	return repo.SaveDocument(doc)
}

// DeleteDocument 删除还没有同意记录的条款，并清掉它的版本和场景条目。
func (i *LegalService) DeleteDocument(ctx context.Context, id uint) error {
	unlock, err := i.lock("legal:document-id:" + cast.ToString(id))
	if err != nil {
		return err
	}
	defer unlock()

	repo := i.repo.WithContext(ctx)
	doc, err := repo.FindDocumentByID(id)
	if err != nil {
		return err
	}
	if doc == nil {
		return LegalDocumentNotFound()
	}
	count, err := repo.CountEventsByDocument(doc.Key)
	if err != nil {
		return err
	}
	if count > 0 {
		return LegalDocumentHasEvents()
	}
	return repo.Transaction(func(tx repository.LegalRepositoryInterface) error {
		if err := tx.DeleteItemsByDocument(doc.ID); err != nil {
			return err
		}
		if err := tx.DeleteVersionsByDocument(doc.ID); err != nil {
			return err
		}
		return tx.DeleteDocument(doc.ID)
	})
}

// RetrieveDocuments 分页返回条款，带上当前版本和所在场景。
func (i *LegalService) RetrieveDocuments(ctx context.Context, req *RetrieveLegalDocumentReq) (int, []LegalDocumentItem, error) {
	if err := i.validate.Struct(req); err != nil {
		return 0, nil, err
	}
	repo := i.repo.WithContext(ctx)
	count, docs, err := repo.ListDocuments(req.Page, req.PageSize, strings.TrimSpace(req.Name), req.Kind, req.Status)
	if err != nil {
		return 0, nil, err
	}
	items, err := i.decorateDocuments(repo, docs)
	if err != nil {
		return 0, nil, err
	}
	return int(count), items, nil
}

// DocumentOptions 返回全部条款，供版本和场景选择。停用的条款带 status。
func (i *LegalService) DocumentOptions(ctx context.Context) ([]LegalOption, error) {
	docs, err := i.repo.WithContext(ctx).ListDocumentOptions()
	if err != nil {
		return nil, err
	}
	out := make([]LegalOption, 0, len(docs))
	for _, doc := range docs {
		out = append(out, LegalOption{
			ID:     cast.ToString(doc.ID),
			Key:    doc.Key,
			Name:   doc.Name,
			Kind:   doc.Kind,
			Status: doc.Status,
		})
	}
	return out, nil
}

func (i *LegalService) decorateDocuments(repo repository.LegalRepositoryInterface, docs []model.LegalDocument) ([]LegalDocumentItem, error) {
	ids := make([]uint, 0, len(docs))
	for _, doc := range docs {
		ids = append(ids, doc.ID)
	}
	published, err := repo.ListPublishedVersions(ids)
	if err != nil {
		return nil, err
	}
	versionByDoc := make(map[uint]string, len(published))
	for _, ver := range published {
		versionByDoc[ver.DocumentID] = ver.Version
	}
	scenes, err := repo.ListSceneOptions()
	if err != nil {
		return nil, err
	}
	sceneName := make(map[uint]string, len(scenes))
	for _, scene := range scenes {
		sceneName[scene.ID] = scene.Name
	}
	links, err := repo.ListItemsByDocumentIDs(ids)
	if err != nil {
		return nil, err
	}
	namesByDoc := make(map[uint][]string)
	seen := make(map[uint]map[string]bool)
	for _, link := range links {
		name := sceneName[link.SceneID]
		if name == "" {
			name = "未知场景(" + cast.ToString(link.SceneID) + ")"
		}
		if seen[link.DocumentID] == nil {
			seen[link.DocumentID] = map[string]bool{}
		}
		if seen[link.DocumentID][name] {
			continue
		}
		seen[link.DocumentID][name] = true
		namesByDoc[link.DocumentID] = append(namesByDoc[link.DocumentID], name)
	}
	out := make([]LegalDocumentItem, 0, len(docs))
	for _, doc := range docs {
		names := namesByDoc[doc.ID]
		if names == nil {
			names = []string{}
		}
		out = append(out, LegalDocumentItem{
			ID:             cast.ToString(doc.ID),
			Key:            doc.Key,
			Name:           doc.Name,
			Kind:           doc.Kind,
			KindLabel:      model.LegalKindLabel(doc.Kind),
			Status:         doc.Status,
			StatusLabel:    model.LegalStatusLabel(doc.Status),
			CurrentVersion: versionByDoc[doc.ID],
			SceneNames:     names,
		})
	}
	return out, nil
}
