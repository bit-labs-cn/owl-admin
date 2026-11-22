<script setup lang="ts">
import { computed, onUnmounted, reactive, ref, watch } from "vue";
import { message } from "@bit-labs.cn/owl-ui/utils/message";
import { previewSchedulerCron } from "@bit-labs.cn/owl-admin-ui/api/scheduler";
import {
  CRON_MODE_OPTIONS,
  WEEKDAY_LABELS,
  buildCron,
  defaultCronConfig,
  describeCron,
  parseCron,
  type CronConfig
} from "./cronPreset";

const props = defineProps<{
  /** 任务时区，用于预览下次执行时间 */
  timezone?: string;
}>();

const emit = defineEmits<{
  (e: "confirm", cron: string): void;
}>();

const visible = ref(false);
const config = reactive<CronConfig>(defaultCronConfig());
const previewTimes = ref<string[]>([]);
const previewLoading = ref(false);
const previewError = ref("");
/** 打开弹窗时的原始表达式，用于判断自定义输入是否被用户改过 */
const originalExpression = ref("");

let previewTimer: ReturnType<typeof setTimeout> | null = null;

const expression = computed(() => buildCron(config));
const readable = computed(() => describeCron(expression.value));

const weekdayModel = computed({
  get: () => config.weekdays,
  set: (value: number[]) => {
    config.weekdays = value.length ? value : [1];
  }
});

function syncConfig(expr: string) {
  const parsed = parseCron(expr);
  Object.assign(config, parsed);
  // 空表达式直接给一个可用的默认（每天 08:00），避免用户面对空表单
  if (!(expr ?? "").trim()) {
    config.mode = "daily";
  }
}

function open(cron: string) {
  originalExpression.value = (cron ?? "").trim();
  syncConfig(cron ?? "");
  visible.value = true;
}

async function refreshPreview() {
  previewError.value = "";
  const expr = expression.value;
  if (!expr) {
    previewTimes.value = [];
    return;
  }
  previewLoading.value = true;
  try {
    const res = await previewSchedulerCron({
      cron: expr,
      timezone: props.timezone?.trim() || undefined,
      count: 5
    });
    const data = (res?.data ?? {}) as { times?: string[] };
    previewTimes.value = data.times ?? [];
  } catch (err) {
    previewTimes.value = [];
    previewError.value = err instanceof Error ? err.message : "表达式无法预览";
  } finally {
    previewLoading.value = false;
  }
}

// 表达式变化后延迟请求，避免连续输入时每敲一个字符都打一次接口
watch(
  () => expression.value,
  () => {
    if (!visible.value) return;
    if (previewTimer) clearTimeout(previewTimer);
    previewTimer = setTimeout(() => refreshPreview(), 350);
  }
);

// 从预设模式切到自定义时，把当前预设生成的表达式填进去；
// 已经手动改过自定义文本的保持原样，避免把用户的输入冲掉。
watch(
  () => config.mode,
  (mode, prev) => {
    if (mode !== "custom") return;
    const custom = config.custom.trim();
    if (custom && custom !== originalExpression.value) return;
    config.custom = buildCron({ ...config, mode: prev ?? "daily" });
    originalExpression.value = config.custom;
  }
);

watch(visible, value => {
  if (value) {
    refreshPreview();
    return;
  }
  // 关闭后清掉待执行的预览请求
  if (previewTimer) {
    clearTimeout(previewTimer);
    previewTimer = null;
  }
});

async function copyExpression() {
  try {
    await navigator.clipboard.writeText(expression.value);
    message("表达式已复制", { type: "success" });
  } catch {
    message("复制失败，请手动选择文本", { type: "error" });
  }
}

function confirm() {
  const expr = expression.value.trim();
  if (!expr) {
    message("表达式不能为空", { type: "warning" });
    return;
  }
  emit("confirm", expr);
  visible.value = false;
}

defineExpose({ open });

onUnmounted(() => {
  if (previewTimer) clearTimeout(previewTimer);
});
</script>

<template>
  <el-dialog
    v-model="visible"
    title="设置执行时间"
    width="640px"
    append-to-body
    align-center
    :close-on-click-modal="false"
    :close-on-press-escape="false"
    class="cron-builder-dialog"
  >
    <div class="cron-builder">
      <el-radio-group v-model="config.mode" class="mode-group">
        <el-radio-button
          v-for="item in CRON_MODE_OPTIONS"
          :key="item.value"
          :value="item.value"
        >
          {{ item.label }}
        </el-radio-button>
      </el-radio-group>

      <div class="mode-panel">
        <template v-if="config.mode === 'minutes'">
          <span class="panel-text">每隔</span>
          <el-input-number v-model="config.minuteInterval" :min="1" :max="59" />
          <span class="panel-text">分钟执行一次</span>
          <span class="panel-hint">填 1 表示每分钟</span>
        </template>

        <template v-else-if="config.mode === 'hourInterval'">
          <span class="panel-text">每隔</span>
          <el-input-number v-model="config.hourInterval" :min="1" :max="23" />
          <span class="panel-text">小时，在第</span>
          <el-input-number v-model="config.minute" :min="0" :max="59" />
          <span class="panel-text">分钟执行</span>
        </template>

        <template v-else-if="config.mode === 'daily'">
          <span class="panel-text">每天</span>
          <el-input-number v-model="config.hour" :min="0" :max="23" />
          <span class="panel-text">时</span>
          <el-input-number v-model="config.minute" :min="0" :max="59" />
          <span class="panel-text">分执行</span>
        </template>

        <template v-else-if="config.mode === 'weekly'">
          <div class="weekly-row">
            <el-checkbox-group v-model="weekdayModel">
              <el-checkbox
                v-for="item in WEEKDAY_LABELS"
                :key="item.value"
                :value="item.value"
              >
                {{ item.label }}
              </el-checkbox>
            </el-checkbox-group>
          </div>
          <div class="time-row">
            <span class="panel-text">时间</span>
            <el-input-number v-model="config.hour" :min="0" :max="23" />
            <span class="panel-text">时</span>
            <el-input-number v-model="config.minute" :min="0" :max="59" />
            <span class="panel-text">分</span>
          </div>
        </template>

        <template v-else-if="config.mode === 'monthly'">
          <span class="panel-text">每月</span>
          <el-input-number v-model="config.monthDay" :min="1" :max="31" />
          <span class="panel-text">号</span>
          <el-input-number v-model="config.hour" :min="0" :max="23" />
          <span class="panel-text">时</span>
          <el-input-number v-model="config.minute" :min="0" :max="59" />
          <span class="panel-text">分执行</span>
          <span class="panel-hint">没有该日期的月份会跳过</span>
        </template>

        <template v-else>
          <el-input
            v-model="config.custom"
            placeholder="例如 0 8 * * 1-5，支持 5 段或 6 段表达式"
            clearable
          />
          <span class="panel-hint">6 段表达式第一段是秒，例如 30 0 8 * * *</span>
        </template>
      </div>

      <el-descriptions :column="1" border size="small" class="result">
        <el-descriptions-item label="执行规则">
          <span class="readable">{{ readable }}</span>
        </el-descriptions-item>
        <el-descriptions-item label="表达式">
          <div class="expression-row">
            <code class="expression">{{ expression || "—" }}</code>
            <el-button link type="primary" size="small" @click="copyExpression">
              复制
            </el-button>
          </div>
        </el-descriptions-item>
        <el-descriptions-item label="接下来">
          <div v-if="previewError" class="preview-box preview-error">{{ previewError }}</div>
          <div v-else-if="previewLoading" class="preview-box preview-empty">正在计算…</div>
          <div v-else-if="previewTimes.length" class="preview-box preview-list">
            <span v-for="time in previewTimes" :key="time">{{ time }}</span>
          </div>
          <div v-else class="preview-box preview-empty">暂无即将执行的时间</div>
        </el-descriptions-item>
      </el-descriptions>
    </div>

    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" @click="confirm">确定</el-button>
    </template>
  </el-dialog>
</template>

<style lang="scss" scoped>
.cron-builder {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.mode-group {
  flex-wrap: wrap;
}

.mode-panel {
  display: flex;
  flex-wrap: wrap;
  /* 各模式的内容行数不同（每周是两行、每月多一行说明），高度统一才不会在切换时让弹窗上下跳动 */
  align-content: center;
  align-items: center;
  box-sizing: border-box;
  width: 100%;
  min-height: 104px;
  padding: 14px 16px;
  gap: 8px;
  background: var(--el-fill-color-lighter);
  border-radius: 6px;
}

.panel-text {
  font-size: 13px;
  color: var(--el-text-color-regular);
}

.panel-hint {
  width: 100%;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.weekly-row {
  width: 100%;
}

.time-row {
  display: flex;
  gap: 8px;
  align-items: center;
}

.result {
  :deep(.el-descriptions__label) {
    width: 92px;
  }
}

.readable {
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.expression-row {
  display: flex;
  gap: 8px;
  align-items: center;
}

.expression {
  padding: 2px 8px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 12px;
  background: var(--el-fill-color-light);
  border-radius: 4px;
}

/* 「接下来」占两行高度：预测结果从「正在计算…」变成 5 条时间时不会撑高整块说明 */
.preview-box {
  display: flex;
  flex-wrap: wrap;
  gap: 6px 14px;
  align-content: flex-start;
  box-sizing: border-box;
  min-height: 42px;
  font-size: 12px;
}

.preview-list {
  color: var(--el-text-color-regular);
}

.preview-empty {
  color: var(--el-text-color-secondary);
}

.preview-error {
  color: var(--el-color-danger);
}
</style>
