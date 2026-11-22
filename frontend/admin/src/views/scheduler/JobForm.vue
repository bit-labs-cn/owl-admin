<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import { message } from "@bit-labs.cn/owl-ui/utils/message";
import {
  getSchedulerActionOptions,
  getSchedulerActions
} from "@bit-labs.cn/owl-admin-ui/api/scheduler";
import {
  SchedulerOverlap,
  SchedulerScriptType,
  overlapOptions,
  scriptTypeOptions
} from "./constants";
import { describeCron } from "./cronPreset";
import CronBuilder from "./CronBuilder.vue";
import ActionPicker from "./ActionPicker.vue";
import type {
  SchedulerAction,
  SchedulerActionOption,
  SchedulerActionParam,
  SchedulerJobFormData
} from "./types";

const props = defineProps<{
  formInline: SchedulerJobFormData;
}>();

const ruleFormRef = ref();
const cronBuilderRef = ref();
const actionPickerRef = ref();

const actions = ref<SchedulerAction[]>([]);
const actionsLoading = ref(false);
const actionOptions = reactive<Record<string, SchedulerActionOption[]>>({});
const optionsLoading = ref<Record<string, boolean>>({});

const newFormInline = reactive<SchedulerJobFormData>({
  id: props.formInline?.id,
  name: props.formInline?.name ?? "",
  description: props.formInline?.description ?? "",
  cron: props.formInline?.cron ?? "",
  timezone: props.formInline?.timezone ?? "",
  timeoutSeconds: props.formInline?.timeoutSeconds ?? 60,
  maxAttempts: props.formInline?.maxAttempts ?? 1,
  retryBackoffSeconds: props.formInline?.retryBackoffSeconds ?? 60,
  overlapPolicy: props.formInline?.overlapPolicy ?? SchedulerOverlap.Skip,
  scriptType: props.formInline?.scriptType ?? SchedulerScriptType.Action,
  actionCode: props.formInline?.actionCode ?? "",
  actionParams: { ...(props.formInline?.actionParams ?? {}) },
  httpMethod: props.formInline?.httpMethod ?? "POST",
  httpURL: props.formInline?.httpURL ?? "",
  httpHeaders: props.formInline?.httpHeaders ?? "",
  httpBody: props.formInline?.httpBody ?? "",
  httpContentType: props.formInline?.httpContentType ?? "application/json",
  httpSuccessStatus: props.formInline?.httpSuccessStatus ?? "",
  shellCommand: props.formInline?.shellCommand ?? "",
  shellWorkingDir: props.formInline?.shellWorkingDir ?? ""
});

const rules = {
  name: [{ required: true, message: "请输入任务名称", trigger: "blur" }],
  cron: [
    {
      required: true,
      message: "请点击「设置时间」配置执行时间",
      trigger: "change"
    }
  ]
};

const selectedAction = computed(() =>
  actions.value.find(item => item.code === newFormInline.actionCode)
);

/** Cron 的人类可读描述，直接显示在表单上，不必让用户读懂表达式。 */
const cronReadable = computed(() => describeCron(newFormInline.cron));

function openCronBuilder() {
  cronBuilderRef.value?.open(newFormInline.cron);
}

function onCronConfirm(cron: string) {
  newFormInline.cron = cron;
  // Cron 与表单校验规则绑定，确认后主动触发一次校验状态刷新
  ruleFormRef.value?.validateField?.("cron");
}

function openActionPicker() {
  actionPickerRef.value?.open(newFormInline.actionCode);
}

function onActionPicked(action: SchedulerAction) {
  if (action.code === newFormInline.actionCode) return;
  onActionChange(action.code);
}

function getRef() {
  return ruleFormRef.value;
}

function paramOptions(param: SchedulerActionParam): SchedulerActionOption[] {
  if (param.dynamic) {
    return actionOptions[param.name] ?? [];
  }
  return param.options ?? [];
}

/** 合并默认值，保证表单提交时参数完整。 */
function applyParamDefaults(action: SchedulerAction) {
  for (const param of action.params) {
    if (newFormInline.actionParams[param.name] == null) {
      newFormInline.actionParams[param.name] = param.default ?? "";
    }
  }
}

async function loadActionOptions() {
  const action = selectedAction.value;
  if (!action) return;
  applyParamDefaults(action);

  const hasDynamic = action.params.some(param => param.dynamic);
  if (!hasDynamic) return;

  try {
    const res = await getSchedulerActionOptions(
      action.code,
      newFormInline.actionParams
    );
    const data = (res?.data ?? {}) as {
      options?: Record<string, SchedulerActionOption[]>;
    };
    for (const [name, options] of Object.entries(data.options ?? {})) {
      actionOptions[name] = options ?? [];
    }
  } catch (err) {
    for (const param of action.params) {
      if (param.dynamic) actionOptions[param.name] = [];
    }
    message(err instanceof Error ? err.message : "方法参数候选加载失败", {
      type: "error"
    });
  } finally {
    for (const param of action.params) {
      if (param.dynamic) optionsLoading.value[param.name] = false;
    }
  }
}

async function loadActions() {
  actionsLoading.value = true;
  try {
    const res = await getSchedulerActions();
    actions.value = (res?.data ?? []) as unknown as SchedulerAction[];
    if (newFormInline.actionCode) {
      await loadActionOptions();
    }
  } catch (err) {
    message(err instanceof Error ? err.message : "可调用方法目录加载失败", {
      type: "error"
    });
  } finally {
    actionsLoading.value = false;
  }
}

function onActionChange(code: string) {
  newFormInline.actionCode = code;
  // 换动作时清空旧参数，避免把上一个动作的参数带过去
  newFormInline.actionParams = {};
  for (const key of Object.keys(actionOptions)) {
    delete actionOptions[key];
  }
  loadActionOptions();
}

function onParamChange(param: SchedulerActionParam) {
  if (param.dynamic) {
    optionsLoading.value[param.name] = true;
  }
  // 任一参数变化都可能影响其它动态候选（如开关 → 命令），整体刷新一次
  loadActionOptions();
}

type PayloadResult =
  | { ok: true; data: Record<string, unknown> }
  | { ok: false; message: string };

function validateActionParams(): string {
  const action = selectedAction.value;
  if (!action) return "请选择要调用的方法";
  for (const param of action.params) {
    const raw = String(newFormInline.actionParams[param.name] ?? "").trim();
    if (param.required && !raw) {
      return `请填写${param.label}`;
    }
    if (param.type === "int" && raw) {
      const value = Number(raw);
      if (!Number.isInteger(value)) return `${param.label}必须是整数`;
      if (param.min != null && value < param.min) return `${param.label}不能小于 ${param.min}`;
      if (param.max != null && value > param.max) return `${param.label}不能大于 ${param.max}`;
    }
    if (
      param.type === "select" &&
      raw &&
      !param.dynamic &&
      !(param.options ?? []).some(option => option.value === raw)
    ) {
      return `${param.label}的取值不在候选范围内`;
    }
  }
  return "";
}

/** 校验脚本配置并组装提交数据。 */
function getPayload(): PayloadResult {
  if (newFormInline.scriptType === SchedulerScriptType.Action) {
    const error = validateActionParams();
    if (error) return { ok: false, message: error };
    return {
      ok: true,
      data: {
        name: newFormInline.name.trim(),
        description: newFormInline.description.trim(),
        cron: newFormInline.cron.trim(),
        timezone: newFormInline.timezone.trim(),
        timeoutSeconds: newFormInline.timeoutSeconds,
        maxAttempts: newFormInline.maxAttempts,
        retryBackoffSeconds: newFormInline.retryBackoffSeconds,
        overlapPolicy: newFormInline.overlapPolicy,
        scriptType: newFormInline.scriptType,
        actionCode: newFormInline.actionCode,
        actionParams: newFormInline.actionParams
      }
    };
  }

  let scriptConfig = "";

  if (newFormInline.scriptType === SchedulerScriptType.Shell) {
    if (!newFormInline.shellCommand.trim()) {
      return { ok: false, message: "请填写要执行的命令" };
    }
    scriptConfig = JSON.stringify({
      command: newFormInline.shellCommand.trim(),
      workingDir: newFormInline.shellWorkingDir.trim()
    });
  } else {
    const url = newFormInline.httpURL.trim();
    if (!url) {
      return { ok: false, message: "请填写请求地址" };
    }
    if (!/^https?:\/\//i.test(url)) {
      return { ok: false, message: "请求地址只支持 http 或 https" };
    }

    let headers: Record<string, string> = {};
    if (newFormInline.httpHeaders.trim()) {
      try {
        const parsed = JSON.parse(newFormInline.httpHeaders);
        if (parsed && typeof parsed === "object" && !Array.isArray(parsed)) {
          headers = parsed as Record<string, string>;
        } else {
          return { ok: false, message: "请求头必须是 JSON 对象" };
        }
      } catch {
        return { ok: false, message: "请求头不是合法的 JSON" };
      }
    }

    const successStatus = newFormInline.httpSuccessStatus
      .split(",")
      .map(item => item.trim())
      .filter(Boolean)
      .map(item => Number(item));
    if (successStatus.some(item => !Number.isInteger(item) || item < 100 || item > 599)) {
      return { ok: false, message: "成功状态码必须是 100-599 的整数，用逗号分隔" };
    }

    scriptConfig = JSON.stringify({
      method: newFormInline.httpMethod,
      url,
      headers,
      body: newFormInline.httpBody,
      contentType: newFormInline.httpContentType.trim(),
      successStatus
    });
  }

  return {
    ok: true,
    data: {
      name: newFormInline.name.trim(),
      description: newFormInline.description.trim(),
      cron: newFormInline.cron.trim(),
      timezone: newFormInline.timezone.trim(),
      timeoutSeconds: newFormInline.timeoutSeconds,
      maxAttempts: newFormInline.maxAttempts,
      retryBackoffSeconds: newFormInline.retryBackoffSeconds,
      overlapPolicy: newFormInline.overlapPolicy,
      scriptType: newFormInline.scriptType,
      scriptConfig
    }
  };
}

onMounted(loadActions);

defineExpose({ getRef, getPayload });
</script>

<template>
  <el-form ref="ruleFormRef" :model="newFormInline" :rules="rules" label-width="110px">
    <el-form-item label="任务名称" prop="name">
      <el-input v-model="newFormInline.name" clearable placeholder="例如：每天 8 点前台开机" />
    </el-form-item>
    <el-form-item label="任务说明" prop="description">
      <el-input
        v-model="newFormInline.description"
        type="textarea"
        :rows="2"
        placeholder="说明这个任务做什么、依赖什么，便于别人接手"
      />
    </el-form-item>

    <el-form-item label="执行时间" prop="cron">
      <div class="cron-field">
        <div class="cron-display">
          <span class="cron-readable">{{ cronReadable }}</span>
          <code class="cron-expression">{{ newFormInline.cron || "未配置" }}</code>
        </div>
        <el-button type="primary" plain @click="openCronBuilder">设置时间</el-button>
      </div>
    </el-form-item>

    <el-form-item label="时区" prop="timezone">
      <el-input
        v-model="newFormInline.timezone"
        clearable
        placeholder="例如 Asia/Shanghai，留空用服务器默认时区"
      />
    </el-form-item>

    <div class="inline-row">
      <el-form-item label="超时（秒）" prop="timeoutSeconds">
        <el-input-number v-model="newFormInline.timeoutSeconds" :min="1" :max="86400" />
      </el-form-item>
      <el-form-item label="最大尝试" prop="maxAttempts">
        <el-input-number v-model="newFormInline.maxAttempts" :min="1" :max="10" />
      </el-form-item>
      <el-form-item label="重试间隔（秒）" prop="retryBackoffSeconds">
        <el-input-number v-model="newFormInline.retryBackoffSeconds" :min="0" :max="86400" />
      </el-form-item>
    </div>

    <el-form-item label="并发策略" prop="overlapPolicy">
      <el-radio-group v-model="newFormInline.overlapPolicy">
        <el-radio v-for="item in overlapOptions" :key="item.value" :value="item.value">
          {{ item.label }}
        </el-radio>
      </el-radio-group>
    </el-form-item>

    <el-form-item label="执行方式" prop="scriptType">
      <el-radio-group v-model="newFormInline.scriptType">
        <el-radio v-for="item in scriptTypeOptions" :key="item.value" :value="item.value">
          {{ item.label }}
        </el-radio>
      </el-radio-group>
    </el-form-item>

    <template v-if="newFormInline.scriptType === SchedulerScriptType.Action">
      <el-form-item label="调用方法" prop="actionCode">
        <div class="action-field">
          <div
            class="action-display"
            :class="{ 'is-empty': !selectedAction }"
            @click="openActionPicker"
          >
            <template v-if="selectedAction">
              <div class="action-display-head">
                <span class="action-display-name">{{ selectedAction.name }}</span>
                <el-tag size="small" type="info">
                  {{ selectedAction.group || "其它" }}
                </el-tag>
              </div>
              <p class="action-display-desc">
                {{ selectedAction.description || "没有填写说明" }}
              </p>
            </template>
            <span v-else class="action-display-placeholder">
              选择子系统登记的方法（可按系统浏览方法与说明）
            </span>
          </div>
          <el-button :loading="actionsLoading" @click="openActionPicker">
            {{ selectedAction ? "更换方法" : "选择方法" }}
          </el-button>
        </div>
      </el-form-item>

      <template v-if="selectedAction">
        <el-form-item
          v-for="param in selectedAction.params"
          :key="param.name"
          :label="param.label"
          :required="param.required"
        >
          <el-select
            v-if="param.type === 'select'"
            v-model="newFormInline.actionParams[param.name]"
            :multiple="param.multiple"
            :loading="optionsLoading[param.name]"
            filterable
            clearable
            placeholder="请选择"
            class="w-full"
            @change="onParamChange(param)"
          >
            <el-option
              v-for="option in paramOptions(param)"
              :key="option.value"
              :label="option.label"
              :value="option.value"
              :disabled="option.disabled"
            />
          </el-select>
          <el-input
            v-else-if="param.type === 'int'"
            v-model="newFormInline.actionParams[param.name]"
            type="number"
            :placeholder="param.min != null && param.max != null ? `${param.min} ~ ${param.max}` : '请输入整数'"
            @change="onParamChange(param)"
          />
          <el-switch
            v-else-if="param.type === 'bool'"
            v-model="newFormInline.actionParams[param.name]"
            active-value="true"
            inactive-value="false"
          />
          <el-input
            v-else
            v-model="newFormInline.actionParams[param.name]"
            :type="param.type === 'text' ? 'textarea' : 'text'"
            :rows="3"
            :placeholder="param.placeholder"
            clearable
          />
          <div v-if="param.description" class="param-hint">{{ param.description }}</div>
        </el-form-item>
      </template>
    </template>

    <template v-else-if="newFormInline.scriptType === SchedulerScriptType.HTTP">
      <el-form-item label="请求方法" prop="httpMethod">
        <el-select v-model="newFormInline.httpMethod" class="w-full">
          <el-option label="GET" value="GET" />
          <el-option label="POST" value="POST" />
          <el-option label="PUT" value="PUT" />
          <el-option label="PATCH" value="PATCH" />
          <el-option label="DELETE" value="DELETE" />
        </el-select>
      </el-form-item>
      <el-form-item label="请求地址" prop="httpURL">
        <el-input v-model="newFormInline.httpURL" clearable placeholder="https://example.com/hook" />
      </el-form-item>
      <el-form-item label="Content-Type" prop="httpContentType">
        <el-input v-model="newFormInline.httpContentType" clearable placeholder="application/json" />
      </el-form-item>
      <el-form-item label="请求头" prop="httpHeaders">
        <el-input
          v-model="newFormInline.httpHeaders"
          type="textarea"
          :rows="3"
          placeholder='JSON 对象，例如 {"X-Token":"xxx"}；敏感值建议走环境变量，不要写死'
        />
      </el-form-item>
      <el-form-item label="请求体" prop="httpBody">
        <el-input
          v-model="newFormInline.httpBody"
          type="textarea"
          :rows="4"
          placeholder="请求体原文；留空表示不带请求体"
        />
      </el-form-item>
      <el-form-item label="成功状态码" prop="httpSuccessStatus">
        <el-input v-model="newFormInline.httpSuccessStatus" clearable placeholder="留空表示 2xx 都算成功" />
      </el-form-item>
    </template>

    <template v-else>
      <el-form-item label="执行命令" prop="shellCommand">
        <el-input
          v-model="newFormInline.shellCommand"
          type="textarea"
          :rows="4"
          placeholder="例如 php think crontab；Windows 用 cmd /C 执行，其它系统用 /bin/sh -c"
        />
      </el-form-item>
      <el-form-item label="工作目录" prop="shellWorkingDir">
        <el-input v-model="newFormInline.shellWorkingDir" clearable placeholder="留空沿用服务进程的工作目录" />
      </el-form-item>
    </template>
  </el-form>

  <CronBuilder
    ref="cronBuilderRef"
    :timezone="newFormInline.timezone"
    @confirm="onCronConfirm"
  />
  <ActionPicker
    ref="actionPickerRef"
    :actions="actions"
    @confirm="onActionPicked"
  />
</template>

<style lang="scss" scoped>
.cron-field {
  display: flex;
  gap: 8px;
  align-items: flex-start;
  width: 100%;
}

.cron-display {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 4px;
  padding: 6px 12px;
  background: var(--el-fill-color-lighter);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 6px;
}

.cron-readable {
  font-size: 13px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.cron-expression {
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.action-field {
  display: flex;
  gap: 8px;
  align-items: flex-start;
  width: 100%;
}

.action-display {
  flex: 1;
  min-height: 56px;
  padding: 8px 12px;
  cursor: pointer;
  background: var(--el-fill-color-lighter);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 6px;
  transition: border-color 0.15s;

  &:hover {
    border-color: var(--el-color-primary-light-5);
  }

  &.is-empty {
    display: flex;
    align-items: center;
  }
}

.action-display-head {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  align-items: center;
}

.action-display-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.action-display-desc {
  margin: 4px 0 0;
  font-size: 12px;
  line-height: 1.6;
  color: var(--el-text-color-secondary);
}

.action-display-placeholder {
  font-size: 13px;
  color: var(--el-text-color-placeholder);
}

.inline-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0 12px;
}

.param-hint {
  width: 100%;
  margin-top: 4px;
  font-size: 12px;
  line-height: 1.5;
  color: var(--el-text-color-secondary);
}
</style>
