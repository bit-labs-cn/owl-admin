import { computed, onMounted, onUnmounted, ref, watch } from "vue";
import dayjs from "dayjs";
import { serverMonitorAPI } from "@bit-labs.cn/owl-admin-ui/api/server-monitor";
import type { ServerSnapshot } from "./types";

const refreshInterval = 3000;
const maxPoints = (15 * 60 * 1000) / refreshInterval;

export interface TrendPoint {
  label: string;
  cpu: number;
  cpuUser: number;
  cpuSystem: number;
  cpuIOWait: number;
  memory: number;
  netIn: number | null;
  netOut: number | null;
  diskRead: number | null;
  diskWrite: number | null;
}

interface CounterSample {
  at: number;
  recv: number;
  sent: number;
  read: number;
  write: number;
}

export function useServerMonitor() {
  const snapshot = ref<ServerSnapshot | null>(null);
  const history = ref<TrendPoint[]>([]);
  const loading = ref(false);
  const loadError = ref("");
  const autoRefresh = ref(true);
  let pending = false;
  let timer = 0;
  let previous: CounterSample | null = null;

  const latest = computed(() => history.value[history.value.length - 1] ?? null);
  const earlier = computed(() => history.value[history.value.length - 2] ?? null);
  const labels = computed(() => history.value.map(point => point.label));
  const healthScore = computed(() => scoreOf(snapshot.value));

  async function load() {
    if (pending) return;
    pending = true;
    if (!snapshot.value) loading.value = true;
    try {
      const res = await serverMonitorAPI.snapshot();
      snapshot.value = res.data ?? null;
      loadError.value = snapshot.value ? "" : "暂时无法获取运行概况";
      if (snapshot.value) pushPoint(snapshot.value);
    } catch {
      loadError.value = "暂时无法获取运行概况";
    } finally {
      pending = false;
      loading.value = false;
    }
  }

  function pushPoint(current: ServerSnapshot) {
    const now = Date.now();
    const elapsed = previous ? (now - previous.at) / 1000 : 0;
    const netReady = Boolean(previous && current.host.netOK && elapsed > 0);
    const diskReady = Boolean(previous && current.host.diskIOOK && elapsed > 0);
    const point: TrendPoint = {
      label: dayjs(current.collectedAt).format("HH:mm:ss"),
      cpu: current.host.cpuPercent,
      cpuUser: current.host.cpuUser,
      cpuSystem: current.host.cpuSystem,
      cpuIOWait: current.host.cpuIOWait,
      memory: current.host.memoryPercent,
      netIn: netReady ? bytesPerSecond(current.host.netRecvBytes, previous!.recv, elapsed) : null,
      netOut: netReady ? bytesPerSecond(current.host.netSentBytes, previous!.sent, elapsed) : null,
      diskRead: diskReady ? bytesPerSecond(current.host.diskReadBytes, previous!.read, elapsed) : null,
      diskWrite: diskReady ? bytesPerSecond(current.host.diskWriteBytes, previous!.write, elapsed) : null
    };
    previous = {
      at: now,
      recv: current.host.netRecvBytes,
      sent: current.host.netSentBytes,
      read: current.host.diskReadBytes,
      write: current.host.diskWriteBytes
    };
    const next = history.value.concat(point);
    history.value = next.length > maxPoints ? next.slice(next.length - maxPoints) : next;
  }

  function restart() {
    window.clearInterval(timer);
    timer = 0;
    if (!autoRefresh.value) return;
    timer = window.setInterval(() => {
      void load();
    }, refreshInterval);
  }

  function formatBytes(value: number) {
    return formatUnit(value, "");
  }

  function formatRate(value: number | null) {
    if (value === null || !Number.isFinite(value)) return "等待采样";
    return formatUnit(value, "/s");
  }

  function formatDuration(seconds: number) {
    const total = Math.max(0, Math.floor(seconds));
    const days = Math.floor(total / 86400);
    const hours = Math.floor((total % 86400) / 3600);
    const minutes = Math.floor((total % 3600) / 60);
    if (days > 0) return `${days} 天 ${hours} 小时`;
    if (hours > 0) return `${hours} 小时 ${minutes} 分钟`;
    if (minutes > 0) return `${minutes} 分钟`;
    return `${total} 秒`;
  }

  function percent(value: number) {
    if (!Number.isFinite(value) || value < 0) return 0;
    if (value > 100) return 100;
    return Math.round(value * 10) / 10;
  }

  function shiftText(current: number, previousValue: number | undefined) {
    if (previousValue === undefined || !Number.isFinite(previousValue)) return "";
    const diff = Math.round((current - previousValue) * 10) / 10;
    if (diff === 0) return "0.0%";
    return `${diff > 0 ? "+" : ""}${diff.toFixed(1)}%`;
  }

  function collectedText() {
    if (!snapshot.value?.collectedAt) return "";
    const parsed = dayjs(snapshot.value.collectedAt);
    return parsed.isValid() ? parsed.format("YYYY-MM-DD HH:mm:ss") : snapshot.value.collectedAt;
  }

  onMounted(() => {
    void load();
    restart();
  });

  onUnmounted(() => {
    window.clearInterval(timer);
  });

  watch(autoRefresh, restart);

  return {
    snapshot,
    history,
    labels,
    latest,
    earlier,
    healthScore,
    loading,
    loadError,
    autoRefresh,
    load,
    formatBytes,
    formatRate,
    formatDuration,
    percent,
    shiftText,
    collectedText
  };
}

function bytesPerSecond(current: number, previous: number, seconds: number) {
  if (seconds <= 0 || current < previous) return 0;
  return (current - previous) / seconds;
}

function formatUnit(value: number, suffix: string) {
  if (!Number.isFinite(value) || value <= 0) return `0 B${suffix}`;
  const units = ["B", "KB", "MB", "GB", "TB"];
  let size = value;
  let index = 0;
  while (size >= 1024 && index < units.length - 1) {
    size /= 1024;
    index += 1;
  }
  const digits = index === 0 ? 0 : 1;
  return `${size.toFixed(digits)} ${units[index]}${suffix}`;
}

function scoreOf(snapshot: ServerSnapshot | null) {
  if (!snapshot) return 0;
  const disk = snapshot.host.disks.reduce((max, item) => Math.max(max, item.percent || 0), 0);
  const pressure = snapshot.host.cpuPercent * 0.45 + snapshot.host.memoryPercent * 0.35 + disk * 0.2;
  return Math.max(0, Math.min(100, Math.round(100 - pressure)));
}
