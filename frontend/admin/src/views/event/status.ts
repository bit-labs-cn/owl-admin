/** 发件箱状态，取值与 owl/provider/event/model.go 的 OutboxStatus* 保持一致 */
export const OutboxStatus = {
  /** 待发送 */
  Pending: 0,
  /** 已发送 */
  Sent: 1,
  /** 投递失败 */
  Failed: 2
} as const;

export type OutboxStatusValue = (typeof OutboxStatus)[keyof typeof OutboxStatus];

export const EventPriority = {
  Normal: 0,
  High: 1
} as const;

export function eventPriorityLabel(priority?: number | null): string {
  return priority === EventPriority.High ? "高优先级" : "普通";
}

export function eventPriorityTag(priority?: number | null): string {
  return priority === EventPriority.High ? "danger" : "info";
}

/** 收件箱状态，取值与 owl/provider/event/model.go 的 InboxStatus* 保持一致 */
export const InboxStatus = {
  /** 消费成功 */
  Success: 1,
  /** 消费失败 */
  Failed: 2
} as const;

export type InboxStatusValue = (typeof InboxStatus)[keyof typeof InboxStatus];

const outboxLabel: Record<number, string> = {
  [OutboxStatus.Pending]: "待发送",
  [OutboxStatus.Sent]: "已发送",
  [OutboxStatus.Failed]: "投递失败"
};

const outboxTag: Record<number, string> = {
  [OutboxStatus.Pending]: "warning",
  [OutboxStatus.Sent]: "success",
  [OutboxStatus.Failed]: "danger"
};

const inboxLabel: Record<number, string> = {
  [InboxStatus.Success]: "消费成功",
  [InboxStatus.Failed]: "消费失败"
};

const inboxTag: Record<number, string> = {
  [InboxStatus.Success]: "success",
  [InboxStatus.Failed]: "danger"
};

/** 发件箱状态文案 */
export function outboxStatusLabel(status?: number | null): string {
  if (status == null) return "—";
  return outboxLabel[status] ?? `未知(${status})`;
}

/** 发件箱状态下 el-tag 类型 */
export function outboxStatusTag(status?: number | null): string {
  if (status == null) return "info";
  return outboxTag[status] ?? "info";
}

/** 收件箱状态文案 */
export function inboxStatusLabel(status?: number | null): string {
  if (status == null) return "—";
  return inboxLabel[status] ?? `未知(${status})`;
}

/** 收件箱状态下 el-tag 类型 */
export function inboxStatusTag(status?: number | null): string {
  if (status == null) return "info";
  return inboxTag[status] ?? "info";
}

/** 发件箱状态筛选项 */
export const outboxStatusOptions = [
  { label: "待发送", value: OutboxStatus.Pending },
  { label: "已发送", value: OutboxStatus.Sent },
  { label: "投递失败", value: OutboxStatus.Failed }
];

/** 收件箱状态筛选项 */
export const inboxStatusOptions = [
  { label: "消费成功", value: InboxStatus.Success },
  { label: "消费失败", value: InboxStatus.Failed }
];
