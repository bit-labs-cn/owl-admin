<script setup lang="ts">
import { h, ref } from "vue";
import { message } from "@bit-labs.cn/owl-ui/utils/message";
import { dialogService } from "@bit-labs.cn/owl-ui/components/ReDialog";
import { PureTableBar } from "@bit-labs.cn/owl-ui/components/RePureTableBar";
import { useRenderIcon } from "@bit-labs.cn/owl-ui/components/ReIcon/src/hooks";
import Refresh from "@iconify-icons/ep/refresh";
import Plus from "@iconify-icons/ep/plus";
import { announcementAPI } from "../../api/notify";
import AnnouncementRead from "./AnnouncementRead.vue";
import type { AnnouncementReadView } from "./readTypes";
import { confirmAction } from "../../confirm";
import { AnnouncementStatusOffline, AnnouncementStatusPublished } from "../../constants";
import { fromRFC3339, tableHeaderStyle, toRFC3339 } from "../../format";
import { announcementStatusOptions } from "../../labels";
import { NOTIFY_PERMS } from "../../permissions";
import AnnouncementForm from "./AnnouncementForm.vue";
import { createAnnouncementColumns } from "./columns";
import { emptyAnnouncementForm, type AnnouncementFormData, type AnnouncementRow } from "./types";
import { useAnnouncementList } from "./useAnnouncementList";

defineOptions({ name: "NotifyAnnouncement" });

const formRef = ref();
const announcementFormRef = ref();
const { form, loading, dataList, pagination, onSearch, resetForm, handleSizeChange, handleCurrentChange } =
  useAnnouncementList();
const columns = createAnnouncementColumns();

function toForm(row?: AnnouncementRow): AnnouncementFormData {
  const base = emptyAnnouncementForm();
  if (!row) return base;
  return {
    ...base,
    id: row.id,
    title: row.title,
    content: row.content,
    level: row.level,
    displayMode: row.displayMode,
    targetScope: row.targetScope || "all",
    startAt: fromRFC3339(row.startAt),
    endAt: fromRFC3339(row.endAt),
    status: row.status
  };
}

function openDialog(title = "新增", row?: AnnouncementRow) {
  dialogService.open({
    title: `${title}公告`,
    width: "680px",
    draggable: true,
    closeOnClickModal: false,
    closeOnPressEscape: false,
    alignCenter: true,
    props: { formInline: toForm(row) },
    contentRenderer: ({ options }) =>
      h(AnnouncementForm, { ref: announcementFormRef, formInline: options.props.formInline }),
    beforeSure: done => {
      const FormRef = announcementFormRef.value.getRef();
      FormRef.validate((valid: boolean) => {
        if (!valid) return;

        const curData = announcementFormRef.value.getFormData();
        const payload = {
          ...curData,
          startAt: toRFC3339(curData.startAt),
          endAt: toRFC3339(curData.endAt)
        };
        const request = curData.id
          ? announcementAPI.update(curData.id, payload)
          : announcementAPI.create(payload);
        request.then(() => {
          message("已保存", { type: "success" });
          done();
          onSearch();
        });
      });
    }
  });
}

async function publish(row: AnnouncementRow) {
  if (!(await confirmAction(`确认发布「${row.title}」吗？`, "发布公告"))) return;
  await announcementAPI.publish(row.id);
  message("已发布", { type: "success" });
  onSearch();
}

async function offline(row: AnnouncementRow) {
  if (!(await confirmAction(`确认下线「${row.title}」吗？`, "下线公告"))) return;
  await announcementAPI.offline(row.id);
  message("已下线", { type: "success" });
  onSearch();
}

const readView = ref<AnnouncementReadView>();

async function showRead(row: AnnouncementRow) {
  readView.value = undefined;
  try {
    const res = await announcementAPI.readState(row.id);
    readView.value = (res?.data ?? undefined) as AnnouncementReadView | undefined;
  } catch {
    message("已读情况加载失败", { type: "error" });
    return;
  }

  dialogService.open({
    title: `已读情况 · ${row.title}`,
    width: "860px",
    hideFooter: true,
    draggable: true,
    closeOnClickModal: false,
    closeOnPressEscape: false,
    alignCenter: true,
    contentRenderer: () => h(AnnouncementRead, { view: readView.value })
  });
}

async function handleDelete(row: AnnouncementRow) {
  if (!(await confirmAction(`确认删除「${row.title}」吗？`, "删除公告"))) return;
  await announcementAPI.remove(row.id);
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
      <el-form-item label="标题：" prop="title">
        <el-input v-model="form.title" clearable class="!w-[200px]" />
      </el-form-item>
      <el-form-item label="状态：" prop="status">
        <el-select v-model="form.status" clearable placeholder="全部" class="!w-[140px]">
          <el-option
            v-for="item in announcementStatusOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" :icon="useRenderIcon('ri:search-line')" :loading="loading" @click="onSearch">
          搜索
        </el-button>
        <el-button :icon="useRenderIcon(Refresh)" @click="resetForm(formRef)">重置</el-button>
      </el-form-item>
    </el-form>
    <PureTableBar title="公告列表" :columns="columns" @refresh="onSearch">
      <template #buttons>
        <Perms :value="NOTIFY_PERMS.announcement.create">
          <el-button type="primary" :icon="useRenderIcon(Plus)" @click="openDialog('新增')">新增公告</el-button>
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
            <Perms :value="NOTIFY_PERMS.announcement.readState">
              <el-button link type="primary" :size="size" @click="showRead(row)">已读情况</el-button>
            </Perms>
            <Perms :value="NOTIFY_PERMS.announcement.publish">
              <el-button
                v-if="row.status !== AnnouncementStatusPublished"
                link
                type="success"
                :size="size"
                @click="publish(row)"
              >
                发布
              </el-button>
            </Perms>
            <Perms :value="NOTIFY_PERMS.announcement.offline">
              <el-button
                v-if="row.status !== AnnouncementStatusOffline"
                link
                type="warning"
                :size="size"
                @click="offline(row)"
              >
                下线
              </el-button>
            </Perms>
            <Perms :value="NOTIFY_PERMS.announcement.update">
              <el-button link type="primary" :size="size" @click="openDialog('编辑', row)">编辑</el-button>
            </Perms>
            <Perms :value="NOTIFY_PERMS.announcement.delete">
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
