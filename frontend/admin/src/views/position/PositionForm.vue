<script setup lang="ts">
import { ref } from "vue";
import type { PositionFormData } from "./types";

const props = defineProps<{
  formInline: PositionFormData;
}>();

const ruleFormRef = ref();
const newFormInline = ref<PositionFormData>({
  id: props.formInline?.id ?? undefined,
  name: props.formInline?.name ?? "",
  remark: props.formInline?.remark ?? "",
  status: props.formInline?.status ?? 1
});

const rules = {
  name: [{ required: true, message: "请输入岗位名称", trigger: "blur" }]
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
    label-width="90px"
  >
    <el-form-item label="岗位名称" prop="name">
      <el-input
        v-model="newFormInline.name"
        clearable
        placeholder="请输入岗位名称"
      />
    </el-form-item>
    <el-form-item label="岗位备注" prop="remark">
      <el-input
        v-model="newFormInline.remark"
        type="textarea"
        clearable
        placeholder="请输入备注"
      />
    </el-form-item>
    <el-form-item label="状态" prop="status">
      <el-radio-group v-model="newFormInline.status">
        <el-radio :label="1">启用</el-radio>
        <el-radio :label="0">停用</el-radio>
      </el-radio-group>
    </el-form-item>
  </el-form>
</template>
