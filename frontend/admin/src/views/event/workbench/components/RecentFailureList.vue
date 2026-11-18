<script setup lang="ts">
import { useRenderIcon } from "@bit-labs.cn/owl-ui/components/ReIcon/src/hooks";
import View from "@iconify-icons/ep/view";
import RefreshRight from "@iconify-icons/ep/refresh-right";
import { formatUnix } from "../../format";
import { eventPriorityLabel, eventPriorityTag } from "../../status";
import { ADMIN_PERMS } from "../../../../permissions";
import type { EventOutboxRow } from "../../types";

defineProps<{ rows: EventOutboxRow[] }>();

const emit = defineEmits<{
  (e: "detail", row: EventOutboxRow): void;
  (e: "retry", row: EventOutboxRow): void;
}>();
</script>

<template>
  <div class="recent-failures">
    <el-empty
      v-if="!rows.length"
      description="没有投递失败的事件"
      :image-size="80"
    />
    <ul v-else class="failure-list">
      <li v-for="row in rows" :key="row.id" class="failure-item">
        <div class="item-head">
          <span class="type" :title="row.eventType">{{ row.eventType || "—" }}</span>
          <el-tag :type="eventPriorityTag(row.priority)" size="small">
            {{ eventPriorityLabel(row.priority) }}
          </el-tag>
          <span class="time">{{ formatUnix(row.createdAt) }}</span>
        </div>
        <p class="error" :title="row.errorMsg">{{ row.errorMsg || "未记录错误信息" }}</p>
        <div class="item-foot">
          <span class="meta">
            {{ row.eventId }}<template v-if="row.retryCount"> · 已重试 {{ row.retryCount }} 次</template>
          </span>
          <span class="actions">
            <Perms :value="ADMIN_PERMS.event.outboxDetail">
              <el-button link type="primary" :icon="useRenderIcon(View)" @click="emit('detail', row)">
                详情
              </el-button>
            </Perms>
            <Perms :value="ADMIN_PERMS.event.retryOutbox">
              <el-button
                link
                type="warning"
                :icon="useRenderIcon(RefreshRight)"
                @click="emit('retry', row)"
              >
                重投
              </el-button>
            </Perms>
          </span>
        </div>
      </li>
    </ul>
  </div>
</template>

<style lang="scss" scoped>
.failure-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
  max-height: 420px;
  margin: 0;
  padding: 0 4px 0 0;
  overflow-y: auto;
  list-style: none;
}

.failure-item {
  padding: 10px 12px;
  border: 1px solid var(--owl-border-color, rgba(0, 0, 0, 0.06));
  border-radius: 6px;
}

.item-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.type {
  font-size: 13px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.time {
  flex-shrink: 0;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.error {
  margin: 6px 0;
  font-size: 12px;
  color: var(--el-color-danger);
  overflow: hidden;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  word-break: break-all;
}

.item-foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.meta {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.actions {
  flex-shrink: 0;
}
</style>
