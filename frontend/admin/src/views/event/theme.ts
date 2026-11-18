/**
 * 图表配色取自 Element Plus 的主题变量，跟随明暗主题与全局主题色；
 * ECharts 用 canvas 渲染，读不到 CSS 变量，必须在渲染时解析成具体色值。
 */
export function readThemeColor(name: string, fallback: string): string {
  if (typeof window === "undefined") return fallback;
  const value = getComputedStyle(document.documentElement)
    .getPropertyValue(name)
    .trim();
  return value || fallback;
}

export function eventPalette() {
  return {
    primary: readThemeColor("--el-color-primary", "#4091f7"),
    success: readThemeColor("--el-color-success", "#22c55e"),
    warning: readThemeColor("--el-color-warning", "#f59e0b"),
    danger: readThemeColor("--el-color-danger", "#ef4444"),
    info: readThemeColor("--el-color-info", "#909399"),
    text: readThemeColor("--el-text-color-regular", "#606266"),
    split: readThemeColor("--owl-border-color", "rgba(0, 0, 0, 0.06)")
  };
}
