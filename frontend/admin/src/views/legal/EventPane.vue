<script setup lang="ts">
import { ref, h, computed } from "vue";
import { useEventList } from "./useEventList";
import { createEventColumns } from "./eventColumns";
import EventDetail from "./EventDetail.vue";
import LegalSwitch from "./LegalSwitch.vue";
import type { LegalEventRow } from "./types";
import { addDialog } from "@bit-labs.cn/owl-ui/components/ReDialog";
import { PureTableBar } from "@bit-labs.cn/owl-ui/components/RePureTableBar";
import { useRenderIcon } from "@bit-labs.cn/owl-ui/components/ReIcon/src/hooks";
import Refresh from "@iconify-icons/ep/refresh";

const formRef = ref();
const { form, documents, scenes, loading, dataList, pagination, onSearch, resetForm, handleSizeChange, handleCurrentChange } =
  useEventList();

function openUser(row: LegalEventRow) {
  addDialog({
    title: row.userLabel || "用户",
    width: "720px",
    draggable: true,
    alignCenter: true,
    hideFooter: true,
    closeOnClickModal: true,
    contentRenderer: () => h(EventDetail, { event: row })
  });
}

const columns = computed(() => createEventColumns(pagination, openUser));
</script>

<template>
  <div>
    <el-form ref="formRef" :inline="true" :model="form" class="search-form bg-bg_color pl-8 pt-[12px]">
      <el-form-item label="条款：" prop="documentKey">
        <el-select v-model="form.documentKey" placeholder="全部" clearable class="!w-[180px]">
          <el-option v-for="item in documents" :key="item.key" :label="item.name" :value="item.key" />
        </el-select>
      </el-form-item>
      <el-form-item label="场景：" prop="sceneKey">
        <el-select v-model="form.sceneKey" placeholder="全部" clearable class="!w-[180px]">
          <el-option v-for="item in scenes" :key="item.key" :label="item.name" :value="item.key" />
        </el-select>
      </el-form-item>
      <el-form-item label="动作：" prop="action">
        <el-select v-model="form.action" placeholder="全部" clearable class="!w-[120px]">
          <el-option label="同意" value="accept" />
          <el-option label="拒绝" value="decline" />
          <el-option label="撤回" value="withdraw" />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" :icon="useRenderIcon('ri:search-line')" :loading="loading" @click="onSearch">
          搜索
        </el-button>
        <el-button :icon="useRenderIcon(Refresh)" @click="resetForm(formRef)">重置</el-button>
      </el-form-item>
    </el-form>
    <PureTableBar title="同意记录" :columns="columns" @refresh="onSearch">
      <template #title>
        <LegalSwitch />
      </template>
      <template #default="{ size, dynamicColumns }">
        <pure-table
          border
          adaptive
          :adaptiveConfig="{ offsetBottom: 160 }"
          align-whole="center"
          row-key="id"
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
        />
      </template>
    </PureTableBar>
  </div>
</template>
