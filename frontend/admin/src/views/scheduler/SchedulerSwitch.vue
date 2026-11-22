<script setup lang="ts">
defineProps<{
  modelValue: "jobs" | "runs";
}>();

const emit = defineEmits<{
  (e: "update:modelValue", value: "jobs" | "runs"): void;
}>();
</script>

<template>
  <el-tabs
    :model-value="modelValue"
    class="scheduler-switch"
    @update:model-value="value => emit('update:modelValue', value as 'jobs' | 'runs')"
  >
    <el-tab-pane label="任务" name="jobs" />
    <el-tab-pane label="执行记录" name="runs" />
  </el-tabs>
</template>

<style lang="scss" scoped>
.scheduler-switch {
  align-self: center;

  :deep(.el-tabs__header) {
    margin: 0;
  }

  :deep(.el-tabs__content) {
    display: none;
  }

  :deep(.el-tabs__nav-wrap::after) {
    height: 0;
  }

  :deep(.el-tabs__item) {
    height: 32px;
    padding: 0 16px;
    font-weight: 600;
    line-height: 32px;
  }

  :deep(.el-tabs__item:nth-child(2)) {
    padding-left: 0;
  }
}
</style>
