<script setup lang="ts">
import { ref } from "vue";
import type { LegalDocumentForm } from "./types";

const props = defineProps<{
  formInline: LegalDocumentForm;
}>();

const ruleFormRef = ref();
const newFormInline = ref<LegalDocumentForm>({
  id: props.formInline?.id,
  key: props.formInline?.key ?? "",
  name: props.formInline?.name ?? "",
  kind: props.formInline?.kind ?? "contract",
  status: props.formInline?.status ?? "enabled"
});

const rules = {
  key: [{ required: true, message: "请输入标识", trigger: "blur" }],
  name: [{ required: true, message: "请输入名称", trigger: "blur" }],
  kind: [{ required: true, message: "请选择性质", trigger: "change" }]
};

function getRef() {
  return ruleFormRef.value;
}

defineExpose({ getRef, getFormData: () => newFormInline.value });
</script>

<template>
  <el-form ref="ruleFormRef" :model="newFormInline" :rules="rules" label-width="72px">
    <el-form-item label="标识" prop="key">
      <el-input
        v-model="newFormInline.key"
        :disabled="!!newFormInline.id"
        clearable
        placeholder="例如 terms，小写字母开头"
      />
    </el-form-item>
    <el-form-item label="名称" prop="name">
      <el-input v-model="newFormInline.name" clearable placeholder="给人看的名称" />
    </el-form-item>
    <el-form-item label="性质" prop="kind">
      <el-select v-model="newFormInline.kind" class="w-full">
        <el-option label="合同" value="contract" />
        <el-option label="告知" value="notice" />
        <el-option label="可选同意" value="optional" />
      </el-select>
    </el-form-item>
  </el-form>
</template>
