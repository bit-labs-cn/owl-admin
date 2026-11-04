package seeder

import (
	"fmt"
	"testing"

	"bit-labs.cn/owl-admin/app/model"
	"bit-labs.cn/owl/utils"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestInitLegalDataTwice(t *testing.T) {
	if err := utils.InitSnowFlakeWorker(1, 1); err != nil {
		t.Fatalf("init snowflake: %v", err)
	}
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	gdb, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Silent),
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := gdb.AutoMigrate(&model.LegalDocument{}, &model.LegalVersion{}, &model.LegalScene{}, &model.LegalSceneItem{}); err != nil {
		t.Fatal(err)
	}
	InitLegalData(gdb)
	InitLegalData(gdb)

	var docs int64
	if err := gdb.Model(&model.LegalDocument{}).Count(&docs).Error; err != nil {
		t.Fatal(err)
	}
	if docs != 3 {
		t.Fatalf("documents = %d", docs)
	}
	var scenes int64
	if err := gdb.Model(&model.LegalScene{}).Count(&scenes).Error; err != nil {
		t.Fatal(err)
	}
	if scenes != 3 {
		t.Fatalf("scenes = %d", scenes)
	}
	var site model.LegalScene
	if err := gdb.Where("key = ?", "site").First(&site).Error; err != nil {
		t.Fatal(err)
	}
	var keys []string
	if err := gdb.Model(&model.LegalSceneItem{}).Where("scene_id = ?", site.ID).Order("sort asc").Pluck("document_key", &keys).Error; err != nil {
		t.Fatal(err)
	}
	if len(keys) != 2 || keys[0] != "terms" || keys[1] != "privacy" {
		t.Fatalf("site items = %v", keys)
	}
	var register model.LegalScene
	if err := gdb.Where("key = ?", "register").First(&register).Error; err != nil {
		t.Fatal(err)
	}
	var registerKeys []string
	if err := gdb.Model(&model.LegalSceneItem{}).Where("scene_id = ?", register.ID).Order("sort asc").Pluck("document_key", &registerKeys).Error; err != nil {
		t.Fatal(err)
	}
	if len(registerKeys) != 3 || registerKeys[0] != "terms" || registerKeys[2] != "privacy" {
		t.Fatalf("register items = %v", registerKeys)
	}
}
