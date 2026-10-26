<script setup lang="ts">
import { computed } from "vue";

const props = defineProps<{
  log: Record<string, any>;
}>();

const userTypeLabel = computed(() => {
  if (props.log.userType === "super_admin") return "超管";
  if (props.log.userType === "user") return "普通用户";
  return props.log.userType || "—";
});
</script>

<template>
  <el-descriptions :column="2" border>
    <el-descriptions-item label="用户名">
      {{ log.userName || "—" }}
    </el-descriptions-item>
    <el-descriptions-item label="用户类型">
      {{ userTypeLabel }}
    </el-descriptions-item>
    <el-descriptions-item label="登录 IP">
      {{ log.ip || "—" }}
    </el-descriptions-item>
    <el-descriptions-item label="地点">
      {{ log.location || "—" }}
    </el-descriptions-item>
    <el-descriptions-item label="登录时间" :span="2">
      {{ log.loginTime || "—" }}
    </el-descriptions-item>
    <el-descriptions-item label="设备 UA" :span="2">
      <span class="wrap">{{ log.userAgent || "—" }}</span>
    </el-descriptions-item>
  </el-descriptions>
</template>

<style scoped>
.wrap {
  white-space: pre-wrap;
  word-break: break-all;
}
</style>
