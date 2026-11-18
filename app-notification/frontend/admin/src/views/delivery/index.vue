<script setup lang="ts">
import { h, ref } from "vue";
import { message } from "@bit-labs.cn/owl-ui/utils/message";
import { dialogService } from "@bit-labs.cn/owl-ui/components/ReDialog";
import { PureTableBar } from "@bit-labs.cn/owl-ui/components/RePureTableBar";
import { useRenderIcon } from "@bit-labs.cn/owl-ui/components/ReIcon/src/hooks";
import Refresh from "@iconify-icons/ep/refresh";
import { deliveryAPI } from "../../api/notify";
import { confirmAction } from "../../confirm";
import { DeliveryStatusFailed } from "../../constants";
import { tableHeaderStyle } from "../../format";
import { channelOptions, deliveryStatusOptions, recipientTypeOptions } from "../../labels";
import { NOTIFY_PERMS } from "../../permissions";
import DeliveryDetail from "./DeliveryDetail.vue";
import { createDeliveryColumns } from "./columns";
import { createDeliveryDetailColumns } from "./detailColumns";
import type { DeliveryBatchRow, DeliveryRow } from "./types";
import { useDeliveryList } from "./useDeliveryList";

defineOptions({ name: "NotifyDelivery" });

const formRef = ref();
const {
  form,
  dataList,
  loading,
  detailList,
  detailLoading,
  pagination,
  onSearch,
  loadDetail,
  resetForm,
  handleSizeChange,
  handleCurrentChange
} = useDeliveryList();
const columns = createDeliveryColumns();
const detailColumns = createDeliveryDetailColumns();

async function openDetail(row: DeliveryBatchRow) {
  await loadDetail(row);
  dialogService.open({
    title: `发送明细 · ${row.title}`,
    width: "980px",
    hideFooter: true,
    draggable: true,
    closeOnClickModal: false,
    closeOnPressEscape: false,
    alignCenter: true,
    contentRenderer: () =>
      h(DeliveryDetail, {
        rows: detailList.value,
        loading: detailLoading.value,
        columns: detailColumns,
        onRetry: (item: DeliveryRow) => retryOne(item)
      })
  });
}

async function retryOne(row: DeliveryRow) {
  if (!(await confirmAction(`确认重试「${row.title}」到 ${row.userName} 吗？`, "重试发送"))) return;
  await deliveryAPI.retry(row.id);
  message("已重新提交", { type: "success" });
  onSearch();
}
</script>

<template>
  <div class="main">
    <el-form
      ref="formRef"
      :inline="true"
      :model="form"
      class="search-form bg-bg_color w-[99/100] pl-8 pt-[12px] overflow-auto"
    >
      <el-form-item label="渠道：" prop="channel">
        <el-select v-model="form.channel" clearable placeholder="全部" class="!w-[140px]">
          <el-option v-for="item in channelOptions" :key="item.value" :label="item.label" :value="item.value" />
        </el-select>
      </el-form-item>
      <el-form-item label="状态：" prop="status">
        <el-select v-model="form.status" clearable placeholder="全部" class="!w-[140px]">
          <el-option v-for="item in deliveryStatusOptions" :key="item.value" :label="item.label" :value="item.value" />
        </el-select>
      </el-form-item>
      <el-form-item label="接收人类型：" prop="recipientType">
        <el-select v-model="form.recipientType" clearable placeholder="全部" class="!w-[140px]">
          <el-option
            v-for="item in recipientTypeOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" :icon="useRenderIcon('ri:search-line')" :loading="loading" @click="onSearch">
          搜索
        </el-button>
        <el-button :icon="useRenderIcon(Refresh)" @click="resetForm(formRef)">重置</el-button>
      </el-form-item>
    </el-form>

    <PureTableBar title="发送记录" :columns="columns" @refresh="onSearch">
      <template #buttons>
        <span class="record-hint">
          一次发送一行，点击「明细」查看该次发送的全部渠道记录。筛选条件作用于明细。
        </span>
      </template>
      <template #default="{ size, dynamicColumns }">
        <pure-table
          border
          adaptive
          :adaptiveConfig="{ offsetBottom: 108 }"
          align-whole="center"
          row-key="messageId"
          showOverflowTooltip
          :loading="loading"
          :size="size"
          :data="dataList"
          :columns="dynamicColumns"
          :pagination="{ ...pagination, size }"
          :header-cell-style="tableHeaderStyle"
          @page-size-change="handleSizeChange"
          @page-current-change="handleCurrentChange"
        >
          <template #empty>
            <el-empty description="还没有发送记录，发一条通知后再来看。" />
          </template>
          <template #operation="{ row }">
            <Perms :value="NOTIFY_PERMS.delivery.detail">
              <el-button link type="primary" :size="size" @click="openDetail(row)">明细</el-button>
            </Perms>
          </template>
        </pure-table>
      </template>
    </PureTableBar>
  </div>
</template>

<style lang="scss" scoped>
.search-form {
  :deep(.el-form-item) {
    margin-bottom: 12px;
  }
}

.record-hint {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.batch-title {
  display: flex;
  flex-direction: column;
  gap: 2px;
  line-height: 1.4;
}

.batch-meta {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.batch-progress {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.batch-progress-text {
  font-size: 12px;
  color: var(--el-text-color-regular);
}

.batch-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.muted {
  color: var(--el-text-color-secondary);
}
</style>
