<script setup lang="ts">
import { ref, h, computed } from "vue";
import { useDocumentList } from "./useDocumentList";
import { createDocumentColumns } from "./documentColumns";
import DocumentForm from "./DocumentForm.vue";
import LegalSwitch from "./LegalSwitch.vue";
import type { LegalDocumentForm, LegalDocumentRow } from "./types";
import { legalAPI } from "@bit-labs.cn/owl-admin-ui/api/legal";
import { addDialog } from "@bit-labs.cn/owl-ui/components/ReDialog";
import { PureTableBar } from "@bit-labs.cn/owl-ui/components/RePureTableBar";
import { useRenderIcon } from "@bit-labs.cn/owl-ui/components/ReIcon/src/hooks";

import AddFill from "@iconify-icons/ri/add-circle-line";
import Refresh from "@iconify-icons/ep/refresh";
import EditPen from "@iconify-icons/ep/edit-pen";
import Delete from "@iconify-icons/ep/delete";

const formRef = ref();
const documentFormRef = ref();

const {
  form,
  loading,
  dataList,
  pagination,
  switchLoadMap,
  onSearch,
  resetForm,
  onChange,
  handleSizeChange,
  handleCurrentChange
} = useDocumentList();

const columns = computed(() => createDocumentColumns({ switchLoadMap, onChange, pagination }));

function openDialog(title = "新增", row?: LegalDocumentRow) {
  addDialog({
    title: `${title}条款`,
    props: {
      formInline: {
        id: row?.id,
        key: row?.key ?? "",
        name: row?.name ?? "",
        kind: row?.kind ?? "contract",
        status: row?.status ?? "enabled"
      }
    },
    width: "520px",
    draggable: true,
    alignCenter: true,
    closeOnClickModal: false,
    contentRenderer: ({ options }) =>
      h(DocumentForm, { ref: documentFormRef, formInline: options.props.formInline }),
    beforeSure: done => {
      const FormRef = documentFormRef.value.getRef();
      const curData = documentFormRef.value.getFormData() as LegalDocumentForm;
      FormRef.validate((valid: boolean) => {
        if (!valid) return;
        const request = curData.id
          ? legalAPI.updateDocument({
              id: curData.id,
              name: curData.name,
              kind: curData.kind,
              status: curData.status
            } as object)
          : legalAPI.createDocument({
              key: curData.key,
              name: curData.name,
              kind: curData.kind
            });
        request.then(() => {
          done();
          onSearch();
        });
      });
    }
  });
}

function handleDelete(row: LegalDocumentRow) {
  legalAPI.deleteDocument(row.id).then(() => onSearch());
}
</script>

<template>
  <div>
    <el-form ref="formRef" :inline="true" :model="form" class="search-form bg-bg_color pl-8 pt-[12px]">
      <el-form-item label="名称：" prop="name">
        <el-input v-model="form.name" placeholder="条款名称" clearable class="!w-[180px]" />
      </el-form-item>
      <el-form-item label="性质：" prop="kind">
        <el-select v-model="form.kind" placeholder="全部" clearable class="!w-[140px]">
          <el-option label="合同" value="contract" />
          <el-option label="告知" value="notice" />
          <el-option label="可选同意" value="optional" />
        </el-select>
      </el-form-item>
      <el-form-item label="状态：" prop="status">
        <el-select v-model="form.status" placeholder="全部" clearable class="!w-[120px]">
          <el-option label="启用" value="enabled" />
          <el-option label="停用" value="disabled" />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" :icon="useRenderIcon('ri:search-line')" :loading="loading" @click="onSearch">
          搜索
        </el-button>
        <el-button :icon="useRenderIcon(Refresh)" @click="resetForm(formRef)">重置</el-button>
      </el-form-item>
    </el-form>
    <PureTableBar title="条款" :columns="columns" @refresh="onSearch">
      <template #title>
        <LegalSwitch />
      </template>
      <template #buttons>
        <el-button type="primary" :icon="useRenderIcon(AddFill)" @click="openDialog()">新增条款</el-button>
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
        >
          <template #operation="{ row }">
            <el-button link type="primary" :size="size" :icon="useRenderIcon(EditPen)" @click="openDialog('修改', row)">
              修改
            </el-button>
            <el-popconfirm :title="`删除条款「${row.name}」？已有同意记录的不能删`" @confirm="handleDelete(row)">
              <template #reference>
                <el-button link type="danger" :size="size" :icon="useRenderIcon(Delete)">删除</el-button>
              </template>
            </el-popconfirm>
          </template>
        </pure-table>
      </template>
    </PureTableBar>
  </div>
</template>
