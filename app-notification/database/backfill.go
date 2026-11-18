package database

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type nullFill struct {
	table  string
	column string
	value  string
}

// nullFills 是历史表里可能已经写入空值、模型又要求非空的列。
var nullFills = []nullFill{
	{"notify_inbox", "principal_type", "1"},
	{"notify_inbox", "principal_id", "0"},
	{"notify_inbox", "task_id", "0"},
	{"notify_inbox", "source", "''"},
	{"notify_inbox", "kind", "1"},
	{"notify_inbox", "category", "1"},
	{"notify_inbox", "title", "''"},
	{"notify_inbox", "biz_key", "''"},
	{"notify_inbox", "todo_status", "0"},
	{"notify_inbox", "read_status", "1"},
	{"notify_announcement_state", "principal_type", "1"},
	{"notify_announcement_state", "principal_id", "0"},
	{"notify_announcement_state", "announcement_id", "0"},
	{"notify_subscription", "principal_type", "1"},
	{"notify_subscription", "principal_id", "0"},
	{"notify_subscription", "category", "1"},
	{"notify_subscription", "channel", "''"},
	{"notify_subscription", "enabled", "1"},
	{"notify_device", "principal_type", "1"},
	{"notify_device", "principal_id", "0"},
	{"notify_device", "platform", "''"},
	{"notify_device", "provider", "''"},
	{"notify_device", "token", "''"},
	{"notify_device", "status", "1"},
	{"notify_delivery", "task_id", "0"},
	{"notify_delivery", "recipient_type", "1"},
	{"notify_delivery", "recipient_id", "0"},
	{"notify_delivery", "channel", "''"},
	{"notify_delivery", "status", "1"},
}

// BackfillNulls 在 AutoMigrate 把列改成非空之前，给已有空值补上默认值。
func BackfillNulls(db *gorm.DB) error {
	for _, fill := range nullFills {
		if !db.Migrator().HasTable(fill.table) || !db.Migrator().HasColumn(fill.table, fill.column) {
			continue
		}
		stmt := fmt.Sprintf(
			"UPDATE %s SET %s = %s WHERE %s IS NULL",
			fill.table, fill.column, fill.value, fill.column,
		)
		if err := db.Exec(stmt).Error; err != nil {
			if isMissingRelation(err) {
				continue
			}
			return err
		}
	}
	return nil
}

func isMissingRelation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "does not exist") || strings.Contains(msg, "42P01")
}
