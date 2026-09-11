<script setup lang="ts">
import { ref, h, computed } from "vue";
import { useAppVersionList } from "./useAppVersionList";
import { createColumns } from "./columns";
import AppVersionForm from "./AppVersionForm.vue";
import type { AppVersionFormData } from "./types";
import { appVersionAPI } from "@bit-labs.cn/owl-admin-ui/api/app-version";
import { addDialog } from "@bit-labs.cn/owl-ui/components/ReDialog";

import { PureTableBar } from "@bit-labs.cn/owl-ui/components/RePureTableBar";
import { useRenderIcon } from "@bit-labs.cn/owl-ui/components/ReIcon/src/hooks";

import AddFill from "@iconify-icons/ri/add-circle-line";
import Refresh from "@iconify-icons/ep/refresh";
import EditPen from "@iconify-icons/ep/edit-pen";
import Delete from "@iconify-icons/ep/delete";

defineOptions({ name: "SystemAppVersion" });

const formRef = ref();
const tableRef = ref();
const appVersionFormRef = ref();

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
} = useAppVersionList();

const columns = computed(() => createColumns({ switchLoadMap, onChange, pagination }));

function openDialog(title = "新增", row?: AppVersionFormData) {
  addDialog({
    title: `${title}APP版本`,
    props: {
      formInline: {
        id: row?.id ?? "",
        version: row?.version ?? "",
        versionName: row?.versionName ?? "",
        apkUrl: row?.apkUrl ?? "",
        apkType: row?.apkType ?? 2,
        content: row?.content ?? "",
        remark: row?.remark ?? "",
        status: row?.status ?? 1
      }
    },
    width: "640px",
    draggable: true,
    closeOnClickModal: false,
    contentRenderer: ({ options }) =>
      h(AppVersionForm, { ref: appVersionFormRef, formInline: options.props.formInline }),
    beforeSure: done => {
      const FormRef = appVersionFormRef.value.getRef();
      const curData = appVersionFormRef.value.getFormData() as AppVersionFormData;
      function chores() {
        done();
        onSearch();
      }
      FormRef.validate((valid: boolean) => {
        if (valid) {
          const api = curData.id
            ? appVersionAPI.update(curData)
            : appVersionAPI.create(curData);
          api.then(() => chores());
        }
      });
    }
  });
}

function handleDelete(row: { id: string; version?: string }) {
  appVersionAPI.remove(row.id).then(() => {
    onSearch();
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
      <el-form-item label="版本号：" prop="version">
        <el-input
          v-model="form.version"
          placeholder="如 100"
          clearable
          class="!w-[200px]"
        />
      </el-form-item>
      <el-form-item label="平台：" prop="apkType">
        <el-select
          v-model="form.apkType"
          placeholder="请选择"
          clearable
          class="!w-[180px]"
        >
          <el-option label="iOS" :value="1" />
          <el-option label="Android" :value="2" />
        </el-select>
      </el-form-item>
      <el-form-item label="状态：" prop="status">
        <el-select
          v-model="form.status"
          placeholder="请选择"
          clearable
          class="!w-[180px]"
        >
          <el-option label="已启用" :value="1" />
          <el-option label="已停用" :value="0" />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-button
          type="primary"
          :icon="useRenderIcon('ri:search-line')"
          :loading="loading"
          @click="onSearch"
        >
          搜索
        </el-button>
        <el-button :icon="useRenderIcon(Refresh)" @click="resetForm(formRef)">
          重置
        </el-button>
      </el-form-item>
    </el-form>

    <PureTableBar title="APP升级" :columns="columns" @refresh="onSearch">
      <template #buttons>
        <el-button
          type="primary"
          :icon="useRenderIcon(AddFill)"
          @click="openDialog()"
        >
          新增版本
        </el-button>
      </template>
      <template v-slot="{ size, dynamicColumns }">
        <pure-table
          border
          ref="tableRef"
          adaptive
          :adaptiveConfig="{ offsetBottom: 108 }"
          align-whole="center"
          table-layout="auto"
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
            <el-button
              class="reset-margin"
              link
              type="primary"
              :size="size"
              :icon="useRenderIcon(EditPen)"
              @click="openDialog('修改', row)"
            >
              修改
            </el-button>
            <el-popconfirm
              :title="`是否确认删除版本 ${row.version}？`"
              @confirm="handleDelete(row)"
            >
              <template #reference>
                <el-button
                  class="reset-margin"
                  link
                  type="danger"
                  :size="size"
                  :icon="useRenderIcon(Delete)"
                >
                  删除
                </el-button>
              </template>
            </el-popconfirm>
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
