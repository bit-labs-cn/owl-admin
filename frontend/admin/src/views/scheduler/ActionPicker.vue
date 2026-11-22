<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { APP_TITLES } from "./constants";
import type { SchedulerAction } from "./types";

const props = defineProps<{
  /** 已加载的方法目录，由新建/修改弹窗传入，避免重复请求 */
  actions: SchedulerAction[];
}>();

const emit = defineEmits<{
  (e: "confirm", action: SchedulerAction): void;
}>();

const visible = ref(false);
const keyword = ref("");
const activeApp = ref("");
const selectedCode = ref("");

/** 按子系统分组，子系统按名称排序，组内保持后端登记顺序。 */
const appGroups = computed(() => {
  const groups = new Map<string, SchedulerAction[]>();
  for (const action of props.actions) {
    const app = action.appName || "unknown";
    if (!groups.has(app)) groups.set(app, []);
    groups.get(app)!.push(action);
  }
  return Array.from(groups.entries())
    .map(([appName, items]) => ({
      appName,
      title: APP_TITLES[appName] ?? appName,
      items
    }))
    .sort((a, b) => a.title.localeCompare(b.title, "zh-CN"));
});

const searchMatched = computed(() => {
  const text = keyword.value.trim().toLowerCase();
  if (!text) return null;
  return props.actions.filter(action => {
    return (
      action.name.toLowerCase().includes(text) ||
      action.description.toLowerCase().includes(text) ||
      action.code.toLowerCase().includes(text) ||
      (APP_TITLES[action.appName] ?? action.appName).toLowerCase().includes(text)
    );
  });
});

/** 搜索时右侧显示跨系统的匹配结果，未搜索时显示选中系统的方法。 */
const visibleActions = computed(() => {
  if (searchMatched.value) return searchMatched.value;
  const group = appGroups.value.find(item => item.appName === activeApp.value);
  return group?.items ?? [];
});

const selectedAction = computed(
  () => props.actions.find(action => action.code === selectedCode.value) ?? null
);

const currentAppTitle = computed(() => {
  if (searchMatched.value) return `搜索结果（${searchMatched.value.length}）`;
  return appGroups.value.find(item => item.appName === activeApp.value)?.title ?? "";
});

function open(code: string) {
  selectedCode.value = code ?? "";
  keyword.value = "";
  const current = props.actions.find(action => action.code === code);
  activeApp.value = current?.appName ?? appGroups.value[0]?.appName ?? "";
  visible.value = true;
}

function selectApp(appName: string) {
  activeApp.value = appName;
  keyword.value = "";
}

function pick(action: SchedulerAction) {
  selectedCode.value = action.code;
}

function confirm() {
  if (!selectedAction.value) return;
  emit("confirm", selectedAction.value);
  visible.value = false;
}

// 目录异步到达时，补上默认选中的系统，避免打开后右侧空白
watch(
  () => appGroups.value.length,
  () => {
    if (!activeApp.value) {
      activeApp.value = appGroups.value[0]?.appName ?? "";
    }
  }
);

defineExpose({ open });
</script>

<template>
  <el-dialog
    v-model="visible"
    title="选择要调用的方法"
    width="900px"
    append-to-body
    align-center
    :close-on-click-modal="false"
    :close-on-press-escape="false"
    class="action-picker-dialog"
  >
    <div class="action-picker">
      <aside class="app-panel">
        <p class="panel-title">所属系统</p>
        <div class="app-list">
          <button
            v-for="group in appGroups"
            :key="group.appName"
            type="button"
            class="app-item"
            :class="{ 'is-active': !searchMatched && activeApp === group.appName }"
            @click="selectApp(group.appName)"
          >
            <span class="app-name">{{ group.title }}</span>
            <span class="app-count">{{ group.items.length }}</span>
          </button>
          <p v-if="!appGroups.length" class="panel-empty">还没有系统登记可调用方法</p>
        </div>
      </aside>

      <section class="action-panel">
        <div class="action-toolbar">
          <span class="panel-title">{{ currentAppTitle || "可选择的方法" }}</span>
          <el-input
            v-model="keyword"
            placeholder="搜索方法名称或说明"
            clearable
            class="!w-[240px]"
          />
        </div>

        <div class="action-list">
          <button
            v-for="action in visibleActions"
            :key="action.code"
            type="button"
            class="action-item"
            :class="{ 'is-selected': selectedCode === action.code }"
            @click="pick(action)"
          >
            <div class="action-head">
              <span class="action-name">{{ action.name }}</span>
              <el-tag size="small" type="info">{{ action.group || "其它" }}</el-tag>
            </div>
            <p class="action-desc">{{ action.description || "没有填写说明" }}</p>
            <div class="action-params">
              <span class="params-label">参数：</span>
              <template v-if="action.params.length">
                <span
                  v-for="param in action.params"
                  :key="param.name"
                  class="param-chip"
                >
                  {{ param.label }}<em v-if="param.required">*</em>
                </span>
              </template>
              <span v-else class="params-empty">无</span>
            </div>
          </button>
          <p v-if="!visibleActions.length" class="panel-empty">
            {{ searchMatched ? "没有匹配的方法" : "该系统还没有登记可调用方法" }}
          </p>
        </div>
      </section>
    </div>

    <template #footer>
      <div class="footer-row">
        <span class="footer-tip">
          {{ selectedAction ? `已选择：${selectedAction.name}` : "请在右侧选择一个方法" }}
        </span>
        <span>
          <el-button @click="visible = false">取消</el-button>
          <el-button type="primary" :disabled="!selectedAction" @click="confirm">
            确定
          </el-button>
        </span>
      </div>
    </template>
  </el-dialog>
</template>

<style lang="scss" scoped>
.action-picker {
  display: flex;
  height: 420px;
  overflow: hidden;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 6px;
}

.app-panel {
  display: flex;
  flex-direction: column;
  width: 190px;
  min-width: 190px;
  padding: 12px 8px;
  background: var(--el-fill-color-lighter);
  border-right: 1px solid var(--el-border-color-lighter);
}

.panel-title {
  margin: 0 0 8px;
  padding: 0 8px;
  font-size: 12px;
  font-weight: 600;
  color: var(--el-text-color-secondary);
}

.app-list {
  flex: 1;
  overflow: auto;
}

.app-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding: 8px 10px;
  margin-bottom: 2px;
  font-size: 13px;
  color: var(--el-text-color-regular);
  text-align: left;
  cursor: pointer;
  background: transparent;
  border: 0;
  border-radius: 6px;

  &:hover {
    background: var(--el-fill-color);
  }

  &.is-active {
    font-weight: 600;
    color: var(--el-color-primary);
    background: var(--el-color-primary-light-9);
  }
}

.app-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.app-count {
  margin-left: 6px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.action-panel {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-width: 0;
  padding: 12px;
}

.action-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;

  .panel-title {
    margin: 0;
    padding: 0;
    font-size: 13px;
  }
}

.action-list {
  flex: 1;
  overflow: auto;
}

.action-item {
  display: block;
  width: 100%;
  padding: 10px 12px;
  margin-bottom: 8px;
  text-align: left;
  cursor: pointer;
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 6px;
  transition: border-color 0.15s;

  &:hover {
    border-color: var(--el-color-primary-light-5);
  }

  &.is-selected {
    background: var(--el-color-primary-light-9);
    border-color: var(--el-color-primary);
  }
}

.action-head {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  align-items: center;
}

.action-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.action-desc {
  margin: 6px 0;
  font-size: 12px;
  line-height: 1.6;
  color: var(--el-text-color-secondary);
}

.action-params {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  align-items: center;
  font-size: 12px;
  color: var(--el-text-color-regular);
}

.params-label {
  color: var(--el-text-color-secondary);
}

.param-chip {
  padding: 1px 8px;
  background: var(--el-fill-color-light);
  border-radius: 10px;

  em {
    font-style: normal;
    color: var(--el-color-danger);
  }
}

.params-empty {
  color: var(--el-text-color-secondary);
}

.panel-empty {
  padding: 12px 8px;
  margin: 0;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.footer-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.footer-tip {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
</style>
