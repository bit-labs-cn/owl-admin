<script setup lang="ts">
import { inject, type Ref } from "vue";

const tab = inject<Ref<string>>("legalTab")!;
</script>

<template>
  <el-tabs v-model="tab" class="legal-switch">
    <el-tab-pane label="条款" name="document" />
    <el-tab-pane label="版本" name="version" />
    <el-tab-pane label="场景" name="scene" />
    <el-tab-pane label="同意记录" name="event" />
  </el-tabs>
</template>

<style lang="scss" scoped>
.legal-switch {
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
