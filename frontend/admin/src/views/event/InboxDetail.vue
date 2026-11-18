<script setup lang="ts">
import { ref, onMounted } from "vue";
import { getInboxDetail } from "@bit-labs.cn/owl-admin-ui/api/event";
import { formatUnix } from "./format";
import { inboxStatusTag } from "./status";
import type { EventInboxRow } from "./types";

const props = defineProps<{
  row: EventInboxRow;
}>();

const detail = ref<EventInboxRow | null>(null);
const loading = ref(true);
const error = ref("");

onMounted(async () => {
  try {
    const res = await getInboxDetail(props.row.id);
    detail.value = (res?.data ?? null) as unknown as EventInboxRow | null;
  } catch (err) {
    error.value = err instanceof Error ? err.message : "消费流水加载失败";
  } finally {
    loading.value = false;
  }
});
</script>

<template>
  <div v-loading="loading" class="inbox-detail">
    <el-alert
      v-if="error"
      :title="error"
      type="error"
      :closable="false"
      show-icon
      class="mb-3"
    />
    <el-descriptions v-if="detail" :column="2" border size="small">
      <el-descriptions-item label="事件ID" :span="2">
        <span class="break-all">{{ detail.eventId || "—" }}</span>
      </el-descriptions-item>
      <el-descriptions-item label="消费组" :span="2">
        <span class="break-all">{{ detail.consumerGroup || "—" }}</span>
      </el-descriptions-item>
      <el-descriptions-item label="状态">
        <el-tag :type="inboxStatusTag(detail.status)" size="small">
          {{ detail.statusLabel || "—" }}
        </el-tag>
      </el-descriptions-item>
      <el-descriptions-item label="写入时间">
        {{ formatUnix(detail.createdAt) }}
      </el-descriptions-item>
      <el-descriptions-item label="消费完成时间" :span="2">
        {{ formatUnix(detail.processedAt) }}
      </el-descriptions-item>
      <el-descriptions-item label="错误信息" :span="2">
        <span v-if="detail.errorMsg" class="error-text">{{ detail.errorMsg }}</span>
        <span v-else>—</span>
      </el-descriptions-item>
    </el-descriptions>
  </div>
</template>

<style lang="scss" scoped>
.inbox-detail {
  min-height: 120px;
}

.break-all {
  word-break: break-all;
}

.error-text {
  color: var(--el-color-danger);
  word-break: break-all;
}
</style>
