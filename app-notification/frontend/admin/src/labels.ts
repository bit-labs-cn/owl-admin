import {
  ChannelDingTalk,
  ChannelEmail,
  ChannelInApp,
  ChannelSMS,
  ChannelWeCom,
  DeliveryStatusFailed,
  DeliveryStatusPending,
  DeliveryStatusSkipped,
  DeliveryStatusSuccess,
  RecipientTypeBroadcast,
  RecipientTypeExternal,
  RecipientTypeUser,
  TypeStatusDisabled,
  TypeStatusEnabled,
  AnnouncementStatusDraft,
  AnnouncementStatusOffline,
  AnnouncementStatusPublished,
  DisplayModeBanner,
  DisplayModeInbox,
  DisplayModePopup,
  KindNotice,
  KindTodo,
  LevelDanger,
  LevelInfo,
  LevelWarning,
  ReadStatusRead,
  ReadStatusUnread,
  ScheduleStatusCanceled,
  ScheduleStatusFailed,
  ScheduleStatusPending,
  ScheduleStatusSending,
  ScheduleStatusSent,
  ScopeAll,
  ScopeDept,
  ScopeRole,
  TodoStatusDone,
  TodoStatusExpired,
  TodoStatusPending
} from "./constants";

export interface EnumOption<T extends string | number = number> {
  value: T;
  label: string;
  tag?: "success" | "warning" | "info" | "danger" | "primary";
}

export const kindOptions: EnumOption[] = [
  { value: KindNotice, label: "系统通知", tag: "info" },
  { value: KindTodo, label: "待办事项", tag: "warning" }
];

export const readStatusOptions: EnumOption[] = [
  { value: ReadStatusUnread, label: "未读", tag: "danger" },
  { value: ReadStatusRead, label: "已读", tag: "success" }
];

export const todoStatusOptions: EnumOption[] = [
  { value: TodoStatusPending, label: "待处理", tag: "warning" },
  { value: TodoStatusDone, label: "已完成", tag: "success" },
  { value: TodoStatusExpired, label: "已过期", tag: "info" }
];

export const announcementStatusOptions: EnumOption[] = [
  { value: AnnouncementStatusDraft, label: "草稿", tag: "info" },
  { value: AnnouncementStatusPublished, label: "已发布", tag: "success" },
  { value: AnnouncementStatusOffline, label: "已下线", tag: "danger" }
];

export const levelOptions: EnumOption[] = [
  { value: LevelInfo, label: "普通知会", tag: "info" },
  { value: LevelWarning, label: "重要提醒", tag: "warning" },
  { value: LevelDanger, label: "紧急通知", tag: "danger" }
];

export const displayModeOptions: EnumOption[] = [
  { value: DisplayModeInbox, label: "仅收件箱" },
  { value: DisplayModeBanner, label: "顶部横幅" },
  { value: DisplayModePopup, label: "弹窗强提示" }
];

export const typeStatusOptions: EnumOption[] = [
  { value: TypeStatusEnabled, label: "启用", tag: "success" },
  { value: TypeStatusDisabled, label: "停用", tag: "info" }
];

export const channelOptions: EnumOption<string>[] = [
  { value: ChannelInApp, label: "站内信" },
  { value: ChannelEmail, label: "邮件" },
  { value: ChannelSMS, label: "短信" },
  { value: ChannelWeCom, label: "企业微信" },
  { value: ChannelDingTalk, label: "钉钉" }
];

export const recipientTypeOptions: EnumOption[] = [
  { value: RecipientTypeUser, label: "站内用户", tag: "primary" },
  { value: RecipientTypeExternal, label: "外部人员", tag: "warning" },
  { value: RecipientTypeBroadcast, label: "群机器人", tag: "info" }
];

export const deliveryStatusOptions: EnumOption[] = [  { value: DeliveryStatusPending, label: "待发送", tag: "warning" },
  { value: DeliveryStatusSuccess, label: "成功", tag: "success" },
  { value: DeliveryStatusFailed, label: "失败", tag: "danger" },
  { value: DeliveryStatusSkipped, label: "跳过", tag: "info" }
];

export const scheduleStatusOptions: EnumOption[] = [
  { value: ScheduleStatusPending, label: "待发送", tag: "warning" },
  { value: ScheduleStatusSending, label: "发送中", tag: "primary" },
  { value: ScheduleStatusSent, label: "已发送", tag: "success" },
  { value: ScheduleStatusCanceled, label: "已取消", tag: "info" },
  { value: ScheduleStatusFailed, label: "发送失败", tag: "danger" }
];

export const scopeOptions: EnumOption<string>[] = [
  { value: ScopeAll, label: "全体人员" },
  { value: ScopeRole, label: "按角色指定" },
  { value: ScopeDept, label: "按部门指定" }
];

export function labelOf<T extends string | number>(options: EnumOption<T>[], value: T | undefined): string {
  if (value === undefined || value === null) return "未知";
  const item = options.find(opt => opt.value === value);
  return item ? item.label : String(value);
}

export function tagOf<T extends string | number>(
  options: EnumOption<T>[],
  value: T | undefined
): "success" | "warning" | "info" | "danger" | "primary" | undefined {
  if (value === undefined || value === null) return undefined;
  const item = options.find(opt => opt.value === value);
  return item?.tag;
}
