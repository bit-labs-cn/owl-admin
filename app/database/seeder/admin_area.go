package seeder

import (
	_ "embed"
	"strings"

	"bit-labs.cn/owl-admin/app/model"
	"bit-labs.cn/owl/utils"
	"gorm.io/gorm"
)

//go:embed admin_area.sql
var AdminAreaSQL string

type AdminAreaSeeder struct {
}

// InitAdminAreaData 初始化行政区划数据
func InitAdminAreaData(db *gorm.DB) {
	var count int64
	if err := db.Model(&model.Area{}).Count(&count).Error; err != nil {
		utils.PrintLnRed("查询行政区划数据失败: %v", err)
		return
	}
	if count > 0 {
		utils.PrintLnGreen("行政区划数据已存在，跳过初始化")
		return
	}

	sql := AdminAreaSQL
	statements := strings.Split(sql, ";")

	err := db.Transaction(func(tx *gorm.DB) error {
		for _, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			if err := tx.Exec(stmt).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		utils.PrintLnRed("初始化行政区划数据失败: %v", err)
	} else {
		utils.PrintLnGreen("行政区划数据初始化成功")
	}
}
