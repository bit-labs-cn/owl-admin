package seeder

import (
	"errors"
	"log"

	"bit-labs.cn/owl/provider/db"

	"bit-labs.cn/owl-admin/app/model"
	"gorm.io/gorm"
)

type DictionarySeeder struct {
}

// InitAllDictData 初始化所有字典数据（包括字典主表和字典项）
func InitAllDictData(db *gorm.DB) {
	// 使用事务确保数据一致性
	err := db.Transaction(func(tx *gorm.DB) error {
		// 先初始化字典主表数据
		if err := initDictDataWithTx(tx); err != nil {
			log.Printf("初始化字典主表数据失败: %v", err)
			return err
		}

		if err := initDictItemDataWithTx(tx); err != nil {
			log.Printf("初始化字典项数据失败: %v", err)
			return err
		}

		return nil
	})

	if err != nil {
		log.Printf("字典数据初始化失败: %v", err)
	}
}

// initDictDataWithTx 在事务中初始化字典主表数据
func initDictDataWithTx(tx *gorm.DB) error {
	var dicts = []model.Dict{
		{
			BaseModel: db.BaseModel{ID: 1},
			Name:      "性别",
			Type:      "gender",
			Status:    1, // 启用状态
			Desc:      "用户性别分类",
			Sort:      1,
		},
		{
			BaseModel: db.BaseModel{ID: 2},
			Name:      "用户状态",
			Type:      "user_status",
			Status:    1,
			Desc:      "用户账号状态",
			Sort:      2,
		},
		{
			BaseModel: db.BaseModel{ID: 3},
			Name:      "在线状态",
			Type:      "online_status",
			Status:    1,
			Desc:      "用户在线状态",
			Sort:      3,
		},
	}

	// 逐个检查并创建字典数据，避免重复插入
	for _, dict := range dicts {
		var existingDict model.Dict
		result := tx.Where("type = ?", dict.Type).First(&existingDict)
		if result.Error != nil {
			// 如果不存在，则创建新记录
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				if err := tx.Model(&model.Dict{}).Create(&dict).Error; err != nil {
					return err
				}
				log.Printf("创建字典: %s (%s)", dict.Name, dict.Type)
			} else {
				return result.Error
			}
		}
	}
	return nil
}

// initDictItemDataWithTx 在事务中初始化字典项数据
func initDictItemDataWithTx(tx *gorm.DB) error {
	var dictItems = []model.DictItem{
		// 性别字典项 (DictID: 1)
		{
			Label:    "男",
			Value:    "1",
			Extend:   "",
			Status:   1,
			Sort:     1,
			DictType: "gender",
			DictID:   1,
		},
		{
			Label:    "女",
			Value:    "2",
			Extend:   "",
			Status:   1,
			Sort:     2,
			DictType: "gender",
			DictID:   1,
		},
		{
			Label:    "未知",
			Value:    "3",
			Extend:   "",
			Status:   1,
			Sort:     3,
			DictType: "gender",
			DictID:   1,
		},

		// 用户状态字典项 (DictID: 2)
		{
			Label:    "正常",
			Value:    "1",
			Extend:   "success",
			Status:   1,
			Sort:     1,
			DictType: "user_status",
			DictID:   2,
		},
		{
			Label:    "禁用",
			Value:    "2",
			Extend:   "danger",
			Status:   1,
			Sort:     2,
			DictType: "user_status",
			DictID:   2,
		},
		{
			Label:    "待审核",
			Value:    "3",
			Extend:   "warning",
			Status:   1,
			Sort:     3,
			DictType: "user_status",
			DictID:   2,
		},

		// 在线状态字典项 (DictID: 3)
		{
			Label:    "在线",
			Value:    "1",
			Extend:   "success",
			Status:   1,
			Sort:     1,
			DictType: "online_status",
			DictID:   3,
		},
		{
			Label:    "离线",
			Value:    "2",
			Extend:   "info",
			Status:   1,
			Sort:     2,
			DictType: "online_status",
			DictID:   3,
		},
		{
			Label:    "忙碌",
			Value:    "3",
			Extend:   "warning",
			Status:   1,
			Sort:     3,
			DictType: "online_status",
			DictID:   3,
		},
		{
			Label:    "隐身",
			Value:    "4",
			Extend:   "default",
			Status:   1,
			Sort:     4,
			DictType: "online_status",
			DictID:   3,
		},
	}

	// 逐个检查并创建字典项数据，避免重复插入
	for _, item := range dictItems {
		var existingItem model.DictItem
		result := tx.Where("dict_type = ? AND value = ?", item.DictType, item.Value).First(&existingItem)
		if result.Error != nil {
			// 如果不存在，则创建新记录
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				if err := tx.Model(&model.DictItem{}).Create(&item).Error; err != nil {
					return err
				}
				log.Printf("创建字典项: %s - %s (%s)", item.DictType, item.Label, item.Value)
			} else {
				return result.Error
			}
		}
	}
	return nil
}
