<script setup lang="ts">
import { computed } from "vue";
import { DeliveryStatusFailed, DeliveryStatusPending } from "../../constants";
import { deliveryStatusOptions, labelOf } from "../../labels";
import { NOTIFY_PERMS } from "../../permissions";
import type { DeliveryRow } from "./types";

const props = defineProps<{
  rows: DeliveryRow[];
  loading: boolean;
  columns: TableColumnList;
}>();

const emit = defineEmits<{ retry: [row: DeliveryRow] }>();

const summary = computed(() => {
  const total = props.rows.length;
  const count = (status: number) => props.rows.filter(item => item.status === status).length;
  return {
    total,
    success: count(2),
    failed: count(DeliveryStatusFailed),
    pending: count(DeliveryStatusPending),
    skipped: count(4)
  };
});

function canRetry(row: DeliveryRow) {
  return row.status === DeliveryStatusFailed;
}
</script>

<template>
  <div class="delivery-detail">
    <div class="summary">
      <span class="item">记录 {{ summary.total }}</span>
      <span class="item ok">成功 {{ summary.success }}</span>
      <span class="item danger">失败 {{ summary.failed }}</span>
      <span class="item warn">待发送 {{ summary.pending }}</span>
      <span class="item muted">跳过 {{ summary.skipped }}</span>
    </div>

    <pure-table
      border
      size="small"
      row-key="id"
      max-height="420"
      :loading="loading"
      :data="rows"
      :columns="columns"
      align-whole="center"
      showOverflowTooltip
    >
      <template #empty>
        <el-empty description="这次发送没有渠道记录" />
      </template>
      <template #operation="{ row }">
        <Perms :value="NOTIFY_PERMS.delivery.retry">
          <el-button v-if="canRetry(row)" link type="primary" size="small" @click="emit('retry', row)">
            重试
          </el-button>
        </Perms>
        <span v-if="!canRetry(row)" class="muted">{{ labelOf(deliveryStatusOptions, row.status) }}</span>
      </template>
    </pure-table>
  </div>
</template>

<style lang="scss" scoped>
.delivery-detail {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.summary {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  padding: 10px 14px;
  border-radius: 8px;
  background: var(--el-fill-color-light);
  font-size: 13px;
}

.ok {
  color: var(--el-color-success);
}

.danger {
  color: var(--el-color-danger);
}

.warn {
  color: var(--el-color-warning);
}

.muted {
  color: var(--el-text-color-secondary);
}
</style>
