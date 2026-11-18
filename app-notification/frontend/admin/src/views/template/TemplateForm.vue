<script setup lang="ts">
import { ref } from "vue";
import { message } from "@bit-labs.cn/owl-ui/utils/message";
import { templateAPI } from "../../api/notify";
import { NOTIFY_PERMS } from "../../permissions";
import RichTextEditor from "../../components/RichTextEditor.vue";
import RichContentView from "../../components/RichContentView.vue";
import type { TemplateFormData, TypeOption } from "./types";

const props = defineProps<{ formInline: TemplateFormData; typeOptions: TypeOption[] }>();
const ruleFormRef = ref();
const newFormInline = ref<TemplateFormData>({ ...props.formInline });
const preview = ref<{ title: string; body: string; link: string }>();

/** 模板变量示例，写在脚本里避免和 Vue 的插值语法打架。 */
const varHint = "{{.Name}}";

const rules = {
  typeCode: [{ required: true, message: "请选择消息类型", trigger: "change" }]
};

function getRef() {
  return ruleFormRef.value;
}

function getFormData() {
  return { ...newFormInline.value };
}

async function runPreview() {
  let vars: Record<string, unknown> = {};
  try {
    vars = newFormInline.value.sample ? JSON.parse(newFormInline.value.sample) : {};
  } catch {
    message("示例变量不是合法 JSON", { type: "warning" });
    return;
  }
  const res = await templateAPI.preview({
    title: newFormInline.value.title,
    body: newFormInline.value.body,
    link: newFormInline.value.link,
    vars
  });
  preview.value = (res?.data ?? {}) as { title: string; body: string; link: string };
}

defineExpose({ getRef, getFormData });
</script>

<template>
  <el-form ref="ruleFormRef" :model="newFormInline" :rules="rules" label-width="96px">
    <el-form-item label="消息类型" prop="typeCode">
      <el-select v-model="newFormInline.typeCode" :disabled="!!newFormInline.id" class="w-full" filterable>
        <el-option v-for="item in typeOptions" :key="item.code" :label="item.name" :value="item.code" />
      </el-select>
      <p class="form-note">一个类型只有一条模板；发往哪些渠道由发送时决定，模板不用选渠道。</p>
    </el-form-item>
    <el-form-item label="标题模板" prop="title">
      <el-input v-model="newFormInline.title" clearable placeholder="例如 单号 {{.OrderNo}}" />
      <p class="form-note">变量写成 {{ varHint }}；标题在所有渠道共用一个。</p>
    </el-form-item>
    <el-form-item label="正文模板" prop="body">
      <RichTextEditor v-model="newFormInline.body" :height="200" placeholder="请输入正文模板，支持富文本" />
      <p class="form-note">
        正文是富文本：站内信和邮件按 HTML 显示，企业微信、短信自动转纯文本。
      </p>
    </el-form-item>
    <el-form-item label="链接模板" prop="link">
      <el-input v-model="newFormInline.link" clearable placeholder="/order/{{.OrderNo}}" />
    </el-form-item>
    <el-form-item label="示例变量" prop="sample">
      <el-input v-model="newFormInline.sample" type="textarea" :rows="3" />
      <p class="form-note">填一份 JSON 示例，用「预览」检查变量替换后的效果。</p>
    </el-form-item>
    <el-form-item>
      <Perms :value="NOTIFY_PERMS.template.preview">
        <el-button @click="runPreview">预览</el-button>
      </Perms>
    </el-form-item>
    <el-form-item v-if="preview" label="预览结果">
      <div class="preview">
        <div>标题：{{ preview.title || "（空）" }}</div>
        <RichContentView class="preview-body" :content="preview.body" empty-text="（空）" :keep-line-break="false" />
        <div>链接：{{ preview.link || "（空）" }}</div>
      </div>
    </el-form-item>
  </el-form>
</template>

<style scoped>
.form-note {
  margin: 6px 0 0;
  color: var(--el-text-color-secondary);
  font-size: 12px;
  line-height: 1.5;
}

.preview {
  width: 100%;
  line-height: 1.6;
}

.preview-body {
  margin: 6px 0;
  padding: 10px 12px;
  border-radius: 6px;
  background: var(--el-fill-color-light);
}
</style>
