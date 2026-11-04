<script setup lang="ts">
import { ref, h, computed } from "vue";
import { useSceneList } from "./useSceneList";
import { createSceneColumns } from "./sceneColumns";
import SceneForm from "./SceneForm.vue";
import LegalSwitch from "./LegalSwitch.vue";
import type { LegalSceneForm, LegalSceneRow } from "./types";
import { legalAPI } from "@bit-labs.cn/owl-admin-ui/api/legal";
import { addDialog } from "@bit-labs.cn/owl-ui/components/ReDialog";
import { PureTableBar } from "@bit-labs.cn/owl-ui/components/RePureTableBar";
import { useRenderIcon } from "@bit-labs.cn/owl-ui/components/ReIcon/src/hooks";
import { ElMessage } from "element-plus";

import AddFill from "@iconify-icons/ri/add-circle-line";
import Refresh from "@iconify-icons/ep/refresh";
import EditPen from "@iconify-icons/ep/edit-pen";
import Delete from "@iconify-icons/ep/delete";

const formRef = ref();
const sceneFormRef = ref();
const { form, loading, dataList, pagination, onSearch, resetForm, handleSizeChange, handleCurrentChange } = useSceneList();
const columns = computed(() => createSceneColumns(pagination));

function openDialog(title = "新增", row?: LegalSceneRow) {
  addDialog({
    title: `${title}场景`,
    props: {
      formInline: {
        id: row?.id,
        key: row?.key ?? "",
        name: row?.name ?? "",
        kind: row?.kind ?? "consent",
        items: row?.items ? row.items.map(item => ({ ...item })) : []
      }
    },
    width: "760px",
    draggable: true,
    alignCenter: true,
    closeOnClickModal: false,
    contentRenderer: ({ options }) => h(SceneForm, { ref: sceneFormRef, formInline: options.props.formInline }),
    beforeSure: done => {
      const FormRef = sceneFormRef.value.getRef();
      const curData = sceneFormRef.value.getFormData() as LegalSceneForm;
      FormRef.validate((valid: boolean) => {
        if (!valid) return;
        if (curData.items.some(item => !item.documentKey)) {
          ElMessage.warning("请为每一行选择条款");
          return;
        }
        if (curData.kind === "consent" && curData.items.some(item => !item.groupKey.trim())) {
          ElMessage.warning("同意场景的每一行都要填写勾选组");
          return;
        }
        const payload = {
          key: curData.key,
          name: curData.name,
          kind: curData.kind,
          items: curData.items
        };
        const request = curData.id ? legalAPI.updateScene({ id: curData.id, ...payload }) : legalAPI.createScene(payload);
        request.then(() => {
          done();
          onSearch();
        });
      });
    }
  });
}

function handleDelete(row: LegalSceneRow) {
  legalAPI.deleteScene(row.id).then(() => onSearch());
}
</script>

<template>
  <div>
    <el-form ref="formRef" :inline="true" :model="form" class="search-form bg-bg_color pl-8 pt-[12px]">
      <el-form-item label="名称：" prop="name">
        <el-input v-model="form.name" placeholder="场景名称" clearable class="!w-[180px]" />
      </el-form-item>
      <el-form-item label="用途：" prop="kind">
        <el-select v-model="form.kind" placeholder="全部" clearable class="!w-[140px]">
          <el-option label="同意" value="consent" />
          <el-option label="阅读" value="reader" />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" :icon="useRenderIcon('ri:search-line')" :loading="loading" @click="onSearch">
          搜索
        </el-button>
        <el-button :icon="useRenderIcon(Refresh)" @click="resetForm(formRef)">重置</el-button>
      </el-form-item>
    </el-form>
    <PureTableBar title="场景" :columns="columns" @refresh="onSearch">
      <template #title>
        <LegalSwitch />
      </template>
      <template #buttons>
        <el-button type="primary" :icon="useRenderIcon(AddFill)" @click="openDialog()">新增场景</el-button>
      </template>
      <template #default="{ size, dynamicColumns }">
        <pure-table
          border
          adaptive
          :adaptiveConfig="{ offsetBottom: 160 }"
          align-whole="center"
          row-key="id"
          :loading="loading"
          :size="size"
          :data="dataList"
          :columns="dynamicColumns"
          :pagination="{ ...pagination, size }"
          :header-cell-style="{
            background: 'var(--el-fill-color-light)',
            color: 'var(--el-text-color-primary)'
          }"
          @page-size-change="handleSizeChange"
          @page-current-change="handleCurrentChange"
        >
          <template #operation="{ row }">
            <el-button link type="primary" :size="size" :icon="useRenderIcon(EditPen)" @click="openDialog('修改', row)">
              修改
            </el-button>
            <el-popconfirm :title="`删除场景「${row.name}」？`" @confirm="handleDelete(row)">
              <template #reference>
                <el-button link type="danger" :size="size" :icon="useRenderIcon(Delete)">删除</el-button>
              </template>
            </el-popconfirm>
          </template>
        </pure-table>
      </template>
    </PureTableBar>
  </div>
</template>
