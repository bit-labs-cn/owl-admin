<script setup lang="ts">
import { ref } from "vue";
import { message as toast } from "@bit-labs.cn/owl-ui/utils/message";
import { PureTableBar } from "@bit-labs.cn/owl-ui/components/RePureTableBar";
import { useRenderIcon } from "@bit-labs.cn/owl-ui/components/ReIcon/src/hooks";
import Refresh from "@iconify-icons/ep/refresh";
import { scheduleAPI } from "../../api/notify";
import { confirmAction } from "../../confirm";
import {
  ScheduleStatusFailed,
  ScheduleStatusPending
} from "../../constants";
import { tableHeaderStyle } from "../../format";
import { scheduleStatusOptions } from "../../labels";
import { NOTIFY_PERMS } from "../../permissions";
import { createScheduleColumns } from "./columns";
import type { ScheduleRow } from "./types";
import { useScheduleList } from "./useScheduleList";

defineOptions({ name: "NotifySchedule" });

const formRef = ref();
const {
  form,
  dataList,
  loading,
  pagination,
  onSearch,
  resetForm,
  handleSizeChange,
  handleCurrentChange
} = useScheduleList();
const columns = createScheduleColumns();

const busyId = ref("");

async function cancel(row: ScheduleRow) {
  if (!(await confirmAction(`确认取消「${row.title}」的定时发送吗？`, "取消定时发送"))) return;
  busyId.value = row.id;
  try {
    await scheduleAPI.cancel(row.id);
    toast("已取消，不会再发送", { type: "success" });
    onSearch();
  } finally {
    busyId.value = "";
  }
}

async function sendNow(row: ScheduleRow) {
  if (!(await confirmAction(`确认立刻发送「${row.title}」吗？不再等到计划时间。`, "立即发送"))) return;
  busyId.value = row.id;
  try {
    await scheduleAPI.sendNow(row.id);
    toast("已发送", { type: "success" });
    onSearch();
  } finally {
    busyId.value = "";
  }
}

async function retry(row: ScheduleRow) {
  busyId.value = row.id;
  try {
    await scheduleAPI.retry(row.id);
    toast("已重新提交发送", { type: "success" });
    onSearch();
  } finally {
    busyId.value = "";
  }
}

function canCancel(row: ScheduleRow) {
  return row.status === ScheduleStatusPending || row.status === ScheduleStatusFailed;
}

function canSendNow(row: ScheduleRow) {
  return row.status === ScheduleStatusPending;
}

function canRetry(row: ScheduleRow) {
  return row.status === ScheduleStatusFailed;
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
      <el-form-item label="标题：" prop="keyword">
        <el-input v-model="form.keyword" clearable class="!w-[200px]" />
      </el-form-item>
      <el-form-item label="状态：" prop="status">
        <el-select v-model="form.status" clearable placeholder="全部" class="!w-[140px]">
          <el-option v-for="item in scheduleStatusOptions" :key="item.value" :label="item.label" :value="item.value" />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" :icon="useRenderIcon('ri:search-line')" :loading="loading" @click="onSearch">
          搜索
        </el-button>
        <el-button :icon="useRenderIcon(Refresh)" @click="resetForm(formRef)">重置</el-button>
      </el-form-item>
    </el-form>

    <PureTableBar title="定时发送" :columns="columns" @refresh="onSearch">
      <template #buttons>
        <span class="record-hint">
          到点由后台调度发送；接收人在发送当刻按组织关系重新解析。失败的任务可以重试。
        </span>
      </template>
      <template #default="{ size, dynamicColumns }">
        <pure-table
          border
          adaptive
          :adaptiveConfig="{ offsetBottom: 108 }"
          align-whole="center"
          row-key="id"
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
            <el-empty description="还没有定时任务，去「发送通知」选定时发送试试。" />
          </template>
          <template #operation="{ row }">
            <Perms :value="NOTIFY_PERMS.schedule.sendNow">
              <el-button
                v-if="canSendNow(row)"
                link
                type="primary"
                :size="size"
                :loading="busyId === row.id"
                @click="sendNow(row)"
              >
                立即发送
              </el-button>
            </Perms>
            <Perms :value="NOTIFY_PERMS.schedule.retry">
              <el-button
                v-if="canRetry(row)"
                link
                type="warning"
                :size="size"
                :loading="busyId === row.id"
                @click="retry(row)"
              >
                重试
              </el-button>
            </Perms>
            <Perms :value="NOTIFY_PERMS.schedule.cancel">
              <el-button
                v-if="canCancel(row)"
                link
                type="danger"
                :size="size"
                :loading="busyId === row.id"
                @click="cancel(row)"
              >
                取消
              </el-button>
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

.schedule-title {
  display: flex;
  flex-direction: column;
  gap: 2px;
  line-height: 1.4;
}

.schedule-meta {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.schedule-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.schedule-error {
  color: var(--el-color-danger);
  font-size: 12px;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.muted {
  color: var(--el-text-color-secondary);
}
</style>
