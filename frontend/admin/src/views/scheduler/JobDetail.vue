<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { getSchedulerJobDetail } from "@bit-labs.cn/owl-admin-ui/api/scheduler";
import {
  formatSchedulerTime,
  jobSourceTag,
  jobStatusTag,
  runStatusTag
} from "./constants";
import { describeCron } from "./cronPreset";
import type { SchedulerJobDetail, SchedulerJobRow } from "./types";

const props = defineProps<{
  row: SchedulerJobRow;
}>();

const detail = ref<SchedulerJobDetail | null>(null);
const loading = ref(true);
const error = ref("");

const cronReadable = computed(() => describeCron(detail.value?.cron ?? props.row.cron));

const scriptText = computed(() => {
  const raw = detail.value?.scriptConfig;
  if (!raw) return "";
  try {
    return JSON.stringify(JSON.parse(raw), null, 2);
  } catch {
    return raw;
  }
});

onMounted(async () => {
  try {
    const res = await getSchedulerJobDetail(props.row.id);
    detail.value = (res?.data ?? null) as unknown as SchedulerJobDetail | null;
  } catch (err) {
    error.value = err instanceof Error ? err.message : "任务详情加载失败";
  } finally {
    loading.value = false;
  }
});
</script>

<template>
  <div v-loading="loading" class="job-detail">
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
          {{ detail.name }}
        </el-descriptions-item>
        <el-descriptions-item label="任务编码">
          <span class="break-all">{{ detail.code }}</span>
        </el-descriptions-item>
        <el-descriptions-item label="来源">
          <el-tag size="small" :type="jobSourceTag(detail.source)">
            {{ detail.sourceText }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="所属子系统">
          {{ detail.ownerApp || "自定义" }}
        </el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag size="small" :type="jobStatusTag(detail.status)">
            {{ detail.statusText }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="执行方式">
          {{ detail.scriptTypeText }}
        </el-descriptions-item>
        <el-descriptions-item label="调用方法">
          {{ detail.actionCode ? detail.actionName || detail.actionCode : "—" }}
        </el-descriptions-item>
        <el-descriptions-item label="参数摘要" :span="2">
          {{ detail.scriptSummary || "—" }}
        </el-descriptions-item>
        <el-descriptions-item label="执行规则">
          <span>{{ cronReadable }}</span>
          <code class="cron-code">{{ detail.cron || "—" }}</code>
        </el-descriptions-item>
        <el-descriptions-item label="时区">
          {{ detail.timezone || "跟随服务器默认" }}
        </el-descriptions-item>
        <el-descriptions-item label="超时">
          {{ detail.timeoutSeconds }} 秒
        </el-descriptions-item>
        <el-descriptions-item label="最大尝试">
          {{ detail.maxAttempts }} 次
        </el-descriptions-item>
        <el-descriptions-item label="重试间隔">
          {{ detail.retryBackoffSeconds }} 秒
        </el-descriptions-item>
        <el-descriptions-item label="并发策略">
          {{ detail.overlapText }}
        </el-descriptions-item>
        <el-descriptions-item label="下次执行">
          {{ formatSchedulerTime(detail.nextRunAt) }}
        </el-descriptions-item>
        <el-descriptions-item label="最近执行">
          <span class="flex items-center gap-2">
            <el-tag
              v-if="detail.lastStatus != null"
              size="small"
              :type="runStatusTag(detail.lastStatus)"
            >
              {{ detail.lastStatusText }}
            </el-tag>
            <span>{{ formatSchedulerTime(detail.lastRunAt) }}</span>
          </span>
        </el-descriptions-item>
        <el-descriptions-item label="执行统计">
          成功 {{ Math.max(detail.runCount - detail.failCount, 0) }} · 失败
          {{ detail.failCount }}
        </el-descriptions-item>
        <el-descriptions-item label="最近错误">
          {{ detail.lastError || "—" }}
        </el-descriptions-item>
        <el-descriptions-item label="任务说明" :span="2">
          {{ detail.description || "—" }}
        </el-descriptions-item>
      </el-descriptions>

      <div class="block">
        <p class="block-title">{{ detail.scriptType === 4 ? "调用参数" : "脚本配置" }}</p>
        <pre v-if="scriptText" class="code">{{ scriptText }}</pre>
        <p v-else class="empty">内置任务的执行逻辑在代码里，没有可查看的脚本配置。</p>
      </div>
    </template>
  </div>
</template>

<style lang="scss" scoped>
.job-detail {
  min-height: 160px;
}

.break-all {
  word-break: break-all;
}

.cron-code {
  margin-left: 8px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.block {
  margin-top: 14px;
}

.block-title {
  margin: 0 0 6px;
  font-size: 13px;
  font-weight: 600;
  color: var(--el-text-color-primary);
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
