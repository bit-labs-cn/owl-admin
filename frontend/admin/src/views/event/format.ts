import dayjs from "dayjs";

/** 后端统一返回 Unix 秒；空值显示为占位符 */
export function formatUnix(value?: number | null, placeholder = "—"): string {
  if (value == null || value <= 0) return placeholder;
  const parsed = dayjs(value * 1000);
  return parsed.isValid() ? parsed.format("YYYY-MM-DD HH:mm:ss") : placeholder;
}

/** 把秒数格式化成「天 / 小时 / 分钟」可读文案 */
export function formatDuration(seconds?: number | null): string {
  if (seconds == null || seconds < 0) return "—";
  const day = Math.floor(seconds / 86400);
  const hour = Math.floor((seconds % 86400) / 3600);
  const minute = Math.floor((seconds % 3600) / 60);

  if (day > 0) return `${day} 天 ${hour} 小时`;
  if (hour > 0) return `${hour} 小时 ${minute} 分钟`;
  if (minute > 0) return `${minute} 分钟`;
  return `${Math.floor(seconds)} 秒`;
}

/** JSON 文本美化；解析失败时原样返回，避免详情页直接报错 */
export function formatJson(raw?: string | null): string {
  if (!raw) return "";
  try {
    return JSON.stringify(JSON.parse(raw), null, 2);
  } catch {
    return String(raw);
  }
}
