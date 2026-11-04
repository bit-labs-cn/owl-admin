export type SettingValueType = "string" | "bool" | "int" | "image";

export interface SettingItem {
  key: string;
  title: string;
  hint: string;
  type: SettingValueType;
  default: string;
  value: string;
  min: number;
  max: number;
  public: boolean;
}

export interface SettingGroup {
  key: string;
  title: string;
  items: SettingItem[];
  groups?: SettingGroup[];
}

export interface SettingSubsystem {
  name: string;
  title: string;
  groups: SettingGroup[];
}

export interface SettingView {
  subsystems: SettingSubsystem[];
}

/** 左侧导航的一个节点。叶子分组带 group，系统和一级目录只用来展开。 */
export interface SettingTreeNode {
  id: string;
  label: string;
  subsystem: string;
  group?: string;
  children?: SettingTreeNode[];
}
