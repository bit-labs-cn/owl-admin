<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import type { UploadFile } from "element-plus";
import { message } from "@bit-labs.cn/owl-ui/utils/message";
import { useRenderIcon } from "@bit-labs.cn/owl-ui/components/ReIcon/src/hooks";
import Refresh from "@iconify-icons/ep/refresh";
import Delete from "@iconify-icons/ep/delete";
import Plus from "@iconify-icons/ep/plus";
import Upload from "@iconify-icons/ep/upload";
import Download from "@iconify-icons/ep/download";
import { systemUserAPI } from "../../api/system";
import {
  downloadExternalTemplate,
  externalTemplateHeaders,
  parseExternalWorkbook,
  type ExternalImportError
} from "../../utils/recipientExcel";
import type { AudienceOption, ExternalRecipient, RecipientOption } from "./useManualSend";

/** 系统用户列表返回的一行，字段见 `app/model/user.go` 的 User。 */
interface SystemUserRow {
  id: string;
  username: string;
  nickname: string;
  email: string;
  phone: string;
  status: number;
}

/** 系统用户状态里「启用」的取值。 */
const UserStatusEnabled = 1;

const props = defineProps<{
  /** 打开弹窗时已经选中的用户 ID。 */
  selectedUserIds: string[];
  /** 打开弹窗时已有的外部人员。 */
  externals: ExternalRecipient[];
  /** 部门树，用来按部门筛选接收人。 */
  depts: AudienceOption[];
  /** 已知用户的名称，用来把已选 ID 显示成人名。 */
  knownUsers: RecipientOption[];
}>();


const PAGE_SIZE = 20;

const keyword = ref("");
const deptId = ref<string>("");

const loading = ref(false);
const list = ref<RecipientOption[]>([]);
const total = ref(0);
const page = ref(1);

/** 选中的用户 ID 与名称缓存：翻页、换筛选条件都不会丢。 */
const selected = ref<Set<string>>(new Set(props.selectedUserIds));
const userCache = reactive(new Map<string, RecipientOption>());
for (const user of props.knownUsers) {
  if (props.selectedUserIds.includes(user.id)) userCache.set(user.id, user);
}

const externals = ref<ExternalRecipient[]>(
  props.externals.map(item => ({ name: item.name, email: item.email, phone: item.phone }))
);

const importVisible = ref(false);

const selectablePageIds = computed(() => list.value.filter(item => !item.disabled).map(item => item.id));
const allPageSelected = computed(
  () => selectablePageIds.value.length > 0 && selectablePageIds.value.every(id => selected.value.has(id))
);
const somePageSelected = computed(
  () => !allPageSelected.value && selectablePageIds.value.some(id => selected.value.has(id))
);

const selectedUsers = computed(() =>
  [...selected.value].map(id => userCache.get(id) ?? { id, name: `未知用户(${id})`, username: "", email: "", phone: "", disabled: false })
);

const validExternals = computed(() => externals.value.filter(item => item.email.trim() || item.phone.trim()));
const targetCount = computed(() => selected.value.size + validExternals.value.length);

function recipientLabel(user: RecipientOption) {
  if (!user.name) return user.username || `未知用户(${user.id})`;
  return user.username && user.name !== user.username ? `${user.name}（${user.username}）` : user.name;
}

async function load() {
  loading.value = true;
  try {
    const res = await systemUserAPI.getList({
      keyword: keyword.value || undefined,
      deptId: deptId.value || undefined,
      page: page.value,
      pageSize: PAGE_SIZE
    });
    const rows = (res?.data?.list ?? []) as SystemUserRow[];
    list.value = rows.map(toRecipient);
    total.value = res?.total ?? 0;
    for (const user of list.value) {
      if (selected.value.has(user.id)) userCache.set(user.id, user);
    }
  } catch {
    list.value = [];
    total.value = 0;
    message("接收人搜索失败，请稍后重试", { type: "error" });
  } finally {
    loading.value = false;
  }
}

/** 系统用户行转成发送页用的候选，禁用账号标出来但仍可见。 */
function toRecipient(row: SystemUserRow): RecipientOption {
  return {
    id: row.id,
    name: row.nickname || row.username,
    username: row.username,
    email: row.email,
    phone: row.phone,
    disabled: row.status !== UserStatusEnabled
  };
}

function search() {
  page.value = 1;
  load();
}

function resetFilters() {
  keyword.value = "";
  deptId.value = "";
  search();
}

function toggle(user: RecipientOption) {
  if (user.disabled) return;
  if (selected.value.has(user.id)) {
    selected.value.delete(user.id);
  } else {
    selected.value.add(user.id);
    userCache.set(user.id, user);
  }
  // Set 的增删不会触发依赖它的 computed，这里换一个新实例
  selected.value = new Set(selected.value);
}

function togglePage(checked: boolean | string | number) {
  const on = Boolean(checked);
  for (const id of selectablePageIds.value) {
    if (on) {
      selected.value.add(id);
      const row = list.value.find(item => item.id === id);
      if (row) userCache.set(id, row);
    } else {
      selected.value.delete(id);
    }
  }
  selected.value = new Set(selected.value);
}

function removeUser(id: string) {
  selected.value.delete(id);
  selected.value = new Set(selected.value);
}

function clearUsers() {
  selected.value = new Set();
}

function clearTargets() {
  selected.value = new Set();
  externals.value = [];
}

function onPageChange(value: number) {
  page.value = value;
  load();
}

function addExternal() {
  externals.value.push({ name: "", email: "", phone: "" });
}

function removeExternal(index: number) {
  externals.value.splice(index, 1);
}

const importing = ref(false);
const importErrors = ref<ExternalImportError[]>([]);

/** 导入会把能用的行并进来；已在列表里的邮箱视为重复，不重复添加。 */
async function onImportFile(uploadFile: UploadFile) {
  const file = uploadFile.raw;
  if (!(file instanceof File)) return;
  if (!file.name.toLowerCase().endsWith(".xlsx")) {
    message("请上传 .xlsx 文件", { type: "warning" });
    return;
  }
  importing.value = true;
  importErrors.value = [];
  try {
    const result = await parseExternalWorkbook(file);
    const exist = new Set(
      externals.value.filter(item => item.email.trim()).map(item => item.email.trim().toLowerCase())
    );
    let added = 0;
    const errors = [...result.errors];
    for (const row of result.rows) {
      const key = row.email.toLowerCase();
      if (exist.has(key)) {
        errors.push({ row: 0, reason: `已在列表里：${row.email}` });
        continue;
      }
      exist.add(key);
      externals.value.push(row);
      added++;
    }
    importErrors.value = errors;
    if (added === 0) {
      message("没有新增外部人员，请检查文件内容", { type: "warning" });
    } else {
      message(`已导入 ${added} 位外部人员${errors.length > 0 ? `，跳过 ${errors.length} 行` : ""}`, {
        type: "success"
      });
    }
  } catch (error) {
    message(error instanceof Error ? error.message : "Excel 解析失败", { type: "error" });
  } finally {
    importing.value = false;
  }
}

async function downloadTemplate() {
  try {
    await downloadExternalTemplate();
  } catch {
    message("模板下载失败", { type: "error" });
  }
}

/** 有行号就说第几行，没有行号的（比如和已有列表重复）只说原因。 */
function importErrorText(item: ExternalImportError) {
  return item.row > 0 ? `第 ${item.row} 行：${item.reason}` : item.reason;
}

/** 交给发送页的最终目标。 */
function getSelection() {
  return {
    userIds: [...selected.value],
    users: [...userCache.values()].filter(user => selected.value.has(user.id)),
    externals: validExternals.value.map(item => ({ ...item }))
  };
}

/** 至少要有站内用户或能投递的外部人员，否则发送没有目标。 */
function validate() {
  if (targetCount.value === 0) {
    return "请至少选择一位接收人，或填写外部人员的邮箱";
  }
  return "";
}

defineExpose({ getSelection, validate });

onMounted(load);
</script>

<template>
  <div class="recipient-picker">
    <div class="picker-toolbar">
      <el-input
        v-model="keyword"
        clearable
        placeholder="用户名、昵称、邮箱或手机号"
        class="keyword"
        @keyup.enter="search"
        @clear="search"
      />

      <el-tree-select
        v-model="deptId"
        :data="depts"
        node-key="id"
        check-strictly
        clearable
        :props="{ label: 'name', children: 'children' }"
        placeholder="部门"
        class="filter"
      />

      <el-button type="primary" :icon="useRenderIcon('ri:search-line')" :loading="loading" @click="search">
        筛选
      </el-button>
      <el-button :icon="useRenderIcon(Refresh)" @click="resetFilters">重置</el-button>
    </div>

    <div class="picker-hint">
      关键词匹配登录名和昵称，部门含下级；已禁用的账号不能选。共 {{ total }} 人符合条件。
    </div>

    <div class="picker-body">
      <div class="picker-left">
        <el-table
          v-loading="loading"
          :data="list"
          border
          size="small"
          max-height="360"
          row-key="id"
          @row-click="toggle"
        >
          <el-table-column width="46">
            <template #header>
              <el-checkbox
                :model-value="allPageSelected"
                :indeterminate="somePageSelected"
                :disabled="selectablePageIds.length === 0"
                @change="togglePage"
              />
            </template>
            <template #default="{ row }">
              <el-checkbox :model-value="selected.has(row.id)" :disabled="row.disabled" @click.stop @change="toggle(row)" />
            </template>
          </el-table-column>
          <el-table-column label="用户" min-width="180">
            <template #default="{ row }">
              <div class="user-cell">
                <span class="user-name">{{ recipientLabel(row) }}</span>
                <span v-if="row.disabled" class="user-disabled">已禁用</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="邮箱" min-width="180">
            <template #default="{ row }">{{ row.email || "—" }}</template>
          </el-table-column>
          <el-table-column label="手机号" width="130">
            <template #default="{ row }">{{ row.phone || "—" }}</template>
          </el-table-column>
        </el-table>

        <el-pagination
          v-if="total > PAGE_SIZE"
          class="picker-pagination"
          layout="total, prev, pager, next"
          background
          :current-page="page"
          :page-size="PAGE_SIZE"
          :total="total"
          @current-change="onPageChange"
        />
      </div>

      <div class="picker-right">
        <div class="right-head">
          <span>已选目标</span>
          <b>{{ targetCount }}</b>
        </div>

        <div class="right-block">
          <div class="block-title">
            <span>站内用户 {{ selectedUsers.length }}</span>
            <el-button v-if="selectedUsers.length > 0" link type="danger" size="small" @click="clearUsers">
              清空
            </el-button>
          </div>
          <div v-if="selectedUsers.length === 0" class="block-empty">在左侧勾选用户</div>
          <div v-else class="chip-list">
            <span v-for="user in selectedUsers" :key="user.id" class="chip">
              {{ recipientLabel(user) }}
              <button type="button" @click="removeUser(user.id)">×</button>
            </span>
          </div>
        </div>

        <div class="right-block">
          <div class="block-title">
            <span>外部人员 {{ validExternals.length }}</span>
            <span class="block-actions">
              <el-button link type="primary" size="small" :icon="useRenderIcon(Plus)" @click="addExternal">添加</el-button>
              <el-button link type="primary" size="small" :icon="useRenderIcon(Upload)" @click="importVisible = !importVisible">
                导入 Excel
              </el-button>
            </span>
          </div>

          <div v-if="importVisible" class="import-box">
            <p class="import-note">
              模板列：{{ externalTemplateHeaders.join(" / ") }}，邮箱必填，称呼和手机号可省。仅支持 .xlsx。
            </p>
            <div class="import-actions">
              <el-button link type="primary" size="small" :icon="useRenderIcon(Download)" @click="downloadTemplate">
                下载模板
              </el-button>
              <el-upload
                :show-file-list="false"
                :auto-upload="false"
                accept=".xlsx"
                :disabled="importing"
                @change="onImportFile"
              >
                <el-button size="small" type="primary" :loading="importing">选择 Excel 导入</el-button>
              </el-upload>
            </div>
            <div v-if="importErrors.length > 0" class="import-errors">
              <span class="import-errors-title">跳过 {{ importErrors.length }} 行</span>
              <span v-for="(item, index) in importErrors.slice(0, 5)" :key="index" class="import-error">
                {{ importErrorText(item) }}
              </span>
              <span v-if="importErrors.length > 5" class="import-error">还有 {{ importErrors.length - 5 }} 行未展示</span>
            </div>
          </div>

          <div v-if="externals.length === 0" class="block-empty">外部人员没有系统账号，只按邮箱投递</div>
          <div v-else class="external-list">
            <div v-for="(item, index) in externals" :key="index" class="external-row">
              <el-input v-model="item.name" size="small" placeholder="称呼" />
              <el-input v-model="item.email" size="small" placeholder="邮箱（必填）" />
              <el-input v-model="item.phone" size="small" placeholder="手机号" />
              <el-button link type="danger" :icon="useRenderIcon(Delete)" @click="removeExternal(index)" />
            </div>
          </div>
        </div>

        <div class="right-foot">
          <el-button link size="small" :disabled="targetCount === 0" @click="clearTargets">清空全部目标</el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.recipient-picker {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.picker-toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;

  .keyword {
    width: 240px;
  }

  .filter {
    width: 150px;
  }
}

.picker-hint {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  align-items: center;
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.picker-body {
  display: grid;
  grid-template-columns: minmax(0, 1.5fr) minmax(300px, 0.9fr);
  gap: 12px;
  align-items: start;
}

@media (max-width: 900px) {
  .picker-body {
    grid-template-columns: minmax(0, 1fr);
  }
}

.picker-left {
  min-width: 0;
}

.user-cell {
  display: flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
}

.user-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.user-disabled {
  flex: 0 0 auto;
  padding: 1px 6px;
  border-radius: 4px;
  background: var(--el-fill-color-light);
  color: var(--el-text-color-secondary);
  font-size: 11px;
}

.picker-pagination {
  margin-top: 8px;
  justify-content: flex-end;
}

.picker-right {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 12px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 8px;
  background: var(--el-fill-color-lighter);
}

.right-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  color: var(--el-text-color-regular);
  font-size: 13px;

  b {
    color: var(--el-color-primary);
    font-size: 18px;
  }
}

.right-block {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding-top: 8px;
  border-top: 1px dashed var(--el-border-color);
}

.block-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  color: var(--el-text-color-primary);
  font-size: 12px;
  font-weight: 600;
}

.block-actions {
  display: inline-flex;
  gap: 4px;
}

.block-empty {
  color: var(--el-text-color-secondary);
  font-size: 12px;
  line-height: 1.5;
}

.chip-list {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  max-height: 120px;
  overflow-y: auto;
}

.chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 8px;
  border: 1px solid var(--el-border-color);
  border-radius: 999px;
  background: var(--el-bg-color);
  font-size: 12px;

  button {
    padding: 0;
    border: 0;
    background: none;
    color: var(--el-text-color-secondary);
    font-size: 13px;
    line-height: 1;
    cursor: pointer;

    &:hover {
      color: var(--el-color-danger);
    }
  }
}

.import-box {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 8px 10px;
  border: 1px dashed var(--el-border-color);
  border-radius: 6px;
  background: var(--el-bg-color);
}

.import-note {
  margin: 0;
  color: var(--el-text-color-secondary);
  font-size: 12px;
  line-height: 1.5;
}

.import-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
  justify-content: flex-end;
}

.import-errors {
  display: flex;
  flex-direction: column;
  gap: 2px;
  max-height: 84px;
  overflow-y: auto;
  color: var(--el-color-warning-dark-2);
  font-size: 12px;
}

.import-errors-title {
  font-weight: 600;
}

.import-error {
  line-height: 1.5;
}

.external-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
  max-height: 168px;
  overflow-y: auto;
}

.external-row {
  display: grid;
  grid-template-columns: minmax(0, 0.8fr) minmax(0, 1.2fr) minmax(0, 0.9fr) auto;
  gap: 4px;
  align-items: center;
}

.right-foot {
  display: flex;
  justify-content: flex-end;
  padding-top: 6px;
  border-top: 1px dashed var(--el-border-color);
}
</style>
