<script setup lang="ts">
import { computed, nextTick, ref, watch } from "vue";
import { useDark, useECharts } from "@pureadmin/utils";
import { eventPalette } from "../../theme";
import type { EventTypeItem } from "../../types";

const props = defineProps<{
  items: EventTypeItem[];
  tone?: "primary" | "success" | "warning";
  emptyText?: string;
}>();

const { isDark } = useDark();
const theme = computed(() => (isDark.value ? "dark" : "light"));

const chartRef = ref();
const { setOptions } = useECharts(chartRef, { theme });

const rows = computed(() => props.items ?? []);
const chartHeight = computed(() => Math.max(240, rows.value.length * 32 + 24));

function render() {
  const palette = eventPalette();
  const colorMap = {
    primary: palette.primary,
    success: palette.success,
    warning: palette.warning
  };
  const color = colorMap[props.tone ?? "primary"];
  // yAxis 自下而上渲染，倒序后最大值才在顶部
  const ordered = [...rows.value].reverse();

  setOptions({
    tooltip: { trigger: "axis", axisPointer: { type: "shadow" } },
    grid: { top: 8, left: 8, right: 24, bottom: 8, containLabel: true },
    xAxis: {
      type: "value",
      minInterval: 1,
      splitLine: { lineStyle: { color: palette.split } },
      axisLabel: { color: palette.text }
    },
    yAxis: {
      type: "category",
      data: ordered.map(item => item.name),
      axisLine: { lineStyle: { color: palette.split } },
      axisLabel: { color: palette.text, width: 160, overflow: "truncate" }
    },
    series: [
      {
        type: "bar",
        barWidth: 12,
        itemStyle: { color, borderRadius: [0, 6, 6, 0] },
        label: {
          show: true,
          position: "right",
          color: palette.text,
          fontSize: 12
        },
        data: ordered.map(item => item.count)
      }
    ]
  });
}

watch(
  [() => props.items, isDark],
  async () => {
    await nextTick();
    render();
  },
  { deep: true, immediate: true }
);
</script>

<template>
  <div class="rank-chart-wrap">
    <el-empty
      v-if="rows.length === 0"
      :description="emptyText || '暂无数据'"
      :image-size="72"
    />
    <div
      v-else
      ref="chartRef"
      :style="{ width: '100%', height: `${chartHeight}px` }"
    />
  </div>
</template>

<style lang="scss" scoped>
.rank-chart-wrap {
  min-height: 240px;
}
</style>
