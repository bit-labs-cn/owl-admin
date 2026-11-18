package service

import (
	"context"

	"bit-labs.cn/owl-admin/app-notification/repository"
	"bit-labs.cn/owl/provider/redis"
)

// SettingValues 是通知中心运行配置。
type SettingValues struct {
	ApprovalEnabled       bool   `json:"approvalEnabled"`
	QuietHours            string `json:"quietHours"`
	RateLimitPerDay       int    `json:"rateLimitPerDay"`
	InboxRetentionDays    int    `json:"inboxRetentionDays"`
	DeliveryRetentionDays int    `json:"deliveryRetentionDays"`
}

// SettingService 读写运行配置。
type SettingService struct {
	repo   repository.SettingRepositoryInterface
	locker redis.LockerFactory
}

func NewSettingService(repo repository.SettingRepositoryInterface, locker redis.LockerFactory) *SettingService {
	return &SettingService{repo: repo, locker: locker}
}

func (s *SettingService) Get(ctx context.Context) (SettingValues, error) {
	out := SettingValues{InboxRetentionDays: 180, DeliveryRetentionDays: 90}
	rows, err := s.repo.WithContext(ctx).List()
	if err != nil {
		return out, err
	}
	for _, row := range rows {
		switch row.Key {
		case settingApproval:
			out.ApprovalEnabled = row.Value == "1" || row.Value == "true"
		case settingQuiet:
			out.QuietHours = row.Value
		case settingRate:
			out.RateLimitPerDay = atoi(row.Value)
		case settingInboxDays:
			out.InboxRetentionDays = atoi(row.Value)
		case settingDeliverDay:
			out.DeliveryRetentionDays = atoi(row.Value)
		}
	}
	return out, nil
}

func (s *SettingService) Save(ctx context.Context, req SettingValues) error {
	l := s.locker.New()
	if err := l.Lock("notify:setting"); err != nil {
		return err
	}
	defer l.Unlock()
	repo := s.repo.WithContext(ctx)
	pairs := map[string]string{
		settingApproval:   boolText(req.ApprovalEnabled),
		settingQuiet:      req.QuietHours,
		settingRate:       itoa(req.RateLimitPerDay),
		settingInboxDays:  itoa(req.InboxRetentionDays),
		settingDeliverDay: itoa(req.DeliveryRetentionDays),
	}
	for k, v := range pairs {
		if err := repo.Set(k, v); err != nil {
			return err
		}
	}
	return nil
}

func boolText(v bool) string {
	if v {
		return "1"
	}
	return "2"
}

func itoa(n int) string {
	if n < 0 {
		n = 0
	}
	return stringsItoa(n)
}

func atoi(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			break
		}
		n = n*10 + int(c-'0')
	}
	return n
}

func stringsItoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [16]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
