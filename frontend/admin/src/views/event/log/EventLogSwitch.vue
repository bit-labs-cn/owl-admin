<script setup lang="ts">
import { inject, type Ref } from "vue";

const tab = inject<Ref<string>>("eventLogTab")!;
</script>

<template>
  <el-tabs v-model="tab" class="event-log-switch">
    <el-tab-pane label="发件箱" name="outbox" />
    <el-tab-pane label="收件箱" name="inbox" />
  </el-tabs>
</template>

<style lang="scss" scoped>
.event-log-switch {
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
