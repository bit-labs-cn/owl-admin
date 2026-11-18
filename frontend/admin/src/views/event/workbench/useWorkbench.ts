import { onActivated, onDeactivated, onMounted, onUnmounted, ref } from "vue";
import { getEventWorkbench } from "@bit-labs.cn/owl-admin-ui/api/event";
import type { EventWorkbenchData } from "../../types";

const AUTO_REFRESH_INTERVAL = 10000;

export function useWorkbench() {
  const days = ref(7);
  const loading = ref(true);
  const error = ref("");
  const data = ref<EventWorkbenchData | null>(null);
  const autoRefresh = ref(false);

  let timer: ReturnType<typeof setInterval> | null = null;

  async function load(silent = false) {
    if (!silent) loading.value = true;
    try {
      const res = await getEventWorkbench({ days: days.value });
      data.value = (res?.data ?? null) as unknown as EventWorkbenchData | null;
      error.value = "";
    } catch (err) {
      error.value = err instanceof Error ? err.message : "工作台数据加载失败";
    } finally {
      loading.value = false;
    }
  }

  function stopTimer() {
    if (timer) {
      clearInterval(timer);
      timer = null;
    }
  }

  function startTimer() {
    stopTimer();
    timer = setInterval(() => load(true), AUTO_REFRESH_INTERVAL);
  }

  function toggleAutoRefresh(value: boolean) {
    autoRefresh.value = value;
    if (value) startTimer();
    else stopTimer();
  }

  onMounted(() => load());
  onActivated(() => {
    if (autoRefresh.value) startTimer();
  });
  // 页面被 keep-alive 缓存时停掉定时器，避免在后台持续请求
  onDeactivated(stopTimer);
  onUnmounted(stopTimer);

  return {
    days,
    loading,
    error,
    data,
    autoRefresh,
    load,
    toggleAutoRefresh
  };
}
