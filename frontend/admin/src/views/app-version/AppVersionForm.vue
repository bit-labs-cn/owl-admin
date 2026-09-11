<script setup lang="ts">
import { computed, ref, watch } from "vue";
import type { UploadProps, UploadUserFile } from "element-plus";
import { ElMessage } from "element-plus";
import type { AppVersionFormData } from "./types";
import { appVersionAPI } from "@bit-labs.cn/owl-admin-ui/api/app-version";

const props = defineProps<{
  formInline: AppVersionFormData;
}>();

const ruleFormRef = ref();
const newFormInline = ref<AppVersionFormData>({
  id: props.formInline?.id ?? undefined,
  version: props.formInline?.version ?? "",
  versionName: props.formInline?.versionName ?? "",
  apkUrl: props.formInline?.apkUrl ?? "",
  apkType: props.formInline?.apkType ?? 2,
  content: props.formInline?.content ?? "",
  remark: props.formInline?.remark ?? "",
  status: props.formInline?.status ?? 1
});

const fileList = ref<UploadUserFile[]>([]);
const maxBytes = 500 * 1024 * 1024;

const accept = computed(() =>
  newFormInline.value.apkType === 1 ? ".ipa" : ".apk,.aab"
);

function fileNameFromUrl(url: string) {
  try {
    const path = url.split("?")[0];
    return decodeURIComponent(path.substring(path.lastIndexOf("/") + 1) || "安装包");
  } catch {
    return "安装包";
  }
}

function syncFileList(url: string) {
  fileList.value = url
    ? [{ name: fileNameFromUrl(url), url, status: "success" }]
    : [];
}

watch(
  () => newFormInline.value.apkUrl,
  url => syncFileList(url),
  { immediate: true }
);

watch(
  () => newFormInline.value.apkType,
  () => {
    const url = newFormInline.value.apkUrl;
    if (!url) return;
    const ext = fileNameFromUrl(url).toLowerCase();
    const iosOk = ext.endsWith(".ipa");
    const androidOk = ext.endsWith(".apk") || ext.endsWith(".aab");
    if (newFormInline.value.apkType === 1 && !iosOk) {
      newFormInline.value.apkUrl = "";
    }
    if (newFormInline.value.apkType === 2 && !androidOk) {
      newFormInline.value.apkUrl = "";
    }
  }
);

const rules = {
  version: [
    { required: true, message: "请输入版本号", trigger: "blur" },
    { pattern: /^[1-9]\d*$/, message: "版本号为递增整数，如 100", trigger: "blur" }
  ],
  versionName: [
    { required: true, message: "请输入版本名称", trigger: "blur" },
    { pattern: /^\d+\.\d+\.\d+$/, message: "版本名称格式为 x.y.z，如 1.0.0", trigger: "blur" }
  ],
  apkType: [{ required: true, message: "请选择安装包类型", trigger: "change" }],
  apkUrl: [{ required: true, message: "请上传安装包", trigger: "change" }]
};

const beforeUpload: UploadProps["beforeUpload"] = rawFile => {
  const ext = rawFile.name.slice(rawFile.name.lastIndexOf(".")).toLowerCase();
  const allowed =
    newFormInline.value.apkType === 1 ? [".ipa"] : [".apk", ".aab"];
  if (!allowed.includes(ext)) {
    ElMessage.warning(
      newFormInline.value.apkType === 1
        ? "iOS 请上传 .ipa 文件"
        : "Android 请上传 .apk 或 .aab 文件"
    );
    return false;
  }
  if (rawFile.size > maxBytes) {
    ElMessage.warning("安装包不能超过 500MB");
    return false;
  }
  return true;
};

const submitPackage: UploadProps["httpRequest"] = async options => {
  const raw = options.file as File & { raw?: File };
  const file = raw?.raw instanceof File ? raw.raw : raw;
  try {
    const result = await appVersionAPI.uploadPackage(file, newFormInline.value.apkType);
    newFormInline.value.apkUrl = result.url;
    fileList.value = [{ name: result.originalName || file.name, url: result.url, status: "success" }];
    ruleFormRef.value?.validateField?.("apkUrl");
    options.onSuccess?.(result);
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : "安装包上传失败");
    options.onError?.(e as Error);
  }
};

function onRemove() {
  newFormInline.value.apkUrl = "";
  fileList.value = [];
}

function getRef() {
  return ruleFormRef.value;
}

defineExpose({ getRef, getFormData: () => newFormInline.value });
</script>

<template>
  <el-form
    ref="ruleFormRef"
    :model="newFormInline"
    :rules="rules"
    label-width="96px"
  >
    <el-form-item label="版本号" prop="version">
      <el-input
        v-model="newFormInline.version"
        clearable
        placeholder="如 100"
      />
    </el-form-item>
    <el-form-item label="版本名称" prop="versionName">
      <el-input
        v-model="newFormInline.versionName"
        clearable
        placeholder="如 1.0.0"
      />
    </el-form-item>
    <el-form-item label="安装包类型" prop="apkType">
      <el-radio-group v-model="newFormInline.apkType">
        <el-radio :value="1">iOS</el-radio>
        <el-radio :value="2">Android</el-radio>
      </el-radio-group>
    </el-form-item>
    <el-form-item label="安装包" prop="apkUrl">
      <el-upload
        v-model:file-list="fileList"
        action=""
        drag
        :limit="1"
        :accept="accept"
        :http-request="submitPackage"
        :before-upload="beforeUpload"
        :on-remove="onRemove"
      >
        <div class="el-upload__text">将安装包拖到此处，或<em>点击上传</em></div>
        <template #tip>
          <div class="el-upload__tip">
            {{
              newFormInline.apkType === 1
                ? "iOS 仅支持 .ipa，不超过 500MB"
                : "Android 支持 .apk / .aab，不超过 500MB"
            }}
          </div>
        </template>
      </el-upload>
    </el-form-item>
    <el-form-item label="更新内容" prop="content">
      <el-input
        v-model="newFormInline.content"
        type="textarea"
        :rows="4"
        clearable
        placeholder="请输入更新内容"
      />
    </el-form-item>
    <el-form-item label="备注" prop="remark">
      <el-input
        v-model="newFormInline.remark"
        type="textarea"
        clearable
        placeholder="请输入备注"
      />
    </el-form-item>
    <el-form-item label="状态" prop="status">
      <el-radio-group v-model="newFormInline.status">
        <el-radio :value="1">启用</el-radio>
        <el-radio :value="0">停用</el-radio>
      </el-radio-group>
    </el-form-item>
  </el-form>
</template>
