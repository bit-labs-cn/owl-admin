<script setup lang="ts">
import { ref, h } from "vue";
import { useDeptList } from "./useDeptList";
import { createColumns } from "./columns";
import { usePublicHooks } from "../hooks";
import DeptForm from "./DeptForm.vue";
import type { DeptFormData } from "./types";
import { deptAPI } from "@bit-labs.cn/owl-admin-ui/api/dept";
import { addDialog } from "@bit-labs.cn/owl-ui/components/ReDialog/index";

import { cloneDeep, deviceDetection } from "@pureadmin/utils";
import { PureTableBar } from "@bit-labs.cn/owl-ui/components/RePureTableBar";
import { useRenderIcon } from "@bit-labs.cn/owl-ui/components/ReIcon/src/hooks";

import Delete from "@iconify-icons/ep/delete";
import EditPen from "@iconify-icons/ep/edit-pen";
import Refresh from "@iconify-icons/ep/refresh";
import AddFill from "@iconify-icons/ri/add-circle-line";

defineOptions({
  name: "SystemDept"
});

const formRef = ref();
const tableRef = ref();
const deptFormRef = ref();

const { tagStyle } = usePublicHooks();
const columns = createColumns(tagStyle);

const {
  form,
  loading,
  dataList,
  onSearch,
  resetForm,
  formatHigherDeptOptions,
  handleSelectionChange
} = useDeptList();

function onFullscreen() {
  tableRef.value.setAdaptive();
}

function openDialog(title = "新增", row?: DeptFormData) {
  const formInline = {
    id: row?.id ?? "0",
    higherDeptOptions: formatHigherDeptOptions(cloneDeep(dataList.value)),
    parentId: row?.parentId ?? "0",
    name: row?.name ?? "",
    principal: row?.principal ?? "",
    phone: row?.phone ?? "",
    email: row?.email ?? "",
    sort: row?.sort ?? 0,
    status: row?.status ?? 1,
    description: row?.description ?? ""
  };
  addDialog({
    title: `${title}部门`,
    props: { formInline },
    width: "40%",
    draggable: true,
    fullscreen: deviceDetection(),
    fullscreenIcon: true,
    closeOnClickModal: false,
    contentRenderer: () => h(DeptForm, { ref: deptFormRef, formInline }),
    beforeSure: (done, { options }) => {
      const FormRef = deptFormRef.value.getRef();
      const curData = options.props.formInline as DeptFormData;

      const chores = () => {
        done();
        onSearch();
      };

      FormRef.validate(valid => {
        if (valid) {
          if (title === "新增") {
            deptAPI.createDept(curData).then(() => chores());
          } else {
            deptAPI.updateDept(curData).then(() => chores());
          }
        }
      });
    }
  });
}

function handleDelete(row) {
  deptAPI.deleteDept(row.id).then(() => {
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
      <el-form-item label="部门名称：" prop="name">
        <el-input
          v-model="form.name"
          placeholder="请输入部门名称"
          clearable
          class="!w-[180px]"
        />
      </el-form-item>
      <el-form-item label="状态：" prop="status">
        <el-select
          v-model="form.status"
          placeholder="请选择状态"
          clearable
          class="!w-[180px]"
        >
          <el-option label="启用" :value="1" />
          <el-option label="停用" :value="0" />
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

    <PureTableBar
      title="部门管理"
      :columns="columns"
      :tableRef="tableRef?.getTableRef()"
      @refresh="onSearch"
      @fullscreen="onFullscreen"
    >
      <template #buttons>
        <el-button
          type="primary"
          :icon="useRenderIcon(AddFill)"
          @click="openDialog()"
        >
          新增部门
        </el-button>
      </template>
      <template v-slot="{ size, dynamicColumns }">
        <pure-table
          ref="tableRef"
          adaptive
          :adaptiveConfig="{ offsetBottom: 45 }"
          align-whole="center"
          row-key="id"
          showOverflowTooltip
          table-layout="auto"
          default-expand-all
          :size="size"
          :loading="loading"
          :data="dataList"
          :columns="dynamicColumns"
          :header-cell-style="{
            background: 'var(--el-fill-color-light)',
            color: 'var(--el-text-color-primary)'
          }"
          @selection-change="handleSelectionChange"
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
            <el-button
              class="reset-margin"
              link
              type="success"
              :size="size"
              :icon="useRenderIcon(AddFill)"
              @click="openDialog('新增', { parentId: row.id } as any)"
            >
              新增
            </el-button>
            <el-popconfirm
              :title="`是否确认删除部门名称为${row.name}的这条数据`"
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

.main-content {
  margin: 24px 24px 0 !important;
}

.search-form {
  :deep(.el-form-item) {
    margin-bottom: 12px;
  }
}
</style>
