<script setup lang="ts">
import { computed, ref } from "vue";
import { PureTableBar } from "@bit-labs.cn/owl-ui/components/RePureTableBar";
import { useRenderIcon } from "@bit-labs.cn/owl-ui/components/ReIcon/src/hooks";
import Refresh from "@iconify-icons/ep/refresh";
import View from "@iconify-icons/ep/view";
import EventLogSwitch from "./EventLogSwitch.vue";
import { useInboxList } from "./useInboxList";
import { createInboxColumns } from "./inboxColumns";
import { inboxStatusOptions } from "../status";
import { openInboxDetail } from "../useEventActions";
import { ADMIN_PERMS } from "../../../permissions";

const props = defineProps<{ initialStatus?: string | number }>();

const formRef = ref();

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
} = useInboxList(props.initialStatus);

const columns = computed(() => createInboxColumns(pagination));
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
      <el-form-item label="消费组：" prop="consumerGroup">
        <el-input
          v-model="form.consumerGroup"
          placeholder="支持模糊匹配"
          clearable
          class="!w-[220px]"
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
            v-for="item in inboxStatusOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
      </el-form-item>
      <el-form-item label="处理时间：" prop="processedAt">
        <el-date-picker
          v-model="form.processedAt"
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

    <PureTableBar title="收件箱" :columns="columns" @refresh="onSearch()">
      <template #title>
        <EventLogSwitch />
      </template>
      <template #buttons>
        <el-switch
          :model-value="autoRefresh"
          active-text="自动刷新"
          @update:model-value="toggleAutoRefresh"
        />
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
            <Perms :value="ADMIN_PERMS.event.inboxDetail">
              <el-button
                class="reset-margin"
                link
                type="primary"
                :size="size"
                :icon="useRenderIcon(View)"
                @click="openInboxDetail(row)"
              >
                详情
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
