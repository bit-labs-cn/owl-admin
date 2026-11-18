/** 与 app-notification/model/const.go、contract 取值严格一致。 */

// 状态
export const StatusEnabled = 1;
export const StatusDisabled = 2;

// 消息类型（1通知，2待办）
export const KindNotice = 1;
export const KindTodo = 2;

// 阅读状态（0未读，1已读）
export const ReadStatusUnread = 0;
export const ReadStatusRead = 1;

// 待办状态（0无，1待处理，2已完成，3已过期）
export const TodoStatusNone = 0;
export const TodoStatusPending = 1;
export const TodoStatusDone = 2;
export const TodoStatusExpired = 3;

// 渠道标识
export const ChannelInApp = "inapp";
export const ChannelEmail = "email";
export const ChannelSMS = "sms";
export const ChannelWeCom = "wecom";
export const ChannelDingTalk = "dingtalk";

export const TypeStatusEnabled = 1;
export const TypeStatusDisabled = 2;

export const DeliveryStatusPending = 1;
export const DeliveryStatusSuccess = 2;
export const DeliveryStatusFailed = 3;
export const DeliveryStatusSkipped = 4;

export const RecipientTypeUser = 1;
export const RecipientTypeExternal = 2;
export const RecipientTypeBroadcast = 3;

// 定时发送任务状态（1待发送，2发送中，3已发送，4已取消，5失败）
export const ScheduleStatusPending = 1;
export const ScheduleStatusSending = 2;
export const ScheduleStatusSent = 3;
export const ScheduleStatusCanceled = 4;
export const ScheduleStatusFailed = 5;

// 公告状态（1草稿，2已发布，3已下线）
export const AnnouncementStatusDraft = 1;
export const AnnouncementStatusPublished = 2;
export const AnnouncementStatusOffline = 3;

// 公告级别（1普通知会，2警告，3紧急）
export const LevelInfo = 1;
export const LevelWarning = 2;
export const LevelDanger = 3;

// 公告展示方式（1收件箱，2顶部横幅，3强提醒弹窗）
export const DisplayModeInbox = 1;
export const DisplayModeBanner = 2;
export const DisplayModePopup = 3;

// 公告受众范围
export const ScopeAll = "all";
export const ScopeDept = "dept";
export const ScopeRole = "role";
