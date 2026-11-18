<script setup lang="ts">
import { computed, h, ref } from "vue";
import { ElMessageBox } from "element-plus";
import { message } from "@bit-labs.cn/owl-ui/utils/message";
import { dialogService } from "@bit-labs.cn/owl-ui/components/ReDialog";
import { PureTableBar } from "@bit-labs.cn/owl-ui/components/RePureTableBar";
import { useRenderIcon } from "@bit-labs.cn/owl-ui/components/ReIcon/src/hooks";
import Refresh from "@iconify-icons/ep/refresh";
import View from "@iconify-icons/ep/view";
import RefreshRight from "@iconify-icons/ep/refresh-right";
import Delete from "@iconify-icons/ep/delete";
import { cleanupEvents } from "@bit-labs.cn/owl-admin-ui/api/event";
import EventLogSwitch from "./EventLogSwitch.vue";
import CleanupForm from "./CleanupForm.vue";
import { useOutboxList } from "./useOutboxList";
import { createOutboxColumns } from "./outboxColumns";
import { OutboxStatus, outboxStatusOptions } from "../status";
import { confirmRetryOutbox, openOutboxDetail } from "../useEventActions";
import { ADMIN_PERMS } from "../../../permissions";
import type { EventOutboxRow } from "../types";

const props = defineProps<{ initialStatus?: string | number }>();

const formRef = ref();
const cleanupFormRef = ref();

const {
  form,
  loading,
  dataList,
  pagination,
  autoRefresh,
  onSearch,
  resetForm,
  handleSizeChange,
  handleCurrentChange,
  toggleAutoRefresh
} = useOutboxList(props.initialStatus);

const columns = computed(() => createOutboxColumns(pagination));

function onRetry(row: EventOutboxRow) {
  confirmRetryOutbox(row, () => onSearch());
}

function openCleanup() {
  dialogService.open({
    title: "清理事件历史",
    props: { formInline: { olderThanDays: 30, targets: ["outbox"] } },
    width: "540px",
    draggable: true,
    closeOnClickModal: false,
    closeOnPressEscape: false,
    contentRenderer: ({ options }) =>
      h(CleanupForm, { ref: cleanupFormRef, formInline: options.props.formInline }),
    beforeSure: done => {
      const formEl = cleanupFormRef.value?.getRef();
      const data = cleanupFormRef.value?.getFormData();
      formEl?.validate(async (valid: boolean) => {
        if (!valid) return;
        try {
          await ElMessageBox.confirm(
            `将删除 ${data.olderThanDays} 天之前已结束的事件记录，删除后不可恢复，确认继续吗？`,
            "确认清理",
            {
              confirmButtonText: "清理",
              cancelButtonText: "取消",
              type: "warning",
              draggable: true,
              closeOnClickModal: false,
              closeOnPressEscape: false
            }
          );
        } catch {
          return;
        }

        try {
          const res = await cleanupEvents({
            olderThanDays: data.olderThanDays,
            targets: data.targets
          });
          const result = (res?.data ?? {}) as unknown as {
            outboxDeleted?: number;
            inboxDeleted?: number;
          };
          done();
          message(
            `已删除：发件箱 ${result.outboxDeleted ?? 0} 条，收件箱 ${result.inboxDeleted ?? 0} 条`,
            { type: "success" }
          );
          onSearch();
        } catch (err) {
          message(err instanceof Error ? err.message : "清理失败", {
            type: "error"
          });
        }
      });
    }
  });
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
      <el-form-item label="事件ID：" prop="eventId">
        <el-input
          v-model="form.eventId"
          placeholder="精确匹配"
          clearable
          class="!w-[240px]"
        />
      </el-form-item>
      <el-form-item label="事件类型：" prop="eventType">
        <el-input
          v-model="form.eventType"
          placeholder="支持模糊匹配"
          clearable
          class="!w-[200px]"
        />
      </el-form-item>
      <el-form-item label="来源：" prop="source">
        <el-input
          v-model="form.source"
          placeholder="来源子系统"
          clearable
          class="!w-[160px]"
        />
      </el-form-item>
      <el-form-item label="业务实体：" prop="subject">
        <el-input
          v-model="form.subject"
          placeholder="如 issue:1001"
          clearable
          class="!w-[180px]"
        />
      </el-form-item>
      <el-form-item label="状态：" prop="status">
        <el-select
          v-model="form.status"
          placeholder="全部"
          clearable
          class="!w-[140px]"
        >
          <el-option
            v-for="item in outboxStatusOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
      </el-form-item>
      <el-form-item label="创建时间：" prop="createdAt">
        <el-date-picker
          v-model="form.createdAt"
          type="datetimerange"
          range-separator="至"
          start-placeholder="开始时间"
          end-placeholder="结束时间"
          value-format="YYYY-MM-DD HH:mm:ss"
          class="!w-[360px]"
          clearable
        />
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

    <PureTableBar title="发件箱" :columns="columns" @refresh="onSearch()">
      <template #title>
        <EventLogSwitch />
      </template>
      <template #buttons>
        <el-switch
          :model-value="autoRefresh"
          active-text="自动刷新"
          class="mr-4"
          @update:model-value="toggleAutoRefresh"
        />
        <Perms :value="ADMIN_PERMS.event.cleanup">
          <el-button :icon="useRenderIcon(Delete)" @click="openCleanup">
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
            <Perms :value="ADMIN_PERMS.event.outboxDetail">
              <el-button
                class="reset-margin"
                link
                type="primary"
                :size="size"
                :icon="useRenderIcon(View)"
                @click="openOutboxDetail(row)"
              >
                详情
              </el-button>
            </Perms>
            <Perms :value="ADMIN_PERMS.event.retryOutbox">
              <el-button
                v-if="row.status !== OutboxStatus.Sent"
                class="reset-margin"
                link
                type="warning"
                :size="size"
                :icon="useRenderIcon(RefreshRight)"
                @click="onRetry(row)"
              >
                重投
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
