<script setup lang="ts">
import { ref } from "vue";
import type { LegalVersionForm } from "./types";

const props = defineProps<{
  formInline: LegalVersionForm;
}>();

const ruleFormRef = ref();
const newFormInline = ref<LegalVersionForm>({
  id: props.formInline?.id,
  documentId: props.formInline?.documentId ?? "",
  version: props.formInline?.version ?? "",
  title: props.formInline?.title ?? "",
  content: props.formInline?.content ?? ""
});

const rules = {
  version: [{ required: true, message: "请输入版本号", trigger: "blur" }],
  title: [{ required: true, message: "请输入标题", trigger: "blur" }],
  content: [{ required: true, message: "请输入正文", trigger: "blur" }]
};

function getRef() {
  return ruleFormRef.value;
}

defineExpose({ getRef, getFormData: () => newFormInline.value });
</script>

<template>
  <el-form ref="ruleFormRef" :model="newFormInline" :rules="rules" label-width="72px">
    <el-form-item label="版本号" prop="version">
      <el-input v-model="newFormInline.version" clearable placeholder="例如 1.0.0" />
    </el-form-item>
    <el-form-item label="标题" prop="title">
      <el-input v-model="newFormInline.title" clearable placeholder="阅读页上的标题" />
    </el-form-item>
    <el-form-item label="正文" prop="content">
      <el-input
        v-model="newFormInline.content"
        type="textarea"
        :rows="12"
        placeholder="Markdown"
      />
    </el-form-item>
  </el-form>
</template>
