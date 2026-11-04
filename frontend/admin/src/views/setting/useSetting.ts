import { computed, ref } from "vue";
import { ElMessageBox } from "element-plus";
import { settingAPI } from "@bit-labs.cn/owl-admin-ui/api/setting";
import { loadPublicSite } from "@bit-labs.cn/owl-admin-ui/loadPublicSite";
import type { SettingGroup, SettingSubsystem, SettingTreeNode } from "./types";

export function useSetting() {
  const loading = ref(false);
  const saving = ref(false);
  const subsystems = ref<SettingSubsystem[]>([]);
  const subsystemName = ref("");
  const groupKey = ref("");

  const currentSubsystem = computed(() => subsystems.value.find(item => item.name === subsystemName.value));
  const groups = computed(() => currentSubsystem.value?.groups ?? []);
  const currentGroup = computed(() => findGroup(groups.value, groupKey.value));
  const currentParentTitle = computed(() => {
    for (const group of groups.value) {
      if (group.groups?.some(child => child.key === groupKey.value)) return group.title;
    }
    return "";
  });
  const items = computed(() => currentGroup.value?.items ?? []);
  const treeData = computed<SettingTreeNode[]>(() =>
    subsystems.value.map(subsystem => ({
      id: subsystem.name,
      label: subsystem.title,
      subsystem: subsystem.name,
      children: subsystem.groups.map(group => toTreeNode(subsystem.name, group))
    }))
  );
  const currentNodeKey = computed(() => nodeIdForGroup(treeData.value, groupKey.value));

  function keepSelection(list: SettingSubsystem[]) {
    if (!list.some(item => item.name === subsystemName.value)) {
      subsystemName.value = list[0]?.name ?? "";
    }
    const nextGroups = list.find(item => item.name === subsystemName.value)?.groups ?? [];
    if (!leafKeys(nextGroups).includes(groupKey.value)) {
      groupKey.value = leafKeys(nextGroups)[0] ?? "";
    }
  }

  async function load() {
    loading.value = true;
    try {
      const res = await settingAPI.get();
      const list = res.data?.subsystems ?? [];
      subsystems.value = list;
      keepSelection(list);
    } finally {
      loading.value = false;
    }
  }

  function onNode(node: SettingTreeNode) {
    const group = firstLeafKey(node);
    if (!group) return;
    subsystemName.value = node.subsystem;
    groupKey.value = group;
  }

  async function save() {
    const current = currentSubsystem.value;
    if (!current) return;
    const values = allItems(current.groups).map(item => ({ key: item.key, value: item.value }));
    saving.value = true;
    try {
      await settingAPI.update({ subsystem: current.name, values });
      await loadPublicSite();
      await load();
    } finally {
      saving.value = false;
    }
  }

  async function restore(group: SettingGroup) {
    try {
      await ElMessageBox.confirm(`确定把「${group.title}」恢复为默认值吗？`, "恢复默认");
    } catch {
      return;
    }
    for (const item of group.items) {
      item.value = item.default;
    }
    await save();
  }

  return {
    loading,
    saving,
    subsystems,
    subsystemName,
    groupKey,
    groups,
    currentGroup,
    currentParentTitle,
    currentSubsystem,
    items,
    treeData,
    currentNodeKey,
    load,
    onNode,
    save,
    restore
  };
}

function toTreeNode(subsystem: string, group: SettingGroup, parentKey?: string): SettingTreeNode {
  const id = parentKey ? `${subsystem}/${parentKey}/${group.key}` : `${subsystem}/${group.key}`;
  if (group.groups?.length) {
    return {
      id,
      label: group.title,
      subsystem,
      children: group.groups.map(child => toTreeNode(subsystem, child, group.key))
    };
  }
  return { id, label: group.title, subsystem, group: group.key };
}

function findGroup(groups: SettingGroup[], key: string): SettingGroup | undefined {
  for (const group of groups) {
    if (group.groups?.length) {
      const nested = findGroup(group.groups, key);
      if (nested) return nested;
      continue;
    }
    if (group.key === key) return group;
  }
  return undefined;
}

function leafKeys(groups: SettingGroup[]): string[] {
  const keys: string[] = [];
  for (const group of groups) {
    if (group.groups?.length) keys.push(...leafKeys(group.groups));
    else keys.push(group.key);
  }
  return keys;
}

function allItems(groups: SettingGroup[]): SettingGroup["items"] {
  return groups.flatMap(group => [...(group.items ?? []), ...allItems(group.groups ?? [])]);
}

function firstLeafKey(node: SettingTreeNode): string | undefined {
  if (node.group) return node.group;
  for (const child of node.children ?? []) {
    const key = firstLeafKey(child);
    if (key) return key;
  }
  return undefined;
}

function nodeIdForGroup(nodes: SettingTreeNode[], key: string): string {
  for (const node of nodes) {
    if (node.group === key) return node.id;
    const nested = nodeIdForGroup(node.children ?? [], key);
    if (nested) return nested;
  }
  return "";
}
