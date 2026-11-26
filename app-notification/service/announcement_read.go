package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"bit-labs.cn/owl-admin/app-notification/model"
	"bit-labs.cn/owl-admin/app-notification/repository"
)

// decodeIDList 解析存成 JSON 数字数组的 ID 列表。
func decodeIDList(raw string) ([]uint, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	var list []uint
	if err := json.Unmarshal([]byte(raw), &list); err != nil {
		return nil, err
	}
	out := make([]uint, 0, len(list))
	for _, id := range list {
		if id > 0 {
			out = append(out, id)
		}
	}

	return out, nil
}

// AdminAnnouncementRead 返回一条公告的已读汇总与每个接收人的状态。
// 公告是读扩散存储，这里把受众解析成用户后与状态表比对，未读即为受众里没有已读记录的人。
func (s *NotificationService) AdminAnnouncementRead(ctx context.Context, id uint) (*repository.AnnouncementReadView, error) {
	ann, err := s.announcements.Detail(ctx, id)
	if err != nil {
		return nil, NotFound("公告不存在")
	}
	values, err := decodeIDList(ann.TargetValues)
	if err != nil {
		return nil, Invalid("公告受众无法读取")
	}

	userIDs, err := s.catalog.ResolveAnnouncementAudience(ctx, ann.TargetScope, values)
	if err != nil {
		return nil, err
	}

	states, err := s.catalog.ListAnnouncementStates(ctx, id)
	if err != nil {
		return nil, err
	}
	stateByUser := make(map[uint]model.NotifyAnnouncementState, len(states))
	for _, state := range states {
		stateByUser[state.UserID] = state
	}

	contacts, err := s.catalog.ListContacts(ctx, dedupeUsers(userIDs))
	if err != nil {
		return nil, err
	}
	contactByID := make(map[uint]repository.UserContact, len(contacts))
	for _, contact := range contacts {
		contactByID[contact.ID] = contact
	}

	view := &repository.AnnouncementReadView{
		AnnouncementID: ann.ID,
		Title:          ann.Title,
		TargetScope:    ann.TargetScope,
		Recipients:     make([]repository.AnnouncementReadRecipient, 0, len(userIDs)),
	}
	for _, uid := range dedupeUsers(userIDs) {
		state, hasState := stateByUser[uid]
		row := repository.AnnouncementReadRecipient{
			UserID:     uid,
			ReadStatus: model.ReadStatusUnread,
		}
		if contact, ok := contactByID[uid]; ok {
			row.UserName = contactName(contact)
			row.Username = contact.Username
			row.Email = contact.Email
		}
		if row.UserName == "" {
			row.UserName = fmt.Sprintf("未知用户(%d)", uid)
		}
		if hasState {
			row.IsDismissed = state.IsDismissed
			if state.ReadAt != nil {
				row.ReadStatus = model.ReadStatusRead
				row.ReadAt = state.ReadAt
			}
			if state.IsDismissed == 1 {
				view.DismissedCount++
			}
		}
		if row.ReadStatus == model.ReadStatusRead {
			view.ReadCount++
		} else {
			view.UnreadCount++
		}
		view.Recipients = append(view.Recipients, row)
	}
	view.RecipientCount = len(view.Recipients)
	return view, nil
}
