/** 发件箱列表行 */
export interface EventOutboxRow {
  id: number;
  eventId: string;
  eventType: string;
  source: string;
  subject: string;
  /** 0 普通，1 高优先级 */
  priority: number;
  status: number;
  statusLabel: string;
  retryCount: number;
  /** Unix 秒 */
  createdAt: number;
  /** Unix 秒 */
  updatedAt: number;
  publishedAt?: number;
  nextRetryAt?: number;
  errorMsg?: string;
}

/** 发件箱详情 */
export interface EventOutboxDetail extends EventOutboxRow {
  payload: string;
  headers: string;
}

/** 收件箱列表行 */
export interface EventInboxRow {
  id: number;
  eventId: string;
  consumerGroup: string;
  status: number;
  statusLabel: string;
  errorMsg?: string;
  /** Unix 秒 */
  processedAt: number;
  /** Unix 秒 */
  createdAt: number;
}

/** Redis Stream 概况 */
export interface EventStreamSnapshot {
  key: string;
  exists: boolean;
  length: number;
  error?: string;
}

/** 消费组积压与消费者活跃情况 */
export interface EventConsumerGroup {
  name: string;
  /** Redis 里累计登记过的消费者个数（含已退出进程留下的记录） */
  totalConsumers: number;
  /** 最近仍在读消息的消费者个数 */
  activeConsumers: number;
  /** 最近一次读取距今秒数，没有消费者时为 -1 */
  lastActiveSeconds: number;
  pending: number;
  /** 该组还未读取的条数（整个 stream，含其它主题） */
  lag: number;
  entriesRead: number;
  lastDeliveredId: string;
  /** 组名去掉应用前缀后的主题 */
  topic: string;
  /** 该主题在发件箱里还没投递成功的条数 */
  topicPending: number;
  /** 是否能从组名反推出主题 */
  topicKnown: boolean;
}

/** 发件箱状态计数 */
export interface EventOutboxSummary {
  total: number;
  pending: number;
  sent: number;
  failed: number;
}

/** 收件箱状态计数 */
export interface EventInboxSummary {
  total: number;
  success: number;
  failed: number;
}

/** 最老待发送事件 */
export interface EventOldestPending {
  id: number;
  eventId: string;
  eventType: string;
  source: string;
  priority: number;
  createdAt: number;
  ageSeconds: number;
}

/** 按天事件量趋势 */
export interface EventTrendPoint {
  date: string;
  published: number;
  consumed: number;
}

/** 事件类型 / 来源聚合项 */
export interface EventTypeItem {
  name: string;
  count: number;
}

/** 工作台聚合响应 */
export interface EventWorkbenchData {
  generatedAt: number;
  /** 应用标识：决定事件 Source、消费组前缀与默认 Stream key */
  appName: string;
  driverName: string;
  rangeDays: number;
  stream: EventStreamSnapshot;
  outbox: EventOutboxSummary;
  inbox: EventInboxSummary;
  oldestPending: EventOldestPending | null;
  trend: EventTrendPoint[];
  trendTruncated: boolean;
  topTypes: EventTypeItem[];
  topSources: EventTypeItem[];
  consumerGroups: EventConsumerGroup[];
  recentFailures: EventOutboxRow[];
}
