package notification

import (
	"context"
	"fmt"
	"strings"
	"time"

	"bit-labs.cn/owl-admin/app-notification/contract"
	"bit-labs.cn/owl-admin/app-notification/repository"
	"bit-labs.cn/owl-admin/app-notification/service"
	"bit-labs.cn/owl-admin/app/model"
	schedulerContract "bit-labs.cn/owl/contract/scheduler"
)

// notificationActionSend 是通知中心暴露给定时任务的动作标识。
const notificationActionSend = "send-message"

// RegisterActions 暴露「发送通知 / 待办」动作，供管理端编排成定时任务。
//
// 参数说明由这里声明，管理端据此渲染表单与校验；执行时走与人工发送相同的 SendDirect，
// 因此渠道、模板、落库与事件投递的行为完全一致。
func (s *SubAppNotification) RegisterActions() []schedulerContract.Action {
	var (
		notifSvc *service.NotificationService
		catalog  repository.CatalogRepositoryInterface
	)
	if err := s.app.Invoke(func(
		svc *service.NotificationService,
		c repository.CatalogRepositoryInterface,
	) {
		notifSvc = svc
		catalog = c
	}); err != nil {
		panic(fmt.Errorf("notify: 注册定时任务动作失败: %w", err))
	}

	return []schedulerContract.Action{{
		Key:         notificationActionSend,
		Name:        "发送通知",
		Group:       "通知中心",
		Description: "按指定接收人与渠道发送一条通知或待办；执行时与人工发送走同一条业务路径。",
		Params: []schedulerContract.Param{
			{
				Name: "kind", Label: "消息类型", Type: schedulerContract.ParamSelect, Required: true,
				Default: fmt.Sprint(contract.KindNotice),
				Options: []schedulerContract.ParamOption{
					{Value: fmt.Sprint(contract.KindNotice), Label: "通知"},
					{Value: fmt.Sprint(contract.KindTodo), Label: "待办"},
				},
			},
			{
				Name: "title", Label: "标题", Type: schedulerContract.ParamString, Required: true,
				Placeholder: "例如：每日巡检提醒",
			},
			{
				Name: "content", Label: "正文", Type: schedulerContract.ParamText, Required: true,
				Placeholder: "支持纯文本；站内信会按正文渲染",
			},
			{
				Name: "link", Label: "跳转链接", Type: schedulerContract.ParamString,
				Placeholder: "例如 /inspection/task/1024，可留空",
			},
			{
				Name: "recipients", Label: "接收人", Type: schedulerContract.ParamSelect, Required: true, Multiple: true, Dynamic: true,
				Description: "最多列出 200 个启用用户；动态候选按用户昵称展示，不显示裸 ID。",
			},
			{
				Name: "channels", Label: "发送渠道", Type: schedulerContract.ParamSelect, Multiple: true,
				Description: "留空时按消息类型的默认渠道，再退回站内信。",
				Options: []schedulerContract.ParamOption{
					{Value: contract.ChannelInApp, Label: "站内信"},
					{Value: contract.ChannelEmail, Label: "邮件"},
					{Value: contract.ChannelSMS, Label: "短信"},
					{Value: contract.ChannelWeCom, Label: "企业微信"},
					{Value: contract.ChannelDingTalk, Label: "钉钉"},
				},
			},
			{
				Name: "typeCode", Label: "消息类型编码", Type: schedulerContract.ParamSelect, Dynamic: true,
				Description: "留空表示不套用类型模板，直接用标题与正文。",
			},
			{
				Name: "dueAt", Label: "待办截止时间", Type: schedulerContract.ParamString,
				Placeholder: "格式 2006-01-02 15:04:05；待办不填时默认 24 小时后到期",
			},
		},
		Options: func(ctx context.Context, params map[string]string) (map[string][]schedulerContract.ParamOption, error) {
			out := map[string][]schedulerContract.ParamOption{}

			contacts, err := catalog.SearchContacts(ctx, "", nil, 200)
			if err != nil {
				return nil, err
			}
			recipients := make([]schedulerContract.ParamOption, 0, len(contacts))
			for _, contact := range contacts {
				// 停用与待审核账号不在候选里，避免定时任务发给收不到消息的人。
				if contact.Status != model.UserStatusEnabled {
					continue
				}
				recipients = append(recipients, schedulerContract.ParamOption{
					Value: fmt.Sprint(contact.ID),
					Label: contactLabel(contact),
				})
			}
			out["recipients"] = recipients

			types, err := catalog.ListEnabledTypes(ctx)
			if err != nil {
				return nil, err
			}
			typeOptions := []schedulerContract.ParamOption{{Value: "", Label: "不套用模板"}}
			for _, item := range types {
				typeOptions = append(typeOptions, schedulerContract.ParamOption{Value: item.Code, Label: item.Name})
			}
			out["typeCode"] = typeOptions
			return out, nil
		},
		Handler: func(ctx context.Context, in schedulerContract.ActionInput) (string, error) {
			if notifSvc == nil {
				return "", fmt.Errorf("通知服务未就绪")
			}
			kind, err := in.Int("kind", contract.KindNotice)
			if err != nil {
				return "", err
			}
			if kind != contract.KindNotice && kind != contract.KindTodo {
				return "", fmt.Errorf("消息类型只能是通知或待办")
			}
			userIDs, err := in.Uints("recipients")
			if err != nil {
				return "", err
			}
			if len(userIDs) == 0 {
				return "", fmt.Errorf("请选择接收人")
			}

			msg := &contract.DirectMessage{
				Kind:         kind,
				TypeCode:     in.String("typeCode", ""),
				Title:        in.String("title", ""),
				Content:      in.String("content", ""),
				Link:         in.String("link", ""),
				UserIDs:      userIDs,
				Channels:     in.Split("channels"),
				SkipTemplate: true,
				// 定时任务由系统发起，业务键用执行记录 ID 保证重试幂等。
				BizKey: fmt.Sprintf("scheduler:%d", in.RunID),
			}
			if kind == contract.KindTodo {
				dueAt, parseErr := parseDueAt(in.String("dueAt", ""))
				if parseErr != nil {
					return "", parseErr
				}
				msg.DueAt = &dueAt
			}
			if err := notifSvc.SendDirect(ctx, msg); err != nil {
				return "", err
			}
			return fmt.Sprintf("已提交通知：%s，接收人 %d 人", msg.Title, len(userIDs)), nil
		},
	}}
}

// contactLabel 用昵称优先拼出可读联系人名称。
func contactLabel(contact repository.UserContact) string {
	name := strings.TrimSpace(contact.Nickname)
	if name == "" {
		name = strings.TrimSpace(contact.Username)
	}
	if name == "" {
		name = fmt.Sprintf("用户(%d)", contact.ID)
	}
	if contact.Username != "" && contact.Nickname != "" && contact.Nickname != contact.Username {
		return fmt.Sprintf("%s（%s）", name, contact.Username)
	}
	return name
}

// parseDueAt 解析待办截止时间；留空时给 24 小时，避免待办一创建就过期。
func parseDueAt(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Now().Add(24 * time.Hour), nil
	}
	due, err := time.ParseInLocation("2006-01-02 15:04:05", raw, time.Local)
	if err != nil {
		return time.Time{}, fmt.Errorf("待办截止时间格式应为 2006-01-02 15:04:05")
	}
	return due, nil
}
