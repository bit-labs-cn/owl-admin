<script setup lang="ts">
defineProps<{
  title: string;
  content: string;
}>();
</script>

<template>
  <div>
    <h3 class="title">{{ title }}</h3>
    <pre class="body">{{ content }}</pre>
  </div>
</template>

<style scoped>
.title {
  margin: 0 0 12px;
  font-size: 16px;
  font-weight: 600;
}

.body {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
  font-family: inherit;
  line-height: 1.6;
}
</style>
