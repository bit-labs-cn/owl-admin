<script setup lang="ts">
import { h, ref } from "vue";
import { message } from "@bit-labs.cn/owl-ui/utils/message";
import { dialogService } from "@bit-labs.cn/owl-ui/components/ReDialog";
import { PureTableBar } from "@bit-labs.cn/owl-ui/components/RePureTableBar";
import { useRenderIcon } from "@bit-labs.cn/owl-ui/components/ReIcon/src/hooks";
import Refresh from "@iconify-icons/ep/refresh";
import { inboxAPI } from "../../api/notify";
import { confirmAction } from "../../confirm";
import { KindTodo, ReadStatusUnread, TodoStatusDone, TodoStatusPending } from "../../constants";
import { tableHeaderStyle } from "../../format";
import { readStatusOptions } from "../../labels";
import InboxDetail from "./InboxDetail.vue";
import { createInboxColumns } from "./columns";
import type { InboxRow } from "./types";
import { useInboxList } from "./useInboxList";

defineOptions({ name: "NotifyInbox" });

const formRef = ref();
const { form, typeOptions, loading, dataList, counts, pagination, onSearch, resetForm, handleSizeChange, handleCurrentChange } =
  useInboxList();
const columns = createInboxColumns();

function openDetail(row: InboxRow) {
  dialogService.open({
    title: "消息详情",
    width: "640px",
    hideFooter: true,
    draggable: true,
    closeOnClickModal: false,
    closeOnPressEscape: false,
    alignCenter: true,
    contentRenderer: () => h(InboxDetail, { row })
  });
  if (row.readStatus === ReadStatusUnread) {
    inboxAPI.read({ ids: [row.id] }).then(() => onSearch());
  }
}

async function markRead(row: InboxRow) {
  await inboxAPI.read({ ids: [row.id] });
  message("已标为已读", { type: "success" });
  onSearch();
}

async function readAll() {
  if (!(await confirmAction("确认把全部消息标为已读吗？", "全部已读"))) return;
  await inboxAPI.readAll();
  message("已全部标为已读", { type: "success" });
  onSearch();
}

async function remove(row: InboxRow) {
  if (!(await confirmAction(`确认删除「${row.title}」吗？`, "删除消息"))) return;
  await inboxAPI.remove(row.id);
  message("已删除", { type: "success" });
  onSearch();
}

async function done(row: InboxRow) {
  await inboxAPI.done(row.id);
  message("待办已完成", { type: "success" });
  onSearch();
}

async function reopen(row: InboxRow) {
  await inboxAPI.reopen(row.id);
  message("待办已重开", { type: "success" });
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
      <el-form-item label="标题：" prop="title">
        <el-input v-model="form.title" clearable class="!w-[200px]" />
      </el-form-item>
      <el-form-item label="消息类型：" prop="typeCode">
        <el-select v-model="form.typeCode" clearable placeholder="全部" class="!w-[180px]">
          <el-option v-for="item in typeOptions" :key="item.code" :label="item.name" :value="item.code" />
        </el-select>
      </el-form-item>
      <el-form-item label="阅读：" prop="readStatus">
        <el-select v-model="form.readStatus" clearable placeholder="全部" class="!w-[140px]">
          <el-option v-for="item in readStatusOptions" :key="item.value" :label="item.label" :value="item.value" />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" :icon="useRenderIcon('ri:search-line')" :loading="loading" @click="onSearch">
          搜索
        </el-button>
        <el-button :icon="useRenderIcon(Refresh)" @click="resetForm(formRef)">重置</el-button>
      </el-form-item>
    </el-form>
    <PureTableBar title="我的消息" :columns="columns" @refresh="onSearch">
      <template #buttons>
        <span class="counts">
          未读通知 {{ counts.unreadNotice ?? counts.notice ?? 0 }} · 待办事项 {{ counts.pendingTodo ?? counts.todo ?? 0 }}
        </span>
        <el-button @click="readAll">全部已读</el-button>
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
          <template #operation="{ row }">
            <el-button link type="primary" :size="size" @click="openDetail(row)">查看</el-button>
            <el-button v-if="row.readStatus === ReadStatusUnread" link type="primary" :size="size" @click="markRead(row)">
              已读
            </el-button>
            <el-button
              v-if="row.kind === KindTodo && row.todoStatus === TodoStatusPending"
              link
              type="success"
              :size="size"
              @click="done(row)"
            >
              完成
            </el-button>
            <el-button
              v-if="row.kind === KindTodo && row.todoStatus === TodoStatusDone"
              link
              type="warning"
              :size="size"
              @click="reopen(row)"
            >
              重开
            </el-button>
            <el-button link type="danger" :size="size" @click="remove(row)">删除</el-button>
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

.counts {
  margin-right: 12px;
  color: var(--el-text-color-secondary);
}
</style>
