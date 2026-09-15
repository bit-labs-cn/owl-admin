<script setup lang="ts">
import { ref } from "vue";

export interface BatchDeptFormData {
  selectedCount: number;
  higherDeptOptions: Record<string, unknown>[];
  parentId: string | number | "";
}

const props = withDefaults(defineProps<{ formInline: BatchDeptFormData }>(), {
  formInline: () => ({
    selectedCount: 0,
    higherDeptOptions: [],
    parentId: ""
  })
});

const formRef = ref();
const newFormInline = ref(props.formInline);

function getRef() {
  return formRef.value;
}

function getFormData() {
  return newFormInline.value;
}

defineExpose({ getRef, getFormData });
</script>

<template>
  <el-form ref="formRef" :model="newFormInline" label-width="88px">
    <el-form-item label="已选用户">
      <el-input :model-value="`${newFormInline.selectedCount} 人`" disabled />
    </el-form-item>
    <el-form-item label="归属部门" prop="parentId">
      <el-cascader
        v-model="newFormInline.parentId"
        class="w-full"
        :options="newFormInline.higherDeptOptions"
        :props="{
          value: 'id',
          label: 'name',
          emitPath: false,
          checkStrictly: true
        }"
        clearable
        filterable
        placeholder="请选择部门（清空表示移除归属部门）"
      >
        <template #default="{ node, data }">
          <span>{{ data.name }}</span>
          <span v-if="!node.isLeaf"> ({{ data.children.length }}) </span>
        </template>
      </el-cascader>
    </el-form-item>
    <el-alert
      type="info"
      :closable="false"
      show-icon
      title="将用所选部门覆盖所选用户的原有归属部门。"
    />
  </el-form>
</template>
