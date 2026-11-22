/**
 * 定时任务的 Cron 可视化配置。
 *
 * 界面只让用户选「多久一次、几点、周几、几号」，由这里生成最终表达式；
 * 反向解析用于编辑已有任务，解析不了的表达式落到「自定义」模式，原样保留。
 */

/** 配置模式。hourly 是 hourInterval 间隔为 1 的形态，不单独列出。 */
export type CronMode =
  | "minutes"
  | "hourInterval"
  | "daily"
  | "weekly"
  | "monthly"
  | "custom";

export interface CronConfig {
  mode: CronMode;
  /** 每隔多少分钟，1 表示每分钟 */
  minuteInterval: number;
  /** 每隔多少小时 */
  hourInterval: number;
  /** 第几分钟（0-59） */
  minute: number;
  /** 第几小时（0-23） */
  hour: number;
  /** 周几，0 表示周日 */
  weekdays: number[];
  /** 每月几号 */
  monthDay: number;
  /** 自定义表达式 */
  custom: string;
}

export const WEEKDAY_LABELS = [
  { value: 1, label: "周一" },
  { value: 2, label: "周二" },
  { value: 3, label: "周三" },
  { value: 4, label: "周四" },
  { value: 5, label: "周五" },
  { value: 6, label: "周六" },
  { value: 0, label: "周日" }
];

export const CRON_MODE_OPTIONS: { value: CronMode; label: string }[] = [
  { value: "minutes", label: "按分钟重复" },
  { value: "hourInterval", label: "按小时重复" },
  { value: "daily", label: "每天" },
  { value: "weekly", label: "每周" },
  { value: "monthly", label: "每月" },
  { value: "custom", label: "自定义表达式" }
];

export function defaultCronConfig(): CronConfig {
  return {
    mode: "daily",
    minuteInterval: 5,
    hourInterval: 1,
    minute: 0,
    hour: 8,
    weekdays: [1, 2, 3, 4, 5],
    monthDay: 1,
    custom: ""
  };
}

function clamp(value: number, min: number, max: number, fallback: number): number {
  if (!Number.isFinite(value)) return fallback;
  if (value < min) return min;
  if (value > max) return max;
  return Math.trunc(value);
}

/** 把界面上的选择生成 5 段 Cron 表达式。 */
export function buildCron(config: CronConfig): string {
  const minute = clamp(config.minute, 0, 59, 0);
  const hour = clamp(config.hour, 0, 23, 0);
  const minuteInterval = clamp(config.minuteInterval, 1, 59, 1);
  const hourInterval = clamp(config.hourInterval, 1, 23, 1);
  const monthDay = clamp(config.monthDay, 1, 31, 1);

  switch (config.mode) {
    case "minutes":
      return minuteInterval === 1 ? "* * * * *" : `*/${minuteInterval} * * * *`;
    case "hourInterval":
      return hourInterval === 1
        ? `${minute} * * * *`
        : `${minute} */${hourInterval} * * *`;
    case "daily":
      return `${minute} ${hour} * * *`;
    case "weekly": {
      const days = (config.weekdays.length ? config.weekdays : [1])
        .slice()
        .sort((a, b) => a - b)
        .map(day => clamp(day, 0, 6, 1))
        .join(",");
      return `${minute} ${hour} * * ${days}`;
    }
    case "monthly":
      return `${minute} ${hour} ${monthDay} * *`;
    case "custom":
      return config.custom.trim();
    default:
      return "";
  }
}

/** 将已有表达式还原成配置。识别不了时落到自定义模式。 */
export function parseCron(expr: string): CronConfig {
  const config = defaultCronConfig();
  const raw = (expr ?? "").trim();
  if (!raw) {
    return config;
  }
  config.custom = raw;

  const fields = raw.split(/\s+/);
  if (fields.length !== 5) {
    config.mode = "custom";
    return config;
  }
  const [minuteField, hourField, dayField, monthField, weekField] = fields;

  if (minuteField === "*" && hourField === "*" && dayField === "*" && monthField === "*" && weekField === "*") {
    config.mode = "minutes";
    config.minuteInterval = 1;
    return config;
  }

  const minuteStep = /^\*\/(\d+)$/.exec(minuteField);
  if (minuteStep && hourField === "*" && dayField === "*" && monthField === "*" && weekField === "*") {
    config.mode = "minutes";
    config.minuteInterval = clamp(Number(minuteStep[1]), 1, 59, 5);
    return config;
  }

  const minuteNumber = /^(\d+)$/.exec(minuteField);
  if (!minuteNumber) {
    config.mode = "custom";
    return config;
  }
  const minute = clamp(Number(minuteNumber[1]), 0, 59, 0);

  if (hourField === "*" && dayField === "*" && monthField === "*" && weekField === "*") {
    config.mode = "hourInterval";
    config.hourInterval = 1;
    config.minute = minute;
    return config;
  }

  const hourStep = /^\*\/(\d+)$/.exec(hourField);
  if (hourStep && dayField === "*" && monthField === "*" && weekField === "*") {
    config.mode = "hourInterval";
    config.hourInterval = clamp(Number(hourStep[1]), 1, 23, 1);
    config.minute = minute;
    return config;
  }

  const hourNumber = /^(\d+)$/.exec(hourField);
  if (!hourNumber) {
    config.mode = "custom";
    return config;
  }
  const hour = clamp(Number(hourNumber[1]), 0, 23, 0);

  if (dayField === "*" && monthField === "*" && weekField === "*") {
    config.mode = "daily";
    config.minute = minute;
    config.hour = hour;
    return config;
  }

  if (dayField === "*" && monthField === "*" && /^[\d,]+$/.test(weekField)) {
    config.mode = "weekly";
    config.minute = minute;
    config.hour = hour;
    config.weekdays = weekField
      .split(",")
      .map(item => clamp(Number(item), 0, 6, 1))
      .filter((value, index, list) => list.indexOf(value) === index);
    return config;
  }

  const monthDay = /^(\d+)$/.exec(dayField);
  if (monthDay && monthField === "*" && weekField === "*") {
    config.mode = "monthly";
    config.minute = minute;
    config.hour = hour;
    config.monthDay = clamp(Number(monthDay[1]), 1, 31, 1);
    return config;
  }

  config.mode = "custom";
  return config;
}

function pad(value: number): string {
  return value < 10 ? `0${value}` : String(value);
}

function timeText(minute: number, hour: number): string {
  return `${pad(hour)}:${pad(minute)}`;
}

function weekdayText(day: number): string {
  return WEEKDAY_LABELS.find(item => item.value === day)?.label ?? `周${day}`;
}

/**
 * 生成人类可读描述，例如「每天 08:00 执行」「每周一、周三 09:30 执行」。
 * 识别不了的自定义表达式原样返回。
 */
export function describeCron(expr: string): string {
  const raw = (expr ?? "").trim();
  if (!raw) return "尚未设置执行时间";

  const fields = raw.split(/\s+/);
  if (fields.length !== 5) {
    return `自定义表达式：${raw}`;
  }
  const config = parseCron(raw);
  switch (config.mode) {
    case "minutes":
      return config.minuteInterval === 1
        ? "每分钟执行"
        : `每隔 ${config.minuteInterval} 分钟执行`;
    case "hourInterval":
      return config.hourInterval === 1
        ? `每小时第 ${config.minute} 分执行`
        : `每隔 ${config.hourInterval} 小时执行一次（每次第 ${config.minute} 分）`;
    case "daily":
      return `每天 ${timeText(config.minute, config.hour)} 执行`;
    case "weekly":
      return `每${config.weekdays.map(weekdayText).join("、")} ${timeText(
        config.minute,
        config.hour
      )} 执行`;
    case "monthly":
      return `每月 ${config.monthDay} 号 ${timeText(config.minute, config.hour)} 执行`;
    default:
      return `自定义表达式：${raw}`;
  }
}
