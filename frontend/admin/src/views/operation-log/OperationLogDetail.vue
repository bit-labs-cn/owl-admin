<script setup lang="ts">
import { computed } from "vue";
import { formatOperateTime, statusTagType } from "./columns";

const props = defineProps<{
  log: Record<string, any>;
}>();

const userTypeLabel = computed(() => {
  if (props.log.userType === "super_admin") return "超级管理员";
  if (props.log.userType === "user") return "用户";
  return props.log.userType || "—";
});

const bodyText = computed(() => {
  const raw = props.log.reqBody;
  if (!raw) return "";
  try {
    return JSON.stringify(JSON.parse(raw), null, 2);
  } catch {
    return String(raw);
  }
});
</script>

<template>
  <el-descriptions :column="2" border>
    <el-descriptions-item label="操作人">
      {{ log.userName || "—" }}
    </el-descriptions-item>
    <el-descriptions-item label="用户类型">
      {{ userTypeLabel }}
    </el-descriptions-item>
    <el-descriptions-item label="请求方式">
      {{ log.method || "—" }}
    </el-descriptions-item>
    <el-descriptions-item label="状态">
      <el-tag :type="statusTagType(Number(log.status))">{{ log.status }}</el-tag>
    </el-descriptions-item>
    <el-descriptions-item label="接口名称">
      {{ log.apiName || "—" }}
    </el-descriptions-item>
    <el-descriptions-item label="耗时">
      {{ log.costMs ?? "—" }} ms
    </el-descriptions-item>
    <el-descriptions-item label="接口地址" :span="2">
      {{ log.path || "—" }}
    </el-descriptions-item>
    <el-descriptions-item label="IP 地址">
      {{ log.ip || "—" }}
    </el-descriptions-item>
    <el-descriptions-item label="操作时间">
      {{ formatOperateTime(log.createdAt) }}
    </el-descriptions-item>
    <el-descriptions-item label="User-Agent" :span="2">
      <span class="wrap">{{ log.userAgent || "—" }}</span>
    </el-descriptions-item>
    <el-descriptions-item label="请求参数" :span="2">
      <pre v-if="bodyText" class="req-body">{{ bodyText }}</pre>
      <span v-else>—</span>
    </el-descriptions-item>
  </el-descriptions>
</template>

<style scoped>
.wrap,
.req-body {
  white-space: pre-wrap;
  word-break: break-all;
}

.req-body {
  margin: 0;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 12px;
  line-height: 1.5;
}
</style>
