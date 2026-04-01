import { reactive, ref, onMounted } from "vue";
import type { PaginationProps } from "@pureadmin/table";
import { isAllEmpty } from "@pureadmin/utils";
import { getOperationLogsList } from "@bit-labs.cn/owl-admin-ui/api/system";

export function useOperationLogList() {
  const form = reactive({
    userName: "",
    path: "",
    method: "",
    status: "" as string | number,
    createdAt: [] as string[]
  });

  const dataList = ref([]);
  const loading = ref(true);
  const pagination = reactive<PaginationProps>({
    total: 0,
    pageSize: 20,
    currentPage: 1,
    background: true
  });

  function resetForm(formEl) {
    if (!formEl) return;
    formEl.resetFields();
    form.userName = "";
    form.path = "";
    form.method = "";
    form.status = "";
    form.createdAt = [];
    onSearch();
  }

  async function onSearch() {
    loading.value = true;
    const payload: Record<string, unknown> = {
      page: pagination.currentPage,
      pageSize: pagination.pageSize
    };
    if (!isAllEmpty(form.userName)) payload.userName = form.userName;
    if (!isAllEmpty(form.path)) payload.path = form.path;
    if (!isAllEmpty(form.method)) payload.method = form.method;
    if (!isAllEmpty(form.status)) payload.status = Number(form.status);
    if (form.createdAt && form.createdAt.length === 2) {
      const start = Math.floor(new Date(form.createdAt[0]).getTime() / 1000);
      const end = Math.floor(new Date(form.createdAt[1]).getTime() / 1000);
      payload.createdAt = `${start},${end}`;
    }

    const { data, pageSize, currentPage, total } =
      await getOperationLogsList(payload);
    dataList.value = data.list ?? [];
    pagination.total = total ?? 0;
    pagination.pageSize = pageSize ?? pagination.pageSize;
    pagination.currentPage = currentPage ?? pagination.currentPage;
    setTimeout(() => (loading.value = false), 300);
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

  onMounted(() => {
    onSearch();
  });

  return {
    form,
    loading,
    dataList,
    pagination,
    onSearch,
    resetForm,
    handleSizeChange,
    handleCurrentChange
  };
}
