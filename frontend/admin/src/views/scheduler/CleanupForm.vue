<script setup lang="ts">
import { reactive, ref } from "vue";

const props = defineProps<{
  formInline: { olderThanDays: number };
}>();

const ruleFormRef = ref();
const form = reactive({
  olderThanDays: props.formInline?.olderThanDays ?? 30
});

const rules = {
  olderThanDays: [
    { required: true, message: "请填写保留天数", trigger: "blur" }
  ]
};

function getRef() {
  return ruleFormRef.value;
}

defineExpose({ getRef, getFormData: () => ({ ...form }) });
</script>

<template>
  <el-form ref="ruleFormRef" :model="form" :rules="rules" label-width="100px">
    <el-form-item label="保留天数" prop="olderThanDays">
      <el-input-number v-model="form.olderThanDays" :min="1" :max="3650" />
    </el-form-item>
    <el-alert
      title="只清理指定天数之前已经结束的记录，未完成与最近的历史会保留。"
      type="info"
      :closable="false"
      show-icon
    />
  </el-form>
</template>
