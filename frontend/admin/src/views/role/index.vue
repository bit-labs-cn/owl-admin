<script setup lang="ts">
import { useRoleList } from "./useRoleList";
import { createColumns } from "./columns";
import RoleForm from "./RoleForm.vue";
import RoleUserListDialog from "./RoleUserListDialog.vue";
import type { RoleFormData } from "./types";
import { roleAPI } from "@bit-labs.cn/owl-admin-ui/api/role";
import { addDialog } from "@bit-labs.cn/owl-ui/components/ReDialog";

import { ref, h, computed, nextTick, onMounted } from "vue";
import { PureTableBar } from "@bit-labs.cn/owl-ui/components/RePureTableBar";
import { useRenderIcon } from "@bit-labs.cn/owl-ui/components/ReIcon/src/hooks";
import {
  delay,
  subBefore,
  deviceDetection,
  useResizeObserver
} from "@pureadmin/utils";
import { usePublicHooks } from "../hooks";

import Database from "@iconify-icons/ri/database-2-line";
import Delete from "@iconify-icons/ep/delete";
import EditPen from "@iconify-icons/ep/edit-pen";
import Refresh from "@iconify-icons/ep/refresh";
import Menu from "@iconify-icons/ep/menu";
import AddFill from "@iconify-icons/ri/add-circle-line";
import Close from "@iconify-icons/ep/close";
import Check from "@iconify-icons/ep/check";
import User from "@iconify-icons/ri/user-3-fill";
import More from "@iconify-icons/ep/more-filled";

defineOptions({
  name: "SystemRole"
});

const iconClass = computed(() => [
  "w-[22px]",
  "h-[22px]",
  "flex",
  "justify-center",
  "items-center",
  "outline-none",
  "rounded-[4px]",
  "cursor-pointer",
  "transition-colors",
  "hover:bg-[#0000000f]",
  "dark:hover:bg-[#ffffff1f]",
  "dark:hover:text-[#ffffffd9]"
]);

const treeRef = ref();
const formRef = ref();
const tableRef = ref();
const contentRef = ref();
const treeHeight = ref();
const roleFormRef = ref();

const { switchStyle } = usePublicHooks();

const {
  form,
  isShow,
  curRow,
  loading,
  rowStyle,
  dataList,
  treeData,
  treeProps,
  pagination,
  isExpandAll,
  isSelectAll,
  treeSearchValue,
  switchLoadMap,
  buttonClass,
  onSearch,
  resetForm,
  onChange,
  handleMenu,
  handleSave,
  handleDelete,
  filterMethod,
  transformI18n,
  onQueryChanged,
  handleSizeChange,
  handleCurrentChange,
  handleSelectionChange
} = useRoleList(treeRef);

const columns = createColumns({ switchLoadMap, switchStyle, onChange });

function openRoleUserDialog(row) {
  addDialog({
    title: `拥有【${row.name}】角色的用户`,
    width: "60%",
    draggable: true,
    alignCenter: true,
    fullscreen: false,
    fullscreenIcon: true,
    hideFooter: true,
    closeOnClickModal: true,
    contentRenderer: () =>
      h(RoleUserListDialog, {
        roleId: row.id,
        roleName: row.name
      })
  });
}

function openDialog(title = "新增", row?: RoleFormData) {
  addDialog({
    title: `${title}角色`,
    props: {
      formInline: {
        id: row?.id ?? "",
        name: row?.name ?? "",
        code: row?.code ?? "",
        remark: row?.remark ?? ""
      }
    },
    width: "40%",
    draggable: true,
    fullscreen: deviceDetection(),
    fullscreenIcon: true,
    closeOnClickModal: false,
    contentRenderer: () =>
      h(RoleForm, { ref: roleFormRef, formInline: null }),
    beforeSure: (done, { options }) => {
      const FormRef = roleFormRef.value.getRef();
      const curData = options.props.formInline as RoleFormData;
      function chores() {
        done();
        onSearch();
      }
      FormRef.validate(valid => {
        if (valid) {
          if (title === "新增") {
            roleAPI.createRole(curData).then(() => chores());
          } else {
            roleAPI.updateRole(curData).then(() => chores());
          }
        }
      });
    }
  });
}

onMounted(() => {
  useResizeObserver(contentRef, async () => {
    await nextTick();
    delay(60).then(() => {
      treeHeight.value = parseFloat(
        subBefore(tableRef.value.getTableDoms().tableWrapper.style.height, "px")
      );
    });
  });
});
</script>

<template>
  <div class="main">
    <el-form
      ref="formRef"
      :inline="true"
      :model="form"
      class="search-form bg-bg_color w-[99/100] pl-8 pt-[12px] overflow-auto"
    >
      <el-form-item label="角色名称：" prop="name">
        <el-input
          v-model="form.name"
          placeholder="请输入角色名称"
          clearable
          class="!w-[180px]"
        />
      </el-form-item>
      <el-form-item label="角色标识：" prop="code">
        <el-input
          v-model="form.code"
          placeholder="请输入角色标识"
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

    <div
      ref="contentRef"
      :class="['flex', deviceDetection() ? 'flex-wrap' : '']"
    >
      <PureTableBar
        :class="[isShow && !deviceDetection() ? '!w-[60vw]' : 'w-full']"
        style="transition: width 220ms cubic-bezier(0.4, 0, 0.2, 1)"
        title="角色管理"
        :columns="columns"
        @refresh="onSearch"
      >
        <template #buttons>
          <el-button
            type="primary"
            :icon="useRenderIcon(AddFill)"
            @click="openDialog()"
          >
            新增角色
          </el-button>
        </template>
        <template v-slot="{ size, dynamicColumns }">
          <pure-table
            ref="tableRef"
            align-whole="center"
            showOverflowTooltip
            table-layout="auto"
            :loading="loading"
            :size="size"
            adaptive
            :row-style="rowStyle"
            :adaptiveConfig="{ offsetBottom: 108 }"
            :data="dataList"
            :columns="dynamicColumns"
            :pagination="{ ...pagination, size }"
            :header-cell-style="{
              background: 'var(--el-fill-color-light)',
              color: 'var(--el-text-color-primary)'
            }"
            @selection-change="handleSelectionChange"
            @page-size-change="handleSizeChange"
            @page-current-change="handleCurrentChange"
          >
            <template #operation="{ row }">
              <el-button
                class="reset-margin"
                link
                type="info"
                :size="size"
                :icon="useRenderIcon(User)"
                @click="openRoleUserDialog(row)"
              >
                用户
              </el-button>
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
                :title="`是否确认删除角色名称为${row.name}的这条数据`"
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
              <el-button
                class="reset-margin"
                link
                type="warning"
                :size="size"
                :icon="useRenderIcon(Menu)"
                @click="handleMenu(row)"
              >
                权限
              </el-button>
              <el-dropdown>
                <el-button
                  class="ml-3 mt-[2px]"
                  link
                  type="primary"
                  :size="size"
                  :icon="useRenderIcon(More)"
                />
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item>
                      <el-button
                        :class="buttonClass"
                        link
                        type="warning"
                        :size="size"
                        :icon="useRenderIcon(Menu)"
                        @click="handleMenu"
                      >
                        菜单权限
                      </el-button>
                    </el-dropdown-item>
                    <el-dropdown-item>
                      <el-button
                        :class="buttonClass"
                        link
                        type="info"
                        :size="size"
                        :icon="useRenderIcon(Database)"
                      >
                        数据权限
                      </el-button>
                    </el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </template>
          </pure-table>
        </template>
      </PureTableBar>

      <div
        v-if="isShow"
        class="!min-w-[calc(100vw-60vw-268px)] w-full mt-2 px-2 pb-2 bg-bg_color ml-2 overflow-auto"
      >
        <div class="flex justify-between w-full px-3 pt-5 pb-4">
          <div class="flex">
            <span :class="iconClass">
              <IconifyIconOffline
                v-tippy="{ content: '关闭' }"
                class="dark:text-white"
                width="18px"
                height="18px"
                :icon="Close"
                @click="handleMenu"
              />
            </span>
            <span :class="[iconClass, 'ml-2']">
              <IconifyIconOffline
                v-tippy="{ content: '保存菜单权限' }"
                class="dark:text-white"
                width="18px"
                height="18px"
                :icon="Check"
                @click="handleSave"
              />
            </span>
          </div>
          <p class="font-bold truncate">
            菜单权限
            {{ `${curRow?.name ? `（${curRow.name}）` : ""}` }}
          </p>
        </div>
        <el-input
          v-model="treeSearchValue"
          placeholder="请输入菜单进行搜索"
          class="mb-1"
          clearable
          @input="onQueryChanged"
        />
        <div class="flex flex-wrap">
          <el-checkbox v-model="isExpandAll" label="展开/折叠" />
          <el-checkbox v-model="isSelectAll" label="全选/全不选" />
        </div>
        <el-tree-v2
          ref="treeRef"
          show-checkbox
          :data="treeData"
          :props="treeProps"
          :height="treeHeight"
          :filter-method="filterMethod"
        >
          <template #default="{ data }">
            <span>{{ transformI18n(data.meta.title) }}</span>
          </template>
        </el-tree-v2>
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
:deep(.el-dropdown-menu__item i) {
  margin: 0;
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
