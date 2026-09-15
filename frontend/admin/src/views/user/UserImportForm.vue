<template>
  <div class="user-import">
    <el-alert
      type="info"
      :closable="false"
      show-icon
      class="import-tip"
      title="请先下载模板，按「用户账号 / 用户昵称 / 初始密码」等列填写后上传。归属部门填完整路径（如 总公司/研发部），角色填角色名称（多个用逗号分隔），均可留空表示不绑定。仅支持 .xlsx，单次最多 2000 行。"
    />

    <div class="import-actions">
      <el-button :loading="downloading" @click="downloadTemplate">
        下载导入模板
      </el-button>
      <el-upload
        :show-file-list="false"
        :auto-upload="false"
        accept=".xlsx"
        :disabled="uploading"
        @change="onFileChange"
      >
        <el-button type="primary" :loading="uploading">选择 Excel 并导入</el-button>
      </el-upload>
    </div>

    <div v-if="result" class="import-result">
      <div class="import-result-header">
        <p>
          合计 {{ result.total }} 行，成功
          <span class="ok">{{ result.created }}</span>
          ，失败
          <span class="bad">{{ result.failed }}</span>
        </p>
        <el-button
          v-if="result.errors?.length"
          type="warning"
          plain
          size="small"
          :loading="exporting"
          @click="exportErrors"
        >
          导出错误
        </el-button>
      </div>
      <el-table
        v-if="result.errors?.length"
        :data="result.errors.slice(0, 50)"
        size="small"
        border
        max-height="260"
      >
        <el-table-column prop="row" label="行号" width="72" />
        <el-table-column prop="username" label="账号" min-width="120" show-overflow-tooltip />
        <el-table-column prop="reason" label="原因" min-width="200" show-overflow-tooltip />
      </el-table>
      <p v-if="(result.errors?.length || 0) > 50" class="more-hint">
        仅展示前 50 条失败原因
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import type { UploadFile } from "element-plus";
import { message } from "@bit-labs.cn/owl-ui/utils/message";
import {
  userManageAPI,
  type ImportUserResult
} from "@bit-labs.cn/owl-admin-ui/api/user";

const emit = defineEmits<{
  (e: "success"): void;
}>();

const downloading = ref(false);
const uploading = ref(false);
const exporting = ref(false);
const result = ref<ImportUserResult | null>(null);

async function downloadTemplate() {
  downloading.value = true;
  try {
    const blob = await userManageAPI.importTemplate();
    if (!(blob instanceof Blob)) {
      message("下载模板失败", { type: "error" });
      return;
    }
    const url = window.URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = "user-import-template.xlsx";
    link.click();
    window.URL.revokeObjectURL(url);
  } catch {
    message("下载模板失败", { type: "error" });
  } finally {
    downloading.value = false;
  }
}

async function onFileChange(uploadFile: UploadFile) {
  const file = uploadFile.raw;
  if (!(file instanceof File)) return;
  const ext = file.name.toLowerCase().split(".").pop();
  if (ext !== "xlsx") {
    message("请上传 Excel 文件（.xlsx）", { type: "warning" });
    return;
  }
  if (file.size > 5 * 1024 * 1024) {
    message("导入文件不能超过 5MB", { type: "warning" });
    return;
  }

  uploading.value = true;
  result.value = null;
  try {
    const res = await userManageAPI.importUsers(file);
    const data = (res?.data as ImportUserResult) ?? {
      total: 0,
      created: 0,
      failed: 0,
      errors: []
    };
    result.value = data;
    if (data.created > 0) {
      emit("success");
    }
  } catch {
    /* http 拦截器已提示 */
  } finally {
    uploading.value = false;
  }
}

async function exportErrors() {
  const errors = result.value?.errors ?? [];
  if (!errors.length) {
    message("没有可导出的错误记录", { type: "warning" });
    return;
  }
  exporting.value = true;
  try {
    const blob = await userManageAPI.exportImportErrors(errors);
    if (!(blob instanceof Blob)) {
      message("导出失败", { type: "error" });
      return;
    }
    // 后端业务失败时也可能以 JSON Blob 返回，做一次探测
    if (blob.type.includes("application/json")) {
      const text = await blob.text();
      try {
        const parsed = JSON.parse(text);
        message(parsed?.msg || "导出失败", { type: "error" });
      } catch {
        message("导出失败", { type: "error" });
      }
      return;
    }
    const url = window.URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = "user-import-errors.xlsx";
    link.click();
    window.URL.revokeObjectURL(url);
  } catch {
    message("导出失败", { type: "error" });
  } finally {
    exporting.value = false;
  }
}
</script>

<style scoped>
.import-tip {
  margin-bottom: 16px;
}
.import-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  align-items: center;
}
.import-result {
  margin-top: 16px;
  font-size: 13px;
}
.import-result-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 8px;
}
.import-result-header p {
  margin: 0;
}
.ok {
  color: var(--el-color-success);
  font-weight: 600;
}
.bad {
  color: var(--el-color-danger);
  font-weight: 600;
}
.more-hint {
  margin: 8px 0 0;
  color: var(--el-text-color-secondary);
  font-size: 12px;
}
</style>
