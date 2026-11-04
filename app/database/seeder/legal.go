package seeder

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"
	"time"

	"bit-labs.cn/owl-admin/app/model"
	"gorm.io/gorm"
)

const (
	legalTermsContent = `# 使用条款

使用本服务前请阅读本条款。你注册或继续使用时，接受的是当时公布的这一版。

## 账号

请提供真实的注册信息，并保管好自己的登录方式。账号只供你本人使用。
`

	legalPrivacyContent = `# 隐私政策

我们只在提供账号服务所必需的范围内处理你的个人信息，例如手机号、邮箱和登录记录。

你可以查阅、更正这些信息，也可以撤回对本政策的同意。撤回之后，我们停止此后基于同意的处理。撤回前已经完成的处理仍然有效。
`

	legalInactiveContent = `# 非活跃账号处理规范

长期没有任何登录或操作、且名下没有需要继续保留的内容时，账号可能被标为非活跃。

标为非活跃之前会尽量通知你。非活跃账号占用的登录名可能被收回。
`
)

// InitLegalData 写入三条初始条款，以及注册、快速创建和网站阅读场景。已有的不覆盖。
func InitLegalData(gdb *gorm.DB) {
	err := gdb.Transaction(func(tx *gorm.DB) error {
		terms, err := ensureLegalDocument(tx, "terms", "使用条款", model.LegalKindContract, legalTermsContent)
		if err != nil {
			return err
		}
		inactive, err := ensureLegalDocument(tx, "inactive-account", "非活跃账号处理规范", model.LegalKindContract, legalInactiveContent)
		if err != nil {
			return err
		}
		privacy, err := ensureLegalDocument(tx, "privacy", "隐私政策", model.LegalKindNotice, legalPrivacyContent)
		if err != nil {
			return err
		}
		account := []model.LegalSceneItem{
			{DocumentID: terms.ID, DocumentKey: terms.Key, Required: true, GroupKey: "account", Wording: "我已阅读并同意", Sort: 1},
			{DocumentID: inactive.ID, DocumentKey: inactive.Key, Required: true, GroupKey: "account", Wording: "我已阅读并同意", Sort: 2},
			{DocumentID: privacy.ID, DocumentKey: privacy.Key, Required: true, GroupKey: "privacy", Wording: "我已阅读", Sort: 3},
		}
		if err := ensureLegalScene(tx, "register", "注册", model.LegalSceneConsent, account); err != nil {
			return err
		}
		if err := ensureLegalScene(tx, "quick-create", "快速创建", model.LegalSceneConsent, account); err != nil {
			return err
		}
		reader := []model.LegalSceneItem{
			{DocumentID: terms.ID, DocumentKey: terms.Key, Sort: 1},
			{DocumentID: privacy.ID, DocumentKey: privacy.Key, Sort: 2},
		}
		return ensureLegalScene(tx, "site", "网站条款", model.LegalSceneReader, reader)
	})
	if err != nil {
		log.Printf("条款数据初始化失败: %v", err)
	}
}

func ensureLegalDocument(tx *gorm.DB, key, name, kind, content string) (*model.LegalDocument, error) {
	var doc model.LegalDocument
	err := tx.Where("key = ?", key).First(&doc).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		doc = model.LegalDocument{Key: key, Name: name, Kind: kind, Status: model.LegalStatusEnabled}
		if err := tx.Create(&doc).Error; err != nil {
			return nil, err
		}
		log.Printf("创建条款: %s", name)
	} else if err != nil {
		return nil, err
	}
	var ver model.LegalVersion
	err = tx.Where("document_id = ? AND version = ?", doc.ID, "1.0.0").First(&ver).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		now := time.Now()
		sum := sha256.Sum256([]byte(content))
		ver = model.LegalVersion{
			DocumentID:    doc.ID,
			Version:       "1.0.0",
			Title:         name,
			Content:       content,
			ContentSHA256: hex.EncodeToString(sum[:]),
			ChangeClass:   model.LegalChangeMaterial,
			Status:        model.LegalVersionPublished,
			EffectiveAt:   &now,
			PublishedAt:   &now,
		}
		if err := tx.Create(&ver).Error; err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	return &doc, nil
}

func ensureLegalScene(tx *gorm.DB, key, name, kind string, items []model.LegalSceneItem) error {
	var scene model.LegalScene
	err := tx.Where("key = ?", key).First(&scene).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	scene = model.LegalScene{Key: key, Name: name, Kind: kind}
	if err := tx.Create(&scene).Error; err != nil {
		return err
	}
	for n := range items {
		items[n].SceneID = scene.ID
		items[n].ID = 0
	}
	if len(items) == 0 {
		return nil
	}
	return tx.Create(&items).Error
}
