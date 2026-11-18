<script setup lang="ts">
import { ref } from "vue";
import {
  announcementStatusOptions,
  displayModeOptions,
  levelOptions,
  scopeOptions
} from "../../labels";
import type { AnnouncementFormData } from "./types";

const props = defineProps<{ formInline: AnnouncementFormData }>();
const ruleFormRef = ref();
const newFormInline = ref<AnnouncementFormData>({ ...props.formInline });

const rules = {
  title: [{ required: true, message: "请输入公告标题", trigger: "blur" }],
  content: [{ required: true, message: "请输入公告正文", trigger: "blur" }]
};

function getRef() {
  return ruleFormRef.value;
}

function getFormData() {
  const form = newFormInline.value;
  return {
    ...form,
    startAt: form.startAt || undefined,
    endAt: form.endAt || undefined
  };
}

defineExpose({ getRef, getFormData });
</script>

<template>
  <el-form ref="ruleFormRef" :model="newFormInline" :rules="rules" label-width="96px">
    <el-form-item label="公告标题" prop="title">
      <el-input v-model="newFormInline.title" clearable placeholder="请输入公告标题" />
    </el-form-item>

    <el-row :gutter="16">
      <el-col :span="12">
        <el-form-item label="级别" prop="level">
          <el-select v-model="newFormInline.level" class="w-full">
            <el-option v-for="item in levelOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="展示方式" prop="displayMode">
          <el-select v-model="newFormInline.displayMode" class="w-full">
            <el-option v-for="item in displayModeOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="受众范围" prop="targetScope">
          <el-select v-model="newFormInline.targetScope" class="w-full">
            <el-option v-for="item in scopeOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="状态" prop="status">
          <el-select v-model="newFormInline.status" class="w-full">
            <el-option
              v-for="item in announcementStatusOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="开始时间" prop="startAt">
          <el-date-picker
            v-model="newFormInline.startAt"
            type="datetime"
            value-format="YYYY-MM-DD HH:mm:ss"
            placeholder="立即生效或选择时间"
            class="w-full"
          />
        </el-form-item>
      </el-col>
      <el-col :span="12">
        <el-form-item label="结束时间" prop="endAt">
          <el-date-picker
            v-model="newFormInline.endAt"
            type="datetime"
            value-format="YYYY-MM-DD HH:mm:ss"
            placeholder="永久有效或选择时间"
            class="w-full"
          />
        </el-form-item>
      </el-col>
    </el-row>

    <el-form-item label="公告正文" prop="content">
      <el-input
        v-model="newFormInline.content"
        type="textarea"
        :rows="6"
        placeholder="支持 Markdown 格式或纯文本内容..."
      />
    </el-form-item>
  </el-form>
</template>

<style scoped>
.w-full {
  width: 100%;
}
</style>
