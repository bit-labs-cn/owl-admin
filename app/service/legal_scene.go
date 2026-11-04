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

// LegalSceneItemReq 是场景里的一条条款。
type LegalSceneItemReq struct {
	DocumentKey string `json:"documentKey" validate:"required,max=64" label:"条款"`
	Required    bool   `json:"required"`
	GroupKey    string `json:"groupKey" validate:"omitempty,max=64" label:"勾选组"`
	Wording     string `json:"wording" validate:"omitempty,max=128" label:"勾选文案"`
	Sort        int    `json:"sort" validate:"min=0,max=9999" label:"排序"`
}

// CreateLegalSceneReq 新建场景。
type CreateLegalSceneReq struct {
	Key   string              `json:"key" validate:"required,max=64" label:"标识"`
	Name  string              `json:"name" validate:"required,max=128" label:"名称"`
	Kind  string              `json:"kind" validate:"required,oneof=consent reader" label:"用途"`
	Items []LegalSceneItemReq `json:"items" validate:"dive" label:"条款"`
}

// UpdateLegalSceneReq 修改场景。标识不能改。
type UpdateLegalSceneReq struct {
	ID    uint                `json:"id,string"`
	Name  string              `json:"name" validate:"required,max=128" label:"名称"`
	Kind  string              `json:"kind" validate:"required,oneof=consent reader" label:"用途"`
	Items []LegalSceneItemReq `json:"items" validate:"dive" label:"条款"`
}

// RetrieveLegalSceneReq 分页查询场景。
type RetrieveLegalSceneReq struct {
	router.PageReq
	Name string `json:"name" form:"name" validate:"omitempty,max=128" label:"名称"`
	Kind string `json:"kind" form:"kind" validate:"omitempty,oneof=consent reader" label:"用途"`
}

// LegalSceneItemView 是场景条目的展示。
type LegalSceneItemView struct {
	DocumentKey  string `json:"documentKey"`
	DocumentName string `json:"documentName"`
	Required     bool   `json:"required"`
	GroupKey     string `json:"groupKey"`
	Wording      string `json:"wording"`
	Sort         int    `json:"sort"`
}

// LegalSceneItemAdmin 是管理端的一条场景。
type LegalSceneItemAdmin struct {
	ID          string               `json:"id"`
	Key         string               `json:"key"`
	Name        string               `json:"name"`
	Kind        string               `json:"kind"`
	KindLabel   string               `json:"kindLabel"`
	ItemSummary string               `json:"itemSummary"`
	Items       []LegalSceneItemView `json:"items"`
}

// LegalSceneDocument 是客户端阅读或勾选时看到的一篇条款。
type LegalSceneDocument struct {
	Key      string `json:"key"`
	Name     string `json:"name"`
	Title    string `json:"title"`
	Version  string `json:"version"`
	Kind     string `json:"kind"`
	Sort     int    `json:"sort"`
	Required bool   `json:"required,omitempty"`
	Wording  string `json:"wording,omitempty"`
}

// LegalSceneGroup 是同意场景里的一个勾选组。
type LegalSceneGroup struct {
	GroupKey  string               `json:"groupKey"`
	Required  bool                 `json:"required"`
	Wording   string               `json:"wording"`
	Documents []LegalSceneDocument `json:"documents"`
}

// LegalSceneView 是手机和 PC 共用的场景结果。
type LegalSceneView struct {
	Key       string               `json:"key"`
	Name      string               `json:"name"`
	Kind      string               `json:"kind"`
	Documents []LegalSceneDocument `json:"documents,omitempty"`
	Groups    []LegalSceneGroup    `json:"groups,omitempty"`
}

// LegalDocumentBody 是一篇已发布或已归档的正文。
type LegalDocumentBody struct {
	Key           string     `json:"key"`
	Name          string     `json:"name"`
	Kind          string     `json:"kind"`
	Title         string     `json:"title"`
	Version       string     `json:"version"`
	PublishedAt   *time.Time `json:"publishedAt,omitempty"`
	Content       string     `json:"content"`
	ContentSHA256 string     `json:"contentSha256"`
}

// LegalVersionRef 是调用方声称用户勾中的一版。
type LegalVersionRef struct {
	DocumentKey string
	Version     string
}

// LegalActor 是写下同意记录的当前用户。
type LegalActor struct {
	UserID    uint
	Username  string
	Nickname  string
	IP        string
	UserAgent string
}

// RecordAcceptanceReq 提交一组同意、拒绝或撤回。
type RecordAcceptanceReq struct {
	SceneKey string                 `json:"sceneKey" validate:"required,max=64" label:"场景"`
	Items    []RecordAcceptanceItem `json:"items" validate:"required,min=1,dive" label:"记录"`
}

// RecordAcceptanceItem 是一次动作里的一篇条款。
type RecordAcceptanceItem struct {
	DocumentKey string `json:"documentKey" validate:"required,max=64" label:"条款"`
	Version     string `json:"version" validate:"required,max=32" label:"版本号"`
	Action      string `json:"action" validate:"required,oneof=accept decline withdraw" label:"动作"`
}

// LegalPending 是当前用户还要补签的实质变更。
type LegalPending struct {
	Key     string `json:"key"`
	Name    string `json:"name"`
	Title   string `json:"title"`
	Version string `json:"version"`
	Kind    string `json:"kind"`
}

// LegalStatus 是当前用户相对已发布条款的补签情况。
type LegalStatus struct {
	Pending []LegalPending `json:"pending"`
}

// RetrieveLegalEventReq 分页查询同意记录。
type RetrieveLegalEventReq struct {
	router.PageReq
	DocumentKey string `json:"documentKey" form:"documentKey" validate:"omitempty,max=64" label:"条款"`
	SceneKey    string `json:"sceneKey" form:"sceneKey" validate:"omitempty,max=64" label:"场景"`
	Action      string `json:"action" form:"action" validate:"omitempty,oneof=accept decline withdraw" label:"动作"`
}

// LegalEventItem 是管理端看到的一条同意记录。
type LegalEventItem struct {
	ID           string     `json:"id"`
	UserLabel    string     `json:"userLabel"`
	Username     string     `json:"username"`
	Nickname     string     `json:"nickname"`
	UserPresent  bool       `json:"userPresent"`
	DocumentName string     `json:"documentName"`
	DocumentKey  string     `json:"documentKey"`
	Version      string     `json:"version"`
	Action       string     `json:"action"`
	ActionLabel  string     `json:"actionLabel"`
	SceneKey     string     `json:"sceneKey"`
	SceneName    string     `json:"sceneName"`
	CreatedAt    *time.Time `json:"createdAt,omitempty"`
	IP           string     `json:"ip"`
	UserAgent    string     `json:"userAgent"`
}

// CreateScene 新建场景并写入条目。
func (i *LegalService) CreateScene(ctx context.Context, req *CreateLegalSceneReq) error {
	if err := i.validate.Struct(req); err != nil {
		return err
	}
	key, err := normalizeLegalKey(req.Key)
	if err != nil {
		return err
	}
	unlock, err := i.lock("legal:scene:" + key)
	if err != nil {
		return err
	}
	defer unlock()

	repo := i.repo.WithContext(ctx)
	existing, err := repo.FindSceneByKey(key)
	if err != nil {
		return err
	}
	if existing != nil {
		return LegalSceneExists()
	}
	items, err := i.buildSceneItems(repo, 0, req.Kind, req.Items)
	if err != nil {
		return err
	}
	scene := &model.LegalScene{Key: key, Name: strings.TrimSpace(req.Name), Kind: req.Kind}
	return repo.Transaction(func(tx repository.LegalRepositoryInterface) error {
		if err := tx.SaveScene(scene); err != nil {
			return err
		}
		for n := range items {
			items[n].SceneID = scene.ID
		}
		return tx.ReplaceSceneItems(scene.ID, items)
	})
}

// UpdateScene 修改场景名称、用途和条目。
func (i *LegalService) UpdateScene(ctx context.Context, req *UpdateLegalSceneReq) error {
	if err := i.validate.Struct(req); err != nil {
		return err
	}
	unlock, err := i.lock("legal:scene-id:" + cast.ToString(req.ID))
	if err != nil {
		return err
	}
	defer unlock()

	repo := i.repo.WithContext(ctx)
	scene, err := repo.FindSceneByID(req.ID)
	if err != nil {
		return err
	}
	if scene == nil {
		return LegalSceneNotFound()
	}
	items, err := i.buildSceneItems(repo, scene.ID, req.Kind, req.Items)
	if err != nil {
		return err
	}
	scene.Name = strings.TrimSpace(req.Name)
	scene.Kind = req.Kind
	return repo.Transaction(func(tx repository.LegalRepositoryInterface) error {
		if err := tx.SaveScene(scene); err != nil {
			return err
		}
		return tx.ReplaceSceneItems(scene.ID, items)
	})
}

// DeleteScene 删除场景及其条目。已经写下的同意记录保留。
func (i *LegalService) DeleteScene(ctx context.Context, id uint) error {
	unlock, err := i.lock("legal:scene-id:" + cast.ToString(id))
	if err != nil {
		return err
	}
	defer unlock()

	repo := i.repo.WithContext(ctx)
	scene, err := repo.FindSceneByID(id)
	if err != nil {
		return err
	}
	if scene == nil {
		return LegalSceneNotFound()
	}
	return repo.DeleteScene(scene.ID)
}

// RetrieveScenes 分页返回场景和它的条目。
func (i *LegalService) RetrieveScenes(ctx context.Context, req *RetrieveLegalSceneReq) (int, []LegalSceneItemAdmin, error) {
	if err := i.validate.Struct(req); err != nil {
		return 0, nil, err
	}
	repo := i.repo.WithContext(ctx)
	count, scenes, err := repo.ListScenes(req.Page, req.PageSize, strings.TrimSpace(req.Name), req.Kind)
	if err != nil {
		return 0, nil, err
	}
	out, err := i.decorateScenes(repo, scenes)
	if err != nil {
		return 0, nil, err
	}
	return int(count), out, nil
}

// SceneOptions 返回全部场景，供筛选。
func (i *LegalService) SceneOptions(ctx context.Context) ([]LegalOption, error) {
	scenes, err := i.repo.WithContext(ctx).ListSceneOptions()
	if err != nil {
		return nil, err
	}
	out := make([]LegalOption, 0, len(scenes))
	for _, scene := range scenes {
		out = append(out, LegalOption{Key: scene.Key, Name: scene.Name, Kind: scene.Kind})
	}
	return out, nil
}

func (i *LegalService) buildSceneItems(repo repository.LegalRepositoryInterface, sceneID uint, kind string, reqs []LegalSceneItemReq) ([]model.LegalSceneItem, error) {
	items := make([]model.LegalSceneItem, 0, len(reqs))
	seenDoc := map[string]bool{}
	type groupState struct {
		required bool
		wording  string
	}
	groups := map[string]groupState{}
	for _, req := range reqs {
		key, err := normalizeLegalKey(req.DocumentKey)
		if err != nil {
			return nil, err
		}
		if seenDoc[key] {
			return nil, LegalSceneItemInvalid("同一场景里条款不能重复")
		}
		seenDoc[key] = true
		doc, err := repo.FindDocumentByKey(key)
		if err != nil {
			return nil, err
		}
		if doc == nil {
			return nil, LegalSceneItemInvalid("条款不存在：" + key)
		}
		if doc.Status != model.LegalStatusEnabled {
			return nil, LegalSceneItemInvalid("条款未启用：" + doc.Name)
		}
		item := model.LegalSceneItem{
			SceneID:     sceneID,
			DocumentID:  doc.ID,
			DocumentKey: doc.Key,
			Sort:        req.Sort,
		}
		if kind == model.LegalSceneReader {
			if req.Required || strings.TrimSpace(req.GroupKey) != "" || strings.TrimSpace(req.Wording) != "" {
				return nil, LegalSceneItemInvalid("阅读场景只排条款顺序")
			}
		} else {
			group := strings.TrimSpace(req.GroupKey)
			if group == "" {
				return nil, LegalSceneItemInvalid("同意场景要填写勾选组")
			}
			wording := strings.TrimSpace(req.Wording)
			if prev, ok := groups[group]; ok {
				if prev.required != req.Required || prev.wording != wording {
					return nil, LegalSceneItemInvalid("同一勾选组的必选和文案要一致")
				}
			} else {
				groups[group] = groupState{required: req.Required, wording: wording}
			}
			item.Required = req.Required
			item.GroupKey = group
			item.Wording = wording
		}
		items = append(items, item)
	}
	return items, nil
}

func (i *LegalService) decorateScenes(repo repository.LegalRepositoryInterface, scenes []model.LegalScene) ([]LegalSceneItemAdmin, error) {
	ids := make([]uint, 0, len(scenes))
	for _, scene := range scenes {
		ids = append(ids, scene.ID)
	}
	links, err := repo.ListItemsBySceneIDs(ids)
	if err != nil {
		return nil, err
	}
	docIDs := make([]uint, 0, len(links))
	for _, link := range links {
		docIDs = append(docIDs, link.DocumentID)
	}
	docs := map[uint]model.LegalDocument{}
	if len(docIDs) > 0 {
		options, err := repo.ListDocumentOptions()
		if err != nil {
			return nil, err
		}
		for _, doc := range options {
			docs[doc.ID] = doc
		}
		// 停用的条款不会出现在选项里，再按主键补一次。
		for _, id := range docIDs {
			if _, ok := docs[id]; ok {
				continue
			}
			doc, err := repo.FindDocumentByID(id)
			if err != nil {
				return nil, err
			}
			if doc != nil {
				docs[doc.ID] = *doc
			}
		}
	}
	byScene := map[uint][]LegalSceneItemView{}
	for _, link := range links {
		name := link.DocumentKey
		if doc, ok := docs[link.DocumentID]; ok && doc.Name != "" {
			name = doc.Name
		} else if name != "" {
			name = "未知条款(" + link.DocumentKey + ")"
		}
		byScene[link.SceneID] = append(byScene[link.SceneID], LegalSceneItemView{
			DocumentKey:  link.DocumentKey,
			DocumentName: name,
			Required:     link.Required,
			GroupKey:     link.GroupKey,
			Wording:      link.Wording,
			Sort:         link.Sort,
		})
	}
	out := make([]LegalSceneItemAdmin, 0, len(scenes))
	for _, scene := range scenes {
		views := byScene[scene.ID]
		if views == nil {
			views = []LegalSceneItemView{}
		}
		names := make([]string, 0, len(views))
		for _, view := range views {
			names = append(names, view.DocumentName)
		}
		out = append(out, LegalSceneItemAdmin{
			ID:          cast.ToString(scene.ID),
			Key:         scene.Key,
			Name:        scene.Name,
			Kind:        scene.Kind,
			KindLabel:   model.LegalSceneKindLabel(scene.Kind),
			ItemSummary: strings.Join(names, "、"),
			Items:       views,
		})
	}
	return out, nil
}

// PublicScene 返回一个场景里当前可阅读的条款。手机和 PC 用同一份结果。
func (i *LegalService) PublicScene(ctx context.Context, key string) (*LegalSceneView, error) {
	key, err := normalizeLegalKey(key)
	if err != nil {
		return nil, err
	}
	repo := i.repo.WithContext(ctx)
	scene, err := repo.FindSceneByKey(key)
	if err != nil {
		return nil, err
	}
	if scene == nil {
		return nil, LegalSceneNotFound()
	}
	items, err := repo.ListSceneItems(scene.ID)
	if err != nil {
		return nil, err
	}
	view := &LegalSceneView{Key: scene.Key, Name: scene.Name, Kind: scene.Kind}
	if scene.Kind == model.LegalSceneReader {
		view.Documents = []LegalSceneDocument{}
	} else {
		view.Groups = []LegalSceneGroup{}
	}
	groupIndex := map[string]int{}
	for _, item := range items {
		doc, err := repo.FindDocumentByID(item.DocumentID)
		if err != nil {
			return nil, err
		}
		if doc == nil || doc.Status != model.LegalStatusEnabled {
			continue
		}
		published, err := repo.FindPublishedVersion(doc.ID)
		if err != nil {
			return nil, err
		}
		if published == nil {
			continue
		}
		entry := LegalSceneDocument{
			Key:     doc.Key,
			Name:    doc.Name,
			Title:   published.Title,
			Version: published.Version,
			Kind:    doc.Kind,
			Sort:    item.Sort,
		}
		if scene.Kind == model.LegalSceneReader {
			view.Documents = append(view.Documents, entry)
			continue
		}
		entry.Required = item.Required
		entry.Wording = item.Wording
		idx, ok := groupIndex[item.GroupKey]
		if !ok {
			idx = len(view.Groups)
			groupIndex[item.GroupKey] = idx
			view.Groups = append(view.Groups, LegalSceneGroup{
				GroupKey:  item.GroupKey,
				Required:  item.Required,
				Wording:   item.Wording,
				Documents: []LegalSceneDocument{},
			})
		}
		view.Groups[idx].Documents = append(view.Groups[idx].Documents, entry)
	}
	return view, nil
}

// PublicDocument 返回条款当前已发布的正文。停用或尚未发布时不返回。
func (i *LegalService) PublicDocument(ctx context.Context, key string) (*LegalDocumentBody, error) {
	key, err := normalizeLegalKey(key)
	if err != nil {
		return nil, err
	}
	repo := i.repo.WithContext(ctx)
	doc, err := repo.FindDocumentByKey(key)
	if err != nil {
		return nil, err
	}
	if doc == nil || doc.Status != model.LegalStatusEnabled {
		return nil, LegalDocumentNotFound()
	}
	published, err := repo.FindPublishedVersion(doc.ID)
	if err != nil {
		return nil, err
	}
	if published == nil {
		return nil, LegalDocumentNotFound()
	}
	return toDocumentBody(doc, published), nil
}

// PublicDocumentVersion 返回某一已发布或已归档版本，供核对当时同意的正文。
func (i *LegalService) PublicDocumentVersion(ctx context.Context, key, version string) (*LegalDocumentBody, error) {
	key, err := normalizeLegalKey(key)
	if err != nil {
		return nil, err
	}
	version, err = normalizeVersionNumber(version)
	if err != nil {
		return nil, err
	}
	repo := i.repo.WithContext(ctx)
	doc, err := repo.FindDocumentByKey(key)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, LegalDocumentNotFound()
	}
	ver, err := repo.FindVersion(doc.ID, version)
	if err != nil {
		return nil, err
	}
	if ver == nil || ver.Status == model.LegalVersionDraft {
		return nil, LegalVersionNotFound()
	}
	return toDocumentBody(doc, ver), nil
}

func toDocumentBody(doc *model.LegalDocument, ver *model.LegalVersion) *LegalDocumentBody {
	return &LegalDocumentBody{
		Key:           doc.Key,
		Name:          doc.Name,
		Kind:          doc.Kind,
		Title:         ver.Title,
		Version:       ver.Version,
		PublishedAt:   ver.PublishedAt,
		Content:       ver.Content,
		ContentSHA256: ver.ContentSHA256,
	}
}

// SceneRequirementsMet 检查同意场景的必选条款是否都出现在这次提交里，且版本是当前已发布版。
// 阅读场景没有必选项。注册接入时调用它，本轮注册接口不调用。
func (i *LegalService) SceneRequirementsMet(ctx context.Context, sceneKey string, accepted []LegalVersionRef) error {
	sceneKey, err := normalizeLegalKey(sceneKey)
	if err != nil {
		return err
	}
	repo := i.repo.WithContext(ctx)
	scene, err := repo.FindSceneByKey(sceneKey)
	if err != nil {
		return err
	}
	if scene == nil {
		return LegalSceneNotFound()
	}
	if scene.Kind != model.LegalSceneConsent {
		return nil
	}
	got := map[string]bool{}
	for _, ref := range accepted {
		key, err := normalizeLegalKey(ref.DocumentKey)
		if err != nil {
			return err
		}
		version, err := normalizeVersionNumber(ref.Version)
		if err != nil {
			return err
		}
		got[key+"\n"+version] = true
	}
	items, err := repo.ListSceneItems(scene.ID)
	if err != nil {
		return err
	}
	for _, item := range items {
		if !item.Required {
			continue
		}
		doc, err := repo.FindDocumentByID(item.DocumentID)
		if err != nil {
			return err
		}
		if doc == nil || doc.Status != model.LegalStatusEnabled {
			name := item.DocumentKey
			if doc != nil && doc.Name != "" {
				name = doc.Name
			}
			return LegalAcceptanceIncomplete(name)
		}
		published, err := repo.FindPublishedVersion(doc.ID)
		if err != nil {
			return err
		}
		if published == nil || !got[doc.Key+"\n"+published.Version] {
			return LegalAcceptanceIncomplete(doc.Name)
		}
	}
	return nil
}

// RecordAcceptances 为当前用户追加同意、拒绝或撤回。同一版本的同一动作再写一次，返回已有记录。
func (i *LegalService) RecordAcceptances(ctx context.Context, actor LegalActor, req *RecordAcceptanceReq) ([]LegalEventItem, error) {
	if err := i.validate.Struct(req); err != nil {
		return nil, err
	}
	sceneKey, err := normalizeLegalKey(req.SceneKey)
	if err != nil {
		return nil, err
	}
	unlock, err := i.lock("legal:accept:" + cast.ToString(actor.UserID))
	if err != nil {
		return nil, err
	}
	defer unlock()

	repo := i.repo.WithContext(ctx)
	scene, err := repo.FindSceneByKey(sceneKey)
	if err != nil {
		return nil, err
	}
	if scene == nil {
		return nil, LegalSceneNotFound()
	}
	events := make([]model.LegalEvent, 0, len(req.Items))
	err = repo.Transaction(func(tx repository.LegalRepositoryInterface) error {
		for _, item := range req.Items {
			event, err := i.prepareEvent(tx, actor, sceneKey, item)
			if err != nil {
				return err
			}
			events = append(events, *event)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	out := make([]LegalEventItem, 0, len(events))
	for _, event := range events {
		item := eventItemFromModel(event, actor)
		item.SceneName = scene.Name
		out = append(out, item)
	}
	return out, nil
}

func (i *LegalService) prepareEvent(repo repository.LegalRepositoryInterface, actor LegalActor, sceneKey string, item RecordAcceptanceItem) (*model.LegalEvent, error) {
	key, err := normalizeLegalKey(item.DocumentKey)
	if err != nil {
		return nil, err
	}
	version, err := normalizeVersionNumber(item.Version)
	if err != nil {
		return nil, err
	}
	doc, err := repo.FindDocumentByKey(key)
	if err != nil {
		return nil, err
	}
	if doc == nil || doc.Status != model.LegalStatusEnabled {
		return nil, LegalDocumentNotFound()
	}
	published, err := repo.FindPublishedVersion(doc.ID)
	if err != nil {
		return nil, err
	}
	if published == nil || published.Version != version {
		return nil, LegalVersionNotCurrent()
	}
	if item.Action == model.LegalActionWithdraw && doc.Kind == model.LegalKindContract {
		return nil, LegalWithdrawDenied()
	}
	existing, err := repo.FindEvent(actor.UserID, published.ID, item.Action)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}
	event := &model.LegalEvent{
		UserID:        actor.UserID,
		Username:      actor.Username,
		Nickname:      actor.Nickname,
		DocumentKey:   doc.Key,
		DocumentName:  doc.Name,
		VersionID:     published.ID,
		Version:       published.Version,
		Title:         published.Title,
		ContentSHA256: published.ContentSHA256,
		Action:        item.Action,
		SceneKey:      sceneKey,
		IP:            actor.IP,
		UserAgent:     actor.UserAgent,
	}
	if err := repo.InsertEvent(event); err != nil {
		return nil, err
	}
	return event, nil
}

func eventItemFromModel(event model.LegalEvent, actor LegalActor) LegalEventItem {
	nickname := event.Nickname
	username := event.Username
	if nickname == "" {
		nickname = actor.Nickname
	}
	if username == "" {
		username = actor.Username
	}
	return LegalEventItem{
		ID:           cast.ToString(event.ID),
		UserLabel:    legalUserLabel(nickname, username),
		Username:     username,
		Nickname:     nickname,
		UserPresent:  true,
		DocumentName: event.DocumentName,
		DocumentKey:  event.DocumentKey,
		Version:      event.Version,
		Action:       event.Action,
		ActionLabel:  model.LegalActionLabel(event.Action),
		SceneKey:     event.SceneKey,
		SceneName:    event.SceneKey,
		CreatedAt:    event.CreatedAt,
		IP:           event.IP,
		UserAgent:    event.UserAgent,
	}
}

// Status 返回当前用户尚未接受的实质变更。错别字变更不列出。
func (i *LegalService) Status(ctx context.Context, userID uint) (*LegalStatus, error) {
	repo := i.repo.WithContext(ctx)
	docs, err := repo.ListDocumentsByStatus(model.LegalStatusEnabled)
	if err != nil {
		return nil, err
	}
	pending := make([]LegalPending, 0)
	for _, doc := range docs {
		published, err := repo.FindPublishedVersion(doc.ID)
		if err != nil {
			return nil, err
		}
		if published == nil || published.ChangeClass != model.LegalChangeMaterial {
			continue
		}
		latest, err := repo.LatestEvent(userID, doc.Key)
		if err != nil {
			return nil, err
		}
		if latest != nil && latest.Action == model.LegalActionAccept && latest.VersionID == published.ID {
			continue
		}
		pending = append(pending, LegalPending{
			Key:     doc.Key,
			Name:    doc.Name,
			Title:   published.Title,
			Version: published.Version,
			Kind:    doc.Kind,
		})
	}
	return &LegalStatus{Pending: pending}, nil
}

// RetrieveEvents 分页返回同意记录。用户显示昵称或账号。
func (i *LegalService) RetrieveEvents(ctx context.Context, req *RetrieveLegalEventReq) (int, []LegalEventItem, error) {
	if err := i.validate.Struct(req); err != nil {
		return 0, nil, err
	}
	count, rows, err := i.repo.WithContext(ctx).ListEvents(req.Page, req.PageSize, strings.TrimSpace(req.DocumentKey), strings.TrimSpace(req.SceneKey), req.Action)
	if err != nil {
		return 0, nil, err
	}
	out := make([]LegalEventItem, 0, len(rows))
	for _, row := range rows {
		nickname := row.Nickname
		username := row.Username
		present := row.LiveUserID != 0
		if present {
			if row.LiveNickname != "" {
				nickname = row.LiveNickname
			}
			if row.LiveUsername != "" {
				username = row.LiveUsername
			}
		}
		sceneName := row.SceneName
		if sceneName == "" && row.SceneKey != "" {
			sceneName = "未知场景(" + row.SceneKey + ")"
		}
		documentName := row.DocumentName
		if documentName == "" {
			documentName = "未知条款(" + row.DocumentKey + ")"
		}
		out = append(out, LegalEventItem{
			ID:           cast.ToString(row.ID),
			UserLabel:    legalUserLabel(nickname, username),
			Username:     username,
			Nickname:     nickname,
			UserPresent:  present,
			DocumentName: documentName,
			DocumentKey:  row.DocumentKey,
			Version:      row.Version,
			Action:       row.Action,
			ActionLabel:  model.LegalActionLabel(row.Action),
			SceneKey:     row.SceneKey,
			SceneName:    sceneName,
			CreatedAt:    row.CreatedAt,
			IP:           row.IP,
			UserAgent:    row.UserAgent,
		})
	}
	return int(count), out, nil
}
