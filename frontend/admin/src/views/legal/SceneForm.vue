<script setup lang="ts">
import { ref, onMounted } from "vue";
import { legalAPI } from "@bit-labs.cn/owl-admin-ui/api/legal";
import type { LegalOption, LegalSceneForm, LegalSceneItemForm } from "./types";

const props = defineProps<{
  formInline: LegalSceneForm;
}>();

const ruleFormRef = ref();
const documents = ref<LegalOption[]>([]);
const newFormInline = ref<LegalSceneForm>({
  id: props.formInline?.id,
  key: props.formInline?.key ?? "",
  name: props.formInline?.name ?? "",
  kind: props.formInline?.kind ?? "consent",
  items: (props.formInline?.items ?? []).map((item, index) => ({
    documentKey: item.documentKey,
    required: item.required,
    groupKey: item.groupKey,
    wording: item.wording,
    sort: item.sort || index + 1
  }))
});

const rules = {
  key: [{ required: true, message: "请输入标识", trigger: "blur" }],
  name: [{ required: true, message: "请输入名称", trigger: "blur" }],
  kind: [{ required: true, message: "请选择用途", trigger: "change" }]
};

function blankItem(): LegalSceneItemForm {
  return {
    documentKey: "",
    required: newFormInline.value.kind === "consent",
    groupKey: "",
    wording: "",
    sort: newFormInline.value.items.length + 1
  };
}

function addItem() {
  newFormInline.value.items.push(blankItem());
}

function removeItem(index: number) {
  newFormInline.value.items.splice(index, 1);
  newFormInline.value.items.forEach((item, i) => {
    item.sort = i + 1;
  });
}

function getRef() {
  return ruleFormRef.value;
}

function getFormData() {
  const data = newFormInline.value;
  const items = data.items.map((item, index) => {
    if (data.kind === "reader") {
      return { documentKey: item.documentKey, required: false, groupKey: "", wording: "", sort: index + 1 };
    }
    return { ...item, sort: index + 1 };
  });
  return { ...data, items };
}

onMounted(async () => {
  const { data } = await legalAPI.getDocumentOptions();
  documents.value = Array.isArray(data) ? data : [];
});

defineExpose({ getRef, getFormData });
</script>

<template>
  <el-form ref="ruleFormRef" :model="newFormInline" :rules="rules" label-width="72px">
    <el-form-item label="标识" prop="key">
      <el-input v-model="newFormInline.key" :disabled="!!newFormInline.id" clearable placeholder="例如 register、site" />
    </el-form-item>
    <el-form-item label="名称" prop="name">
      <el-input v-model="newFormInline.name" clearable placeholder="给人看的名称" />
    </el-form-item>
    <el-form-item label="用途" prop="kind">
      <el-radio-group v-model="newFormInline.kind">
        <el-radio value="consent">同意，用来勾选</el-radio>
        <el-radio value="reader">阅读，用来排顶栏</el-radio>
      </el-radio-group>
    </el-form-item>
    <el-form-item label="条款">
      <div class="items">
        <div v-for="(item, index) in newFormInline.items" :key="index" class="item">
          <el-select v-model="item.documentKey" placeholder="选择条款" class="doc">
            <el-option
              v-for="doc in documents.filter(item => item.status === 'enabled')"
              :key="doc.key"
              :label="doc.name"
              :value="doc.key"
            />
          </el-select>
          <template v-if="newFormInline.kind === 'consent'">
            <el-input v-model="item.groupKey" placeholder="勾选组，同组一个框" class="group" />
            <el-input v-model="item.wording" placeholder="勾选文案" class="wording" />
            <el-checkbox v-model="item.required">必选</el-checkbox>
          </template>
          <el-button link type="danger" @click="removeItem(index)">移除</el-button>
        </div>
        <el-button type="primary" link @click="addItem">添加条款</el-button>
      </div>
    </el-form-item>
  </el-form>
</template>

<style scoped>
.items {
  width: 100%;
}

.item {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
  margin-bottom: 8px;
}

.doc {
  width: 180px;
}

.group {
  width: 160px;
}

.wording {
  width: 180px;
}
</style>
