<script setup lang="ts">
import { ref, onMounted } from "vue";
import { getOutboxDetail } from "@bit-labs.cn/owl-admin-ui/api/event";
import { formatJson, formatUnix } from "./format";
import { eventPriorityLabel, eventPriorityTag, outboxStatusTag } from "./status";
import type { EventOutboxDetail, EventOutboxRow } from "./types";

const props = defineProps<{
  row: EventOutboxRow;
}>();

const detail = ref<EventOutboxDetail | null>(null);
const loading = ref(true);
const error = ref("");

const payloadText = ref("");
const headersText = ref("");

onMounted(async () => {
  try {
    const res = await getOutboxDetail(props.row.id);
    const data = (res?.data ?? null) as unknown as EventOutboxDetail | null;
    detail.value = data;
    payloadText.value = formatJson(data?.payload);
    headersText.value = formatJson(data?.headers);
  } catch (err) {
    error.value = err instanceof Error ? err.message : "事件详情加载失败";
  } finally {
    loading.value = false;
  }
});
</script>

<template>
  <div v-loading="loading" class="outbox-detail">
    <el-alert
      v-if="error"
      :title="error"
      type="error"
      :closable="false"
      show-icon
      class="mb-3"
    />
    <template v-if="detail">
      <el-descriptions :column="2" border size="small">
        <el-descriptions-item label="事件类型" :span="2">
          {{ detail.eventType || "—" }}
        </el-descriptions-item>
        <el-descriptions-item label="事件ID" :span="2">
          <span class="break-all">{{ detail.eventId || "—" }}</span>
        </el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="outboxStatusTag(detail.status)" size="small">
            {{ detail.statusLabel || "—" }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="优先级">
          <el-tag :type="eventPriorityTag(detail.priority)" size="small">
            {{ eventPriorityLabel(detail.priority) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="重试次数">
          {{ detail.retryCount }}
        </el-descriptions-item>
        <el-descriptions-item label="来源">
          {{ detail.source || "—" }}
        </el-descriptions-item>
        <el-descriptions-item label="业务实体">
          {{ detail.subject || "—" }}
        </el-descriptions-item>
        <el-descriptions-item label="创建时间">
          {{ formatUnix(detail.createdAt) }}
        </el-descriptions-item>
        <el-descriptions-item label="更新时间">
          {{ formatUnix(detail.updatedAt) }}
        </el-descriptions-item>
        <el-descriptions-item label="投递成功时间">
          {{ formatUnix(detail.publishedAt) }}
        </el-descriptions-item>
        <el-descriptions-item label="下次重试时间">
          {{ formatUnix(detail.nextRetryAt) }}
        </el-descriptions-item>
        <el-descriptions-item label="错误信息" :span="2">
          <span v-if="detail.errorMsg" class="error-text">{{ detail.errorMsg }}</span>
          <span v-else>—</span>
        </el-descriptions-item>
      </el-descriptions>

      <div class="block">
        <p class="block-title">事件载荷</p>
        <pre v-if="payloadText" class="code">{{ payloadText }}</pre>
        <p v-else class="empty">—</p>
      </div>

      <div class="block">
        <p class="block-title">元数据（Headers）</p>
        <pre v-if="headersText" class="code">{{ headersText }}</pre>
        <p v-else class="empty">—</p>
      </div>
    </template>
  </div>
</template>

<style lang="scss" scoped>
.outbox-detail {
  min-height: 160px;
}

.break-all {
  word-break: break-all;
}

.error-text {
  color: var(--el-color-danger);
  word-break: break-all;
}

.block {
  margin-top: 14px;
}

.block-title {
  margin: 0 0 6px;
  font-size: 13px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.code {
  margin: 0;
  max-height: 260px;
  padding: 10px 12px;
  overflow: auto;
  border-radius: 6px;
  background: var(--el-fill-color-light);
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 12px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-all;
}

.empty {
  margin: 0;
  color: var(--el-text-color-secondary);
}
</style>
