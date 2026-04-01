<script setup lang="ts">
import { ref } from "vue";
import { useLoginLogList } from "./useLoginLogList";
import { columns } from "./columns";
import { PureTableBar } from "@bit-labs.cn/owl-ui/components/RePureTableBar";
import { useRenderIcon } from "@bit-labs.cn/owl-ui/components/ReIcon/src/hooks";

import Refresh from "@iconify-icons/ep/refresh";

defineOptions({
  name: "SystemLoginLog"
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
} = useLoginLogList();
</script>

<template>
  <div class="main">
    <el-form
      ref="formRef"
      :inline="true"
      :model="form"
      class="search-form bg-bg_color w-[99/100] pl-8 pt-[12px] overflow-auto"
    >
      <el-form-item label="用户名：" prop="userName">
        <el-input
          v-model="form.userName"
          placeholder="请输入用户名"
          clearable
          class="!w-[180px]"
        />
      </el-form-item>
      <el-form-item label="IP 地址：" prop="ip">
        <el-input
          v-model="form.ip"
          placeholder="请输入 IP 地址"
          clearable
          class="!w-[180px]"
        />
      </el-form-item>
      <el-form-item label="用户类型：" prop="userType">
        <el-select
          v-model="form.userType"
          placeholder="请选择"
          clearable
          class="!w-[120px]"
        >
          <el-option label="普通用户" value="user" />
          <el-option label="超管" value="super_admin" />
        </el-select>
      </el-form-item>
      <el-form-item label="登录时间：" prop="loginTime">
        <el-date-picker
          v-model="form.loginTime"
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

    <PureTableBar title="登录日志" :columns="columns" @refresh="onSearch">
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
