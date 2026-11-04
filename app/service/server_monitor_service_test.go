package service

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestServerSnapshotKeepsIndependentFailures(t *testing.T) {
	gdb := openMonitorDB(t)
	client := redis.NewClient(&redis.Options{
		Addr:         "127.0.0.1:1",
		DialTimeout:  50 * time.Millisecond,
		ReadTimeout:  50 * time.Millisecond,
		WriteTimeout: 50 * time.Millisecond,
		MaxRetries:   -1,
	})
	t.Cleanup(func() { _ = client.Close() })

	snap := NewServerMonitorService(gdb, client).Snapshot(context.Background())
	if snap.CollectedAt == "" {
		t.Fatal("missing collectedAt")
	}
	if snap.Runtime.GoVersion == "" || snap.Runtime.Goroutines < 1 {
		t.Fatalf("runtime = %+v", snap.Runtime)
	}
	if !snap.Database.OK {
		t.Fatalf("database = %+v", snap.Database)
	}
	if snap.Redis.OK || snap.Redis.Error == "" {
		t.Fatalf("redis = %+v", snap.Redis)
	}
	if snap.Host.HostError == "" && snap.Host.Hostname == "" {
		t.Fatal("host produced neither name nor error")
	}
}

func TestParseRedisInfo(t *testing.T) {
	raw := "# Server\r\nredis_version:7.2.0\r\nused_memory:2048\r\nkeyspace_hits:3\r\nkeyspace_misses:1\r\n"
	info := parseRedisInfo(raw)
	if info["redis_version"] != "7.2.0" || parseUint64(info["used_memory"]) != 2048 {
		t.Fatalf("info = %#v", info)
	}
	if parseInt64(info["keyspace_hits"]) != 3 {
		t.Fatalf("hits = %s", info["keyspace_hits"])
	}
	if strings.Contains(info["# Server"], "Server") {
		t.Fatal("comment parsed as field")
	}
}

func openMonitorDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	gdb, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	return gdb
}
