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
import { getInboxList } from "@bit-labs.cn/owl-admin-ui/api/event";
import type { EventInboxRow } from "../types";

const AUTO_REFRESH_INTERVAL = 5000;

export function useInboxList(initialStatus?: string | number) {
  const form = reactive({
    eventId: "",
    consumerGroup: "",
    // 工作台 KPI 会把状态放在 query 里传进来，转成数字才能和下拉选项对上
    status: (initialStatus === "" || initialStatus == null
      ? ""
      : Number(initialStatus)) as number | string,
    processedAt: [] as string[]
  });

  const dataList = ref<EventInboxRow[]>([]);
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
    if (raw.eventId) params.eventId = raw.eventId.trim();
    if (raw.consumerGroup) params.consumerGroup = raw.consumerGroup.trim();
    if (raw.status !== "" && raw.status != null) params.status = Number(raw.status);
    if (raw.processedAt && raw.processedAt.length === 2) {
      const start = Math.floor(new Date(raw.processedAt[0]).getTime() / 1000);
      const end = Math.floor(new Date(raw.processedAt[1]).getTime() / 1000);
      params.processedAt = `${start},${end}`;
    }
    return params;
  }

  async function onSearch(silent = false) {
    if (!silent) loading.value = true;
    try {
      const res = await getInboxList(buildParams());
      dataList.value = (res?.data?.list ?? []) as EventInboxRow[];
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
    form.eventId = "";
    form.consumerGroup = "";
    form.status = "";
    form.processedAt = [];
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
    loading,
    dataList,
    pagination,
    autoRefresh,
    onSearch,
    resetForm,
    handleSizeChange,
    handleCurrentChange,
    toggleAutoRefresh
  };
}
