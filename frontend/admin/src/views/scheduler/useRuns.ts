import {
  onActivated,
  onDeactivated,
  onMounted,
  onUnmounted,
  reactive,
  ref,
  toRaw
} from "vue";
import type { PaginationProps } from "@pureadmin/table";
import { getSchedulerRuns } from "@bit-labs.cn/owl-admin-ui/api/scheduler";
import type { SchedulerRunRow } from "./types";

const AUTO_REFRESH_INTERVAL = 5000;

export function useRuns() {
  const form = reactive({
    jobCode: "",
    status: "" as number | string,
    triggerType: "" as number | string
  });

  const jobFilter = ref<{ jobId: string; jobCode: string } | null>(null);
  const dataList = ref<SchedulerRunRow[]>([]);
  const loading = ref(true);
  const autoRefresh = ref(false);
  const pagination = reactive<PaginationProps>({
    total: 0,
    pageSize: 20,
    currentPage: 1,
    background: true
  });

  let timer: ReturnType<typeof setInterval> | null = null;

  function buildParams() {
    const raw = toRaw(form);
    const params: Record<string, unknown> = {
      page: pagination.currentPage,
      pageSize: pagination.pageSize
    };
    if (jobFilter.value?.jobId) params.jobId = jobFilter.value.jobId;
    if (jobFilter.value?.jobCode) params.jobCode = jobFilter.value.jobCode;
    if (raw.jobCode) params.jobCode = raw.jobCode.trim();
    if (raw.status !== "" && raw.status != null) params.status = String(raw.status);
    if (raw.triggerType !== "" && raw.triggerType != null) {
      params.triggerType = Number(raw.triggerType);
    }
    return params;
  }

  async function onSearch(silent = false) {
    if (!silent) loading.value = true;
    try {
      const res = await getSchedulerRuns(buildParams());
      dataList.value = (res?.data?.list ?? []) as SchedulerRunRow[];
      pagination.total = res?.total ?? 0;
      pagination.pageSize = res?.pageSize ?? pagination.pageSize;
      pagination.currentPage = res?.currentPage ?? pagination.currentPage;
    } finally {
      loading.value = false;
    }
  }

  /** 从任务列表跳到执行记录时，按任务过滤。 */
  function filterByJob(jobId: string, jobCode: string) {
    jobFilter.value = { jobId, jobCode };
    form.jobCode = jobCode;
    pagination.currentPage = 1;
  }

  function clearJobFilter() {
    jobFilter.value = null;
  }

  function resetForm(formEl) {
    if (!formEl) return;
    formEl.resetFields();
    form.jobCode = "";
    form.status = "";
    form.triggerType = "";
    jobFilter.value = null;
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

  function stopTimer() {
    if (timer) {
      clearInterval(timer);
      timer = null;
    }
  }

  function toggleAutoRefresh(value: boolean) {
    autoRefresh.value = value;
    stopTimer();
    if (value) timer = setInterval(() => onSearch(true), AUTO_REFRESH_INTERVAL);
  }

  onMounted(() => onSearch());
  onActivated(() => {
    if (autoRefresh.value && !timer) {
      timer = setInterval(() => onSearch(true), AUTO_REFRESH_INTERVAL);
    }
  });
  // 页面被 keep-alive 缓存时停掉定时器，避免在后台持续请求
  onDeactivated(stopTimer);
  onUnmounted(stopTimer);

  return {
    form,
    jobFilter,
    loading,
    dataList,
    autoRefresh,
    pagination,
    onSearch,
    filterByJob,
    clearJobFilter,
    resetForm,
    handleSizeChange,
    handleCurrentChange,
    toggleAutoRefresh
  };
}
