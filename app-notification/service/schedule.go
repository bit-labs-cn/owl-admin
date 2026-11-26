package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"bit-labs.cn/owl-admin/app-notification/contract"
	"bit-labs.cn/owl-admin/app-notification/model"
	"bit-labs.cn/owl-admin/app-notification/repository"
)

// scheduleBizKeyPrefix 是定时任务的业务键前缀。
// 到点发送时用「前缀 + 任务ID」当业务键，重试或重复领取都不会发出第二条消息。
const scheduleBizKeyPrefix = "notify-schedule:"

func scheduleBizKey(id uint) string {
	return fmt.Sprintf("%s%d", scheduleBizKeyPrefix, id)
}

// ScheduleInput 是创建一条定时发送任务的参数。
// 受众只保存选择条件（用户、角色、部门、用户组、外部人员），到点发送时才解析成具体用户。
type ScheduleInput struct {
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
	PlanAt             time.Time
	SenderID           uint
}

// scheduleSummaryNameLimit 是接收人摘要里最多列出的姓名个数，其余用「等 N 人」概括。
const scheduleSummaryNameLimit = 3

// CreateSchedule 校验并保存一条定时发送任务。
func (s *NotificationService) CreateSchedule(ctx context.Context, in ScheduleInput) (*model.NotifySchedule, error) {
	typeCode := strings.TrimSpace(in.TypeCode)
	kind, err := s.resolveKind(ctx, in.Kind, typeCode)
	if err != nil {
		return nil, err
	}
	title := strings.TrimSpace(in.Title)
	if title == "" {
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
	if in.PlanAt.IsZero() {
		return nil, Invalid("请填写定时发送时间")
	}
	if !in.PlanAt.After(time.Now()) {
		return nil, Invalid("定时发送时间必须晚于当前时间")
	}

	externals := normalizeExternals(in.ExternalRecipients)
	selection := repository.AudienceSelection{
		UserIDs:      in.UserIDs,
		RoleIDs:      in.RoleIDs,
		DeptIDs:      in.DeptIDs,
		GroupIDs:     in.GroupIDs,
		IncludeChild: in.IncludeDeptChild,
	}
	if len(selection.UserIDs) == 0 && len(selection.RoleIDs) == 0 &&
		len(selection.DeptIDs) == 0 && len(selection.GroupIDs) == 0 && len(externals) == 0 {
		return nil, Invalid("请选择接收人，或填写外部人员的邮箱")
	}

	// 计划时解析一次，只为给列表留下可读的接收人摘要；真正发送时按当时组织关系重算。
	userIDs, err := s.catalog.ResolveAudience(ctx, selection)
	if err != nil {
		return nil, err
	}
	summary, err := s.scheduleAudienceSummary(ctx, userIDs, externals)
	if err != nil {
		return nil, err
	}

	row := &model.NotifySchedule{
		Kind:     kind,
		TypeCode: typeCode,
		Title:    title,
		Content:  ContentHTML(in.Content),
		Link:     strings.TrimSpace(in.Link),
		Channels: toJSON(channels),
		Audience: toJSON(repository.ScheduleAudience{
			UserIDs:      selection.UserIDs,
			RoleIDs:      selection.RoleIDs,
			DeptIDs:      selection.DeptIDs,
			GroupIDs:     selection.GroupIDs,
			IncludeChild: selection.IncludeChild,
			Externals:    externals,
		}),
		AudienceSummary: summary,
		RecipientCount:  len(userIDs) + len(externals),
		DueAt:           in.DueAt,
		PlanAt:          in.PlanAt,
		Status:          model.ScheduleStatusPending,
		SenderID:        in.SenderID,
	}

	if err := s.catalog.CreateSchedule(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

// scheduleAudienceSummary 生成「张三、李四 等 12 人；外部人员 2 人」这样的可读摘要。
func (s *NotificationService) scheduleAudienceSummary(ctx context.Context, userIDs []uint, externals []contract.ExternalRecipient) (string, error) {
	parts := make([]string, 0, 2)
	if len(userIDs) > 0 {
		contacts, err := s.catalog.ListContacts(ctx, userIDs)
		if err != nil {
			return "", err
		}
		names := make([]string, 0, scheduleSummaryNameLimit)
		for _, contact := range contacts {
			if len(names) >= scheduleSummaryNameLimit {
				break
			}
			names = append(names, contactName(contact))
		}
		switch {
		case len(names) == 0:
			parts = append(parts, fmt.Sprintf("%d 人", len(userIDs)))
		case len(userIDs) > len(names):
			parts = append(parts, fmt.Sprintf("%s 等 %d 人", strings.Join(names, "、"), len(userIDs)))
		default:
			parts = append(parts, strings.Join(names, "、"))
		}
	}
	if len(externals) > 0 {
		parts = append(parts, fmt.Sprintf("外部人员 %d 人", len(externals)))
	}
	return strings.Join(parts, "；"), nil
}

// AdminSchedules 返回管理端定时发送列表，补齐类型名与发送人。
func (s *NotificationService) AdminSchedules(ctx context.Context, q repository.ScheduleQuery) (int64, []repository.ScheduleDTO, error) {
	total, list, err := s.catalog.ListSchedules(ctx, q)
	if err != nil {
		return 0, nil, err
	}

	for i := range list {
		row := &list[i]
		if row.TypeCode != "" && row.TypeName == "" {
			row.TypeName = fmt.Sprintf("未知类型(%s)", row.TypeCode)
		}
		if row.SenderID == 0 {
			row.SenderName = "系统"
		} else if row.SenderName == "" {
			row.SenderName = fmt.Sprintf("未知用户(%d)", row.SenderID)
		}
	}

	return total, list, nil
}

// CancelSchedule 取消一条还没发出去的定时任务。
func (s *NotificationService) CancelSchedule(ctx context.Context, id uint) error {
	row, err := s.catalog.FindSchedule(ctx, id)
	if err != nil {
		return err
	}
	if row == nil {
		return NotFound("定时任务不存在")
	}

	switch row.Status {
	case model.ScheduleStatusCanceled:
		return Invalid("该定时任务已取消")
	case model.ScheduleStatusSent:
		return Invalid("已发送的定时任务不能取消")
	case model.ScheduleStatusSending:
		return Invalid("任务正在发送，无法取消")
	}

	row.Status = model.ScheduleStatusCanceled
	row.LastError = ""

	return s.catalog.SaveSchedule(ctx, row)
}

// SendScheduleNow 跳过计划时间，立刻发送一条待发送或失败的定时任务。
func (s *NotificationService) SendScheduleNow(ctx context.Context, id uint) error {
	row, err := s.catalog.FindSchedule(ctx, id)
	if err != nil {
		return err
	}
	if row == nil {
		return NotFound("定时任务不存在")
	}
	if row.Status == model.ScheduleStatusSent {
		return Invalid("该定时任务已经发送")
	}
	if row.Status == model.ScheduleStatusCanceled {
		return Invalid("已取消的定时任务不能发送")
	}
	if row.Status == model.ScheduleStatusSending {
		return Invalid("任务正在发送，请稍后刷新")
	}
	now := time.Now()
	claimed, err := s.catalog.ClaimSchedule(ctx, row.ID, &now)
	if err != nil {
		return err
	}
	if !claimed {
		return Invalid("任务状态已变化，请刷新后重试")
	}
	row.Status = model.ScheduleStatusSending
	row.Attempts++
	row.PlanAt = now
	return s.dispatchSchedule(ctx, row)
}

// RetrySchedule 重试一条发送失败的定时任务，清零尝试次数后立刻重发。
func (s *NotificationService) RetrySchedule(ctx context.Context, id uint) error {
	row, err := s.catalog.FindSchedule(ctx, id)
	if err != nil {
		return err
	}
	if row == nil {
		return NotFound("定时任务不存在")
	}
	if row.Status != model.ScheduleStatusFailed {
		return Invalid("只有失败的定时任务可以重试")
	}
	// 先把尝试次数清零，否则一次失败就把重试额度用光了
	row.Attempts = 0
	row.LastError = ""
	row.MessageID = 0
	row.SentAt = nil

	if err := s.catalog.SaveSchedule(ctx, row); err != nil {
		return err
	}
	now := time.Now()
	claimed, err := s.catalog.ClaimSchedule(ctx, row.ID, &now)
	if err != nil {
		return err
	}
	if !claimed {
		return Invalid("任务状态已变化，请刷新后重试")
	}
	row.Status = model.ScheduleStatusSending
	row.Attempts = 1
	row.PlanAt = now
	return s.dispatchSchedule(ctx, row)
}

// DispatchDueSchedules 领取所有到点的定时任务并发送，供后台调度协程调用。
func (s *NotificationService) DispatchDueSchedules(ctx context.Context, limit int) {
	rows, err := s.catalog.ClaimDueSchedules(ctx, time.Now(), limit)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("notify: claim scheduled messages failed: ", err)
		}
		return
	}
	for i := range rows {
		if err := s.dispatchSchedule(ctx, &rows[i]); err != nil && s.logger != nil {
			s.logger.Error("notify: send scheduled message failed: ", rows[i].ID, " ", err)
		}
	}
}

// dispatchSchedule 按任务里保存的条件解析受众并发送，然后回写结果。
func (s *NotificationService) dispatchSchedule(ctx context.Context, row *model.NotifySchedule) error {
	var audience repository.ScheduleAudience
	if err := json.Unmarshal([]byte(row.Audience), &audience); err != nil {
		return s.markScheduleFailure(ctx, row, Invalid("定时任务的接收人条件无法读取"))
	}
	userIDs, err := s.catalog.ResolveAudience(ctx, audience.Selection())
	if err != nil {
		return s.markScheduleFailure(ctx, row, err)
	}

	var channels []string
	if row.Channels != "" {
		_ = json.Unmarshal([]byte(row.Channels), &channels)
	}

	err = s.SendDirect(ctx, &contract.DirectMessage{
		Kind:               row.Kind,
		TypeCode:           row.TypeCode,
		Title:              row.Title,
		Content:            row.Content,
		Link:               row.Link,
		BizKey:             scheduleBizKey(row.ID),
		DueAt:              row.DueAt,
		UserIDs:            userIDs,
		Channels:           channels,
		SkipTemplate:       true,
		SenderID:           row.SenderID,
		ExternalRecipients: audience.Externals,
	})
	if err != nil {
		return s.markScheduleFailure(ctx, row, err)
	}

	now := time.Now()
	row.Status = model.ScheduleStatusSent
	row.SentAt = &now
	row.LastError = ""
	// 业务键就是任务的幂等键，用它能找回这次真正落库的消息，方便从定时任务跳到消息记录。
	if msg, findErr := s.messages.FindByBizKey(ctx, nil, scheduleBizKey(row.ID)); findErr == nil && msg != nil {
		row.MessageID = msg.ID
	}

	return s.catalog.SaveSchedule(ctx, row)
}

// markScheduleFailure 记录失败原因。未超过上限时按尝试次数退避，避免连续几秒内把额度烧光。
func (s *NotificationService) markScheduleFailure(ctx context.Context, row *model.NotifySchedule, cause error) error {
	row.LastError = truncateRunes(cause.Error(), 512)
	if row.Attempts >= model.ScheduleMaxAttempts {
		row.Status = model.ScheduleStatusFailed
	} else {
		row.Status = model.ScheduleStatusPending
		// 第 n 次失败后等 n 分钟再试，给依赖方恢复的时间
		row.PlanAt = time.Now().Add(time.Duration(row.Attempts) * time.Minute)
	}

	if err := s.catalog.SaveSchedule(ctx, row); err != nil {
		return err
	}
	return cause
}

// truncateRunes 按字符数截断，避免把多字节字符截成半个。
func truncateRunes(text string, limit int) string {
	runes := []rune(text)
	if len(runes) <= limit {
		return text
	}
	return string(runes[:limit])
}
