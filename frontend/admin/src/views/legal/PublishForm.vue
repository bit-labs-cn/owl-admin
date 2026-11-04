<script setup lang="ts">
import { ref } from "vue";

const props = defineProps<{
  changeClass?: string;
}>();

const ruleFormRef = ref();
const form = ref({
  changeClass: props.changeClass || "material"
});

const rules = {
  changeClass: [{ required: true, message: "请选择变更级别", trigger: "change" }]
};

function getRef() {
  return ruleFormRef.value;
}

defineExpose({ getRef, getFormData: () => form.value });
</script>

<template>
  <el-form ref="ruleFormRef" :model="form" :rules="rules" label-width="88px">
    <el-form-item label="变更级别" prop="changeClass">
      <el-radio-group v-model="form.changeClass">
        <el-radio value="editorial">错别字，已同意的人不用重签</el-radio>
        <el-radio value="material">实质变更，未同意的人需要补签</el-radio>
      </el-radio-group>
    </el-form-item>
  </el-form>
</template>
