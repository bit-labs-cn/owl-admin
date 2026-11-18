package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"bit-labs.cn/owl-admin/app-notification/contract"
	notifEvent "bit-labs.cn/owl-admin/app-notification/event"
	"bit-labs.cn/owl-admin/app-notification/model"
	"bit-labs.cn/owl-admin/app-notification/repository"
	contractEvent "bit-labs.cn/owl/contract/event"
	"bit-labs.cn/owl/provider/redis"
	"bit-labs.cn/owl/provider/router"
	validatorv10 "github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/spf13/cast"
	"gorm.io/gorm"
)

const (
	settingApproval   = "approvalEnabled"
	settingQuiet      = "quietHours"
	settingRate       = "rateLimitPerDay"
	settingInboxDays  = "inboxRetentionDays"
	settingDeliverDay = "deliveryRetentionDays"
	maxRetry          = 5
)

// CreateTaskReq 是后台手工创建通知的入参。
type CreateTaskReq struct {
	Title          string             `json:"title" validate:"required,max=255" label:"标题"`
	EventCode      string             `json:"eventCode" validate:"omitempty,max=64" label:"事件编码"`
	Source         string             `json:"source" validate:"omitempty,max=64" label:"来源"`
	Category       int                `json:"category" validate:"required,min=1,max=4" label:"分类"`
	Kind           int                `json:"kind" validate:"required,min=1,max=3" label:"类型"`
	Channels       []string           `json:"channels" validate:"required,min=1" label:"渠道"`
	Audience       contract.Audience  `json:"audience" validate:"required" label:"受众"`
	TemplateCode   string             `json:"templateCode" validate:"omitempty,max=64" label:"模板编码"`
	ContentHTML    string             `json:"contentHtml" validate:"omitempty" label:"正文"`
	ContentText    string             `json:"contentText" validate:"omitempty" label:"纯文本"`
	Variables      map[string]string  `json:"variables" label:"变量"`
	Payload        map[string]string  `json:"payload" label:"附加内容"`
	Link           *contract.DeepLink `json:"link" label:"跳转"`
	Actions        []contract.Action  `json:"actions" label:"动作"`
	SMSTemplate    string             `json:"smsTemplate" validate:"omitempty,max=64" label:"短信模板"`
	BizKey         string             `json:"bizKey" validate:"omitempty,max=128" label:"业务主键"`
	DueAt          *time.Time         `json:"dueAt" label:"截止时间"`
	ScheduleAt     *time.Time         `json:"scheduleAt" label:"定时发送"`
	IdempotencyKey string             `json:"idempotencyKey" validate:"omitempty,max=128" label:"幂等键"`
	Origin         string             `json:"origin" validate:"omitempty,max=255" label:"站点地址"`
	Submit         bool               `json:"submit" label:"立即发送"`
}

// RetrieveTaskReq 是任务列表查询。
type RetrieveTaskReq struct {
	router.PageReq
	Title  string `json:"title" form:"title" validate:"omitempty,max=255" label:"标题"`
	Kind   int    `json:"kind" form:"kind" validate:"omitempty,min=1,max=3" label:"类型"`
	Status int    `json:"status" form:"status" validate:"omitempty,min=1,max=8" label:"状态"`
}

// TaskService 负责创建、预估和分发通知任务。
type TaskService struct {
	tasks      repository.TaskRepositoryInterface
	templates  repository.TemplateRepositoryInterface
	deliveries repository.DeliveryRepositoryInterface
	inbox      repository.InboxRepositoryInterface
	devices    repository.DeviceRepositoryInterface
	accounts   repository.ChannelAccountRepositoryInterface
	subs       repository.SubscriptionRepositoryInterface
	settings   repository.SettingRepositoryInterface
	audience   *AudienceResolver
	hub        *Hub
	email      *EmailChannel
	sms        *SMSChannel
	app        *AppChannel
	locker     redis.LockerFactory
	validate   *validatorv10.Validate
	bus        contractEvent.EventBus
	db         *gorm.DB
}

func NewTaskService(
	tasks repository.TaskRepositoryInterface,
	templates repository.TemplateRepositoryInterface,
	deliveries repository.DeliveryRepositoryInterface,
	inbox repository.InboxRepositoryInterface,
	devices repository.DeviceRepositoryInterface,
	accounts repository.ChannelAccountRepositoryInterface,
	subs repository.SubscriptionRepositoryInterface,
	settings repository.SettingRepositoryInterface,
	audience *AudienceResolver,
	hub *Hub,
	email *EmailChannel,
	sms *SMSChannel,
	app *AppChannel,
	locker redis.LockerFactory,
	validate *validatorv10.Validate,
	bus contractEvent.EventBus,
	db *gorm.DB,
) *TaskService {
	return &TaskService{
		tasks: tasks, templates: templates, deliveries: deliveries, inbox: inbox, devices: devices,
		accounts: accounts, subs: subs, settings: settings, audience: audience, hub: hub,
		email: email, sms: sms, app: app, locker: locker, validate: validate,
		bus: bus, db: db,
	}
}

func (s *TaskService) Create(ctx context.Context, req *CreateTaskReq) (*model.Task, error) {
	if err := s.validate.Struct(req); err != nil {
		return nil, err
	}
	if req.IdempotencyKey != "" {
		old, err := s.tasks.WithContext(ctx).FindByIdempotency(req.IdempotencyKey)
		if err == nil {
			return old, nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}
	row, err := s.build(ctx, req)
	if err != nil {
		return nil, err
	}
	l := s.locker.New()
	if err := l.Lock("notify:task:create"); err != nil {
		return nil, err
	}
	defer l.Unlock()
	if err := s.tasks.WithContext(ctx).Create(row); err != nil {
		return nil, err
	}
	if req.Submit {
		if err := s.Submit(ctx, row.ID); err != nil {
			return row, err
		}
		return s.tasks.WithContext(ctx).Detail(row.ID)
	}
	return row, nil
}

func (s *TaskService) SendMessage(ctx context.Context, msg *contract.Message) (*model.Task, error) {
	if msg == nil {
		return nil, Invalid("通知内容为空")
	}
	req := &CreateTaskReq{
		Title: msg.Title, EventCode: msg.EventCode, Source: msg.Source, Category: msg.Category, Kind: msg.Kind,
		Channels: msg.Channels, Audience: msg.Audience, TemplateCode: msg.TemplateCode, ContentHTML: msg.ContentHTML,
		ContentText: msg.ContentText, Variables: msg.Variables, Payload: msg.Payload, Link: msg.Link, Actions: msg.Actions,
		BizKey: msg.BizKey, DueAt: msg.DueAt, ScheduleAt: msg.ScheduleAt, IdempotencyKey: msg.IdempotencyKey, Submit: msg.ScheduleAt == nil,
	}
	if req.Category == 0 {
		req.Category = model.CategoryBiz
	}
	if req.Kind == 0 {
		req.Kind = model.KindMessage
	}
	if len(req.Channels) == 0 {
		req.Channels = []string{model.ChannelInApp}
	}
	return s.Create(ctx, req)
}

func (s *TaskService) build(ctx context.Context, req *CreateTaskReq) (*model.Task, error) {
	title := strings.TrimSpace(req.Title)
	htmlBody := SanitizeHTML(req.ContentHTML)
	textBody := strings.TrimSpace(req.ContentText)
	smsCode := strings.TrimSpace(req.SMSTemplate)
	link := req.Link
	actions := req.Actions
	if req.TemplateCode != "" {
		tpl, err := s.templates.WithContext(ctx).FindByCode(req.TemplateCode)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, NotFound("通知模板不存在")
		}
		if err != nil {
			return nil, err
		}
		if tpl.Status != model.StatusEnabled {
			return nil, NotAllowed("通知模板已停用")
		}
		if title == "" {
			title = applyVars(tpl.Title, req.Variables)
		}
		if htmlBody == "" {
			htmlBody = SanitizeHTML(applyVars(tpl.ContentHTML, req.Variables))
		}
		if textBody == "" {
			textBody = applyVars(tpl.ContentText, req.Variables)
		}
		if smsCode == "" {
			smsCode = tpl.SMSTemplate
		}
		if link == nil && tpl.LinkJSON != "" {
			link = &contract.DeepLink{}
			fromJSON(tpl.LinkJSON, link)
		}
		if len(actions) == 0 && tpl.ActionJSON != "" {
			fromJSON(tpl.ActionJSON, &actions)
		}
	}
	if textBody == "" {
		textBody = HTMLToText(htmlBody)
	}
	if title == "" {
		return nil, Invalid("标题不能为空")
	}
	status := model.TaskStatusDraft
	if req.ScheduleAt != nil {
		status = model.TaskStatusPending
	}
	return &model.Task{
		BizNo: uuid.NewString(), Title: title, EventCode: req.EventCode, Source: req.Source, Category: req.Category,
		Kind: req.Kind, ChannelsJSON: toJSON(req.Channels), AudienceJSON: toJSON(req.Audience), TemplateCode: req.TemplateCode,
		ContentHTML: AbsolutizeHTML(req.Origin, htmlBody), ContentText: textBody, VariableJSON: toJSON(req.Variables),
		PayloadJSON: toJSON(req.Payload), LinkJSON: toJSON(link), ActionJSON: toJSON(actions), SMSTemplate: smsCode,
		BizKey: req.BizKey, DueAt: req.DueAt, ScheduleAt: req.ScheduleAt, Status: status, IdempotencyKey: req.IdempotencyKey, Origin: req.Origin,
	}, nil
}

func (s *TaskService) Update(ctx context.Context, id uint, req *CreateTaskReq) error {
	if err := s.validate.Struct(req); err != nil {
		return err
	}
	row, err := s.tasks.WithContext(ctx).Detail(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return NotFound("通知任务不存在")
	}
	if err != nil {
		return err
	}
	if row.Status != model.TaskStatusDraft && row.Status != model.TaskStatusPendingReview {
		return NotAllowed("只有草稿和待审核任务可以修改")
	}
	next, err := s.build(ctx, req)
	if err != nil {
		return err
	}
	next.ID = row.ID
	next.BizNo = row.BizNo
	next.Status = row.Status
	next.BaseModel = row.BaseModel
	return s.tasks.WithContext(ctx).Save(next)
}

func (s *TaskService) Delete(ctx context.Context, id uint) error {
	row, err := s.tasks.WithContext(ctx).Detail(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return NotFound("通知任务不存在")
	}
	if err != nil {
		return err
	}
	if row.Status == model.TaskStatusSending {
		return NotAllowed("发送中的任务不能删除")
	}
	return s.tasks.WithContext(ctx).Delete(id)
}

func (s *TaskService) Detail(ctx context.Context, id uint) (*model.Task, error) {
	row, err := s.tasks.WithContext(ctx).Detail(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, NotFound("通知任务不存在")
	}
	return row, err
}

func (s *TaskService) Retrieve(ctx context.Context, req *RetrieveTaskReq) (int64, []model.Task, error) {
	if err := s.validate.Struct(req); err != nil {
		return 0, nil, err
	}
	return s.tasks.WithContext(ctx).Retrieve(req.Page, req.PageSize, req.Title, req.Kind, req.Status)
}

// Titles 返回任务标题，供投递记录这类关联列表显示可读名称。
func (s *TaskService) Titles(ctx context.Context, ids []uint) (map[uint]string, error) {
	return s.tasks.WithContext(ctx).Titles(ids)
}

func (s *TaskService) Estimate(ctx context.Context, audience contract.Audience) (AudienceEstimate, error) {
	_, est, err := s.audience.Resolve(ctx, audience)
	return est, err
}

func (s *TaskService) Submit(ctx context.Context, id uint) error {
	l := s.locker.New()
	if err := l.Lock("notify:task:submit:" + cast.ToString(id)); err != nil {
		return err
	}
	defer l.Unlock()
	row, err := s.tasks.WithContext(ctx).Detail(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return NotFound("通知任务不存在")
	}
	if err != nil {
		return err
	}
	if row.Status == model.TaskStatusSuccess || row.Status == model.TaskStatusCancelled {
		return NotAllowed("任务当前状态不能发送")
	}
	if row.ScheduleAt != nil && row.ScheduleAt.After(time.Now()) {
		row.Status = model.TaskStatusPending
		return s.tasks.WithContext(ctx).Save(row)
	}
	var audience contract.Audience
	fromJSON(row.AudienceJSON, &audience)
	if s.approvalOn(ctx) && audience.All && row.Status != model.TaskStatusPendingReview {
		row.Status = model.TaskStatusPendingReview
		return s.tasks.WithContext(ctx).Save(row)
	}
	row.Status = model.TaskStatusSending
	if err := s.tasks.WithContext(ctx).Save(row); err != nil {
		return err
	}

	// 异步削峰：通过事件总线发布任务提交事件，交由后台流式消费受众解析与渠道扇出
	if s.bus != nil {
		evt, err := contractEvent.NewEvent(notifEvent.TopicTaskSubmit, notifEvent.TaskSubmitPayload{
			TaskID: row.ID,
		}, contractEvent.WithSource("owl-notification"), contractEvent.WithSubject(fmt.Sprintf("task:%d", row.ID)), contractEvent.WithPriority(contractEvent.PriorityHigh))
		if err != nil {
			return err
		}
		if err := s.bus.PublishTx(ctx, s.db, evt); err != nil {
			return err
		}
		return nil
	}

	// 无事件总线时回退同步扇出
	return s.FanoutAsync(ctx, row.ID)
}

// FanoutAsync 由事件总线消费者调用，异步解析受众、写入站内信并扇出各渠道投递事件。
func (s *TaskService) FanoutAsync(ctx context.Context, taskID uint) error {
	row, err := s.tasks.WithContext(ctx).Detail(taskID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if row.Status == model.TaskStatusCancelled {
		return nil
	}

	var audience contract.Audience
	fromJSON(row.AudienceJSON, &audience)

	recipients, _, err := s.audience.Resolve(ctx, audience)
	if err != nil {
		row.Status = model.TaskStatusFailed
		_ = s.tasks.WithContext(ctx).Save(row)
		return err
	}

	var channels []string
	fromJSON(row.ChannelsJSON, &channels)

	for _, person := range recipients {
		s.fanout(ctx, row, person, channels)
	}

	return s.refreshStatus(ctx, row.ID)
}

func (s *TaskService) Approve(ctx context.Context, id uint) error {
	row, err := s.tasks.WithContext(ctx).Detail(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return NotFound("通知任务不存在")
	}
	if err != nil {
		return err
	}
	if row.Status != model.TaskStatusPendingReview {
		return NotAllowed("任务不在待审核状态")
	}
	row.Status = model.TaskStatusPending
	if err := s.tasks.WithContext(ctx).Save(row); err != nil {
		return err
	}
	return s.Submit(ctx, id)
}

func (s *TaskService) Cancel(ctx context.Context, id uint) error {
	row, err := s.tasks.WithContext(ctx).Detail(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return NotFound("通知任务不存在")
	}
	if err != nil {
		return err
	}
	row.Status = model.TaskStatusCancelled
	return s.tasks.WithContext(ctx).Save(row)
}

func (s *TaskService) fanout(ctx context.Context, task *model.Task, person Recipient, channels []string) {
	for _, channel := range channels {
		if !s.allow(ctx, task, person, channel) {
			s.writeDelivery(ctx, task, person, channel, "", model.DeliveryStatusSkipped, "用户已关闭该渠道")
			continue
		}
		switch channel {
		case model.ChannelInApp:
			if person.Type != model.RecipientTypeUser {
				s.writeDelivery(ctx, task, person, channel, "", model.DeliveryStatusSkipped, "外部联系人没有站内信")
				continue
			}
			s.writeInbox(ctx, task, person)
		case model.ChannelEmail:
			if person.Email == "" {
				s.writeDelivery(ctx, task, person, channel, "", model.DeliveryStatusSkipped, "没有邮箱")
				continue
			}
			s.writeDelivery(ctx, task, person, channel, person.Email, model.DeliveryStatusPending, "")
		case model.ChannelSMS:
			if person.Phone == "" {
				s.writeDelivery(ctx, task, person, channel, "", model.DeliveryStatusSkipped, "没有手机号")
				continue
			}
			s.writeDelivery(ctx, task, person, channel, person.Phone, model.DeliveryStatusPending, "")
		case model.ChannelApp:
			if person.Type != model.RecipientTypeUser {
				s.writeDelivery(ctx, task, person, channel, "", model.DeliveryStatusSkipped, "外部联系人没有 App")
				continue
			}
			s.writeDelivery(ctx, task, person, channel, "", model.DeliveryStatusPending, "")
		default:
			s.writeDelivery(ctx, task, person, channel, "", model.DeliveryStatusSkipped, "未知渠道")
		}
	}
}

func (s *TaskService) allow(ctx context.Context, task *model.Task, person Recipient, channel string) bool {
	if task.Category == model.CategorySecurity || channel == model.ChannelInApp {
		return true
	}
	if person.Type != model.RecipientTypeUser {
		return true
	}
	ok, err := s.subs.WithContext(ctx).Allows(model.PrincipalTypeAdmin, person.ID, task.Category, channel)
	return err != nil || ok
}

func (s *TaskService) writeInbox(ctx context.Context, task *model.Task, person Recipient) {
	todo := 0
	if task.Kind == model.KindTodo {
		todo = model.TodoStatusPending
	}
	row := &model.Inbox{
		PrincipalType: model.PrincipalTypeAdmin, PrincipalID: person.ID, TaskID: task.ID, Source: task.Source,
		Kind: task.Kind, Category: task.Category, Title: task.Title, ContentHTML: task.ContentHTML, ContentText: task.ContentText,
		PayloadJSON: task.PayloadJSON, LinkJSON: task.LinkJSON, ActionJSON: task.ActionJSON, BizKey: task.BizKey,
		TodoStatus: todo, DueAt: task.DueAt, ReadStatus: model.ReadStatusUnread,
	}
	if _, err := s.inbox.WithContext(ctx).FindByPrincipalTask(row.PrincipalType, row.PrincipalID, row.TaskID); err == nil {
		s.writeDelivery(ctx, task, person, model.ChannelInApp, person.Name, model.DeliveryStatusSent, "")
		return
	}
	if err := s.inbox.WithContext(ctx).Create(row); err != nil {
		s.writeDelivery(ctx, task, person, model.ChannelInApp, person.Name, model.DeliveryStatusFailed, err.Error())
		return
	}
	s.writeDelivery(ctx, task, person, model.ChannelInApp, person.Name, model.DeliveryStatusSent, "")
	s.hub.Publish(ctx, PushEvent{PrincipalType: row.PrincipalType, PrincipalID: row.PrincipalID, Event: "inbox", Payload: map[string]any{"id": cast.ToString(row.ID), "title": row.Title, "kind": row.Kind}})
}

func (s *TaskService) writeDelivery(ctx context.Context, task *model.Task, person Recipient, channel, address string, status int, reason string) {
	row := &model.Delivery{
		TaskID: task.ID, RecipientType: person.Type, RecipientID: person.ID, RecipientName: person.Name,
		Channel: channel, Address: address, Status: status, ErrorMsg: trimErr(reason),
	}
	if status == model.DeliveryStatusSent {
		now := time.Now()
		row.SentAt = &now
	}
	if status == model.DeliveryStatusPending {
		if next := s.quietUntil(ctx, time.Now()); next != nil {
			row.NextRetryAt = next
		}
	}
	_ = s.deliveries.WithContext(ctx).Create(row)

	// 如果处于待发送状态且不在免打扰时段，通过事件总线触发各渠道异步消费发送
	if row.Status == model.DeliveryStatusPending && row.NextRetryAt == nil && s.bus != nil {
		var topic string
		switch channel {
		case model.ChannelEmail:
			topic = notifEvent.TopicDispatchEmail
		case model.ChannelSMS:
			topic = notifEvent.TopicDispatchSMS
		case model.ChannelApp:
			topic = notifEvent.TopicDispatchApp
		}
		if topic != "" {
			evt, err := contractEvent.NewEvent(topic, notifEvent.DispatchPayload{
				DeliveryID:    row.ID,
				TaskID:        task.ID,
				RecipientType: person.Type,
				RecipientID:   person.ID,
				Channel:       channel,
				Address:       address,
			}, contractEvent.WithSource("owl-notification"), contractEvent.WithSubject(fmt.Sprintf("delivery:%d", row.ID)), contractEvent.WithPriority(contractEvent.PriorityHigh))
			if err == nil {
				_ = s.bus.PublishTx(ctx, s.db, evt)
			}
		}
	}
}

func (s *TaskService) refreshStatus(ctx context.Context, taskID uint) error {
	_, rows, err := s.deliveries.WithContext(ctx).Retrieve(1, 10000, taskID, "", 0)
	if err != nil {
		return err
	}
	task, err := s.tasks.WithContext(ctx).Detail(taskID)
	if err != nil {
		return err
	}
	success, fail, skip, pending := 0, 0, 0, 0
	for _, row := range rows {
		switch row.Status {
		case model.DeliveryStatusSent:
			success++
		case model.DeliveryStatusFailed:
			fail++
		case model.DeliveryStatusSkipped:
			skip++
		default:
			pending++
		}
	}
	task.SuccessCount, task.FailCount, task.SkipCount = success, fail, skip
	switch {
	case pending > 0:
		task.Status = model.TaskStatusSending
	case success > 0 && fail > 0:
		task.Status = model.TaskStatusPartial
	case success > 0:
		task.Status = model.TaskStatusSuccess
	case fail > 0:
		task.Status = model.TaskStatusFailed
	default:
		task.Status = model.TaskStatusSuccess
	}
	return s.tasks.WithContext(ctx).Save(task)
}

func (s *TaskService) approvalOn(ctx context.Context) bool {
	v, ok, err := s.settings.WithContext(ctx).Get(settingApproval)
	return err == nil && ok && (v == "1" || v == "true")
}

func (s *TaskService) quietUntil(ctx context.Context, now time.Time) *time.Time {
	raw, ok, err := s.settings.WithContext(ctx).Get(settingQuiet)
	if err != nil || !ok || !strings.Contains(raw, "-") {
		return nil
	}
	parts := strings.Split(raw, "-")
	if len(parts) != 2 {
		return nil
	}
	start, err1 := time.Parse("15:04", strings.TrimSpace(parts[0]))
	end, err2 := time.Parse("15:04", strings.TrimSpace(parts[1]))
	if err1 != nil || err2 != nil {
		return nil
	}
	cur := time.Date(0, 1, 1, now.Hour(), now.Minute(), 0, 0, time.UTC)
	ss := time.Date(0, 1, 1, start.Hour(), start.Minute(), 0, 0, time.UTC)
	ee := time.Date(0, 1, 1, end.Hour(), end.Minute(), 0, 0, time.UTC)
	inside := false
	if ss.Before(ee) {
		inside = !cur.Before(ss) && cur.Before(ee)
	} else {
		inside = !cur.Before(ss) || cur.Before(ee)
	}
	if !inside {
		return nil
	}
	next := time.Date(now.Year(), now.Month(), now.Day(), end.Hour(), end.Minute(), 0, 0, now.Location())
	if !next.After(now) {
		next = next.Add(24 * time.Hour)
	}
	return &next
}

func applyVars(content string, vars map[string]string) string {
	if content == "" || len(vars) == 0 {
		return content
	}
	out := content
	for k, v := range vars {
		out = strings.ReplaceAll(out, "${"+k+"}", v)
	}
	return out
}

func maskAddress(s string) string {
	if strings.Contains(s, "@") {
		parts := strings.SplitN(s, "@", 2)
		name := parts[0]
		if len(name) > 2 {
			name = name[:1] + "***" + name[len(name)-1:]
		}
		return name + "@" + parts[1]
	}
	r := []rune(s)
	if len(r) < 7 {
		return s
	}
	return string(r[:3]) + "****" + string(r[len(r)-4:])
}

func trimErr(s string) string {
	r := []rune(s)
	if len(r) > 500 {
		return string(r[:500])
	}
	return s
}
