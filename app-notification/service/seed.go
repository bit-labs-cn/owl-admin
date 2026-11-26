package service

import (
	"context"

	"bit-labs.cn/owl-admin/app-notification/model"
)

const (
	builtinNoticeType = "system.notice"
	builtinTodoType   = "system.todo"
)

// EnsureBuiltinTypes 在缺少内置类型时补上系统通知和待办。已有记录不覆盖。
func (s *NotificationService) EnsureBuiltinTypes(ctx context.Context) error {
	seeds := []model.NotifyType{
		{
			Code:            builtinNoticeType,
			Name:            "系统通知",
			Description:     "系统产生的知会",
			Kind:            model.KindNotice,
			DefaultChannels: encodeChannels([]string{model.ChannelInApp}),
			Status:          model.TypeStatusEnabled,
		},
		{
			Code:            builtinTodoType,
			Name:            "待办",
			Description:     "需要处理的待办",
			Kind:            model.KindTodo,
			DefaultChannels: encodeChannels([]string{model.ChannelInApp}),
			Status:          model.TypeStatusEnabled,
		},
	}
	for i := range seeds {
		exist, err := s.catalog.FindTypeByCode(ctx, seeds[i].Code)
		if err != nil {
			return err
		}
		if exist != nil {
			continue
		}

		if err := s.catalog.SaveType(ctx, &seeds[i]); err != nil {
			return err
		}
	}

	return nil
}
