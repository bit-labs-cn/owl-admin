<script setup lang="ts">
import { computed } from "vue";
import { ReNormalCountTo } from "@bit-labs.cn/owl-ui/components/ReCountTo";

const props = defineProps<{
  title: string;
  value: number;
  unit?: string;
  hint?: string;
  tone?: "primary" | "success" | "warning" | "danger" | "info";
  clickable?: boolean;
}>();

const emit = defineEmits<{ (e: "click"): void }>();

const toneColor = computed(() => {
  const map: Record<string, string> = {
    primary: "var(--el-color-primary)",
    success: "var(--el-color-success)",
    warning: "var(--el-color-warning)",
    danger: "var(--el-color-danger)",
    info: "var(--el-color-info)"
  };
  return map[props.tone ?? "primary"];
});

function onClick() {
  if (props.clickable) emit("click");
}
</script>

<template>
  <el-card
    shadow="never"
    class="kpi-card"
    :class="{ clickable }"
    :style="{ '--kpi-color': toneColor }"
    @click="onClick"
  >
    <div class="kpi-head">
      <span class="kpi-title">{{ title }}</span>
      <span class="kpi-dot" />
    </div>
    <div class="kpi-value">
      <ReNormalCountTo :key="value" :start-val="0" :end-val="value" :duration="600" />
      <span v-if="unit" class="kpi-unit">{{ unit }}</span>
    </div>
    <p class="kpi-hint">{{ hint || "—" }}</p>
  </el-card>
</template>

<style lang="scss" scoped>
.kpi-card {
  position: relative;
  height: 100%;
  overflow: hidden;
  border-color: var(--owl-border-color, rgba(0, 0, 0, 0.06));

  &::before {
    content: "";
    position: absolute;
    top: 0;
    right: 0;
    width: 46%;
    height: 100%;
    background: linear-gradient(
      135deg,
      transparent 0%,
      color-mix(in srgb, var(--kpi-color) 8%, transparent) 100%
    );
    pointer-events: none;
  }

  &.clickable {
    cursor: pointer;
    transition: border-color 0.2s ease;

    &:hover {
      border-color: color-mix(in srgb, var(--kpi-color) 45%, transparent);
    }
  }
}

.kpi-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.kpi-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.kpi-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--kpi-color);
}

.kpi-value {
  display: flex;
  align-items: baseline;
  gap: 4px;
  margin-top: 10px;
  font-size: 26px;
  font-weight: 700;
  color: var(--el-text-color-primary);
}

.kpi-unit {
  font-size: 13px;
  font-weight: 500;
  color: var(--el-text-color-secondary);
}

.kpi-hint {
  margin: 6px 0 0;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
</style>
