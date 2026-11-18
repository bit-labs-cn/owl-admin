<script setup lang="ts">
import { computed, nextTick, ref, watch } from "vue";
import { useDark, useECharts } from "@pureadmin/utils";
import { eventPalette } from "../../theme";
import type { EventTrendPoint } from "../../types";

const props = defineProps<{ data: EventTrendPoint[] }>();

const { isDark } = useDark();
const theme = computed(() => (isDark.value ? "dark" : "light"));

const chartRef = ref();
const { setOptions } = useECharts(chartRef, { theme });

function render() {
  const palette = eventPalette();
  const points = props.data ?? [];

  setOptions({
    tooltip: { trigger: "axis" },
    legend: {
      data: ["发布", "消费成功"],
      top: 0,
      right: 0,
      textStyle: { color: palette.text }
    },
    grid: { top: 36, left: 8, right: 16, bottom: 4, containLabel: true },
    xAxis: {
      type: "category",
      boundaryGap: false,
      data: points.map(item => item.date.slice(5)),
      axisLine: { lineStyle: { color: palette.split } },
      axisLabel: { color: palette.text }
    },
    yAxis: {
      type: "value",
      minInterval: 1,
      splitLine: { lineStyle: { color: palette.split } },
      axisLabel: { color: palette.text }
    },
    series: [
      {
        name: "发布",
        type: "line",
        smooth: true,
        showSymbol: false,
        itemStyle: { color: palette.primary },
        areaStyle: { opacity: 0.14 },
        data: points.map(item => item.published)
      },
      {
        name: "消费成功",
        type: "line",
        smooth: true,
        showSymbol: false,
        itemStyle: { color: palette.success },
        data: points.map(item => item.consumed)
      }
    ]
  });
}

watch(
  [() => props.data, isDark],
  async () => {
    await nextTick();
    render();
  },
  { deep: true, immediate: true }
);
</script>

<template>
  <div ref="chartRef" class="trend-chart" />
</template>

<style lang="scss" scoped>
.trend-chart {
  width: 100%;
  height: 300px;
}
</style>
