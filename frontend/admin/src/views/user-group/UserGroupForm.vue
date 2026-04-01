<script setup lang="ts">
import { ref, reactive, onMounted, watch, nextTick } from "vue";
import type { UserGroupFormData } from "./types";
import { userManageAPI } from "@bit-labs.cn/owl-admin-ui/api/user";

const props = withDefaults(
  defineProps<{
    formInline: UserGroupFormData;
    initialUserIDs?: string[];
  }>(),
  {
    formInline: () => ({
      id: "",
      name: "",
      code: "",
      status: 1,
      remark: ""
    }),
    initialUserIDs: () => []
  }
);

const ruleFormRef = ref();
const newFormInline = ref(props.formInline);

const rules = {
  name: [{ required: true, message: "请输入用户组名称", trigger: "blur" }],
  code: [{ required: true, message: "请输入用户组编码", trigger: "blur" }]
};

const userTableRef = ref();
const userList = ref<any[]>([]);
const userLoading = ref(false);
const selectedUserIDs = ref<Set<string>>(new Set(props.initialUserIDs));
const userSearchKeyword = ref("");
let isSyncingSelection = false;

const userPagination = reactive({
  total: 0,
  pageSize: 10,
  currentPage: 1
});

async function fetchUsers() {
  userLoading.value = true;
  const res = await userManageAPI.getUserList({
    page: userPagination.currentPage,
    pageSize: userPagination.pageSize,
    username: userSearchKeyword.value || undefined
  });
  userList.value = res?.data?.list ?? [];
  userPagination.total = res?.total ?? 0;
  userPagination.pageSize = res?.pageSize ?? userPagination.pageSize;
  userPagination.currentPage =
    res?.currentPage ?? userPagination.currentPage;
  userLoading.value = false;

  await nextTick();
  syncTableSelection();
}

function syncTableSelection() {
  if (!userTableRef.value) return;
  isSyncingSelection = true;
  const tableEl = userTableRef.value;
  userList.value.forEach(row => {
    const isSelected = selectedUserIDs.value.has(String(row.id));
    tableEl.toggleRowSelection(row, isSelected);
  });
  isSyncingSelection = false;
}

function handleSelectionChange(selection: any[]) {
  if (isSyncingSelection) return;
  const currentPageIds = new Set(userList.value.map(r => String(r.id)));
  currentPageIds.forEach(id => selectedUserIDs.value.delete(id));
  selection.forEach(row => selectedUserIDs.value.add(String(row.id)));
}

function handleUserSizeChange(pageSize: number) {
  userPagination.pageSize = pageSize;
  userPagination.currentPage = 1;
  fetchUsers();
}

function handleUserPageChange(page: number) {
  userPagination.currentPage = page;
  fetchUsers();
}

function searchUsers() {
  userPagination.currentPage = 1;
  fetchUsers();
}

function getRef() {
  return ruleFormRef.value;
}

function getSelectedUserIDs(): string[] {
  return Array.from(selectedUserIDs.value);
}

defineExpose({ getRef, getSelectedUserIDs });

onMounted(() => {
  fetchUsers();
});

watch(
  () => props.initialUserIDs,
  ids => {
    selectedUserIDs.value = new Set(ids);
    nextTick(() => syncTableSelection());
  }
);
</script>

<template>
  <div class="user-group-form">
    <el-form
      ref="ruleFormRef"
      :model="newFormInline"
      :rules="rules"
      label-width="100px"
    >
      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item label="用户组名称" prop="name">
            <el-input
              v-model="newFormInline.name"
              clearable
              placeholder="请输入用户组名称"
            />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="用户组编码" prop="code">
            <el-input
              v-model="newFormInline.code"
              clearable
              placeholder="请输入用户组编码"
            />
          </el-form-item>
        </el-col>
      </el-row>
      <el-form-item label="备注">
        <el-input
          v-model="newFormInline.remark"
          placeholder="请输入备注信息"
          type="textarea"
          :rows="2"
        />
      </el-form-item>
    </el-form>

    <div class="user-select-section">
      <div class="user-select-header">
        <span class="user-select-title">
          选择成员
          <el-tag size="small" type="info" class="ml-2">
            已选 {{ selectedUserIDs.size }} 人
          </el-tag>
        </span>
        <div class="user-search">
          <el-input
            v-model="userSearchKeyword"
            placeholder="搜索用户名"
            clearable
            size="small"
            class="!w-[200px]"
            @keyup.enter="searchUsers"
            @clear="searchUsers"
          />
          <el-button size="small" type="primary" @click="searchUsers">
            搜索
          </el-button>
        </div>
      </div>

      <el-table
        ref="userTableRef"
        :data="userList"
        v-loading="userLoading"
        border
        size="small"
        row-key="id"
        max-height="320"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="45" reserve-selection />
        <el-table-column prop="username" label="账号" min-width="100" />
        <el-table-column prop="nickname" label="昵称" min-width="100" />
        <el-table-column prop="phone" label="手机号" min-width="120" />
        <el-table-column label="状态" width="70">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
              {{ row.status === 1 ? "启用" : "停用" }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>

      <div class="flex justify-end mt-2">
        <el-pagination
          small
          background
          layout="total, prev, pager, next, sizes"
          :total="userPagination.total"
          :current-page="userPagination.currentPage"
          :page-size="userPagination.pageSize"
          :page-sizes="[10, 20, 50]"
          @size-change="handleUserSizeChange"
          @current-change="handleUserPageChange"
        />
      </div>
    </div>
  </div>
</template>

<style scoped>
.user-group-form {
  padding: 0 4px;
}

.user-select-section {
  margin-top: 8px;
  border-top: 1px solid var(--el-border-color-lighter);
  padding-top: 12px;
}

.user-select-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}

.user-select-title {
  font-size: 14px;
  font-weight: 500;
  color: var(--el-text-color-primary);
}

.user-search {
  display: flex;
  align-items: center;
  gap: 8px;
}
</style>
