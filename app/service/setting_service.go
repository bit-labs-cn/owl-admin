package service

import (
	"context"
	"strconv"

	"bit-labs.cn/owl-admin/app/model"
	"bit-labs.cn/owl-admin/app/repository"
	"bit-labs.cn/owl/provider/redis"
	"bit-labs.cn/owl/provider/setting"
	validatorv10 "github.com/go-playground/validator/v10"
)

// SettingValueReq 是保存时的一项。
type SettingValueReq struct {
	Key   string `json:"key" validate:"required,max=64" label:"配置键"`
	Value string `json:"value" label:"配置值"`
}

// UpdateSettingsReq 只更新一个子系统里已经登记的键。
type UpdateSettingsReq struct {
	Subsystem string            `json:"subsystem" validate:"required,max=64" label:"子系统"`
	Values    []SettingValueReq `json:"values" validate:"required,min=1,dive" label:"配置值"`
}

// SettingItemView 是返回给页面的一项配置，不含校验函数。
type SettingItemView struct {
	Key     string            `json:"key"`
	Title   string            `json:"title"`
	Hint    string            `json:"hint"`
	Type    setting.ValueType `json:"type"`
	Default string            `json:"default"`
	Value   string            `json:"value"`
	Min     int               `json:"min"`
	Max     int               `json:"max"`
	Public  bool              `json:"public"`
}

// SettingGroupView 是一个分组。Groups 最多再有一级。
type SettingGroupView struct {
	Key    string             `json:"key"`
	Title  string             `json:"title"`
	Items  []SettingItemView  `json:"items"`
	Groups []SettingGroupView `json:"groups,omitempty"`
}

// SettingSubsystemView 是一个子系统。
type SettingSubsystemView struct {
	Name   string             `json:"name"`
	Title  string             `json:"title"`
	Groups []SettingGroupView `json:"groups"`
}

// SettingView 是配置页和公开接口的树。
type SettingView struct {
	Subsystems []SettingSubsystemView `json:"subsystems"`
}

// SettingService 按注册表合并覆盖值。
type SettingService struct {
	repo     repository.SettingRepositoryInterface
	registry *setting.Registry
	locker   redis.LockerFactory
	validate *validatorv10.Validate
}

// NewSettingService 创建配置服务。
func NewSettingService(
	repo repository.SettingRepositoryInterface,
	registry *setting.Registry,
	locker redis.LockerFactory,
	validate *validatorv10.Validate,
) *SettingService {
	return &SettingService{repo: repo, registry: registry, locker: locker, validate: validate}
}

// View 返回全部已登记配置及当前值。
func (s *SettingService) View(ctx context.Context) (SettingView, error) {
	rows, err := s.repo.WithContext(ctx).All()
	if err != nil {
		return SettingView{}, err
	}
	return s.view(rows, false), nil
}

// Public 只返回登记为公开的配置。
func (s *SettingService) Public(ctx context.Context) (SettingView, error) {
	rows, err := s.repo.WithContext(ctx).All()
	if err != nil {
		return SettingView{}, err
	}
	return s.view(rows, true), nil
}

// Update 保存一个子系统提交的值。未知键和越界值会被拒绝。
func (s *SettingService) Update(ctx context.Context, req *UpdateSettingsReq) error {
	if err := s.validate.Struct(req); err != nil {
		return err
	}
	sub, ok := s.registry.Find(req.Subsystem)
	if !ok {
		return SettingUnknown("子系统不存在")
	}

	l := s.locker.New()
	if err := l.Lock("setting:update:" + req.Subsystem); err != nil {
		return err
	}
	defer l.Unlock()

	seen := make(map[string]struct{}, len(req.Values))
	for _, v := range req.Values {
		if _, dup := seen[v.Key]; dup {
			return SettingInvalid("配置项重复")
		}
		seen[v.Key] = struct{}{}
		if _, ok := s.registry.Item(req.Subsystem, v.Key); !ok {
			return SettingUnknown("配置项不存在")
		}
	}

	rows, err := s.repo.WithContext(ctx).List(req.Subsystem)
	if err != nil {
		return err
	}
	stored := mapRows(rows)
	next := make(map[string]string)
	eachSettingItem(sub.Groups, func(item setting.Item) {
		next[item.Key] = effective(item, stored)
	})
	for _, v := range req.Values {
		item, _ := s.registry.Item(req.Subsystem, v.Key)
		if err := item.Accepts(v.Value); err != nil {
			return SettingInvalid(err.Error())
		}
		next[v.Key] = v.Value
	}

	checked := map[string]struct{}{}
	for _, v := range req.Values {
		group, ok := s.registry.GroupOf(req.Subsystem, v.Key)
		if !ok || group.Validate == nil {
			continue
		}
		if _, done := checked[group.Key]; done {
			continue
		}
		checked[group.Key] = struct{}{}
		values := make(map[string]string, len(group.Items))
		for _, item := range group.Items {
			values[item.Key] = next[item.Key]
		}
		if err := group.Validate(values); err != nil {
			return SettingInvalid(err.Error())
		}
	}

	upserts := map[string]string{}
	var drops []string
	for _, v := range req.Values {
		item, _ := s.registry.Item(req.Subsystem, v.Key)
		if v.Value == item.Default {
			drops = append(drops, v.Key)
			continue
		}
		upserts[v.Key] = v.Value
	}
	return s.repo.WithContext(ctx).Apply(req.Subsystem, upserts, drops)
}

// String 读取文本配置。未覆盖或值损坏时返回默认值。
func (s *SettingService) String(ctx context.Context, subsystem, key string) (string, error) {
	item, raw, err := s.lookup(ctx, subsystem, key)
	if err != nil {
		return "", err
	}
	if item.Type != setting.TypeString {
		return "", SettingInvalid("配置项类型不是文本")
	}
	return raw, nil
}

// Bool 读取布尔配置。
func (s *SettingService) Bool(ctx context.Context, subsystem, key string) (bool, error) {
	item, raw, err := s.lookup(ctx, subsystem, key)
	if err != nil {
		return false, err
	}
	if item.Type != setting.TypeBool {
		return false, SettingInvalid("配置项类型不是布尔")
	}
	return raw == "true", nil
}

// Int 读取整数配置。
func (s *SettingService) Int(ctx context.Context, subsystem, key string) (int, error) {
	item, raw, err := s.lookup(ctx, subsystem, key)
	if err != nil {
		return 0, err
	}
	if item.Type != setting.TypeInt {
		return 0, SettingInvalid("配置项类型不是整数")
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0, SettingInvalid("配置值不是整数")
	}
	return n, nil
}

func (s *SettingService) lookup(ctx context.Context, subsystem, key string) (setting.Item, string, error) {
	item, ok := s.registry.Item(subsystem, key)
	if !ok {
		return setting.Item{}, "", SettingUnknown("配置项不存在")
	}
	row, found, err := s.repo.WithContext(ctx).Get(subsystem, key)
	if err != nil {
		return setting.Item{}, "", err
	}
	if !found {
		return item, item.Default, nil
	}
	return item, effective(item, map[string]string{key: row.Value}), nil
}

func (s *SettingService) view(rows []model.Setting, publicOnly bool) SettingView {
	storedBySub := map[string]map[string]string{}
	for _, row := range rows {
		if storedBySub[row.Subsystem] == nil {
			storedBySub[row.Subsystem] = map[string]string{}
		}
		storedBySub[row.Subsystem][row.Key] = row.Value
	}
	view := SettingView{}
	for _, sub := range s.registry.Subsystems() {
		stored := storedBySub[sub.Name]
		out := SettingSubsystemView{Name: sub.Name, Title: sub.Title}
		for _, group := range sub.Groups {
			gv, ok := buildGroupView(group, stored, publicOnly)
			if !ok {
				continue
			}
			out.Groups = append(out.Groups, gv)
		}
		if len(out.Groups) == 0 {
			continue
		}
		view.Subsystems = append(view.Subsystems, out)
	}
	if view.Subsystems == nil {
		view.Subsystems = []SettingSubsystemView{}
	}
	return view
}

func buildGroupView(group setting.Group, stored map[string]string, publicOnly bool) (SettingGroupView, bool) {
	gv := SettingGroupView{Key: group.Key, Title: group.Title, Items: []SettingItemView{}}
	for _, item := range group.Items {
		if publicOnly && !item.Public {
			continue
		}
		gv.Items = append(gv.Items, SettingItemView{
			Key: item.Key, Title: item.Title, Hint: item.Hint, Type: item.Type,
			Default: item.Default, Value: effective(item, stored),
			Min: item.Min, Max: item.Max, Public: item.Public,
		})
	}
	for _, child := range group.Groups {
		nested, ok := buildGroupView(child, stored, publicOnly)
		if !ok {
			continue
		}
		gv.Groups = append(gv.Groups, nested)
	}
	if len(gv.Items) == 0 && len(gv.Groups) == 0 {
		return SettingGroupView{}, false
	}
	return gv, true
}

func eachSettingItem(groups []setting.Group, fn func(setting.Item)) {
	for _, group := range groups {
		for _, item := range group.Items {
			fn(item)
		}
		eachSettingItem(group.Groups, fn)
	}
}

func mapRows(rows []model.Setting) map[string]string {
	out := make(map[string]string, len(rows))
	for _, row := range rows {
		out[row.Key] = row.Value
	}
	return out
}

func effective(item setting.Item, stored map[string]string) string {
	raw, ok := stored[item.Key]
	if !ok || item.Accepts(raw) != nil {
		return item.Default
	}
	return raw
}
