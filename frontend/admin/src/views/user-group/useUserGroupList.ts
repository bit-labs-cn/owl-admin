import { reactive, ref, toRaw, computed, onMounted } from "vue";
import type { PaginationProps } from "@pureadmin/table";
import { ElMessageBox } from "element-plus";
import { userGroupAPI } from "@bit-labs.cn/owl-admin-ui/api/user-group";

export function useUserGroupList() {
  const form = reactive({ name: "", code: "", status: "" });
  const dataList = ref([]);
  const loading = ref(true);
  const switchLoadMap = ref({});
  const pagination = reactive<PaginationProps>({
    total: 0,
    pageSize: 10,
    currentPage: 1,
    background: true
  });

  const buttonClass = computed(() => [
    "!h-[20px]",
    "reset-margin",
    "!text-gray-500",
    "dark:!text-white",
    "dark:hover:!text-primary"
  ]);

  function onBeforeStatusChange({ row, index }): Promise<boolean> {
    const actionText = row.status === 1 ? "停用" : "启用";
    const newStatus = row.status === 1 ? 2 : 1;
    return ElMessageBox.confirm(
      `确认要<strong>${actionText}</strong><strong style='color:var(--el-color-primary)'>${row.name}</strong>吗?`,
      "系统提示",
      {
        confirmButtonText: "确定",
        cancelButtonText: "取消",
        type: "warning",
        dangerouslyUseHTMLString: true,
        draggable: true
      }
    ).then(() => {
      switchLoadMap.value[index] = Object.assign(
        {},
        switchLoadMap.value[index],
        { loading: true }
      );
      userGroupAPI
        .changeStatus({ ...row, status: newStatus })
        .finally(() => {
          switchLoadMap.value[index] = Object.assign(
            {},
            switchLoadMap.value[index],
            { loading: false }
          );
        });
      return true;
    });
  }

  function handleDelete(row) {
    userGroupAPI.remove(row.id).then(() => {
      onSearch();
    });
  }

  function handleSizeChange(val: number) {
    pagination.pageSize = val;
    pagination.currentPage = 1;
    onSearch();
  }

  function handleCurrentChange(val: number) {
    pagination.currentPage = val;
    onSearch();
  }

  async function onSearch() {
    loading.value = true;
    const res = await userGroupAPI.getList({
      ...toRaw(form),
      page: pagination.currentPage,
      pageSize: pagination.pageSize
    });
    dataList.value = res?.data?.list ?? [];
    pagination.total = res?.total ?? 0;
    pagination.pageSize = res?.pageSize ?? pagination.pageSize;
    pagination.currentPage = res?.currentPage ?? pagination.currentPage;
    setTimeout(() => {
      loading.value = false;
    }, 500);
  }

  function resetForm(formEl) {
    if (!formEl) return;
    formEl.resetFields();
    onSearch();
  }

  onMounted(() => {
    onSearch();
  });

  return {
    form,
    loading,
    dataList,
    pagination,
    switchLoadMap,
    buttonClass,
    onSearch,
    resetForm,
    onBeforeStatusChange,
    handleDelete,
    handleSizeChange,
    handleCurrentChange
  };
}
