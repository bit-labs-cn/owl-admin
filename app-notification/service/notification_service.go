package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"bit-labs.cn/owl-admin/app-notification/contract"
	notifEvent "bit-labs.cn/owl-admin/app-notification/event"
	"bit-labs.cn/owl-admin/app-notification/model"
	"bit-labs.cn/owl-admin/app-notification/repository"
	adminsvc "bit-labs.cn/owl-admin/app/service"
	contractEvent "bit-labs.cn/owl/contract/event"
	"bit-labs.cn/owl/contract/log"
	"gorm.io/gorm"
)

// NotificationService 核心通知服务，实现 contract.Notifier。
type NotificationService struct {
	db            *gorm.DB
	bus           contractEvent.EventBus
	messages      repository.MessageRepositoryInterface
	announcements repository.AnnouncementRepositoryInterface
	catalog       repository.CatalogRepositoryInterface
	settings      *adminsvc.SettingService
	hub           *Hub
	logger        log.Logger

	mu       sync.RWMutex
	channels map[string]contract.ChannelSender
}

func NewNotificationService(
	db *gorm.DB,
	bus contractEvent.EventBus,
	messages repository.MessageRepositoryInterface,
	announcements repository.AnnouncementRepositoryInterface,
	catalog repository.CatalogRepositoryInterface,
	settings *adminsvc.SettingService,
	hub *Hub,
	logger log.Logger,
) *NotificationService {
	s := &NotificationService{
		db:            db,
		bus:           bus,
		messages:      messages,
		announcements: announcements,
		catalog:       catalog,
		settings:      settings,
		hub:           hub,
		logger:        logger,
		channels:      make(map[string]contract.ChannelSender),
	}
	return s
}

// RegisterChannel 注册外部渠道插件 (SPI)
func (s *NotificationService) RegisterChannel(sender contract.ChannelSender) {
	if sender == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.channels[sender.ChannelCode()] = sender
}

func (s *NotificationService) getChannel(code string) contract.ChannelSender {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.channels[code]
}

// SendDirect 发送定向消息（通知或待办，写扩散双表落地）。
// 外部渠道写入发送记录，并在同一事务里发布事件，由消费者实际投递。
func (s *NotificationService) SendDirect(ctx context.Context, msg *contract.DirectMessage) error {
	if msg == nil {
		return Invalid("消息不能为空")
	}
	title := strings.TrimSpace(msg.Title)
	userIDs := dedupeUsers(msg.UserIDs)
	externals := normalizeExternals(msg.ExternalRecipients)
	if len(userIDs) == 0 && len(externals) == 0 {
		return Invalid("接收人不能为空")
	}

	kind := msg.Kind
	if kind != contract.KindTodo {
		kind = contract.KindNotice
	}
	typeCode := strings.TrimSpace(msg.TypeCode)
	bizKey := strings.TrimSpace(msg.BizKey)

	var defaults []string
	var tpl *model.NotifyTemplate
	if typeCode != "" {
		typ, err := s.catalog.FindTypeByCode(ctx, typeCode)
		if err != nil {
			return err
		}
		if typ == nil || typ.Status != model.TypeStatusEnabled {
			return NotFound("消息类型不存在或已停用")
		}
		defaults, err = decodeChannels(typ.DefaultChannels)
		if err != nil {
			return Invalid("消息类型的默认渠道无法读取")
		}
		if !msg.SkipTemplate {
			tpl, err = s.catalog.FindTemplate(ctx, typeCode)
			if err != nil {
				return err
			}
		}
	}

	emailOn, wecomOn, _ := s.channelGates(ctx)
	plan := chooseChannels(userIDs, msg.Channels, defaults, typeCode != "", emailOn, wecomOn, func(code string) bool {
		return s.getChannel(code) != nil
	})

	// 正文统一成 HTML 再分发：站内信和邮件直接用它，纯文本渠道由 forChannel 转回文字
	content := ContentHTML(msg.Content)
	rendered, err := renderTemplate(tpl, msg.Vars, title, content, msg.Link)
	if err != nil {
		return Invalid(err.Error())
	}
	inAppCopy := rendered.forChannel(model.ChannelInApp)
	if strings.TrimSpace(inAppCopy.Title) == "" {
		return Invalid("消息标题不能为空；用消息类型时请给类型模板或标题")
	}

	contacts, err := s.catalog.ListContacts(ctx, userIDs)
	if err != nil {
		return err
	}
	contactByID := make(map[uint]repository.UserContact, len(contacts))
	for _, contact := range contacts {
		contactByID[contact.ID] = contact
	}
	sentAt := time.Now()

	inAppUsers := make([]uint, 0, len(plan.Users))
	emailUsers := make([]uint, 0, len(plan.Users))
	var deliveries []model.NotifyDelivery
	for _, user := range plan.Users {
		if user.InApp {
			inAppUsers = append(inAppUsers, user.UserID)
			deliveries = append(deliveries, inAppDelivery(user.UserID, inAppCopy, contactByID[user.UserID], sentAt))
		}
		if user.Email {
			emailUsers = append(emailUsers, user.UserID)
		}
		for _, skip := range user.Skips {
			copy := rendered.forChannel(skip.Channel)
			delivery := skippedDelivery(0, user.UserID, skip.Channel, copy, skip.Reason)
			if contact, ok := contactByID[user.UserID]; ok {
				delivery.RecipientName = contactName(contact)
				delivery.RecipientEmail = contact.Email
			}
			deliveries = append(deliveries, delivery)
		}
	}
	if len(emailUsers) > 0 {
		copy := rendered.forChannel(model.ChannelEmail)
		for _, uid := range emailUsers {
			delivery := pendingDelivery(0, uid, model.ChannelEmail, copy)
			if contact, ok := contactByID[uid]; ok {
				delivery.RecipientName = contactName(contact)
				delivery.RecipientEmail = strings.TrimSpace(contact.Email)
			}
			deliveries = append(deliveries, delivery)
		}
	}
	if plan.WeCom {
		copy := rendered.forChannel(model.ChannelWeCom)
		delivery := pendingDelivery(0, 0, model.ChannelWeCom, copy)
		delivery.RecipientType = model.RecipientTypeBroadcast
		delivery.RecipientName = "群机器人"
		deliveries = append(deliveries, delivery)
	}
	deliveries = append(deliveries, externalDeliveries(externals, rendered, effectiveChannels(msg.Channels, defaults, typeCode != ""), emailOn)...)
	if len(inAppUsers) == 0 && len(deliveries) == 0 {
		return nil
	}
	payloadJSON := toJSON(msg.Payload)
	for i := range deliveries {
		deliveries[i].Payload = payloadJSON
	}

	row := &model.NotifyMessage{
		Kind:     kind,
		TypeCode: typeCode,
		BizKey:   bizKey,
		Title:    inAppCopy.Title,
		Content:  inAppCopy.Content,
		Link:     inAppCopy.Link,
		DueAt:    msg.DueAt,
		SenderID: msg.SenderID,
	}

	var duplicated bool
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if bizKey != "" {
			exist, findErr := s.messages.FindByBizKey(ctx, tx, bizKey)
			if findErr != nil {
				return findErr
			}
			if exist != nil {
				duplicated = true
				return nil
			}
		}
		if err := s.messages.CreateDirect(ctx, tx, row, inAppUsers); err != nil {
			return err
		}
		for i := range deliveries {
			deliveries[i].MessageID = row.ID
		}
		if err := s.catalog.CreateDeliveries(ctx, tx, deliveries); err != nil {
			return err
		}
		for i := range deliveries {
			if deliveries[i].Status != model.DeliveryStatusPending {
				continue
			}
			evt, evtErr := contractEvent.NewEvent(notifEvent.TopicNotifyChannelSend, notifEvent.ChannelSendPayload{
				DeliveryID: deliveries[i].ID,
			}, contractEvent.WithSource("notify"), contractEvent.WithSubject(fmt.Sprintf("delivery:%d", deliveries[i].ID)))
			if evtErr != nil {
				return evtErr
			}
			if err := s.bus.PublishTx(ctx, tx, evt); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	if duplicated {
		return nil
	}

	for _, uid := range inAppUsers {
		s.hub.Publish(ctx, PushEvent{
			PrincipalType: 1,
			PrincipalID:   uid,
			Event:         "inbox",
			Payload: map[string]any{
				"id":    row.ID,
				"title": row.Title,
				"kind":  row.Kind,
			},
		})
	}
	return nil
}

// CompleteByBizKey 根据业务主键办结待办并触发销项事件
func (s *NotificationService) CompleteByBizKey(ctx context.Context, bizKey string) error {
	bizKey = strings.TrimSpace(bizKey)
	if bizKey == "" {
		return nil
	}

	if err := s.messages.CompleteByBizKey(ctx, bizKey); err != nil {
		return err
	}

	// 触发实时推送让在线客户端刷新待办列表与红点
	s.hub.Publish(ctx, PushEvent{
		PrincipalType: 1,
		PrincipalID:   0, // 全员广播刷新
		Event:         "todo",
		Payload: map[string]any{
			"bizKey": bizKey,
		},
	})

	return nil
}

// BroadcastAnnouncement 发布全员或定向公告（读扩散模式，仅插入 1 条主表数据）
func (s *NotificationService) BroadcastAnnouncement(ctx context.Context, ann *contract.AnnouncementMessage) error {
	if ann == nil {
		return errors.New("announcement cannot be nil")
	}
	title := strings.TrimSpace(ann.Title)
	if title == "" {
		return errors.New("announcement title is required")
	}

	scope := ann.Scope
	if scope == "" {
		scope = contract.ScopeAll
	}
	level := ann.Level
	if level <= 0 {
		level = contract.LevelInfo
	}
	displayMode := ann.DisplayMode
	if displayMode <= 0 {
		displayMode = contract.DisplayModeInbox
	}

	targetJSON := ""
	if len(ann.TargetValues) > 0 {
		bytes, _ := json.Marshal(ann.TargetValues)
		targetJSON = string(bytes)
	}

	row := &model.NotifyAnnouncement{
		Title:        title,
		Content:      ann.Content,
		Level:        level,
		DisplayMode:  displayMode,
		TargetScope:  scope,
		TargetValues: targetJSON,
		StartAt:      ann.StartAt,
		EndAt:        ann.EndAt,
		Status:       model.AnnouncementStatusPublished,
		CreatedBy:    ann.SenderID,
	}

	if err := s.announcements.Create(ctx, row); err != nil {
		return err
	}

	// 广播 SSE 告知所有在线浏览器拉取新公告
	s.hub.Publish(ctx, PushEvent{
		PrincipalType: 1,
		PrincipalID:   0,
		Event:         "announcement",
		Payload: map[string]any{
			"id":          row.ID,
			"title":       row.Title,
			"level":       row.Level,
			"displayMode": row.DisplayMode,
		},
	})

	return nil
}

// --- 用户端端点查询逻辑 ---

func (s *NotificationService) MyCounts(ctx context.Context, userID uint) (unreadNotice int64, pendingTodo int64, err error) {
	return s.messages.UnreadCounts(ctx, userID)
}

func (s *NotificationService) MyMessages(ctx context.Context, userID uint, kind int, readStatus *int, typeCode string, page, pageSize int) (int64, []repository.UserMessageDTO, error) {
	total, list, err := s.messages.RetrieveUserMessages(ctx, userID, kind, readStatus, strings.TrimSpace(typeCode), page, pageSize)
	if err != nil {
		return 0, nil, err
	}
	for i := range list {
		if list[i].TypeCode != "" && list[i].TypeName == "" {
			list[i].TypeName = fmt.Sprintf("未知类型(%s)", list[i].TypeCode)
		}
	}
	return total, list, nil
}

func (s *NotificationService) MarkRead(ctx context.Context, userID uint, recipientIDs []uint) error {
	return s.messages.MarkAsRead(ctx, userID, recipientIDs)
}

func (s *NotificationService) MarkAllRead(ctx context.Context, userID uint, kind int) error {
	return s.messages.MarkAllAsRead(ctx, userID, kind)
}

func (s *NotificationService) CompleteByID(ctx context.Context, userID, recipientID uint) error {
	if err := s.messages.CompleteByID(ctx, userID, recipientID); err != nil {
		return err
	}
	s.hub.Publish(ctx, PushEvent{PrincipalType: 1, PrincipalID: userID, Event: "todo"})
	return nil
}

func (s *NotificationService) ReopenByID(ctx context.Context, userID, recipientID uint) error {
	if err := s.messages.ReopenByID(ctx, userID, recipientID); err != nil {
		return err
	}
	s.hub.Publish(ctx, PushEvent{PrincipalType: 1, PrincipalID: userID, Event: "todo"})
	return nil
}

func (s *NotificationService) DeleteUserMessage(ctx context.Context, userID, recipientID uint) error {
	return s.messages.DeleteUserMessage(ctx, userID, recipientID)
}

func (s *NotificationService) MyAnnouncements(ctx context.Context, userID uint, page, pageSize int) (int64, []repository.UserAnnouncementDTO, error) {
	return s.announcements.RetrieveUserAnnouncements(ctx, userID, page, pageSize)
}

func (s *NotificationService) DismissAnnouncement(ctx context.Context, announcementID, userID uint) error {
	return s.announcements.Dismiss(ctx, announcementID, userID)
}

// --- 管理后台公告管理逻辑 ---

func (s *NotificationService) AdminAnnouncements(ctx context.Context, page, pageSize int, keyword string, status int) (int64, []model.NotifyAnnouncement, error) {
	return s.announcements.RetrieveAdmin(ctx, page, pageSize, keyword, status)
}

func (s *NotificationService) AdminAnnouncementDetail(ctx context.Context, id uint) (*model.NotifyAnnouncement, error) {
	return s.announcements.Detail(ctx, id)
}

func (s *NotificationService) SaveAnnouncement(ctx context.Context, ann *model.NotifyAnnouncement) error {
	if ann.ID == 0 {
		return s.announcements.Create(ctx, ann)
	}
	return s.announcements.Update(ctx, ann)
}

func (s *NotificationService) DeleteAnnouncement(ctx context.Context, id uint) error {
	return s.announcements.Delete(ctx, id)
}

// --- 事件总线监听适配 ---

// HandleDirectEvent 监听通用定向消息事件
func (s *NotificationService) HandleDirectEvent(ctx context.Context, evt *notifEvent.DirectMessagePayload) error {
	return s.SendDirect(ctx, evt)
}

// HandleAnnouncementEvent 监听通用公告广播事件
func (s *NotificationService) HandleAnnouncementEvent(ctx context.Context, evt *notifEvent.AnnouncementPayload) error {
	return s.BroadcastAnnouncement(ctx, evt)
}
