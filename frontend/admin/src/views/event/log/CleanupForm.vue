<script setup lang="ts">
import { ref } from "vue";

const props = defineProps<{
  formInline: { olderThanDays: number; targets: string[] };
}>();

const ruleFormRef = ref();
const newFormInline = ref({
  olderThanDays: props.formInline?.olderThanDays ?? 30,
  targets: props.formInline?.targets ?? ["outbox"]
});

const rules = {
  olderThanDays: [
    { required: true, message: "请输入保留天数", trigger: "blur" },
    {
      type: "number",
      min: 1,
      max: 365,
      message: "保留天数需在 1～365 之间",
      trigger: "blur"
    }
  ],
  targets: [
    {
      type: "array",
      required: true,
      message: "请选择至少一个清理目标",
      trigger: "change"
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
    <el-form-item label="保留天数" prop="olderThanDays">
      <el-input-number
        v-model="newFormInline.olderThanDays"
        :min="1"
        :max="365"
        controls-position="right"
        class="!w-[160px]"
      />
      <span class="tip">只删除这个天数之前的已结束记录</span>
    </el-form-item>
    <el-form-item label="清理目标" prop="targets">
      <el-checkbox-group v-model="newFormInline.targets">
        <el-checkbox value="outbox">发件箱（已发送）</el-checkbox>
        <el-checkbox value="inbox">收件箱（消费成功）</el-checkbox>
      </el-checkbox-group>
    </el-form-item>
    <el-alert
      title="待发送、投递失败、消费失败的流水不会被删除，方便继续排查与重投。"
      type="info"
      :closable="false"
      show-icon
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
