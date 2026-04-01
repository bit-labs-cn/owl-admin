<script setup lang="ts">
import { ref } from "vue";
import { useOperationLogList } from "./useOperationLogList";
import { columns } from "./columns";
import { PureTableBar } from "@bit-labs.cn/owl-ui/components/RePureTableBar";
import { useRenderIcon } from "@bit-labs.cn/owl-ui/components/ReIcon/src/hooks";

import Refresh from "@iconify-icons/ep/refresh";

defineOptions({
  name: "SystemOperationLog"
});

const formRef = ref();
const tableRef = ref();
const {
  form,
  loading,
  dataList,
  pagination,
  onSearch,
  resetForm,
  handleSizeChange,
  handleCurrentChange
} = useOperationLogList();
</script>

<template>
  <div class="main">
    <el-form
      ref="formRef"
      :inline="true"
      :model="form"
      class="search-form bg-bg_color w-[99/100] pl-8 pt-[12px] overflow-auto"
    >
      <el-form-item label="操作人：" prop="userName">
        <el-input
          v-model="form.userName"
          placeholder="请输入操作人"
          clearable
          class="!w-[160px]"
        />
      </el-form-item>
      <el-form-item label="请求路径：" prop="path">
        <el-input
          v-model="form.path"
          placeholder="请输入请求路径"
          clearable
          class="!w-[180px]"
        />
      </el-form-item>
      <el-form-item label="请求方法：" prop="method">
        <el-select
          v-model="form.method"
          placeholder="请选择"
          clearable
          class="!w-[100px]"
        >
          <el-option label="GET" value="GET" />
          <el-option label="POST" value="POST" />
          <el-option label="PUT" value="PUT" />
          <el-option label="DELETE" value="DELETE" />
        </el-select>
      </el-form-item>
      <el-form-item label="状态码：" prop="status">
        <el-select
          v-model="form.status"
          placeholder="请选择"
          clearable
          class="!w-[100px]"
        >
          <el-option label="200" :value="200" />
          <el-option label="400" :value="400" />
          <el-option label="500" :value="500" />
        </el-select>
      </el-form-item>
      <el-form-item label="操作时间：" prop="createdAt">
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
          @click="onSearch"
        >
          搜索
        </el-button>
        <el-button :icon="useRenderIcon(Refresh)" @click="resetForm(formRef)">
          重置
        </el-button>
      </el-form-item>
    </el-form>

    <PureTableBar title="操作日志" :columns="columns" @refresh="onSearch">
      <template v-slot="{ size, dynamicColumns }">
        <pure-table
          ref="tableRef"
          adaptive
          :adaptiveConfig="{ offsetBottom: 108 }"
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
        />
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
