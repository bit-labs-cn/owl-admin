package contract

import "context"

// Notifier 是业务子系统与通知中心的交互入口。
type Notifier interface {
	// SendDirect 发送定向消息（包含通知与待办，写扩散双表存储）
	SendDirect(ctx context.Context, msg *DirectMessage) error

	// CompleteByBizKey 根据业务主键一键办结待办（销项闭环）
	CompleteByBizKey(ctx context.Context, bizKey string) error

	// BroadcastAnnouncement 发布全员或定向公告（读扩散存储，杜绝大写入锁表）
	BroadcastAnnouncement(ctx context.Context, ann *AnnouncementMessage) error
}
