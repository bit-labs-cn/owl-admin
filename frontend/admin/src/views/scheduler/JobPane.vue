<script setup lang="ts">
import { computed, ref } from "vue";
import { PureTableBar } from "@bit-labs.cn/owl-ui/components/RePureTableBar";
import { useRenderIcon } from "@bit-labs.cn/owl-ui/components/ReIcon/src/hooks";
import { message } from "@bit-labs.cn/owl-ui/utils/message";
import { changeSchedulerJobStatus } from "@bit-labs.cn/owl-admin-ui/api/scheduler";
import { ADMIN_PERMS } from "../../permissions";
import SchedulerSwitch from "./SchedulerSwitch.vue";
import { useJobs } from "./useJobs";
import { createJobColumns } from "./columns";
import {
  SchedulerJobSource,
  SchedulerJobStatus,
  canTriggerJob,
  jobSourceOptions,
  jobStatusOptions
} from "./constants";
import {
  confirmDeleteJob,
  openJobDetail,
  openJobForm,
  openTriggerDialog
} from "./useSchedulerActions";
import type { SchedulerJobRow } from "./types";

import AddFill from "@iconify-icons/ri/add-circle-line";
import Refresh from "@iconify-icons/ep/refresh";
import View from "@iconify-icons/ep/view";
import EditPen from "@iconify-icons/ep/edit-pen";
import Delete from "@iconify-icons/ep/delete";
import VideoPlay from "@iconify-icons/ep/video-play";
import List from "@iconify-icons/ep/list";

const props = defineProps<{
  tab: "jobs" | "runs";
}>();

const emit = defineEmits<{
  (e: "update:tab", value: "jobs" | "runs"): void;
  (e: "open-runs", row: SchedulerJobRow): void;
  (e: "changed"): void;
}>();

const formRef = ref();

const {
  form,
  loading,
  dataList,
  switchLoadMap,
  pagination,
  onSearch,
  resetForm,
  handleSizeChange,
  handleCurrentChange
} = useJobs();

const columns = computed(() =>
  createJobColumns({ pagination, switchLoadMap, onChangeStatus })
);

function reload() {
  onSearch(true);
  emit("changed");
}

function onChangeStatus({ row, index }: { row: SchedulerJobRow; index: number }) {
  const target = row.status;
  const previous =
    target === SchedulerJobStatus.Enabled
      ? SchedulerJobStatus.Disabled
      : SchedulerJobStatus.Enabled;

  switchLoadMap.value[index] = { loading: true };
  changeSchedulerJobStatus(row.id, target)
    .then(() => {
      message(target === SchedulerJobStatus.Enabled ? "已启用" : "已停用", {
        type: "success"
      });
      reload();
    })
    .catch(err => {
      // 接口失败时把开关拨回去，避免界面状态与后端不一致
      row.status = previous;
      message(err instanceof Error ? err.message : "状态修改失败", {
        type: "error"
      });
    })
    .finally(() => {
      switchLoadMap.value[index] = { loading: false };
    });
}

function handleCreate() {
  openJobForm(null, reload);
}

function handleEdit(row: SchedulerJobRow) {
  openJobForm(row, reload);
}

function handleDelete(row: SchedulerJobRow) {
  confirmDeleteJob(row, reload);
}

function handleTrigger(row: SchedulerJobRow) {
  // 触发后直接跳到该任务的执行记录，方便立即看到结果
  openTriggerDialog(row, () => {
    emit("changed");
    handleOpenRuns(row);
  });
}

function handleOpenRuns(row: SchedulerJobRow) {
  emit("open-runs", row);
}
</script>

<template>
  <div>
    <el-form
      ref="formRef"
      :inline="true"
      :model="form"
      class="search-form bg-bg_color w-[99/100] pl-8 pt-[12px] overflow-auto"
    >
      <el-form-item label="任务名称：" prop="name">
        <el-input
          v-model="form.name"
          placeholder="支持模糊匹配"
          clearable
          class="!w-[200px]"
        />
      </el-form-item>
      <el-form-item label="所属子系统：" prop="ownerApp">
        <el-input
          v-model="form.ownerApp"
          placeholder="如 cms"
          clearable
          class="!w-[160px]"
        />
      </el-form-item>
      <el-form-item label="来源：" prop="source">
        <el-select v-model="form.source" placeholder="全部" clearable class="!w-[130px]">
          <el-option
            v-for="item in jobSourceOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
      </el-form-item>
      <el-form-item label="状态：" prop="status">
        <el-select v-model="form.status" placeholder="全部" clearable class="!w-[130px]">
          <el-option
            v-for="item in jobStatusOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-button
          type="primary"
          :icon="useRenderIcon('ri:search-line')"
          :loading="loading"
          @click="onSearch()"
        >
          搜索
        </el-button>
        <el-button :icon="useRenderIcon(Refresh)" @click="resetForm(formRef)">
          重置
        </el-button>
      </el-form-item>
    </el-form>

    <PureTableBar title="定时任务" :columns="columns" @refresh="reload">
      <template #title>
        <SchedulerSwitch
          :model-value="props.tab"
          @update:model-value="value => emit('update:tab', value)"
        />
      </template>
      <template #buttons>
        <Perms :value="ADMIN_PERMS.scheduler.create">
          <el-button type="primary" :icon="useRenderIcon(AddFill)" @click="handleCreate">
            新建任务
          </el-button>
        </Perms>
      </template>
      <template #default="{ size, dynamicColumns }">
        <pure-table
          border
          adaptive
          :adaptiveConfig="{ offsetBottom: 160 }"
          align-whole="center"
          row-key="id"
          showOverflowTooltip
          table-layout="auto"
          :loading="loading"
          :size="size"
          :data="dataList"
          :columns="dynamicColumns"
          :pagination="{ ...pagination, size }"
          :header-cell-style="{
            background: 'var(--el-fill-color-light)',
            color: 'var(--el-text-color-primary)'
          }"
          @page-size-change="handleSizeChange"
          @page-current-change="handleCurrentChange"
        >
          <template #operation="{ row }">
            <Perms :value="ADMIN_PERMS.scheduler.detail">
              <el-button
                class="reset-margin"
                link
                type="primary"
                :size="size"
                :icon="useRenderIcon(View)"
                @click="openJobDetail(row)"
              >
                详情
              </el-button>
            </Perms>
            <Perms :value="ADMIN_PERMS.scheduler.trigger">
              <el-button
                class="reset-margin"
                link
                type="success"
                :size="size"
                :icon="useRenderIcon(VideoPlay)"
                :disabled="!canTriggerJob(row)"
                @click="handleTrigger(row)"
              >
                触发
              </el-button>
            </Perms>
            <Perms :value="ADMIN_PERMS.scheduler.runs">
              <el-button
                class="reset-margin"
                link
                type="info"
                :size="size"
                :icon="useRenderIcon(List)"
                @click="handleOpenRuns(row)"
              >
                记录
              </el-button>
            </Perms>
            <Perms :value="ADMIN_PERMS.scheduler.update">
              <el-button
                v-if="row.source === SchedulerJobSource.Custom"
                class="reset-margin"
                link
                type="primary"
                :size="size"
                :icon="useRenderIcon(EditPen)"
                @click="handleEdit(row)"
              >
                编辑
              </el-button>
            </Perms>
            <Perms :value="ADMIN_PERMS.scheduler.delete">
              <el-button
                v-if="row.source === SchedulerJobSource.Custom"
                class="reset-margin"
                link
                type="danger"
                :size="size"
                :icon="useRenderIcon(Delete)"
                @click="handleDelete(row)"
              >
                删除
              </el-button>
            </Perms>
          </template>
        </pure-table>
      </template>
    </PureTableBar>
  </div>
</template>

<style lang="scss" scoped>
:deep(.el-table__inner-wrapper::before) {
  height: 0;
}

.search-form {
  :deep(.el-form-item) {
    margin-bottom: 12px;
  }
}
</style>
