<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from "vue";
import { message } from "@bit-labs.cn/owl-ui/utils/message";
import { getSchedulerRunDetail } from "@bit-labs.cn/owl-admin-ui/api/scheduler";
import {
  SchedulerRunStatus,
  formatDuration,
  formatSchedulerTime,
  runStatusTag
} from "./constants";
import type { SchedulerRunRow } from "./types";

const props = defineProps<{
  row: SchedulerRunRow;
}>();

const detail = ref<SchedulerRunRow | null>(null);
const loading = ref(true);
const error = ref("");

let timer: ReturnType<typeof setInterval> | null = null;

const payloadText = computed(() => pretty(detail.value?.payload));
const outputText = computed(() => detail.value?.output ?? "");
const stackText = computed(() => detail.value?.stack ?? "");

function pretty(raw?: string): string {
  if (!raw) return "";
  try {
    return JSON.stringify(JSON.parse(raw), null, 2);
  } catch {
    return raw;
  }
}

async function copy(text: string, label: string) {
  if (!text) return;
  try {
    await navigator.clipboard.writeText(text);
    message(`${label}已复制`, { type: "success" });
  } catch {
    message("复制失败，请手动选择文本", { type: "error" });
  }
}

async function load(silent = false) {
  if (!silent) loading.value = true;
  try {
    const res = await getSchedulerRunDetail(props.row.id);
    detail.value = (res?.data ?? null) as unknown as SchedulerRunRow | null;
    error.value = "";
  } catch (err) {
    error.value = err instanceof Error ? err.message : "执行记录加载失败";
  } finally {
    loading.value = false;
  }
  // 待执行 / 执行中的记录继续轮询，出结果后自动停下
  if (!isActive(detail.value?.status)) stopTimer();
}

function isActive(status?: number | null): boolean {
  return status === SchedulerRunStatus.Pending || status === SchedulerRunStatus.Running;
}

function startTimer() {
  stopTimer();
  timer = setInterval(() => load(true), 5000);
}

function stopTimer() {
  if (timer) {
    clearInterval(timer);
    timer = null;
  }
}

onMounted(async () => {
  await load();
  if (isActive(detail.value?.status)) startTimer();
});
onUnmounted(stopTimer);
</script>

<template>
  <div v-loading="loading" class="run-detail">
    <el-alert
      v-if="error"
      :title="error"
      type="error"
      :closable="false"
      show-icon
      class="mb-3"
    />
    <template v-if="detail">
      <el-descriptions :column="2" border size="small">
        <el-descriptions-item label="任务名称">
          {{ detail.jobName || "—" }}
        </el-descriptions-item>
        <el-descriptions-item label="任务编码">
          <span class="break-all">{{ detail.jobCode }}</span>
        </el-descriptions-item>
        <el-descriptions-item label="触发方式">
          {{ detail.triggerTypeText || "—" }}
        </el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag size="small" :type="runStatusTag(detail.status)">
            {{ detail.statusText }}
          </el-tag>
          <span v-if="isActive(detail.status)" class="refresh-hint">每 5 秒自动刷新</span>
        </el-descriptions-item>
        <el-descriptions-item label="计划时间">
          {{ formatSchedulerTime(detail.plannedAt) }}
        </el-descriptions-item>
        <el-descriptions-item label="开始时间">
          {{ formatSchedulerTime(detail.startedAt) }}
        </el-descriptions-item>
        <el-descriptions-item label="结束时间">
          {{ formatSchedulerTime(detail.finishedAt) }}
        </el-descriptions-item>
        <el-descriptions-item label="耗时">
          {{ formatDuration(detail.durationMs) }}
        </el-descriptions-item>
        <el-descriptions-item label="尝试次数">
          {{ detail.attemptText || `${detail.attempt}/${detail.maxAttempts}` }}
        </el-descriptions-item>
        <el-descriptions-item label="重试来源">
          {{ detail.retryOfRunId && detail.retryOfRunId !== "0" ? `#${detail.retryOfRunId}` : "—" }}
        </el-descriptions-item>
        <el-descriptions-item label="执行实例" :span="2">
          <span class="break-all">{{ detail.instanceId || "—" }}</span>
        </el-descriptions-item>
      </el-descriptions>

      <div class="block">
        <p class="block-title">
          <span>执行参数</span>
          <el-button
            v-if="payloadText"
            link
            type="primary"
            size="small"
            @click="copy(payloadText, '执行参数')"
          >
            复制
          </el-button>
        </p>
        <pre v-if="payloadText" class="code">{{ payloadText }}</pre>
        <p v-else class="empty">—</p>
      </div>

      <div class="block">
        <p class="block-title">
          <span>执行输出</span>
          <el-button
            v-if="outputText"
            link
            type="primary"
            size="small"
            @click="copy(outputText, '执行输出')"
          >
            复制
          </el-button>
        </p>
        <pre v-if="outputText" class="code">{{ outputText }}</pre>
        <p v-else class="empty">—</p>
      </div>

      <div class="block">
        <p class="block-title">错误信息</p>
        <p v-if="detail.errorMsg" class="error-text">{{ detail.errorMsg }}</p>
        <p v-else class="empty">—</p>
      </div>

      <div v-if="stackText" class="block">
        <p class="block-title">
          <span>错误堆栈</span>
          <el-button link type="primary" size="small" @click="copy(stackText, '错误堆栈')">
            复制
          </el-button>
        </p>
        <pre class="code">{{ stackText }}</pre>
      </div>
    </template>
  </div>
</template>

<style lang="scss" scoped>
.run-detail {
  min-height: 160px;
}

.break-all {
  word-break: break-all;
}

.refresh-hint {
  margin-left: 8px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.block {
  margin-top: 14px;
}

.block-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin: 0 0 6px;
  font-size: 13px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.error-text {
  margin: 0;
  color: var(--el-color-danger);
  word-break: break-all;
}

.code {
  margin: 0;
  max-height: 260px;
  padding: 10px 12px;
  overflow: auto;
  border-radius: 6px;
  background: var(--el-fill-color-light);
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 12px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-all;
}

.empty {
  margin: 0;
  color: var(--el-text-color-secondary);
}
</style>
