<script setup lang="ts">
import { h, ref } from "vue";
import { message as toast } from "@bit-labs.cn/owl-ui/utils/message";
import { PureTableBar } from "@bit-labs.cn/owl-ui/components/RePureTableBar";
import { dialogService } from "@bit-labs.cn/owl-ui/components/ReDialog";
import { useRenderIcon } from "@bit-labs.cn/owl-ui/components/ReIcon/src/hooks";
import Refresh from "@iconify-icons/ep/refresh";
import { tableHeaderStyle } from "../../format";
import { NOTIFY_PERMS } from "../../permissions";
import MessageDetail from "./MessageDetail.vue";
import { createMessageColumns } from "./columns";
import type { MessageRow } from "./types";
import { useMessageList } from "./useMessageList";

defineOptions({ name: "NotifyMessage" });

const formRef = ref();
const {
  form,
  typeOptions,
  dataList,
  loading,
  detail,
  pagination,
  onSearch,
  loadDetail,
  resetForm,
  handleSizeChange,
  handleCurrentChange
} = useMessageList();
const columns = createMessageColumns();

async function showDetail(row: MessageRow) {
  try {
    await loadDetail(row);
  } catch {
    toast("消息详情加载失败", { type: "error" });
    return;
  }
  dialogService.open({
    title: "消息接收情况",
    width: "860px",
    hideFooter: true,
    draggable: true,
    closeOnClickModal: false,
    closeOnPressEscape: false,
    alignCenter: true,
    contentRenderer: () => h(MessageDetail, { detail: detail.value })
  });
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
      <el-form-item label="标题：" prop="title">
        <el-input v-model="form.title" clearable class="!w-[200px]" />
      </el-form-item>
      <el-form-item label="消息类型：" prop="typeCode">
        <el-select v-model="form.typeCode" clearable placeholder="全部" class="!w-[180px]">
          <el-option v-for="item in typeOptions" :key="item.code" :label="item.name" :value="item.code" />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" :icon="useRenderIcon('ri:search-line')" :loading="loading" @click="onSearch">
          搜索
        </el-button>
        <el-button :icon="useRenderIcon(Refresh)" @click="resetForm(formRef)">重置</el-button>
      </el-form-item>
    </el-form>

    <PureTableBar title="消息记录" :columns="columns" @refresh="onSearch">
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
            <el-empty description="还没有发过消息，去「发送通知」发一条试试。" />
          </template>
          <template #operation="{ row }">
            <Perms :value="NOTIFY_PERMS.message.detail">
              <el-button link type="primary" :size="size" @click="showDetail(row)">接收情况</el-button>
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

.read-cell {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.read-text {
  font-size: 12px;
  color: var(--el-text-color-regular);
}

.unread {
  margin-left: 6px;
  color: var(--el-color-warning);
}

.all-read {
  margin-left: 6px;
  color: var(--el-color-success);
}

.muted {
  color: var(--el-text-color-secondary);
}
</style>
