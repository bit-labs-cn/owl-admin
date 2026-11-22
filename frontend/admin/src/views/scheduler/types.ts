/** 定时任务列表行，字段与 owl/provider/scheduler/model.go 的 Job 一致 */
export interface SchedulerJobRow {
  /** 雪花 ID，后端以字符串返回 */
  id: string;
  /** 任务编码：builtin:<子系统>:<key> 或 custom:<ID> */
  code: string;
  name: string;
  description: string;
  ownerApp: string;
  jobKey: string;
  /** 1 内置，2 自定义 */
  source: number;
  cron: string;
  timezone: string;
  /** 1 内置处理函数，2 HTTP，3 Shell */
  scriptType: number;
  /** 「调用子系统方法」任务绑定的动作编码与名称快照 */
  actionCode?: string;
  actionName?: string;
  /** 参数摘要，例如「开关：前台；命令：开机」 */
  scriptSummary?: string;
  timeoutSeconds: number;
  maxAttempts: number;
  retryBackoffSeconds: number;
  /** 1 跳过，2 允许并发 */
  overlapPolicy: number;
  /** 1 启用，2 停用，3 已下线 */
  status: number;
  nextRunAt?: string | null;
  lastRunAt?: string | null;
  lastStatus?: number | null;
  lastError: string;
  runCount: number;
  failCount: number;
  createdAt: string;
  updatedAt: string;

  sourceText?: string;
  statusText?: string;
  scriptTypeText?: string;
  overlapText?: string;
  lastStatusText?: string;
}

/** 任务详情，额外包含脚本配置 */
export interface SchedulerJobDetail extends SchedulerJobRow {
  scriptConfig: string;
}

/** 执行记录列表行 */
export interface SchedulerRunRow {
  id: string;
  jobId: string;
  jobCode: string;
  jobName: string;
  /** 1 定时，2 手动，3 重试 */
  triggerType: number;
  /** 1 待执行，2 执行中，3 成功，4 失败，5 已取消，6 超时，7 已跳过 */
  status: number;
  attempt: number;
  maxAttempts: number;
  plannedAt: string;
  startedAt?: string | null;
  finishedAt?: string | null;
  durationMs: number;
  retryOfRunId: string;
  payload: string;
  output: string;
  errorMsg: string;
  stack: string;
  instanceId: string;
  leaseExpiresAt?: string | null;
  createdAt: string;
  updatedAt: string;

  statusText?: string;
  triggerTypeText?: string;
  attemptText?: string;
}

/** 任务汇总 */
export interface SchedulerStats {
  jobTotal: number;
  jobEnabled: number;
  jobDisabled: number;
  jobRetired: number;
  pending: number;
  running: number;
  successToday: number;
  failedToday: number;
  recentFailures?: SchedulerRunRow[];
}

/** Cron 预览响应 */
export interface SchedulerCronPreview {
  times: string[];
}

/** 新建/修改任务的表单数据 */
export interface SchedulerJobFormData {
  id?: string;
  name: string;
  description: string;
  cron: string;
  timezone: string;
  timeoutSeconds: number;
  maxAttempts: number;
  retryBackoffSeconds: number;
  overlapPolicy: number;
  scriptType: number;
  /** 调用子系统方法：动作编码与参数 */
  actionCode: string;
  actionParams: Record<string, string>;
  /** HTTP 任务字段 */
  httpMethod: string;
  httpURL: string;
  httpHeaders: string;
  httpBody: string;
  httpContentType: string;
  httpSuccessStatus: string;
  /** Shell 任务字段 */
  shellCommand: string;
  shellWorkingDir: string;
}

/** 动作参数声明，字段与 owl/contract/scheduler 的 Param 一致 */
export interface SchedulerActionParam {
  name: string;
  label: string;
  type: "string" | "text" | "int" | "bool" | "select";
  required?: boolean;
  default?: string;
  placeholder?: string;
  description?: string;
  multiple?: boolean;
  options?: SchedulerActionOption[];
  dynamic?: boolean;
  min?: number | null;
  max?: number | null;
}

/** 下拉候选 */
export interface SchedulerActionOption {
  value: string;
  label: string;
  disabled?: boolean;
}

/** 子系统登记的可调用方法 */
export interface SchedulerAction {
  code: string;
  key: string;
  name: string;
  description: string;
  group: string;
  appName: string;
  params: SchedulerActionParam[];
}

/** 动态候选响应 */
export interface SchedulerActionOptions {
  options: Record<string, SchedulerActionOption[]>;
}
