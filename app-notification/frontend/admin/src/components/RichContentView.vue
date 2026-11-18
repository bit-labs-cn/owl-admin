<script setup lang="ts">
import { computed } from "vue";
import { isRichContent } from "../format";

const props = withDefaults(
  defineProps<{
    content?: string | null;
    /** 正文为空时的占位文案。 */
    emptyText?: string;
    /** 纯文本是否保留换行；富文本本来就按 HTML 排版。 */
    keepLineBreak?: boolean;
  }>(),
  {
    emptyText: "无正文",
    keepLineBreak: true
  }
);

const isRich = computed(() => isRichContent(props.content));
</script>

<template>
  <div v-if="isRich" class="rich-content" v-html="content" />
  <div v-else class="rich-content is-plain" :class="{ 'keep-break': keepLineBreak }">
    {{ content || emptyText }}
  </div>
</template>

<style lang="scss" scoped>
.rich-content {
  color: var(--el-text-color-primary);
  font-size: 13px;
  line-height: 1.7;
  word-break: break-word;

  &.is-plain.keep-break {
    white-space: pre-wrap;
  }

  :deep(p) {
    margin: 0 0 8px;
  }

  :deep(p:last-child) {
    margin-bottom: 0;
  }

  :deep(h1),
  :deep(h2),
  :deep(h3),
  :deep(h4) {
    margin: 12px 0 8px;
    font-weight: 600;
    line-height: 1.4;
  }

  :deep(h1) {
    font-size: 18px;
  }

  :deep(h2) {
    font-size: 16px;
  }

  :deep(h3) {
    font-size: 15px;
  }

  :deep(h4) {
    font-size: 14px;
  }

  :deep(ul),
  :deep(ol) {
    margin: 6px 0;
    padding-left: 22px;
  }

  :deep(ul) {
    list-style: disc;
  }

  :deep(ol) {
    list-style: decimal;
  }

  :deep(li) {
    margin: 2px 0;
  }

  :deep(blockquote) {
    margin: 8px 0;
    padding: 6px 12px;
    border-left: 3px solid var(--el-color-primary-light-5);
    background: var(--el-fill-color-light);
    color: var(--el-text-color-regular);
  }

  :deep(a) {
    color: var(--el-color-primary);
    text-decoration: underline;
  }

  :deep(img) {
    max-width: 100%;
    height: auto;
    border-radius: 6px;
  }

  :deep(hr) {
    margin: 12px 0;
    border: 0;
    border-top: 1px solid var(--el-border-color-lighter);
  }

  :deep(table) {
    width: 100%;
    margin: 8px 0;
    border-collapse: collapse;
    font-size: 12px;
  }

  :deep(th),
  :deep(td) {
    padding: 6px 8px;
    border: 1px solid var(--el-border-color-lighter);
    text-align: left;
  }

  :deep(th) {
    background: var(--el-fill-color-light);
    font-weight: 600;
  }

  :deep(code) {
    padding: 1px 4px;
    border-radius: 4px;
    background: var(--el-fill-color-light);
    font-size: 12px;
  }
}
</style>
