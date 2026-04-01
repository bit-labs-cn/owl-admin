<script setup lang="ts">
import { ref, h } from "vue";
import { usePositionList } from "./usePositionList";
import { createColumns } from "./columns";
import PositionForm from "./PositionForm.vue";
import type { PositionFormData } from "./types";
import { positionAPI } from "@bit-labs.cn/owl-admin-ui/api/position";
import { addDialog } from "@bit-labs.cn/owl-ui/components/ReDialog";

import { PureTableBar } from "@bit-labs.cn/owl-ui/components/RePureTableBar";
import { useRenderIcon } from "@bit-labs.cn/owl-ui/components/ReIcon/src/hooks";

import AddFill from "@iconify-icons/ri/add-circle-line";
import Refresh from "@iconify-icons/ep/refresh";
import EditPen from "@iconify-icons/ep/edit-pen";
import Delete from "@iconify-icons/ep/delete";

defineOptions({ name: "SystemPosition" });

const formRef = ref();
const tableRef = ref();
const positionFormRef = ref();

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
} = usePositionList();

const columns = createColumns({ switchLoadMap, onChange });

function openDialog(title = "新增", row?: PositionFormData) {
  addDialog({
    title: `${title}岗位`,
    props: {
      formInline: {
        id: row?.id ?? "",
        name: row?.name ?? "",
        remark: row?.remark ?? "",
        status: row?.status ?? 1
      }
    },
    width: "40%",
    draggable: true,
    closeOnClickModal: false,
    contentRenderer: ({ options }) =>
      h(PositionForm, { ref: positionFormRef, formInline: options.props.formInline }),
    beforeSure: (done) => {
      const FormRef = positionFormRef.value.getRef();
      const curData = positionFormRef.value.getFormData() as PositionFormData;
      function chores() {
        done();
        onSearch();
      }
      FormRef.validate(valid => {
        if (valid) {
          if (title === "新增") {
            positionAPI.createPosition(curData).then(() => chores());
          } else {
            positionAPI.updatePosition(curData).then(() => chores());
          }
        }
      });
    }
  });
}

function handleDelete(row) {
  positionAPI.deletePosition(row.id).then(() => {
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
      <el-form-item label="岗位名称：" prop="name">
        <el-input
          v-model="form.name"
          placeholder="请输入岗位名称"
          clearable
          class="!w-[200px]"
        />
      </el-form-item>
      <el-form-item label="状态：" prop="status">
        <el-select
          v-model="form.status"
          placeholder="请选择"
          clearable
          class="!w-[180px]"
        >
          <el-option label="已启用" value="1" />
          <el-option label="已停用" value="0" />
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

    <PureTableBar title="岗位管理" :columns="columns" @refresh="onSearch">
      <template #buttons>
        <el-button
          type="primary"
          :icon="useRenderIcon(AddFill)"
          @click="openDialog()"
        >
          新增岗位
        </el-button>
      </template>
      <template v-slot="{ size, dynamicColumns }">
        <pure-table
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
              :title="`是否确认删除岗位名称为${row.name}的这条数据`"
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
