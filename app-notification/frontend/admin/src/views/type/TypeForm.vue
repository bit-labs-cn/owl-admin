<script setup lang="ts">
import { ref } from "vue";
import { channelOptions, kindOptions, typeStatusOptions } from "../../labels";
import type { TypeFormData } from "./types";

const props = defineProps<{ formInline: TypeFormData }>();
const ruleFormRef = ref();
const newFormInline = ref<TypeFormData>({ ...props.formInline, defaultChannels: [...props.formInline.defaultChannels] });

const rules = {
  code: [{ required: true, message: "请输入类型编码", trigger: "blur" }],
  name: [{ required: true, message: "请输入类型名称", trigger: "blur" }]
};

function getRef() {
  return ruleFormRef.value;
}

function getFormData() {
  return { ...newFormInline.value };
}

defineExpose({ getRef, getFormData });
</script>

<template>
  <el-form ref="ruleFormRef" :model="newFormInline" :rules="rules" label-width="96px">
    <el-form-item label="类型编码" prop="code">
      <el-input v-model="newFormInline.code" :disabled="!!newFormInline.id" clearable placeholder="例如 approval.todo" />
    </el-form-item>
    <el-form-item label="类型名称" prop="name">
      <el-input v-model="newFormInline.name" clearable placeholder="请输入类型名称" />
    </el-form-item>
    <el-form-item label="说明" prop="description">
      <el-input v-model="newFormInline.description" type="textarea" :rows="2" />
    </el-form-item>
    <el-form-item label="种类" prop="kind">
      <el-select v-model="newFormInline.kind" class="w-full">
        <el-option v-for="item in kindOptions" :key="item.value" :label="item.label" :value="item.value" />
      </el-select>
    </el-form-item>
    <el-form-item label="默认渠道" prop="defaultChannels">
      <el-checkbox-group v-model="newFormInline.defaultChannels">
        <el-checkbox v-for="item in channelOptions" :key="item.value" :value="item.value">{{ item.label }}</el-checkbox>
      </el-checkbox-group>
    </el-form-item>
    <el-form-item label="状态" prop="status">
      <el-select v-model="newFormInline.status" class="w-full">
        <el-option v-for="item in typeStatusOptions" :key="item.value" :label="item.label" :value="item.value" />
      </el-select>
    </el-form-item>
  </el-form>
</template>
