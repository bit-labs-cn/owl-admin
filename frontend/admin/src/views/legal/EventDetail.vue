<script setup lang="ts">
import { computed } from "vue";
import dayjs from "dayjs";
import type { LegalEventRow } from "./types";

const props = defineProps<{
  event: LegalEventRow;
}>();

const createdAt = computed(() =>
  props.event.createdAt ? dayjs(props.event.createdAt).format("YYYY-MM-DD HH:mm:ss") : "—"
);
</script>

<template>
  <div>
    <el-alert
      v-if="!event.userPresent"
      title="该用户已不在系统中，下面是同意当时留下的名称。"
      type="info"
      :closable="false"
      class="mb-3"
    />
    <el-descriptions :column="2" border>
      <el-descriptions-item label="昵称">{{ event.nickname || "—" }}</el-descriptions-item>
      <el-descriptions-item label="账号">{{ event.username || "—" }}</el-descriptions-item>
      <el-descriptions-item label="条款">{{ event.documentName }}</el-descriptions-item>
      <el-descriptions-item label="版本">{{ event.version }}</el-descriptions-item>
      <el-descriptions-item label="动作">{{ event.actionLabel }}</el-descriptions-item>
      <el-descriptions-item label="场景">{{ event.sceneName }}</el-descriptions-item>
      <el-descriptions-item label="时间" :span="2">{{ createdAt }}</el-descriptions-item>
      <el-descriptions-item label="IP">{{ event.ip || "—" }}</el-descriptions-item>
      <el-descriptions-item label="设备">{{ event.userAgent || "—" }}</el-descriptions-item>
    </el-descriptions>
  </div>
</template>
