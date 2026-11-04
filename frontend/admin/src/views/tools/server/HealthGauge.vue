<script setup lang="ts">
import { computed } from "vue";

const props = defineProps<{ score: number }>();

const arcLength = Math.PI * 78;

const dash = computed(() => {
  const ratio = Math.max(0, Math.min(100, props.score)) / 100;
  return `${arcLength * ratio} ${arcLength}`;
});

const color = computed(() => {
  if (props.score >= 80) return "var(--el-color-success)";
  if (props.score >= 60) return "var(--el-color-warning)";
  return "var(--el-color-danger)";
});
</script>

<template>
  <svg class="gauge" viewBox="0 0 220 132" role="img" :aria-label="`综合得分 ${score}`">
    <path
      d="M 32 108 A 78 78 0 0 1 188 108"
      fill="none"
      stroke="var(--el-fill-color-dark)"
      stroke-width="14"
      stroke-linecap="round"
    />
    <path
      d="M 32 108 A 78 78 0 0 1 188 108"
      fill="none"
      :stroke="color"
      stroke-width="14"
      stroke-linecap="round"
      :stroke-dasharray="dash"
    />
    <text x="110" y="84" text-anchor="middle" class="gauge-score">{{ score }}</text>
    <text x="110" y="104" text-anchor="middle" class="gauge-caption">综合得分</text>
  </svg>
</template>

<style scoped>
.gauge {
  display: block;
  width: min(100%, 240px);
  margin: 0 auto;
}

.gauge-score {
  fill: var(--el-text-color-primary);
  font-size: 36px;
  font-weight: 600;
}

.gauge-caption {
  fill: var(--el-text-color-secondary);
  font-size: 13px;
}
</style>
