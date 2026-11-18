<script setup lang="ts">
import { computed, nextTick, ref, watch } from "vue";
import { useDark, useECharts } from "@pureadmin/utils";
import { eventPalette } from "../../theme";
import { eventPriorityLabel, eventPriorityTag } from "../../status";
import { formatDuration, formatUnix } from "../../format";
import type { EventOldestPending, EventOutboxSummary } from "../../types";

const props = defineProps<{
  summary: EventOutboxSummary;
  oldest: EventOldestPending | null;
}>();

const { isDark } = useDark();
const theme = computed(() => (isDark.value ? "dark" : "light"));

const chartRef = ref();
const { setOptions } = useECharts(chartRef, { theme });

const total = computed(() => props.summary?.total ?? 0);

function percent(value: number) {
  if (!total.value) return "0%";
  return `${((value / total.value) * 100).toFixed(1)}%`;
}

function render() {
  const palette = eventPalette();

  setOptions({
    tooltip: { trigger: "item", formatter: "{b}：{c} 条（{d}%）" },
    legend: {
      bottom: 0,
      itemWidth: 10,
      itemHeight: 10,
      textStyle: { color: palette.text }
    },
    series: [
      {
        type: "pie",
        radius: ["52%", "72%"],
        center: ["50%", "44%"],
        avoidLabelOverlap: true,
        label: { show: false },
        data: [
          {
            name: "待发送",
            value: props.summary?.pending ?? 0,
            itemStyle: { color: palette.warning }
          },
          {
            name: "已发送",
            value: props.summary?.sent ?? 0,
            itemStyle: { color: palette.success }
          },
          {
            name: "投递失败",
            value: props.summary?.failed ?? 0,
            itemStyle: { color: palette.danger }
          }
        ]
      }
    ]
  });
}

watch(
  [() => props.summary, isDark],
  async () => {
    await nextTick();
    render();
  },
  { deep: true, immediate: true }
);
</script>

<template>
  <div class="status-donut">
    <div ref="chartRef" class="donut" />
    <ul class="legend">
      <li>
        <span class="dot warning" />
        <span class="label">待发送</span>
        <span class="value">{{ summary?.pending ?? 0 }}</span>
        <span class="percent">{{ percent(summary?.pending ?? 0) }}</span>
      </li>
      <li>
        <span class="dot success" />
        <span class="label">已发送</span>
        <span class="value">{{ summary?.sent ?? 0 }}</span>
        <span class="percent">{{ percent(summary?.sent ?? 0) }}</span>
      </li>
      <li>
        <span class="dot danger" />
        <span class="label">投递失败</span>
        <span class="value">{{ summary?.failed ?? 0 }}</span>
        <span class="percent">{{ percent(summary?.failed ?? 0) }}</span>
      </li>
    </ul>
    <div class="oldest">
      <template v-if="oldest">
        <div class="oldest-head">
          <span>最老待发送</span>
          <el-tag type="warning" size="small">
            {{ formatDuration(oldest.ageSeconds) }}
          </el-tag>
        </div>
        <p class="oldest-type">{{ oldest.eventType || "—" }}</p>
        <el-tag :type="eventPriorityTag(oldest.priority)" size="small">
          {{ eventPriorityLabel(oldest.priority) }}
        </el-tag>
        <p class="oldest-meta">
          {{ oldest.eventId }} · {{ formatUnix(oldest.createdAt) }}
        </p>
      </template>
      <p v-else class="oldest-empty">当前没有待发送事件</p>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.status-donut {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.donut {
  width: 100%;
  height: 200px;
}

.legend {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin: 0;
  padding: 0 4px;
  list-style: none;

  li {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 13px;
    color: var(--el-text-color-regular);
  }

  .label {
    flex: 1;
  }

  .value {
    font-weight: 600;
    color: var(--el-text-color-primary);
  }

  .percent {
    width: 52px;
    text-align: right;
    color: var(--el-text-color-secondary);
  }
}

.dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;

  &.warning {
    background: var(--el-color-warning);
  }

  &.success {
    background: var(--el-color-success);
  }

  &.danger {
    background: var(--el-color-danger);
  }
}

.oldest {
  margin-top: auto;
  padding: 12px;
  border-radius: 6px;
  background: var(--owl-primary-bg-light, var(--el-fill-color-light));
}

.oldest-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 13px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.oldest-type {
  margin: 8px 0 2px;
  font-size: 13px;
  word-break: break-all;
}

.oldest-meta {
  margin: 0;
  font-size: 12px;
  color: var(--el-text-color-secondary);
  word-break: break-all;
}

.oldest-empty {
  margin: 0;
  font-size: 13px;
  color: var(--el-text-color-secondary);
}
</style>
