import { reactive, ref, onMounted, toRaw } from "vue";
import type { PaginationProps } from "@pureadmin/table";
import { appVersionAPI } from "@bit-labs.cn/owl-admin-ui/api/app-version";

export function useAppVersionList() {
  const form = reactive({
    version: "",
    apkType: "" as number | string,
    status: "" as number | string
  });

  const dataList = ref([]);
  const loading = ref(true);
  const switchLoadMap = ref<Record<number, { loading: boolean }>>({});
  const pagination = reactive<PaginationProps>({
    total: 0,
    pageSize: 10,
    currentPage: 1,
    background: true
  });

  function onChange({ row, index }) {
    switchLoadMap.value[index] = Object.assign({}, switchLoadMap.value[index], {
      loading: true
    });
    appVersionAPI
      .changeStatus(row.id, row.status)
      .then(() => {
        onSearch();
      })
      .finally(() => {
        switchLoadMap.value[index] = Object.assign(
          {},
          switchLoadMap.value[index],
          { loading: false }
        );
      });
  }

  async function onSearch() {
    loading.value = true;
    const raw = toRaw(form);
    const payload: Record<string, unknown> = {
      page: pagination.currentPage,
      pageSize: pagination.pageSize
    };
    if (raw.version) payload.version = raw.version;
    if (raw.apkType !== "" && raw.apkType != null) payload.apkType = raw.apkType;
    if (raw.status !== "" && raw.status != null) payload.status = raw.status;
    const res = await appVersionAPI.getList(payload);
    dataList.value = (res?.data?.list ?? []).map((row: Record<string, unknown>) => ({
      ...row,
      status: row.status == null ? 1 : row.status
    }));
    pagination.total = res?.total ?? 0;
    pagination.pageSize = res?.pageSize ?? pagination.pageSize;
    pagination.currentPage = res?.currentPage ?? pagination.currentPage;
    loading.value = false;
  }

  function resetForm(formEl) {
    if (!formEl) return;
    formEl.resetFields();
    onSearch();
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
    switchLoadMap,
    onSearch,
    resetForm,
    onChange,
    handleSizeChange,
    handleCurrentChange
  };
}
