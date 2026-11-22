<script setup lang="ts">
import { computed } from "vue";
import type { SchedulerStats } from "./types";

const props = defineProps<{
  stats: SchedulerStats | null;
}>();

const items = computed(() => [
  { label: "任务总数", value: props.stats?.jobTotal ?? 0, tone: "" },
  { label: "已启用", value: props.stats?.jobEnabled ?? 0, tone: "text-[var(--el-color-success)]" },
  { label: "已停用", value: props.stats?.jobDisabled ?? 0, tone: "" },
  { label: "待执行", value: props.stats?.pending ?? 0, tone: "" },
  { label: "执行中", value: props.stats?.running ?? 0, tone: "text-[var(--el-color-warning)]" },
  { label: "今日成功", value: props.stats?.successToday ?? 0, tone: "text-[var(--el-color-success)]" },
  { label: "今日失败", value: props.stats?.failedToday ?? 0, tone: "text-[var(--el-color-danger)]" }
]);
</script>

<template>
  <div class="scheduler-summary bg-bg_color">
    <div v-for="item in items" :key="item.label" class="summary-item">
      <span class="summary-label">{{ item.label }}</span>
      <span class="summary-value" :class="item.tone">{{ item.value }}</span>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.scheduler-summary {
  display: flex;
  flex-wrap: wrap;
  gap: 12px 32px;
  align-items: center;
  padding: 12px 20px;
  margin-bottom: 12px;
  border-radius: 6px;
}

.summary-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 64px;
}

.summary-label {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.summary-value {
  font-size: 20px;
  font-weight: 600;
  line-height: 1.2;
  color: var(--el-text-color-primary);
}
</style>
