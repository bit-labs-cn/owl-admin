<script setup lang="ts">
import { onMounted, reactive, ref } from "vue";
import type { PaginationProps } from "@pureadmin/table";
import { roleAPI } from "@bit-labs.cn/owl-admin-ui/api/role";

const props = defineProps<{
  roleId: string;
  roleName?: string;
}>();

const loading = ref(false);
const dataList = ref<any[]>([]);
const pagination = reactive<PaginationProps>({
  total: 0,
  pageSize: 10,
  currentPage: 1,
  background: true
});

const columns = [
  { label: "用户编号", prop: "id",width: 180},
  { label: "账号", prop: "username", minWidth: 100 },
  { label: "用户昵称", prop: "nickname", minWidth: 120 },
  { label: "手机号码", prop: "phone", minWidth: 120 },
  {
    label: "角色",
    prop: "roles",
    minWidth: 180,
    formatter: ({ roles }) =>
      Array.isArray(roles) && roles.length
        ? roles.map((r: any) => r.name).join("，")
        : "未分配角色"
  }
];

async function fetchUserList() {
  if (!props.roleId) return;
  loading.value = true;
  const res = await roleAPI.getUsersByRoleId(props.roleId, {
    page: pagination.currentPage,
    pageSize: pagination.pageSize
  });
  // 后端返回 { data: { list }, total, currentPage, pageSize }，total 在顶层
  dataList.value = res?.data?.list ?? [];
  pagination.total = res?.total ?? 0;
  pagination.pageSize = res?.pageSize ?? pagination.pageSize;
  pagination.currentPage = res?.currentPage ?? pagination.currentPage;
  loading.value = false;
}

function handleSizeChange(pageSize: number) {
  pagination.pageSize = pageSize;
  pagination.currentPage = 1;
  fetchUserList();
}

function handleCurrentChange(page: number) {
  pagination.currentPage = page;
  fetchUserList();
}

onMounted(() => {
  fetchUserList();
});
</script>

<template>
  <div class="role-user-list-dialog">
    <div class="table-wrap">
      <el-table :data="dataList" v-loading="loading" border>
        <el-table-column
          v-for="column in columns"
          :key="column.prop"
          v-bind="column"
        />
      </el-table>
    </div>
    <div class="flex justify-end mt-3 flex-shrink-0">
      <el-pagination
        background
        layout="total, prev, pager, next, sizes"
        :total="pagination.total"
        :current-page="pagination.currentPage"
        :page-size="pagination.pageSize"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </div>
  </div>
</template>

<style scoped>
.role-user-list-dialog {
  padding: 12px 4px;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.table-wrap {
  max-height: min(60vh, 480px);
  overflow: auto;
  flex: 1 1 auto;
  min-height: 0;
}

.table-wrap :deep(.el-table) {
  --el-table-border-color: var(--el-border-color-lighter);
}
</style>

