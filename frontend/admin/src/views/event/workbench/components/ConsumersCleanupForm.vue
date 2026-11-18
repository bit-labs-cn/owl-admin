<script setup lang="ts">
import { ref } from "vue";

const props = defineProps<{
  formInline: { idleDays: number };
}>();

const ruleFormRef = ref();
const newFormInline = ref({
  idleDays: props.formInline?.idleDays ?? 7
});

const rules = {
  idleDays: [
    { required: true, message: "请输入空闲天数", trigger: "blur" },
    {
      type: "number",
      min: 1,
      max: 365,
      message: "空闲天数需在 1～365 之间",
      trigger: "blur"
    }
  ]
};

function getRef() {
  return ruleFormRef.value;
}

defineExpose({ getRef, getFormData: () => newFormInline.value });
</script>

<template>
  <el-form
    ref="ruleFormRef"
    :model="newFormInline"
    :rules="rules"
    label-width="96px"
  >
    <el-form-item label="空闲天数" prop="idleDays">
      <el-input-number
        v-model="newFormInline.idleDays"
        :min="1"
        :max="365"
        controls-position="right"
        class="!w-[160px]"
      />
      <span class="tip">只清理超过这个天数没读过消息的消费者</span>
    </el-form-item>
    <el-alert
      title="Redis 不会自动回收消费者：进程每次重启都会按新 pid 生成新名字，旧记录会一直留在消费组里，把消费者数堆大。删除这些记录不影响正在运行的消费者，也不会丢消息。"
      type="info"
      :closable="false"
      show-icon
    />
    <el-alert
      title="还有未确认消息（pending > 0）的消费者会被跳过，避免丢掉待重试的消息。"
      type="warning"
      :closable="false"
      show-icon
      class="mt-2"
    />
  </el-form>
</template>

<style lang="scss" scoped>
.tip {
  margin-left: 10px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
</style>
