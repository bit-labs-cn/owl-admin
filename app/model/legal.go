package model

import (
	"time"

	"bit-labs.cn/owl/provider/db"
)

// 条款性质。合同必须同意才能使用对应服务，不能撤回。告知是隐私政策这类说明。可选同意不挡住注册，可以撤回。
const (
	LegalKindContract = "contract"
	LegalKindNotice   = "notice"
	LegalKindOptional = "optional"
)

// 条款是否还在对外使用。
const (
	LegalStatusEnabled  = "enabled"
	LegalStatusDisabled = "disabled"
)

// 版本状态。草稿可改，发布后冻结，被新版本替换后归档。
const (
	LegalVersionDraft     = "draft"
	LegalVersionPublished = "published"
	LegalVersionArchived  = "archived"
)

// 发布时的变更级别。错别字不要求老用户重签，实质变更要重签。
const (
	LegalChangeEditorial = "editorial"
	LegalChangeMaterial  = "material"
)

// 场景用途。同意场景用来勾选，阅读场景用来排顶栏。
const (
	LegalSceneConsent = "consent"
	LegalSceneReader  = "reader"
)

// 同意动作。撤回是新的一行，不改原来的同意。
const (
	LegalActionAccept   = "accept"
	LegalActionDecline  = "decline"
	LegalActionWithdraw = "withdraw"
)

// LegalKindLabel 返回条款性质的显示名。
func LegalKindLabel(kind string) string {
	switch kind {
	case LegalKindContract:
		return "合同"
	case LegalKindNotice:
		return "告知"
	case LegalKindOptional:
		return "可选同意"
	default:
		return kind
	}
}

// LegalStatusLabel 返回启用状态的显示名。
func LegalStatusLabel(status string) string {
	switch status {
	case LegalStatusEnabled:
		return "启用"
	case LegalStatusDisabled:
		return "停用"
	default:
		return status
	}
}

// LegalVersionStatusLabel 返回版本状态的显示名。
func LegalVersionStatusLabel(status string) string {
	switch status {
	case LegalVersionDraft:
		return "草稿"
	case LegalVersionPublished:
		return "已发布"
	case LegalVersionArchived:
		return "已归档"
	default:
		return status
	}
}

// LegalChangeLabel 返回变更级别的显示名。
func LegalChangeLabel(change string) string {
	switch change {
	case LegalChangeEditorial:
		return "错别字"
	case LegalChangeMaterial:
		return "实质变更"
	default:
		return ""
	}
}

// LegalSceneKindLabel 返回场景用途的显示名。
func LegalSceneKindLabel(kind string) string {
	switch kind {
	case LegalSceneConsent:
		return "同意"
	case LegalSceneReader:
		return "阅读"
	default:
		return kind
	}
}

// LegalActionLabel 返回同意动作的显示名。
func LegalActionLabel(action string) string {
	switch action {
	case LegalActionAccept:
		return "同意"
	case LegalActionDecline:
		return "拒绝"
	case LegalActionWithdraw:
		return "撤回"
	default:
		return action
	}
}

// LegalDocument 是一条稳定的条款，正文在版本上。
type LegalDocument struct {
	db.BaseModel
	Key    string `gorm:"size:64;uniqueIndex;comment:稳定标识" json:"key"`
	Name   string `gorm:"size:128;comment:显示名" json:"name"`
	Kind   string `gorm:"size:32;comment:性质" json:"kind"`
	Status string `gorm:"size:32;comment:启用状态" json:"status"`
}

func (LegalDocument) TableName() string {
	return "admin_legal_document"
}

// LegalVersion 是条款的一版正文。发布后不可改。
type LegalVersion struct {
	db.BaseModel
	DocumentID    uint       `gorm:"uniqueIndex:uk_legal_version,priority:1;index;comment:条款" json:"documentId,string"`
	Version       string     `gorm:"size:32;uniqueIndex:uk_legal_version,priority:2;comment:版本号" json:"version"`
	Title         string     `gorm:"size:128;comment:标题" json:"title"`
	Content       string     `gorm:"type:text;comment:正文 Markdown" json:"content"`
	ContentSHA256 string     `gorm:"size:64;comment:正文哈希" json:"contentSha256"`
	ChangeClass   string     `gorm:"size:32;comment:变更级别" json:"changeClass"`
	Status        string     `gorm:"size:32;index;comment:版本状态" json:"status"`
	EffectiveAt   *time.Time `gorm:"comment:生效时间" json:"effectiveAt,omitempty"`
	PublishedAt   *time.Time `gorm:"comment:发布时间" json:"publishedAt,omitempty"`
}

func (LegalVersion) TableName() string {
	return "admin_legal_version"
}

// LegalScene 是一处出示条款的入口，例如注册或网站顶栏。
type LegalScene struct {
	db.BaseModel
	Key  string `gorm:"size:64;uniqueIndex;comment:场景标识" json:"key"`
	Name string `gorm:"size:128;comment:场景名称" json:"name"`
	Kind string `gorm:"size:32;comment:用途" json:"kind"`
}

func (LegalScene) TableName() string {
	return "admin_legal_scene"
}

// LegalSceneItem 是场景里的一条条款，顺序和勾选组都在这里。
type LegalSceneItem struct {
	db.BaseModel
	SceneID     uint   `gorm:"index;comment:场景" json:"sceneId,string"`
	DocumentID  uint   `gorm:"index;comment:条款" json:"documentId,string"`
	DocumentKey string `gorm:"size:64;comment:条款标识" json:"documentKey"`
	Required    bool   `gorm:"comment:是否必选" json:"required"`
	GroupKey    string `gorm:"size:64;comment:勾选组" json:"groupKey"`
	Wording     string `gorm:"size:128;comment:勾选文案" json:"wording"`
	Sort        int    `gorm:"comment:排序" json:"sort"`
}

func (LegalSceneItem) TableName() string {
	return "admin_legal_scene_item"
}

// LegalEvent 是一条只追加的同意记录。
type LegalEvent struct {
	db.BaseModel
	UserID        uint   `gorm:"uniqueIndex:uk_legal_event_action,priority:1;index:idx_legal_event_user_doc,priority:1;comment:用户" json:"userId,string"`
	Username      string `gorm:"size:128;comment:账号快照" json:"username"`
	Nickname      string `gorm:"size:128;comment:昵称快照" json:"nickname"`
	DocumentKey   string `gorm:"size:64;index:idx_legal_event_user_doc,priority:2;comment:条款标识" json:"documentKey"`
	DocumentName  string `gorm:"size:128;comment:条款名称快照" json:"documentName"`
	VersionID     uint   `gorm:"uniqueIndex:uk_legal_event_action,priority:2;comment:版本" json:"versionId,string"`
	Version       string `gorm:"size:32;comment:版本号快照" json:"version"`
	Title         string `gorm:"size:128;comment:标题快照" json:"title"`
	ContentSHA256 string `gorm:"size:64;comment:正文哈希快照" json:"contentSha256"`
	Action        string `gorm:"size:32;uniqueIndex:uk_legal_event_action,priority:3;comment:动作" json:"action"`
	SceneKey      string `gorm:"size:64;index;comment:场景标识" json:"sceneKey"`
	IP            string `gorm:"size:64;comment:IP" json:"ip"`
	UserAgent     string `gorm:"size:512;comment:User-Agent" json:"userAgent"`
}

func (LegalEvent) TableName() string {
	return "admin_legal_event"
}
