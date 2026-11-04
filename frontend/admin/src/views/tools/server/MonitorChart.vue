<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from "vue";
import echarts from "@bit-labs.cn/owl-ui/plugins/echarts";

const props = defineProps<{
  labels: string[];
  series: { name: string; color: string; data: Array<number | null>; area?: boolean }[];
  yMax?: number;
  formatValue?: (value: number) => string;
}>();

const chartRef = ref<HTMLDivElement>();
let chart: ReturnType<typeof echarts.init> | null = null;
let dark = false;
let observer: MutationObserver | null = null;

function isDark() {
  return document.documentElement.classList.contains("dark");
}

function render() {
  if (!chartRef.value) return;
  const nextDark = isDark();
  if (!chart || nextDark !== dark) {
    dark = nextDark;
    chart?.dispose();
    chart = echarts.init(chartRef.value, dark ? "dark" : undefined);
  }
  chart.setOption(
    {
      animationDurationUpdate: 280,
      grid: { top: 32, right: 12, bottom: 24, left: 58 },
      legend: { top: 0, right: 0, icon: "circle", itemWidth: 8, itemHeight: 8, textStyle: { fontSize: 12 } },
      tooltip: {
        trigger: "axis",
        valueFormatter: (value: number | string) => {
          const numeric = typeof value === "number" ? value : Number(value);
          if (!Number.isFinite(numeric)) return "";
          return props.formatValue ? props.formatValue(numeric) : String(numeric);
        }
      },
      xAxis: {
        type: "category",
        boundaryGap: false,
        data: props.labels,
        axisLabel: { fontSize: 11, hideOverlap: true }
      },
      yAxis: {
        type: "value",
        min: props.yMax ? 0 : undefined,
        max: props.yMax,
        axisLabel: {
          fontSize: 11,
          formatter: (value: number) => (props.formatValue ? props.formatValue(value) : String(value))
        },
        splitLine: { lineStyle: { type: "dashed" } }
      },
      series: props.series.map(item => ({
        name: item.name,
        type: "line",
        smooth: true,
        showSymbol: false,
        connectNulls: false,
        data: item.data,
        color: item.color,
        lineStyle: { width: 2 },
        areaStyle: item.area ? { opacity: 0.16 } : undefined
      }))
    },
    true
  );
}

function resize() {
  chart?.resize();
}

onMounted(() => {
  render();
  window.addEventListener("resize", resize);
  observer = new MutationObserver(() => render());
  observer.observe(document.documentElement, { attributes: true, attributeFilter: ["class"] });
});

onBeforeUnmount(() => {
  window.removeEventListener("resize", resize);
  observer?.disconnect();
  chart?.dispose();
  chart = null;
});

watch(() => [props.labels, props.series, props.yMax], render, { deep: true });
</script>

<template>
  <div ref="chartRef" class="monitor-chart" />
</template>

<style scoped>
.monitor-chart {
  width: 100%;
  height: 240px;
}
</style>
