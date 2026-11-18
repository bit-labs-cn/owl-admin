package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"bit-labs.cn/owl-admin/app-notification/contract"
	"bit-labs.cn/owl-admin/app-notification/model"
	"bit-labs.cn/owl-admin/app-notification/repository"
	adminmodel "bit-labs.cn/owl-admin/app/model"
)

// ManualSendInput 是管理端人工发送一条通知或待办的参数。
type ManualSendInput struct {
	Kind               int
	TypeCode           string
	Title              string
	Content            string
	Link               string
	Channels           []string
	UserIDs            []uint
	RoleIDs            []uint
	DeptIDs            []uint
	GroupIDs           []uint
	IncludeDeptChild   bool
	ExternalRecipients []contract.ExternalRecipient
	DueAt              *time.Time
	// PlanAt 非空时不是立即发送，而是登记一条定时任务，到点由后台调度发送。
	PlanAt   *time.Time
	SenderID uint
}

// RecipientOption 是发送页的收件人候选。
type RecipientOption struct {
	ID       uint   `json:"id,string"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Disabled bool   `json:"disabled"`
}

// resolveKind 用消息类型自带的种类决定这条消息是通知还是待办。
// 选了类型就以类型为准，避免调用方传的 kind 和类型对不上；没选类型时才看调用方传的值。
func (s *NotificationService) resolveKind(ctx context.Context, kind int, typeCode string) (int, error) {
	if typeCode == "" {
		if kind == model.KindTodo {
			return model.KindTodo, nil
		}
		return model.KindNotice, nil
	}
	typ, err := s.catalog.FindTypeByCode(ctx, typeCode)
	if err != nil {
		return 0, err
	}
	if typ == nil || typ.Status != model.TypeStatusEnabled {
		return 0, NotFound("消息类型不存在或已停用")
	}
	if typ.Kind == model.KindTodo {
		return model.KindTodo, nil
	}
	return model.KindNotice, nil
}

// ManualSend 校验并发送一条人工消息，复用业务通道的同一条发送管线。
// 传了 PlanAt 时不立即发送，而是登记一条定时任务并返回它。
func (s *NotificationService) ManualSend(ctx context.Context, in ManualSendInput) (*model.NotifySchedule, error) {
	if in.PlanAt != nil {
		return s.CreateSchedule(ctx, ScheduleInput{
			Kind:               in.Kind,
			TypeCode:           in.TypeCode,
			Title:              in.Title,
			Content:            in.Content,
			Link:               in.Link,
			Channels:           in.Channels,
			UserIDs:            in.UserIDs,
			RoleIDs:            in.RoleIDs,
			DeptIDs:            in.DeptIDs,
			GroupIDs:           in.GroupIDs,
			IncludeDeptChild:   in.IncludeDeptChild,
			ExternalRecipients: in.ExternalRecipients,
			DueAt:              in.DueAt,
			PlanAt:             *in.PlanAt,
			SenderID:           in.SenderID,
		})
	}

	userIDs, err := s.catalog.ResolveAudience(ctx, repository.AudienceSelection{
		UserIDs:      in.UserIDs,
		RoleIDs:      in.RoleIDs,
		DeptIDs:      in.DeptIDs,
		GroupIDs:     in.GroupIDs,
		IncludeChild: in.IncludeDeptChild,
	})
	if err != nil {
		return nil, err
	}
	externals := normalizeExternals(in.ExternalRecipients)
	if len(userIDs) == 0 && len(externals) == 0 {
		return nil, Invalid("请选择接收人，或填写外部人员的邮箱")
	}

	typeCode := strings.TrimSpace(in.TypeCode)
	kind, err := s.resolveKind(ctx, in.Kind, typeCode)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(in.Title) == "" {
		return nil, Invalid("请填写标题")
	}
	channels := normalizeChannels(in.Channels)
	for _, channel := range channels {
		if !knownChannel(channel) {
			return nil, Invalid("渠道无效")
		}
	}
	if kind == model.KindTodo && in.DueAt == nil {
		return nil, Invalid("待办需要填写截止时间")
	}

	return nil, s.SendDirect(ctx, &contract.DirectMessage{
		Kind:               kind,
		TypeCode:           typeCode,
		Title:              strings.TrimSpace(in.Title),
		Content:            in.Content,
		Link:               strings.TrimSpace(in.Link),
		DueAt:              in.DueAt,
		UserIDs:            userIDs,
		Channels:           channels,
		SkipTemplate:       true,
		SenderID:           in.SenderID,
		ExternalRecipients: externals,
	})
}

// RecipientOptions 返回收件人候选。传 ID 时按 ID 回显，否则按关键字搜索。
func (s *NotificationService) RecipientOptions(ctx context.Context, keyword string, userIDs []uint) ([]RecipientOption, error) {
	contacts, err := s.catalog.SearchContacts(ctx, keyword, userIDs, 50)
	if err != nil {
		return nil, err
	}
	return s.toRecipientOptions(contacts), nil
}

// toRecipientOptions 把联系人转成发送页用的候选，并标出已禁用账号。
func (s *NotificationService) toRecipientOptions(contacts []repository.UserContact) []RecipientOption {
	out := make([]RecipientOption, 0, len(contacts))
	for _, contact := range contacts {
		out = append(out, RecipientOption{
			ID:       contact.ID,
			Name:     contactName(contact),
			Username: contact.Username,
			Email:    contact.Email,
			Phone:    contact.Phone,
			Disabled: contact.Status != adminmodel.UserStatusEnabled,
		})
	}
	return out
}

// AudienceEstimate 是发送页的接收人预估结果。
// 只做统计，不写库；真正发送时按当刻的组织关系重新解析。
type AudienceEstimate struct {
	UserCount     int `json:"userCount"`
	ExternalCount int `json:"externalCount"`
	Total         int `json:"total"`
}

// EstimateAudience 预估当前受众条件下的接收人数。
// 外部人员只按邮箱投递，没有填邮箱的不计入预计接收。
func (s *NotificationService) EstimateAudience(
	ctx context.Context,
	sel repository.AudienceSelection,
	externals []contract.ExternalRecipient,
) (*AudienceEstimate, error) {
	userIDs, err := s.catalog.ResolveAudience(ctx, sel)
	if err != nil {
		return nil, err
	}
	externalCount := 0
	for _, item := range normalizeExternals(externals) {
		if strings.TrimSpace(item.Email) != "" {
			externalCount++
		}
	}
	return &AudienceEstimate{
		UserCount:     len(userIDs),
		ExternalCount: externalCount,
		Total:         len(userIDs) + externalCount,
	}, nil
}

// AudienceOptions 返回角色、部门树和用户组选项，供发送页选择受众。
func (s *NotificationService) AudienceOptions(ctx context.Context) (*repository.AudienceOptions, error) {
	return s.catalog.AudienceOptions(ctx)
}

// AdminMessages 返回管理端消息记录及接收统计。
func (s *NotificationService) AdminMessages(ctx context.Context, q repository.MessageQuery) (int64, []repository.MessageSummaryDTO, error) {
	total, list, err := s.catalog.ListMessages(ctx, q)
	if err != nil {
		return 0, nil, err
	}
	for i := range list {
		if list[i].TypeCode != "" && list[i].TypeName == "" {
			list[i].TypeName = fmt.Sprintf("未知类型(%s)", list[i].TypeCode)
		}
		if list[i].SenderID > 0 && list[i].SenderName == "" {
			list[i].SenderName = fmt.Sprintf("未知用户(%d)", list[i].SenderID)
		}
		if list[i].SenderID == 0 {
			list[i].SenderName = "系统"
		}
	}
	return total, list, nil
}

// AdminMessageDetail 返回一条消息的正文与接收人已读、待办明细。
func (s *NotificationService) AdminMessageDetail(ctx context.Context, id uint) (*repository.MessageDetailDTO, error) {
	detail, err := s.catalog.FindMessageDetail(ctx, id)
	if err != nil {
		return nil, err
	}
	if detail == nil {
		return nil, NotFound("消息不存在")
	}
	if detail.TypeCode != "" && detail.TypeName == "" {
		detail.TypeName = fmt.Sprintf("未知类型(%s)", detail.TypeCode)
	}
	if detail.SenderID == 0 {
		detail.SenderName = "系统"
	} else if detail.SenderName == "" {
		detail.SenderName = fmt.Sprintf("未知用户(%d)", detail.SenderID)
	}
	for i := range detail.Recipients {
		row := &detail.Recipients[i]
		if row.UserName == "" {
			row.UserName = fmt.Sprintf("未知用户(%d)", row.UserID)
		}
	}
	return detail, nil
}

// normalizeExternals 去掉空行，要求至少有一个联系方式，并按邮箱或手机号去重。
func normalizeExternals(items []contract.ExternalRecipient) []contract.ExternalRecipient {
	out := make([]contract.ExternalRecipient, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		email := strings.TrimSpace(item.Email)
		phone := strings.TrimSpace(item.Phone)
		name := strings.TrimSpace(item.Name)
		if email == "" && phone == "" {
			continue
		}
		key := strings.ToLower(email) + "\x00" + phone
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		if name == "" {
			name = email
			if name == "" {
				name = phone
			}
		}
		out = append(out, contract.ExternalRecipient{Name: name, Email: email, Phone: phone})
	}
	return out
}
