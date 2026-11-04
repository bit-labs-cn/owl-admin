import type { SettingGroup, SettingSubsystem } from "./types";

/** 把配置树收成「子系统.分组.配置键」。有二级分组时路径多一段。 */
export function indexSettingValues(subsystems: SettingSubsystem[] | undefined): Record<string, string> {
  const out: Record<string, string> = {};
  for (const subsystem of subsystems ?? []) {
    collect(out, subsystem.name, subsystem.groups);
  }
  return out;
}

function collect(out: Record<string, string>, prefix: string, groups: SettingGroup[] | undefined) {
  for (const group of groups ?? []) {
    const path = `${prefix}.${group.key}`;
    for (const item of group.items ?? []) {
      out[`${path}.${item.key}`] = item.value ?? "";
    }
    collect(out, path, group.groups);
  }
}
