<script setup lang="ts">
import { useRoleList } from "./useRoleList";
import { createColumns } from "./columns";
import RoleForm from "./RoleForm.vue";
import RoleUserListDialog from "./RoleUserListDialog.vue";
import type { RoleFormData } from "./types";
import { roleAPI } from "@bit-labs.cn/owl-admin-ui/api/role";
import { addDialog } from "@bit-labs.cn/owl-ui/components/ReDialog";
import {
  ref,
  h,
  computed,
  nextTick,
  onBeforeUnmount,
  watch
} from "vue";
import { PureTableBar } from "@bit-labs.cn/owl-ui/components/RePureTableBar";
import { useRenderIcon } from "@bit-labs.cn/owl-ui/components/ReIcon/src/hooks";
import { deviceDetection } from "@pureadmin/utils";
import { usePublicHooks } from "../hooks";

import Delete from "@iconify-icons/ep/delete";
import EditPen from "@iconify-icons/ep/edit-pen";
import Refresh from "@iconify-icons/ep/refresh";
import AddFill from "@iconify-icons/ri/add-circle-line";
import Menu from "@iconify-icons/ep/menu";
import Close from "@iconify-icons/ep/close";
import Check from "@iconify-icons/ep/check";
import User from "@iconify-icons/ri/user-3-fill";

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
const treeWrapRef = ref<HTMLElement | null>(null);
const treeWrapHeight = ref(320);
let treeWrapResizeObserver: ResizeObserver | null = null;
const roleFormRef = ref();
const { switchStyle } = usePublicHooks();

/** 固定引用，避免模板里每次 `useRenderIcon(...)` 新建组件导致图标重复渲染 */
const roleUserOpIcon = useRenderIcon(User);
const roleEditOpIcon = useRenderIcon(EditPen);
const roleDeleteOpIcon = useRenderIcon(Delete);
const roleMenuPermIcon = useRenderIcon(Menu);

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

const columns = computed(() => createColumns({ switchLoadMap, switchStyle, onChange, pagination }));

function tearDownTreeWrapObserver() {
  treeWrapResizeObserver?.disconnect();
  treeWrapResizeObserver = null;
}

function updateTreeWrapHeightFromEl() {
  const el = treeWrapRef.value;
  if (!el) return;
  const h = Math.floor(el.getBoundingClientRect().height);
  treeWrapHeight.value = Math.max(160, h);
}

function setupTreeWrapObserver() {
  tearDownTreeWrapObserver();
  const el = treeWrapRef.value;
  if (!el) return;
  updateTreeWrapHeightFromEl();
  treeWrapResizeObserver = new ResizeObserver(() => {
    updateTreeWrapHeightFromEl();
  });
  treeWrapResizeObserver.observe(el);
}

watch(isShow, async show => {
  if (!show) {
    tearDownTreeWrapObserver();
    return;
  }
  await nextTick();
  requestAnimationFrame(() => {
    setupTreeWrapObserver();
    requestAnimationFrame(() => updateTreeWrapHeightFromEl());
  });
});

onBeforeUnmount(() => {
  tearDownTreeWrapObserver();
});

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

</script>

<template>
  <div class="main flex min-h-0 flex-1 flex-col">
    <el-form
      ref="formRef"
      :inline="true"
      :model="form"
      class="search-form bg-bg_color w-[99/100] shrink-0 pl-8 pt-[12px] overflow-auto"
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
      :class="[
        'flex min-h-0 flex-1 items-stretch',
        deviceDetection() ? 'flex-wrap' : ''
      ]"
    >
      <PureTableBar
        :class="[
          isShow && !deviceDetection() ? '!w-[60vw]' : 'w-full',
          'min-h-0 min-w-0'
        ]"
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
            border
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
              <div class="role-op-actions">
                <el-button
                  link
                  type="info"
                  :size="size"
                  :icon="roleUserOpIcon"
                  @click="openRoleUserDialog(row)"
                >
                  用户
                </el-button>
                <el-button
                  link
                  type="primary"
                  :size="size"
                  :icon="roleEditOpIcon"
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
                      link
                      type="danger"
                      :size="size"
                      :icon="roleDeleteOpIcon"
                    >
                      删除
                    </el-button>
                  </template>
                </el-popconfirm>
                <el-button
                  link
                  type="warning"
                  :size="size"
                  :icon="roleMenuPermIcon"
                  @click="handleMenu(row)"
                >
                  权限
                </el-button>
              </div>
            </template>
          </pure-table>
        </template>
      </PureTableBar>

      <div
        v-if="isShow"
        class="!min-w-[calc(100vw-60vw-268px)] ml-2 mt-2 flex w-full min-h-0 flex-1 flex-col bg-bg_color px-2 pb-2 overflow-hidden"
      >
        <div class="flex w-full shrink-0 justify-between px-3 pb-4 pt-5">
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
          class="mb-1 shrink-0"
          clearable
          @input="onQueryChanged"
        />
        <div class="flex shrink-0 flex-wrap">
          <el-checkbox v-model="isExpandAll" label="展开/折叠" />
          <el-checkbox v-model="isSelectAll" label="全选/全不选" />
        </div>
        <div
          ref="treeWrapRef"
          class="min-h-[200px] min-w-0 flex-1 overflow-hidden"
        >
          <el-tree-v2
            ref="treeRef"
            show-checkbox
            :data="treeData"
            :props="treeProps"
            :height="treeWrapHeight"
            :filter-method="filterMethod"
          >
            <template #default="{ data }">
              <span
                class="inline-flex min-w-0 max-w-full flex-wrap items-center gap-1.5"
              >
                <span class="truncate">{{
                  transformI18n(data.meta.title)
                }}</span>
                <el-tag
                  v-if="data.rolePermMenuTypeShowTag"
                  size="small"
                  effect="plain"
                  :type="data.rolePermMenuTypeTagType"
                >
                  {{ data.menuType }}
                </el-tag>
              </span>
            </template>
          </el-tree-v2>
        </div>
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.role-op-actions {
  display: inline-flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: flex-start;
  gap: 4px;
  max-width: 100%;
  vertical-align: middle;
}

.role-op-actions :deep(.el-button),
.role-op-actions :deep(.el-button + .el-button) {
  margin-left: 0;
  margin-right: 0;
}

.role-op-actions :deep(.el-popconfirm) {
  display: inline-flex;
  vertical-align: middle;
}

/* 固定右侧列单元格默认 overflow:hidden，会把「权限」等尾部裁掉 */
:deep(.el-table__fixed-right .el-table__cell .cell) {
  overflow: visible;
}

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
