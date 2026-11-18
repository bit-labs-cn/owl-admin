<script setup lang="ts">
import { h, ref } from "vue";
import { message } from "@bit-labs.cn/owl-ui/utils/message";
import { dialogService } from "@bit-labs.cn/owl-ui/components/ReDialog";
import { PureTableBar } from "@bit-labs.cn/owl-ui/components/RePureTableBar";
import { useRenderIcon } from "@bit-labs.cn/owl-ui/components/ReIcon/src/hooks";
import Refresh from "@iconify-icons/ep/refresh";
import Plus from "@iconify-icons/ep/plus";
import { templateAPI } from "../../api/notify";
import { confirmAction } from "../../confirm";
import { tableHeaderStyle } from "../../format";
import { NOTIFY_PERMS } from "../../permissions";
import TemplateForm from "./TemplateForm.vue";
import { createTemplateColumns } from "./columns";
import { emptyTemplateForm, type TemplateFormData, type TemplateRow } from "./types";
import { useTemplateList } from "./useTemplateList";

defineOptions({ name: "NotifyTemplate" });

const formRef = ref();
const templateFormRef = ref();
const {
  form,
  loading,
  dataList,
  typeOptions,
  pagination,
  onSearch,
  resetForm,
  handleSizeChange,
  handleCurrentChange
} = useTemplateList();
const columns = createTemplateColumns(() => typeOptions.value);

function toForm(row?: TemplateRow): TemplateFormData {
  const base = emptyTemplateForm();
  if (!row) return base;
  return { ...base, id: row.id, typeCode: row.typeCode, title: row.title, body: row.body, link: row.link };
}

function openDialog(title = "新增", row?: TemplateRow) {
  dialogService.open({
    title: `${title}消息模板`,
    width: "720px",
    draggable: true,
    closeOnClickModal: false,
    closeOnPressEscape: false,
    alignCenter: true,
    props: { formInline: toForm(row), typeOptions: typeOptions.value },
    contentRenderer: ({ options }) =>
      h(TemplateForm, {
        ref: templateFormRef,
        formInline: options.props.formInline,
        typeOptions: options.props.typeOptions
      }),
    beforeSure: done => {
      const FormRef = templateFormRef.value.getRef();
      FormRef.validate((valid: boolean) => {
        if (!valid) return;
        const curData = templateFormRef.value.getFormData();
        const payload = {
          typeCode: curData.typeCode,
          title: curData.title,
          body: curData.body,
          link: curData.link
        };
        const request = curData.id ? templateAPI.update(curData.id, payload) : templateAPI.create(payload);
        request.then(() => {
          message(`${title}成功`, { type: "success" });
          done();
          onSearch();
        });
      });
    }
  });
}

async function handleDelete(row: TemplateRow) {
  const name = typeOptions.value.find(item => item.code === row.typeCode)?.name || row.typeCode;
  if (!(await confirmAction(`确认删除「${name}」的模板吗？`, "删除消息模板"))) return;
  await templateAPI.remove(row.id);
  message("已删除", { type: "success" });
  onSearch();
}
</script>

<template>
  <div class="main">
    <el-form
      ref="formRef"
      :inline="true"
      :model="form"
      class="search-form bg-bg_color w-[99/100] pl-8 pt-[12px] overflow-auto"
    >
      <el-form-item label="消息类型：" prop="typeCode">
        <el-select v-model="form.typeCode" clearable placeholder="全部" class="!w-[180px]">
          <el-option v-for="item in typeOptions" :key="item.code" :label="item.name" :value="item.code" />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" :icon="useRenderIcon('ri:search-line')" :loading="loading" @click="onSearch">
          搜索
        </el-button>
        <el-button :icon="useRenderIcon(Refresh)" @click="resetForm(formRef)">重置</el-button>
      </el-form-item>
    </el-form>
    <PureTableBar title="消息模板" :columns="columns" @refresh="onSearch">
      <template #buttons>
        <Perms :value="NOTIFY_PERMS.template.create">
          <el-button type="primary" :icon="useRenderIcon(Plus)" @click="openDialog('新增')">新增模板</el-button>
        </Perms>
      </template>
      <template #default="{ size, dynamicColumns }">
        <pure-table
          border
          adaptive
          :adaptiveConfig="{ offsetBottom: 108 }"
          align-whole="center"
          row-key="id"
          showOverflowTooltip
          :loading="loading"
          :size="size"
          :data="dataList"
          :columns="dynamicColumns"
          :pagination="{ ...pagination, size }"
          :header-cell-style="tableHeaderStyle"
          @page-size-change="handleSizeChange"
          @page-current-change="handleCurrentChange"
        >
          <template #operation="{ row }">
            <Perms :value="NOTIFY_PERMS.template.update">
              <el-button link type="primary" :size="size" @click="openDialog('编辑', row)">编辑</el-button>
            </Perms>
            <Perms :value="NOTIFY_PERMS.template.delete">
              <el-button link type="danger" :size="size" @click="handleDelete(row)">删除</el-button>
            </Perms>
          </template>
        </pure-table>
      </template>
    </PureTableBar>
  </div>
</template>

<style lang="scss" scoped>
.search-form {
  :deep(.el-form-item) {
    margin-bottom: 12px;
  }
}
</style>
