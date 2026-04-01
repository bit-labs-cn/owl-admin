import { reactive, ref, onMounted } from "vue";
import type { PaginationProps } from "@pureadmin/table";
import { isAllEmpty } from "@pureadmin/utils";
import { getLoginLogsList } from "@bit-labs.cn/owl-admin-ui/api/system";

export function useLoginLogList() {
  const form = reactive({
    userName: "",
    ip: "",
    userType: "",
    loginTime: [] as string[]
  });

  const dataList = ref([]);
  const loading = ref(true);
  const pagination = reactive<PaginationProps>({
    total: 0,
    pageSize: 10,
    currentPage: 1,
    background: true
  });

  function resetForm(formEl) {
    if (!formEl) return;
    formEl.resetFields();
    form.userName = "";
    form.ip = "";
    form.userType = "";
    form.loginTime = [];
    onSearch();
  }

  async function onSearch() {
    loading.value = true;
    const payload: Record<string, unknown> = {
      page: pagination.currentPage,
      pageSize: pagination.pageSize
    };
    if (!isAllEmpty(form.userName)) payload.userName = form.userName;
    if (!isAllEmpty(form.ip)) payload.ip = form.ip;
    if (!isAllEmpty(form.userType)) payload.userType = form.userType;
    if (form.loginTime && form.loginTime.length === 2) {
      const start = Math.floor(new Date(form.loginTime[0]).getTime() / 1000);
      const end = Math.floor(new Date(form.loginTime[1]).getTime() / 1000);
      payload.loginTime = `${start},${end}`;
    }

    const { data, pageSize, currentPage, total } =
      await getLoginLogsList(payload);

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
