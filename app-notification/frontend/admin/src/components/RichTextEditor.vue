<script setup lang="ts">
import "@wangeditor/editor/dist/css/style.css";
import { onBeforeUnmount, shallowRef } from "vue";
import { message } from "@bit-labs.cn/owl-ui/utils/message";
import { Editor, Toolbar } from "@wangeditor/editor-for-vue";
import type { IDomEditor, IEditorConfig, IToolbarConfig } from "@wangeditor/editor";
import { uploadNotifyImage } from "../api/notify";

const props = withDefaults(
  defineProps<{
    modelValue: string;
    placeholder?: string;
    height?: number;
  }>(),
  {
    placeholder: "请输入正文，可以加粗、分点、插图",
    height: 240
  }
);

const emit = defineEmits<{
  (e: "update:modelValue", value: string): void;
}>();

const editorRef = shallowRef<IDomEditor>();

/** 只留能安全投递到站内信和邮件的菜单。 */
const toolbarConfig: Partial<IToolbarConfig> = {
  toolbarKeys: [
    "headerSelect",
    "bold",
    "italic",
    "underline",
    "through",
    "|",
    "color",
    "|",
    "bulletedList",
    "numberedList",
    "blockquote",
    "|",
    "insertLink",
    "uploadImage",
    "|",
    "undo",
    "redo",
    "clearStyle"
  ]
};

function insertImage(file: File, insertFn: (url: string, alt: string, href: string) => void) {
  uploadNotifyImage(file)
    .then(url => insertFn(url, file.name || "图片", url))
    .catch((error: unknown) => {
      message(error instanceof Error ? error.message : "图片上传失败", { type: "error" });
    });
}

const editorConfig: Partial<IEditorConfig> = {
  placeholder: props.placeholder,
  MENU_CONF: {
    uploadImage: { customUpload: insertImage }
  }
};

function handleCreated(editor: IDomEditor) {
  editorRef.value = editor;
}

onBeforeUnmount(() => {
  editorRef.value?.destroy();
  editorRef.value = undefined;
});
</script>

<template>
  <div class="rich-editor">
    <Toolbar class="rich-toolbar" :editor="editorRef" :default-config="toolbarConfig" mode="default" />
    <Editor
      :model-value="modelValue"
      class="rich-body"
      :style="{ height: `${height}px` }"
      :default-config="editorConfig"
      mode="default"
      @on-created="handleCreated"
      @update:model-value="emit('update:modelValue', $event)"
    />
  </div>
</template>

<style lang="scss" scoped>
.rich-editor {
  width: 100%;
  border: 1px solid var(--el-border-color);
  border-radius: 8px;
  overflow: hidden;
  transition: border-color 0.15s ease;

  &:focus-within {
    border-color: var(--el-color-primary);
    box-shadow: 0 0 0 3px var(--el-color-primary-light-9);
  }
}

.rich-toolbar {
  border-bottom: 1px solid var(--el-border-color-lighter);
  background: var(--el-fill-color-lighter);
}

.rich-body {
  overflow-y: auto;
  background: var(--el-bg-color);
}

.rich-editor :deep(.w-e-text-container) {
  background: var(--el-bg-color);
  color: var(--el-text-color-primary);
}

.rich-editor :deep(.w-e-text-placeholder) {
  color: var(--el-text-color-placeholder);
}

.rich-editor :deep(.w-e-scroll) {
  padding: 10px 12px;
  font-size: 14px;
  line-height: 1.7;
}

// 弹层挂在 body 上，跟着主题走
:global(.w-e-menu-tooltip-v5) {
  z-index: 3000;
}
</style>
