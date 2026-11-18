<script setup lang="ts">
import { h, ref } from "vue";
import { message } from "@bit-labs.cn/owl-ui/utils/message";
import { dialogService } from "@bit-labs.cn/owl-ui/components/ReDialog";
import { PureTableBar } from "@bit-labs.cn/owl-ui/components/RePureTableBar";
import { useRenderIcon } from "@bit-labs.cn/owl-ui/components/ReIcon/src/hooks";
import Refresh from "@iconify-icons/ep/refresh";
import Plus from "@iconify-icons/ep/plus";
import { typeAPI } from "../../api/notify";
import { confirmAction } from "../../confirm";
import { parseJson, tableHeaderStyle } from "../../format";
import { typeStatusOptions } from "../../labels";
import { NOTIFY_PERMS } from "../../permissions";
import TypeForm from "./TypeForm.vue";
import { createTypeColumns } from "./columns";
import { emptyTypeForm, type TypeFormData, type TypeRow } from "./types";
import { useTypeList } from "./useTypeList";

defineOptions({ name: "NotifyType" });

const formRef = ref();
const typeFormRef = ref();
const { form, loading, dataList, pagination, onSearch, resetForm, handleSizeChange, handleCurrentChange } =
  useTypeList();
const columns = createTypeColumns();

function toForm(row?: TypeRow): TypeFormData {
  const base = emptyTypeForm();
  if (!row) return base;
  return {
    id: row.id,
    code: row.code,
    name: row.name,
    description: row.description,
    kind: row.kind,
    defaultChannels: parseJson<string[]>(row.defaultChannels, []),
    status: row.status
  };
}

function openDialog(title = "新增", row?: TypeRow) {
  dialogService.open({
    title: `${title}消息类型`,
    width: "640px",
    draggable: true,
    closeOnClickModal: false,
    closeOnPressEscape: false,
    alignCenter: true,
    props: { formInline: toForm(row) },
    contentRenderer: ({ options }) => h(TypeForm, { ref: typeFormRef, formInline: options.props.formInline }),
    beforeSure: done => {
      const FormRef = typeFormRef.value.getRef();
      FormRef.validate((valid: boolean) => {
        if (!valid) return;
        const curData = typeFormRef.value.getFormData();
        const request = curData.id ? typeAPI.update(curData.id, curData) : typeAPI.create(curData);
        request.then(() => {
          message(`${title}成功`, { type: "success" });
          done();
          onSearch();
        });
      });
    }
  });
}

async function handleDelete(row: TypeRow) {
  if (!(await confirmAction(`确认删除「${row.name}」吗？`, "删除消息类型"))) return;
  await typeAPI.remove(row.id);
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
      <el-form-item label="名称：" prop="name">
        <el-input v-model="form.name" clearable class="!w-[200px]" />
      </el-form-item>
      <el-form-item label="状态：" prop="status">
        <el-select v-model="form.status" clearable placeholder="全部" class="!w-[140px]">
          <el-option v-for="item in typeStatusOptions" :key="item.value" :label="item.label" :value="item.value" />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" :icon="useRenderIcon('ri:search-line')" :loading="loading" @click="onSearch">
          搜索
        </el-button>
        <el-button :icon="useRenderIcon(Refresh)" @click="resetForm(formRef)">重置</el-button>
      </el-form-item>
    </el-form>
    <PureTableBar title="消息类型" :columns="columns" @refresh="onSearch">
      <template #buttons>
        <Perms :value="NOTIFY_PERMS.type.create">
          <el-button type="primary" :icon="useRenderIcon(Plus)" @click="openDialog('新增')">新增类型</el-button>
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
            <Perms :value="NOTIFY_PERMS.type.update">
              <el-button link type="primary" :size="size" @click="openDialog('编辑', row)">编辑</el-button>
            </Perms>
            <Perms :value="NOTIFY_PERMS.type.delete">
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
