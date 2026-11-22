<script setup lang="ts">
import { reactive } from "vue";
import type { SchedulerJobRow } from "./types";

const props = defineProps<{
  job: SchedulerJobRow;
}>();

const form = reactive({
  payload: ""
});

type PayloadResult = { ok: true; payload: string } | { ok: false; message: string };

/** 手动触发时可附带一段 JSON 参数，内置任务会在 Handler 的 Input.Payload 里收到。 */
function getPayload(): PayloadResult {
  // 只有内置处理函数会读取 Payload；动作 / HTTP / Shell 按任务里保存的配置执行。
  if (props.job.scriptType !== 1) {
    return { ok: true, payload: "" };
  }
  const raw = form.payload.trim();
  if (!raw) {
    return { ok: true, payload: "" };
  }
  try {
    JSON.parse(raw);
  } catch {
    return { ok: false, message: "执行参数必须是合法的 JSON" };
  }
  return { ok: true, payload: raw };
}

defineExpose({ getPayload });
</script>

<template>
  <div class="trigger-dialog">
    <el-alert
      :title="`即将立即执行「${props.job.name}」`"
      type="warning"
      :closable="false"
      show-icon
      class="mb-3"
    />
    <el-alert
      v-if="props.job.scriptType !== 1"
      title="该任务按任务里保存的配置执行，手动触发不会改变参数。"
      type="info"
      :closable="false"
      show-icon
      class="mb-3"
    />
    <el-form v-else label-width="90px">
      <el-form-item label="执行参数">
        <el-input
          v-model="form.payload"
          type="textarea"
          :rows="5"
          placeholder='可选，JSON 对象；例如 {"source":"manual"}'
        />
      </el-form-item>
    </el-form>
  </div>
</template>

<style lang="scss" scoped>
.trigger-dialog {
  min-height: 120px;
}
</style>
