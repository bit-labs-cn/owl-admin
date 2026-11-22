import { onActivated, onMounted, reactive, ref, toRaw } from "vue";
import type { PaginationProps } from "@pureadmin/table";
import { getSchedulerJobs } from "@bit-labs.cn/owl-admin-ui/api/scheduler";
import type { SchedulerJobRow } from "./types";

export function useJobs() {
  const form = reactive({
    name: "",
    ownerApp: "",
    source: "" as number | string,
    status: "" as number | string
  });

  const dataList = ref<SchedulerJobRow[]>([]);
  const loading = ref(true);
  const switchLoadMap = ref<Record<number, { loading: boolean }>>({});
  const pagination = reactive<PaginationProps>({
    total: 0,
    pageSize: 10,
    currentPage: 1,
    background: true
  });

  function buildParams() {
    const raw = toRaw(form);
    const params: Record<string, unknown> = {
      page: pagination.currentPage,
      pageSize: pagination.pageSize
    };
    if (raw.name) params.name = raw.name.trim();
    if (raw.ownerApp) params.ownerApp = raw.ownerApp.trim();
    if (raw.source !== "" && raw.source != null) params.source = Number(raw.source);
    if (raw.status !== "" && raw.status != null) params.status = String(raw.status);
    return params;
  }

  async function onSearch(silent = false) {
    if (!silent) loading.value = true;
    try {
      const res = await getSchedulerJobs(buildParams());
      dataList.value = (res?.data?.list ?? []) as SchedulerJobRow[];
      pagination.total = res?.total ?? 0;
      pagination.pageSize = res?.pageSize ?? pagination.pageSize;
      pagination.currentPage = res?.currentPage ?? pagination.currentPage;
    } finally {
      loading.value = false;
    }
  }

  function resetForm(formEl) {
    if (!formEl) return;
    formEl.resetFields();
    form.name = "";
    form.ownerApp = "";
    form.source = "";
    form.status = "";
    pagination.currentPage = 1;
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

  onMounted(() => onSearch());
  onActivated(() => onSearch(true));

  return {
    form,
    loading,
    dataList,
    switchLoadMap,
    pagination,
    onSearch,
    resetForm,
    handleSizeChange,
    handleCurrentChange
  };
}
