import { onActivated, onMounted, ref } from "vue";
import { getSchedulerStats } from "@bit-labs.cn/owl-admin-ui/api/scheduler";
import type { SchedulerStats } from "./types";

/** 汇总数据只在页面激活时拉一次，失败不打断列表使用。 */
export function useSchedulerStats() {
  const stats = ref<SchedulerStats | null>(null);
  const loading = ref(false);

  async function load() {
    loading.value = true;
    try {
      const res = await getSchedulerStats();
      stats.value = (res?.data ?? null) as unknown as SchedulerStats | null;
    } catch {
      stats.value = null;
    } finally {
      loading.value = false;
    }
  }

  onMounted(load);
  onActivated(load);

  return { stats, loading, load };
}
