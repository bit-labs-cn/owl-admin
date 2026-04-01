<script setup lang="ts">
import { ref, h } from "vue";
import { useMenuList } from "./useMenuList";
import { columns } from "./columns";
import MenuForm from "./MenuForm.vue";
import type { MenuFormData } from "./types";
import { addDialog } from "@bit-labs.cn/owl-ui/components/ReDialog";
import { message } from "@bit-labs.cn/owl-ui/utils/message";
import { transformI18n } from "@bit-labs.cn/owl-ui/plugins/i18n";
import { cloneDeep, deviceDetection } from "@pureadmin/utils";
import { PureTableBar } from "@bit-labs.cn/owl-ui/components/RePureTableBar";
import { useRenderIcon } from "@bit-labs.cn/owl-ui/components/ReIcon/src/hooks";

import Delete from "@iconify-icons/ep/delete";
import EditPen from "@iconify-icons/ep/edit-pen";
import Refresh from "@iconify-icons/ep/refresh";
import AddFill from "@iconify-icons/ri/add-circle-line";

defineOptions({
  name: "SystemMenu"
});

const formRef = ref();
const tableRef = ref();
const menuFormRef = ref();

const {
  form,
  loading,
  dataList,
  onSearch,
  resetForm,
  formatHigherMenuOptions,
  handleSelectionChange
} = useMenuList();

function onFullscreen() {
  tableRef.value.setAdaptive();
}

function openDialog(title = "新增", row?: MenuFormData) {
  addDialog({
    title: `${title}菜单`,
    props: {
      formInline: {
        menuType: row?.menuType ?? 0,
        higherMenuOptions: formatHigherMenuOptions(cloneDeep(dataList.value)),
        parentId: row?.parentId ?? 0,
        title: row?.title ?? "",
        name: row?.name ?? "",
        path: row?.path ?? "",
        component: row?.component ?? "",
        rank: row?.rank ?? 99,
        redirect: row?.redirect ?? "",
        icon: row?.icon ?? "",
        extraIcon: row?.extraIcon ?? "",
        enterTransition: row?.enterTransition ?? "",
        leaveTransition: row?.leaveTransition ?? "",
        activePath: row?.activePath ?? "",
        auths: row?.auths ?? "",
        frameSrc: row?.frameSrc ?? "",
        frameLoading: row?.frameLoading ?? true,
        keepAlive: row?.keepAlive ?? false,
        hiddenTag: row?.hiddenTag ?? false,
        fixedTag: row?.fixedTag ?? false,
        showLink: row?.showLink ?? true,
        showParent: row?.showParent ?? false
      }
    },
    width: "45%",
    draggable: true,
    fullscreen: deviceDetection(),
    fullscreenIcon: true,
    closeOnClickModal: false,
    contentRenderer: () =>
      h(MenuForm, { ref: menuFormRef, formInline: null }),
    beforeSure: (done, { options }) => {
      const FormRef = menuFormRef.value.getRef();
      const curData = options.props.formInline as MenuFormData;
      function chores() {
        message(
          `您${title}了菜单名称为${transformI18n(curData.title)}的这条数据`,
          { type: "success" }
        );
        done();
        onSearch();
      }
      FormRef.validate(valid => {
        if (valid) {
          if (title === "新增") {
            chores();
          } else {
            chores();
          }
        }
      });
    }
  });
}

function handleDelete(row) {
  message(`您删除了菜单名称为${transformI18n(row.title)}的这条数据`, {
    type: "success"
  });
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
      <el-form-item label="菜单名称：" prop="title">
        <el-input
          v-model="form.title"
          placeholder="请输入菜单名称"
          clearable
          class="!w-[180px]"
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

    <PureTableBar
      title="只显示菜单，不做管理，后台统一维护"
      :columns="columns"
      :isExpandAll="false"
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
          新增菜单
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
          :loading="loading"
          :size="size"
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
              v-show="row.menuType !== 3"
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
              :title="`是否确认删除菜单名称为${transformI18n(row.title)}的这条数据${row?.children?.length > 0 ? '。注意下级菜单也会一并删除，请谨慎操作' : ''}`"
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
