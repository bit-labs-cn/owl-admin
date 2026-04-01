<script setup lang="ts">
import { ref, h } from "vue";
import { useUserGroupList } from "./useUserGroupList";
import { createColumns } from "./columns";
import UserGroupForm from "./UserGroupForm.vue";
import UserGroupUserListDialog from "./UserGroupUserListDialog.vue";
import type { UserGroupFormData } from "./types";
import { userGroupAPI } from "@bit-labs.cn/owl-admin-ui/api/user-group";
import { addDialog } from "@bit-labs.cn/owl-ui/components/ReDialog";
import { PureTableBar } from "@bit-labs.cn/owl-ui/components/RePureTableBar";
import { useRenderIcon } from "@bit-labs.cn/owl-ui/components/ReIcon/src/hooks";
import { deviceDetection } from "@pureadmin/utils";
import { usePublicHooks } from "../hooks";

import Delete from "@iconify-icons/ep/delete";
import EditPen from "@iconify-icons/ep/edit-pen";
import Refresh from "@iconify-icons/ep/refresh";
import AddFill from "@iconify-icons/ri/add-circle-line";
import User from "@iconify-icons/ri/user-3-fill";

defineOptions({
  name: "SystemUserGroup"
});

const formRef = ref();
const userGroupFormRef = ref();

const { switchStyle } = usePublicHooks();

const {
  form,
  loading,
  dataList,
  pagination,
  switchLoadMap,
  onSearch,
  resetForm,
  onBeforeStatusChange,
  handleDelete,
  handleSizeChange,
  handleCurrentChange
} = useUserGroupList();

const columns = createColumns({ switchLoadMap, switchStyle, onBeforeStatusChange });

function openUserListDialog(row) {
  addDialog({
    title: `【${row.name}】用户组的成员`,
    width: "60%",
    draggable: true,
    alignCenter: true,
    fullscreen: false,
    fullscreenIcon: true,
    hideFooter: true,
    closeOnClickModal: true,
    contentRenderer: () =>
      h(UserGroupUserListDialog, {
        groupId: row.id,
        groupName: row.name
      })
  });
}

async function openDialog(title = "新增", row?: UserGroupFormData) {
  let initialUserIDs: string[] = [];
  if (title !== "新增" && row?.id) {
    const res = await userGroupAPI.getUserIdsByGroupId(String(row.id));
    initialUserIDs = res?.data ?? [];
  }

  addDialog({
    title: `${title}用户组`,
    props: {
      formInline: {
        id: row?.id ?? "",
        name: row?.name ?? "",
        code: row?.code ?? "",
        status: row?.status ?? 1,
        remark: row?.remark ?? ""
      },
      initialUserIDs
    },
    width: "680px",
    draggable: true,
    fullscreen: deviceDetection(),
    fullscreenIcon: true,
    closeOnClickModal: false,
    contentRenderer: () =>
      h(UserGroupForm, {
        ref: userGroupFormRef,
        formInline: null,
        initialUserIDs: null
      }),
    beforeSure: (done, { options }) => {
      const FormRef = userGroupFormRef.value.getRef();
      const curData = options.props.formInline as UserGroupFormData;
      const selectedUserIDs =
        userGroupFormRef.value.getSelectedUserIDs() as string[];

      function chores() {
        done();
        onSearch();
      }

      FormRef.validate(valid => {
        if (valid) {
          const payload = { ...curData, userIDs: selectedUserIDs };
          if (title === "新增") {
            userGroupAPI.create(payload).then(() => chores());
          } else {
            userGroupAPI.update(payload).then(() => chores());
          }
        }
      });
    }
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
      <el-form-item label="用户组名称：" prop="name">
        <el-input
          v-model="form.name"
          placeholder="请输入用户组名称"
          clearable
          class="!w-[180px]"
        />
      </el-form-item>
      <el-form-item label="用户组编码：" prop="code">
        <el-input
          v-model="form.code"
          placeholder="请输入用户组编码"
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
          <el-option label="已停用" value="2" />
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

    <PureTableBar title="用户组管理" :columns="columns" @refresh="onSearch">
      <template #buttons>
        <el-button
          type="primary"
          :icon="useRenderIcon(AddFill)"
          @click="openDialog()"
        >
          新增用户组
        </el-button>
      </template>
      <template v-slot="{ size, dynamicColumns }">
        <pure-table
          align-whole="center"
          showOverflowTooltip
          table-layout="auto"
          :loading="loading"
          :size="size"
          adaptive
          :adaptiveConfig="{ offsetBottom: 108 }"
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
              type="info"
              :size="size"
              :icon="useRenderIcon(User)"
              @click="openUserListDialog(row)"
            >
              成员
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
              :title="`是否确认删除用户组【${row.name}】`"
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
