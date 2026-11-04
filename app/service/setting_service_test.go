package service

import (
	"context"
	"fmt"
	"testing"

	"bit-labs.cn/owl-admin/app/model"
	"bit-labs.cn/owl-admin/app/repository"
	"bit-labs.cn/owl/provider/setting"
	"bit-labs.cn/owl/utils"
	"github.com/glebarez/sqlite"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupSettingDB(t *testing.T) *gorm.DB {
	t.Helper()
	if err := utils.InitSnowFlakeWorker(1, 1); err != nil {
		t.Fatalf("init snowflake: %v", err)
	}
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	gdb, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Silent),
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := gdb.AutoMigrate(&model.Setting{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return gdb
}

func newSettingService(t *testing.T, gdb *gorm.DB) *SettingService {
	t.Helper()
	reg := setting.NewRegistry()
	if err := reg.Add(subsystemAdmin, AdminSettingCatalog()); err != nil {
		t.Fatal(err)
	}
	return NewSettingService(repository.NewSettingRepository(gdb), reg, testLockerFactory{}, validator.New())
}

func TestSettingDefaultsAndPublic(t *testing.T) {
	svc := newSettingService(t, setupSettingDB(t))
	ctx := context.Background()

	view, err := svc.View(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(view.Subsystems) != 1 || view.Subsystems[0].Name != subsystemAdmin {
		t.Fatalf("subsystems = %+v", view.Subsystems)
	}
	var sawLogo, sawCopyright bool
	for _, group := range view.Subsystems[0].Groups {
		for _, item := range group.Items {
			if item.Key == keySiteLogo {
				sawLogo = item.Public
			}
			if item.Key == keyCopyright {
				sawCopyright = item.Public && item.Value == ""
			}
		}
	}
	if !sawLogo || !sawCopyright {
		t.Fatal("site logo and copyright must be public")
	}
	seconds, err := svc.Int(ctx, subsystemAdmin, keyRegisterCooldownSeconds)
	if err != nil || seconds != 60 {
		t.Fatalf("cooldown = %d err=%v", seconds, err)
	}

	pub, err := svc.Public(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, group := range pub.Subsystems[0].Groups {
		for _, item := range group.Items {
			if !item.Public {
				t.Fatalf("public view leaked %s", item.Key)
			}
		}
	}
}

func TestSettingUpdateRejectsUnknownAndDropsDefault(t *testing.T) {
	svc := newSettingService(t, setupSettingDB(t))
	ctx := context.Background()

	err := svc.Update(ctx, &UpdateSettingsReq{
		Subsystem: subsystemAdmin,
		Values:    []SettingValueReq{{Key: "nope", Value: "1"}},
	})
	if err == nil || err.Error() == "" {
		t.Fatal("expected unknown key")
	}

	err = svc.Update(ctx, &UpdateSettingsReq{
		Subsystem: subsystemAdmin,
		Values:    []SettingValueReq{{Key: keyRegisterCooldownSeconds, Value: "10"}},
	})
	if err == nil {
		t.Fatal("expected cooldown below minimum")
	}

	err = svc.Update(ctx, &UpdateSettingsReq{
		Subsystem: subsystemAdmin,
		Values:    []SettingValueReq{{Key: keyRegisterCooldownSeconds, Value: "90"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	seconds, err := svc.Int(ctx, subsystemAdmin, keyRegisterCooldownSeconds)
	if err != nil || seconds != 90 {
		t.Fatalf("cooldown = %d err=%v", seconds, err)
	}

	err = svc.Update(ctx, &UpdateSettingsReq{
		Subsystem: subsystemAdmin,
		Values:    []SettingValueReq{{Key: keyRegisterEnabled, Value: "false"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	open, err := svc.Bool(ctx, subsystemAdmin, keyRegisterEnabled)
	if err != nil || open {
		t.Fatalf("register = %v err=%v", open, err)
	}

	err = svc.Update(ctx, &UpdateSettingsReq{
		Subsystem: subsystemAdmin,
		Values:    []SettingValueReq{{Key: keyRegisterCooldownSeconds, Value: "60"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	rows, err := svc.repo.WithContext(ctx).List(subsystemAdmin)
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		if row.Key == keyRegisterCooldownSeconds {
			t.Fatal("default value should not be stored")
		}
	}
}

func TestSettingCorruptValueFallsBack(t *testing.T) {
	gdb := setupSettingDB(t)
	svc := newSettingService(t, gdb)
	ctx := context.Background()
	if err := gdb.Create(&model.Setting{Subsystem: subsystemAdmin, Key: keyRegisterCooldownSeconds, Value: "nope"}).Error; err != nil {
		t.Fatal(err)
	}
	seconds, err := svc.Int(ctx, subsystemAdmin, keyRegisterCooldownSeconds)
	if err != nil || seconds != 60 {
		t.Fatalf("cooldown = %d err=%v", seconds, err)
	}
}

func TestSettingImageLogo(t *testing.T) {
	svc := newSettingService(t, setupSettingDB(t))
	ctx := context.Background()
	const png = "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg=="

	err := svc.Update(ctx, &UpdateSettingsReq{
		Subsystem: subsystemAdmin,
		Values:    []SettingValueReq{{Key: keySiteLogo, Value: "https://example.com/a.png"}},
	})
	if err == nil {
		t.Fatal("expected url logo to be rejected")
	}

	err = svc.Update(ctx, &UpdateSettingsReq{
		Subsystem: subsystemAdmin,
		Values:    []SettingValueReq{{Key: keySiteLogo, Value: png}},
	})
	if err != nil {
		t.Fatal(err)
	}
	view, err := svc.View(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var got string
	for _, group := range view.Subsystems[0].Groups {
		for _, item := range group.Items {
			if item.Key == keySiteLogo {
				got = item.Value
			}
		}
	}
	if got != png {
		t.Fatalf("logo = %q", got)
	}
}
