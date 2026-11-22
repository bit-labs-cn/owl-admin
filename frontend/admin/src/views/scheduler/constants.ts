import dayjs from "dayjs";

/** 任务来源，与后端 JobSource* 一致 */
export const SchedulerJobSource = {
  Builtin: 1,
  Custom: 2
} as const;

/** 任务状态，与后端 JobStatus* 一致 */
export const SchedulerJobStatus = {
  Enabled: 1,
  Disabled: 2,
  Retired: 3
} as const;

/** 执行方式，与 contract/scheduler 的 ScriptType 一致 */
export const SchedulerScriptType = {
  Handler: 1,
  HTTP: 2,
  Shell: 3,
  /** 调用子系统登记的可调用方法 */
  Action: 4
} as const;

/** 并发策略，与 contract/scheduler 的 OverlapPolicy 一致 */
export const SchedulerOverlap = {
  Skip: 1,
  Allow: 2
} as const;

/** 执行记录状态，与后端 RunStatus* 一致 */
export const SchedulerRunStatus = {
  Pending: 1,
  Running: 2,
  Success: 3,
  Failed: 4,
  Canceled: 5,
  Timeout: 6,
  Skipped: 7
} as const;

/** 触发方式，与 contract/scheduler 的 TriggerType 一致 */
export const SchedulerTriggerType = {
  Cron: 1,
  Manual: 2,
  Retry: 3
} as const;

export const jobSourceOptions = [
  { label: "内置", value: SchedulerJobSource.Builtin },
  { label: "自定义", value: SchedulerJobSource.Custom }
];

export const jobStatusOptions = [
  { label: "启用", value: SchedulerJobStatus.Enabled },
  { label: "停用", value: SchedulerJobStatus.Disabled },
  { label: "已下线", value: SchedulerJobStatus.Retired }
];

export const runStatusOptions = [
  { label: "待执行", value: SchedulerRunStatus.Pending },
  { label: "执行中", value: SchedulerRunStatus.Running },
  { label: "成功", value: SchedulerRunStatus.Success },
  { label: "失败", value: SchedulerRunStatus.Failed },
  { label: "已取消", value: SchedulerRunStatus.Canceled },
  { label: "超时", value: SchedulerRunStatus.Timeout },
  { label: "已跳过", value: SchedulerRunStatus.Skipped }
];

export const triggerTypeOptions = [
  { label: "定时", value: SchedulerTriggerType.Cron },
  { label: "手动", value: SchedulerTriggerType.Manual },
  { label: "重试", value: SchedulerTriggerType.Retry }
];

export const scriptTypeOptions = [
  { label: "调用子系统方法", value: SchedulerScriptType.Action },
  { label: "HTTP 请求", value: SchedulerScriptType.HTTP },
  { label: "Shell 命令", value: SchedulerScriptType.Shell }
];

export const overlapOptions = [
  { label: "跳过本次", value: SchedulerOverlap.Skip },
  { label: "允许并发", value: SchedulerOverlap.Allow }
];

/** 子系统名称映射。可调用方法按子系统分组展示，未收录的子系统直接显示 appName。 */
export const APP_TITLES: Record<string, string> = {
  admin: "系统管理",
  notify: "通知中心",
  cms: "内容管理",
  coordinator: "vlink 协调器",
  forge: "Forge 流水线",
  sms: "短信平台",
  inspection: "巡检",
  workorder: "工单",
  knowledge: "知识库",
  mall: "商城",
  portal: "门户",
  device: "设备管理",
  warning: "告警",
  metrics: "指标"
};

const runStatusLabel: Record<number, string> = {
  [SchedulerRunStatus.Pending]: "待执行",
  [SchedulerRunStatus.Running]: "执行中",
  [SchedulerRunStatus.Success]: "成功",
  [SchedulerRunStatus.Failed]: "失败",
  [SchedulerRunStatus.Canceled]: "已取消",
  [SchedulerRunStatus.Timeout]: "超时",
  [SchedulerRunStatus.Skipped]: "已跳过"
};

const runStatusTagMap: Record<number, string> = {
  [SchedulerRunStatus.Pending]: "info",
  [SchedulerRunStatus.Running]: "warning",
  [SchedulerRunStatus.Success]: "success",
  [SchedulerRunStatus.Failed]: "danger",
  [SchedulerRunStatus.Canceled]: "info",
  [SchedulerRunStatus.Timeout]: "danger",
  [SchedulerRunStatus.Skipped]: "info"
};

const jobStatusTagMap: Record<number, string> = {
  [SchedulerJobStatus.Enabled]: "success",
  [SchedulerJobStatus.Disabled]: "warning",
  [SchedulerJobStatus.Retired]: "info"
};

const jobSourceTagMap: Record<number, string> = {
  [SchedulerJobSource.Builtin]: "info",
  [SchedulerJobSource.Custom]: "success"
};

/** 执行状态文案 */
export function runStatusLabelOf(status?: number | null): string {
  if (status == null) return "—";
  return runStatusLabel[status] ?? `未知(${status})`;
}

/** 执行状态对应的 el-tag 类型 */
export function runStatusTag(status?: number | null): string {
  if (status == null) return "info";
  return runStatusTagMap[status] ?? "info";
}

/** 任务状态对应的 el-tag 类型 */
export function jobStatusTag(status?: number | null): string {
  if (status == null) return "info";
  return jobStatusTagMap[status] ?? "info";
}

/** 任务来源对应的 el-tag 类型 */
export function jobSourceTag(source?: number | null): string {
  if (source == null) return "info";
  return jobSourceTagMap[source] ?? "info";
}

/** 时间格式化，空值显示占位符 */
export function formatSchedulerTime(value?: string | null): string {
  if (!value) return "—";
  const parsed = dayjs(value);
  return parsed.isValid() ? parsed.format("YYYY-MM-DD HH:mm:ss") : "—";
}

/** 耗时格式化 */
export function formatDuration(ms?: number | null): string {
  if (!ms || ms <= 0) return "—";
  if (ms < 1000) return `${ms} 毫秒`;
  const seconds = ms / 1000;
  if (seconds < 60) return `${seconds.toFixed(1)} 秒`;
  const minutes = Math.floor(seconds / 60);
  const rest = Math.round(seconds % 60);
  return `${minutes} 分 ${rest} 秒`;
}

/** 任务是否还能手动触发 */
export function canTriggerJob(row: { status: number }): boolean {
  return row.status === SchedulerJobStatus.Enabled;
}

/** 执行记录是否可以重试 */
export function canRetryRun(status: number): boolean {
  return (
    status === SchedulerRunStatus.Failed ||
    status === SchedulerRunStatus.Timeout ||
    status === SchedulerRunStatus.Canceled
  );
}

/** 执行记录是否可以取消 */
export function canCancelRun(status: number): boolean {
  return (
    status === SchedulerRunStatus.Pending ||
    status === SchedulerRunStatus.Running
  );
}
