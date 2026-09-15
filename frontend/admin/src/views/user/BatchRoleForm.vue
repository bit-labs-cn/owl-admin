<script setup lang="ts">
import { ref } from "vue";

export interface BatchRoleFormData {
  selectedCount: number;
  roleOptions: Array<{ id: string | number; name: string }>;
  ids: Array<string | number>;
}

const props = withDefaults(defineProps<{ formInline: BatchRoleFormData }>(), {
  formInline: () => ({
    selectedCount: 0,
    roleOptions: [],
    ids: []
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
    <el-form-item label="角色列表" prop="ids">
      <el-select
        v-model="newFormInline.ids"
        placeholder="请选择角色（可清空表示移除全部角色）"
        class="w-full"
        clearable
        multiple
        filterable
      >
        <el-option
          v-for="item in newFormInline.roleOptions"
          :key="item.id"
          :value="item.id"
          :label="item.name"
        />
      </el-select>
    </el-form-item>
    <el-alert
      type="info"
      :closable="false"
      show-icon
      title="将用上方勾选的角色覆盖所选用户的原有角色。"
    />
  </el-form>
</template>
