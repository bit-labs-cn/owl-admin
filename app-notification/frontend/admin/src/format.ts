export function formatTime(value?: string | null) {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`;
}

export function toRFC3339(value?: string) {
  if (!value) return undefined;
  const date = new Date(value.replace(" ", "T"));
  if (Number.isNaN(date.getTime())) return undefined;
  return date.toISOString();
}

export function fromRFC3339(value?: string | null) {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`;
}

export function parseJson<T>(raw: string | undefined, fallback: T): T {
  if (!raw) return fallback;
  try {
    return JSON.parse(raw) as T;
  } catch {
    return fallback;
  }
}

export function pairsToText(raw?: string) {
  const value = parseJson<Record<string, string>>(raw, {});
  return Object.entries(value)
    .map(([key, item]) => `${key}=${item ?? ""}`)
    .join("\n");
}

export function textToPairs(text: string) {
  const out: Record<string, string> = {};
  text.split("\n").forEach(line => {
    const index = line.indexOf("=");
    if (index <= 0) return;
    const key = line.slice(0, index).trim();
    if (!key) return;
    out[key] = line.slice(index + 1).trim();
  });
  return out;
}

export function applyPage(
  pagination: { total: number; pageSize: number; currentPage: number },
  res?: { total?: number; pageSize?: number; currentPage?: number }
) {
  pagination.total = res?.total ?? 0;
  pagination.pageSize = res?.pageSize ?? pagination.pageSize;
  pagination.currentPage = res?.currentPage ?? pagination.currentPage;
}

export const tableHeaderStyle = {
  background: "var(--el-fill-color-light)",
  color: "var(--el-text-color-primary)"
};

// 正文可能是富文本，也可能还是早期存的纯文本。判断方式与后端 IsHTML 一致。
const richTagRe = /<\/?(p|br|div|span|strong|b|em|i|u|s|ul|ol|li|h[1-6]|blockquote|pre|code|table|thead|tbody|tr|th|td|a|img|hr)\b/i;

export function isRichContent(raw?: string | null) {
  return !!raw && richTagRe.test(raw);
}

/** 把富文本取成纯文本，给企业微信、短信这类只吃文字的渠道预览用。 */
export function richToText(raw?: string | null) {
  if (!raw) return "";
  if (!isRichContent(raw)) return raw.replace(/\s+/g, " ").trim();
  const div = document.createElement("div");
  div.innerHTML = raw;
  return (div.textContent ?? "").replace(/\s+/g, " ").trim();
}

/** 富文本里是否真有内容：只有图片也算有内容，空段落不算。 */
export function hasRichContent(raw?: string | null) {
  if (!raw) return false;
  if (/<img\b/i.test(raw)) return true;
  return richToText(raw).length > 0;
}
