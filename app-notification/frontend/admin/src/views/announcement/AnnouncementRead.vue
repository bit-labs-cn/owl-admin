<script setup lang="ts">
import { computed } from "vue";
import { formatTime } from "../../format";
import { readStatusOptions, scopeOptions, labelOf, tagOf } from "../../labels";
import type { AnnouncementReadRecipient, AnnouncementReadView } from "./readTypes";

const props = defineProps<{ view?: AnnouncementReadView }>();

const unreadList = computed(() => (props.view?.recipients ?? []).filter(item => item.readStatus === 0));

const percent = computed(() => {
  const v = props.view;
  if (!v || v.recipientCount === 0) return 0;
  return Math.round((v.readCount / v.recipientCount) * 100);
});

function recipientName(row: AnnouncementReadRecipient) {
  return row.userName || (row.userId ? `未知用户(${row.userId})` : "未知接收人");
}
</script>

<template>
  <div v-if="view" class="ann-read">
    <div class="summary">
      <div class="summary-row">
        <span class="label">受众范围</span>
        <span>{{ labelOf(scopeOptions, view.targetScope) }}</span>
      </div>
      <div class="summary-row">
        <span class="label">接收人</span>
        <span>{{ view.recipientCount }}</span>
      </div>
      <div class="summary-row">
        <span class="label">已读</span>
        <span class="ok">{{ view.readCount }}</span>
      </div>
      <div class="summary-row">
        <span class="label">未读</span>
        <span :class="{ danger: view.unreadCount > 0 }">{{ view.unreadCount }}</span>
      </div>
      <div class="summary-row">
        <span class="label">已关闭横幅</span>
        <span>{{ view.dismissedCount }}</span>
      </div>
    </div>

    <div class="progress">
      <el-progress :percentage="percent" :stroke-width="10" />
    </div>

    <div v-if="unreadList.length > 0" class="unread-block">
      <div class="unread-title">未读 {{ unreadList.length }} 人</div>
      <div class="unread-tags">
        <el-tag v-for="item in unreadList" :key="item.userId" type="warning" effect="plain">
          {{ recipientName(item) }}
        </el-tag>
      </div>
    </div>

    <el-table :data="view.recipients" border size="small" max-height="380">
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
      <el-table-column label="阅读时间" min-width="170">
        <template #default="{ row }">{{ formatTime(row.readAt) || "—" }}</template>
      </el-table-column>
      <el-table-column label="关闭横幅" width="100">
        <template #default="{ row }">{{ row.isDismissed === 1 ? "已关闭" : "—" }}</template>
      </el-table-column>
    </el-table>
  </div>
  <el-empty v-else description="加载中" />
</template>

<style lang="scss" scoped>
.ann-read {
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
</style>
