<script setup lang="ts">
import { reactive, ref } from "vue";
import { useRenderIcon } from "@bit-labs.cn/owl-ui/components/ReIcon/src/hooks";
import { DictTable } from "./controller";
import { Delete } from "@element-plus/icons-vue";
import { DictAPI } from "@bit-labs.cn/owl-admin-ui/api/dict";
import { isAllEmpty } from "@pureadmin/utils";

const dict = new DictTable(new DictAPI());
const formRef = ref();
const form = reactive({ name: "", type: "" });

function onSearch() {
  const params: { name?: string; type?: string } = {};
  if (!isAllEmpty(form.name)) params.name = form.name;
  if (!isAllEmpty(form.type)) params.type = form.type;
  dict.loadData(params);
}

function resetForm(el) {
  if (!el) return;
  el.resetFields();
  form.name = "";
  form.type = "";
  dict.loadData();
}
</script>

<template>
  <div>
    <el-alert
      type="warning"
      :closable="false"
      show-icon
      description="Tip：双击修改,拖拽排序,按回车键保存."
      style="margin: 10px 0"
    />
    <div class="flex">
      <div class="!w-[60vw]">
        <el-form
          ref="formRef"
          :inline="true"
          :model="form"
          class="search-form bg-bg_color w-[99/100] pl-8 pt-[12px] overflow-auto"
        >
          <el-form-item label="字典名称：" prop="name">
            <el-input
              v-model="form.name"
              placeholder="请输入字典名称"
              clearable
              class="!w-[180px]"
            />
          </el-form-item>
          <el-form-item label="字典类型：" prop="type">
            <el-input
              v-model="form.type"
              placeholder="请输入字典类型"
              clearable
              class="!w-[180px]"
            />
          </el-form-item>
          <el-form-item>
            <el-button
              type="primary"
              :icon="useRenderIcon('ri:search-line')"
              @click="onSearch"
            >
              搜索
            </el-button>
            <el-button @click="resetForm(formRef)">重置</el-button>
          </el-form-item>
        </el-form>
        <div class="flex justify-between p-2.5 bg-white">
          <div>字典列表</div>
          <div>
            <el-button type="primary" size="small" @click="dict.createDict()"
              >新增字典
            </el-button>
          </div>
        </div>
        <pure-table
          id="dictTable"
          row-key="id"
          :data="dict.data()"
          :columns="dict.columns"
          :row-class-name="dict.rowClass"
          :loading="dict.isLoading()"
          @row-click="dict.chooseDict"
        >
          <template #operation="{ index }">
            <el-button
              type="danger"
              :icon="Delete"
              circle
              size="small"
              @click="dict.deleteDict(index)"
            />
          </template>
        </pure-table>
      </div>
      &nbsp;&nbsp;
      <div class="!w-[40vw]">
        <div class="flex justify-between p-2.5 bg-white">
          <div>字典详情（{{ dict.choseDict.value?.name }}）</div>
          <div>
            <el-button
              type="primary"
              size="small"
              :disabled="!dict.choseDict.value"
              @click="dict.createDictItem()"
              >新增字典项
            </el-button>
          </div>
        </div>
        <pure-table
          id="dictItemTable"
          row-key="id"
          :data="dict.itemData()"
          :columns="dict.itemColumns()"
          :loading="dict.itemIsLoading()"
        >
          <template #operation="{ index }">
            <el-button
              type="danger"
              :icon="Delete"
              circle
              size="small"
              @click="dict.deleteDictItem(index)"
            />
          </template>
        </pure-table>
      </div>
    </div>
  </div>
</template>

<style>
.add-bg {
  background: #ececec !important;
}
</style>
