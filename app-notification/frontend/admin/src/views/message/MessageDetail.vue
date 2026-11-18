<script setup lang="ts">
import { computed } from "vue";
import { formatTime } from "../../format";
import { KindTodo } from "../../constants";
import { kindOptions, labelOf, readStatusOptions, tagOf, todoStatusOptions } from "../../labels";
import RichContentView from "../../components/RichContentView.vue";
import type { MessageDetail, MessageRecipient } from "./types";

const props = defineProps<{ detail?: MessageDetail }>();

const isTodo = computed(() => props.detail?.kind === KindTodo);
const unreadList = computed(() => (props.detail?.recipients ?? []).filter(item => item.readStatus === 0));

function recipientName(row: MessageRecipient) {
  return row.userName || (row.userId ? `未知用户(${row.userId})` : "未知接收人");
}
</script>

<template>
  <div v-if="detail" class="message-detail">
    <div class="summary">
      <div class="summary-row">
        <span class="label">接收人</span>
        <span>{{ detail.recipientCount }}</span>
      </div>
      <div class="summary-row">
        <span class="label">已读</span>
        <span class="ok">{{ detail.readCount }}</span>
      </div>
      <div class="summary-row">
        <span class="label">未读</span>
        <span :class="{ danger: detail.unreadCount > 0 }">{{ detail.unreadCount }}</span>
      </div>
      <div v-if="isTodo" class="summary-row">
        <span class="label">待办</span>
        <span>待处理 {{ detail.todoPending }} / 已完成 {{ detail.todoDone }}</span>
      </div>
      <div v-if="detail.deliveryFailed > 0" class="summary-row">
        <span class="label">投递失败</span>
        <span class="danger">{{ detail.deliveryFailed }}</span>
      </div>
    </div>

    <div class="head">
      <h3 class="title">{{ detail.title }}</h3>
      <div class="meta">
        <span>{{ detail.typeName || "未分类" }}</span>
        <span>{{ labelOf(kindOptions, detail.kind) }}</span>
        <span>{{ detail.senderName || "系统" }}</span>
        <span>{{ formatTime(detail.createdAt) }}</span>
        <span v-if="detail.dueAt">截止 {{ formatTime(detail.dueAt) }}</span>
      </div>
    </div>

    <RichContentView class="content" :content="detail.content" empty-text="无正文" />

    <div v-if="detail.link" class="link-line">
      <el-link type="primary" :href="detail.link" target="_blank" :underline="false">{{ detail.link }}</el-link>
    </div>

    <div v-if="unreadList.length > 0" class="unread-block">
      <div class="unread-title">未读 {{ unreadList.length }} 人</div>
      <div class="unread-tags">
        <el-tag v-for="item in unreadList" :key="item.id" type="warning" effect="plain">
          {{ recipientName(item) }}
        </el-tag>
      </div>
    </div>

    <el-table :data="detail.recipients" border size="small" class="recipient-table" max-height="360">
      <el-table-column label="接收人" min-width="160">
        <template #default="{ row }">
          <div class="recipient-cell">
            <span>{{ recipientName(row) }}</span>
            <span v-if="row.username && row.username !== row.userName" class="sub">{{ row.username }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="邮箱" min-width="180">
        <template #default="{ row }">{{ row.email || "—" }}</template>
      </el-table-column>
      <el-table-column label="阅读状态" width="110">
        <template #default="{ row }">
          <el-tag :type="tagOf(readStatusOptions, row.readStatus)">
            {{ labelOf(readStatusOptions, row.readStatus) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="阅读时间" min-width="160">
        <template #default="{ row }">{{ formatTime(row.readAt) || "—" }}</template>
      </el-table-column>
      <el-table-column v-if="isTodo" label="待办状态" width="110">
        <template #default="{ row }">{{ labelOf(todoStatusOptions, row.todoStatus) }}</template>
      </el-table-column>
      <el-table-column v-if="isTodo" label="完成时间" min-width="160">
        <template #default="{ row }">{{ formatTime(row.doneAt) || "—" }}</template>
      </el-table-column>
    </el-table>
  </div>
  <el-empty v-else description="加载中" />
</template>

<style lang="scss" scoped>
.message-detail {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.summary {
  display: flex;
  flex-wrap: wrap;
  gap: 18px;
  padding: 10px 14px;
  border-radius: 8px;
  background: var(--el-fill-color-light);
  font-size: 13px;
}

.summary-row {
  display: flex;
  gap: 6px;
}

.label {
  color: var(--el-text-color-secondary);
}

.ok {
  color: var(--el-color-success);
}

.danger {
  color: var(--el-color-danger);
}

.title {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
}

.meta {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-top: 6px;
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.content {
  padding: 12px;
  border-radius: 8px;
  background: var(--el-fill-color-lighter);
  font-size: 13px;
  line-height: 1.7;
  word-break: break-word;
}

.link-line {
  font-size: 12px;
  word-break: break-all;
}

.unread-block {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.unread-title {
  color: var(--el-color-warning-dark-2);
  font-size: 13px;
  font-weight: 600;
}

.unread-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.recipient-cell {
  display: flex;
  flex-direction: column;
  line-height: 1.4;
}

.sub {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.recipient-table {
  margin-top: 4px;
}
</style>
