<script setup lang="ts">
import { computed, ref } from "vue";
import { PureTableBar } from "@bit-labs.cn/owl-ui/components/RePureTableBar";
import { useRenderIcon } from "@bit-labs.cn/owl-ui/components/ReIcon/src/hooks";
import { ADMIN_PERMS } from "../../permissions";
import SchedulerSwitch from "./SchedulerSwitch.vue";
import { useRuns } from "./useRuns";
import { createRunColumns } from "./runColumns";
import {
  canCancelRun,
  canRetryRun,
  runStatusOptions,
  triggerTypeOptions
} from "./constants";
import {
  confirmCancelRun,
  confirmRetryRun,
  openCleanupRuns,
  openRunDetail
} from "./useSchedulerActions";

import Refresh from "@iconify-icons/ep/refresh";
import View from "@iconify-icons/ep/view";
import RefreshRight from "@iconify-icons/ep/refresh-right";
import CircleClose from "@iconify-icons/ep/circle-close";
import Delete from "@iconify-icons/ep/delete";

const props = defineProps<{
  tab: "jobs" | "runs";
  jobFilter?: { jobId: string; jobCode: string } | null;
}>();

const emit = defineEmits<{
  (e: "update:tab", value: "jobs" | "runs"): void;
  (e: "changed"): void;
}>();

const formRef = ref();

const {
  form,
  jobFilter,
  loading,
  dataList,
  autoRefresh,
  pagination,
  onSearch,
  resetForm,
  handleSizeChange,
  handleCurrentChange,
  toggleAutoRefresh
} = useRuns();

// 从任务列表「记录」跳进来时，按该任务过滤；组件是 v-if 挂载的，初始化时直接生效
if (props.jobFilter) {
  jobFilter.value = { ...props.jobFilter };
  form.jobCode = props.jobFilter.jobCode;
}

const columns = computed(() => createRunColumns(pagination));

function reload() {
  onSearch(true);
  emit("changed");
}

function clearFilter() {
  jobFilter.value = null;
  form.jobCode = "";
  pagination.currentPage = 1;
  onSearch();
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
      <el-form-item label="任务编码：" prop="jobCode">
        <el-input
          v-model="form.jobCode"
          placeholder="精确匹配"
          clearable
          class="!w-[220px]"
        />
      </el-form-item>
      <el-form-item label="状态：" prop="status">
        <el-select v-model="form.status" placeholder="全部" clearable class="!w-[140px]">
          <el-option
            v-for="item in runStatusOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
      </el-form-item>
      <el-form-item label="触发方式：" prop="triggerType">
        <el-select
          v-model="form.triggerType"
          placeholder="全部"
          clearable
          class="!w-[130px]"
        >
          <el-option
            v-for="item in triggerTypeOptions"
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
        <el-button v-if="jobFilter" link type="primary" @click="clearFilter">
          清除任务筛选
        </el-button>
      </el-form-item>
    </el-form>

    <PureTableBar title="执行记录" :columns="columns" @refresh="reload">
      <template #title>
        <SchedulerSwitch
          :model-value="props.tab"
          @update:model-value="value => emit('update:tab', value)"
        />
      </template>
      <template #buttons>
        <el-switch
          :model-value="autoRefresh"
          active-text="自动刷新"
          class="mr-4"
          @update:model-value="toggleAutoRefresh"
        />
        <Perms :value="ADMIN_PERMS.scheduler.cleanup">
          <el-button :icon="useRenderIcon(Delete)" @click="openCleanupRuns(reload)">
            清理历史
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
            <Perms :value="ADMIN_PERMS.scheduler.runDetail">
              <el-button
                class="reset-margin"
                link
                type="primary"
                :size="size"
                :icon="useRenderIcon(View)"
                @click="openRunDetail(row)"
              >
                详情
              </el-button>
            </Perms>
            <Perms :value="ADMIN_PERMS.scheduler.retryRun">
              <el-button
                v-if="canRetryRun(row.status)"
                class="reset-margin"
                link
                type="warning"
                :size="size"
                :icon="useRenderIcon(RefreshRight)"
                @click="confirmRetryRun(row, reload)"
              >
                重试
              </el-button>
            </Perms>
            <Perms :value="ADMIN_PERMS.scheduler.cancelRun">
              <el-button
                v-if="canCancelRun(row.status)"
                class="reset-margin"
                link
                type="danger"
                :size="size"
                :icon="useRenderIcon(CircleClose)"
                @click="confirmCancelRun(row, reload)"
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
:deep(.el-table__inner-wrapper::before) {
  height: 0;
}

.search-form {
  :deep(.el-form-item) {
    margin-bottom: 12px;
  }
}
</style>
