<script setup lang="ts">
import type { InboxRow } from "./types";
import { formatTime } from "../../format";
import { kindOptions, labelOf, todoStatusOptions } from "../../labels";
import { KindTodo } from "../../constants";
import RichContentView from "../../components/RichContentView.vue";

defineProps<{ row: InboxRow }>();
</script>

<template>
  <div class="inbox-detail">
    <div class="header mb-4">
      <h3 class="title text-lg font-bold">{{ row.title }}</h3>
      <div class="meta text-xs text-gray-500 mt-1 flex gap-4">
        <span>消息类型：{{ row.typeName || (row.typeCode ? `未知类型(${row.typeCode})` : "未分类") }}</span>
        <span>种类：{{ labelOf(kindOptions, row.kind) }}</span>
        <span v-if="row.kind === KindTodo">状态：{{ labelOf(todoStatusOptions, row.todoStatus) }}</span>
        <span v-if="row.dueAt">截止：{{ formatTime(row.dueAt) }}</span>
        <span>时间：{{ formatTime(row.createdAt) }}</span>
      </div>
    </div>

    <RichContentView class="content" :content="row.content" empty-text="无详细正文" />

    <div v-if="row.link" class="action mt-4 pt-3 border-t">
      <el-link type="primary" :href="row.link" target="_blank" :underline="false">
        前往业务页面处理 &raquo;
      </el-link>
    </div>
  </div>
</template>

<style scoped>
.title {
  margin: 0;
}

.content {
  padding: 12px;
  border-radius: 6px;
  background: var(--el-fill-color-light);
}
</style>
