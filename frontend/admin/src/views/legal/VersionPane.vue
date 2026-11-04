<script setup lang="ts">
import { ref, h, computed } from "vue";
import { useVersionList } from "./useVersionList";
import { createVersionColumns } from "./versionColumns";
import VersionForm from "./VersionForm.vue";
import PublishForm from "./PublishForm.vue";
import VersionContent from "./VersionContent.vue";
import LegalSwitch from "./LegalSwitch.vue";
import type { LegalVersionForm, LegalVersionRow } from "./types";
import { legalAPI } from "@bit-labs.cn/owl-admin-ui/api/legal";
import { addDialog } from "@bit-labs.cn/owl-ui/components/ReDialog";
import { PureTableBar } from "@bit-labs.cn/owl-ui/components/RePureTableBar";
import { useRenderIcon } from "@bit-labs.cn/owl-ui/components/ReIcon/src/hooks";

import AddFill from "@iconify-icons/ri/add-circle-line";
import Refresh from "@iconify-icons/ep/refresh";
import EditPen from "@iconify-icons/ep/edit-pen";
import Delete from "@iconify-icons/ep/delete";
import View from "@iconify-icons/ep/view";
import Upload from "@iconify-icons/ep/upload";

const versionFormRef = ref();
const publishFormRef = ref();
const { documents, form, loading, dataList, pagination, loadDocuments, onSearch, handleSizeChange, handleCurrentChange } =
  useVersionList();
const columns = computed(() => createVersionColumns(pagination));

function openDialog(title = "新增", row?: LegalVersionRow) {
  addDialog({
    title: `${title}草稿`,
    props: {
      formInline: {
        id: row?.id,
        documentId: form.documentId,
        version: row?.version ?? "",
        title: row?.title ?? "",
        content: row?.content ?? ""
      }
    },
    width: "720px",
    draggable: true,
    alignCenter: true,
    closeOnClickModal: false,
    contentRenderer: ({ options }) =>
      h(VersionForm, { ref: versionFormRef, formInline: options.props.formInline }),
    beforeSure: done => {
      const FormRef = versionFormRef.value.getRef();
      const curData = versionFormRef.value.getFormData() as LegalVersionForm;
      FormRef.validate((valid: boolean) => {
        if (!valid) return;
        const request = curData.id
          ? legalAPI.updateVersion({
              id: curData.id,
              version: curData.version,
              title: curData.title,
              content: curData.content
            })
          : legalAPI.createVersion({
              documentId: form.documentId,
              version: curData.version,
              title: curData.title,
              content: curData.content
            });
        request.then(() => {
          done();
          onSearch();
        });
      });
    }
  });
}

function openPublish(row: LegalVersionRow) {
  addDialog({
    title: `发布 ${row.version}`,
    width: "520px",
    draggable: true,
    alignCenter: true,
    closeOnClickModal: false,
    contentRenderer: () => h(PublishForm, { ref: publishFormRef }),
    beforeSure: done => {
      const FormRef = publishFormRef.value.getRef();
      FormRef.validate((valid: boolean) => {
        if (!valid) return;
        const changeClass = publishFormRef.value.getFormData().changeClass as string;
        legalAPI.publishVersion(row.id, changeClass).then(() => {
          done();
          onSearch();
        });
      });
    }
  });
}

function openContent(row: LegalVersionRow) {
  addDialog({
    title: "查看正文",
    width: "720px",
    draggable: true,
    alignCenter: true,
    hideFooter: true,
    closeOnClickModal: true,
    contentRenderer: () => h(VersionContent, { title: row.title, content: row.content })
  });
}

function handleDelete(row: LegalVersionRow) {
  legalAPI.deleteVersion(row.id).then(() => onSearch());
}
</script>

<template>
  <div>
    <el-form :inline="true" class="search-form bg-bg_color pl-8 pt-[12px]">
      <el-form-item label="条款：">
        <el-select v-model="form.documentId" placeholder="选择条款" class="!w-[240px]" @visible-change="open => open && loadDocuments()">
          <el-option
            v-for="item in documents"
            :key="item.id"
            :label="item.status === 'disabled' ? `${item.name}（停用）` : item.name"
            :value="item.id"
          />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-button :icon="useRenderIcon(Refresh)" @click="loadDocuments().then(() => onSearch())">刷新</el-button>
      </el-form-item>
    </el-form>
    <PureTableBar title="版本" :columns="columns" @refresh="onSearch">
      <template #title>
        <LegalSwitch />
      </template>
      <template #buttons>
        <el-button type="primary" :icon="useRenderIcon(AddFill)" :disabled="!form.documentId" @click="openDialog()">
          新增草稿
        </el-button>
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
            <el-button link type="primary" :size="size" :icon="useRenderIcon(View)" @click="openContent(row)">
              查看
            </el-button>
            <el-button
              v-if="row.status === 'draft'"
              link
              type="primary"
              :size="size"
              :icon="useRenderIcon(EditPen)"
              @click="openDialog('修改', row)"
            >
              修改
            </el-button>
            <el-button
              v-if="row.status === 'draft'"
              link
              type="primary"
              :size="size"
              :icon="useRenderIcon(Upload)"
              @click="openPublish(row)"
            >
              发布
            </el-button>
            <el-popconfirm v-if="row.status === 'draft'" :title="`删除草稿 ${row.version}？`" @confirm="handleDelete(row)">
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
