<script setup lang="ts">
import { onMounted } from "vue";
import { ElMessage } from "element-plus";
import { useSetting } from "./useSetting";
import type { SettingItem, SettingTreeNode } from "./types";

const imageTypes = ["image/png", "image/jpeg", "image/gif", "image/webp"];

defineOptions({ name: "SystemSetting" });

const {
  loading,
  saving,
  currentSubsystem,
  currentParentTitle,
  currentGroup,
  items,
  treeData,
  currentNodeKey,
  load,
  onNode,
  save,
  restore
} = useSetting();

onMounted(load);

function isStacked(item: SettingItem) {
  return item.type !== "bool" && item.type !== "int";
}

function systemNodes(system: SettingTreeNode) {
  return system.children ?? [];
}

function onInt(item: SettingItem, value: number | null | undefined) {
  if (value === undefined || value === null || Number.isNaN(value)) return;
  item.value = String(value);
}

function onBool(item: SettingItem, value: string | number | boolean) {
  item.value = value === true ? "true" : "false";
}

function imageLimit(item: SettingItem) {
  return item.max > 0 ? item.max : 200 * 1024;
}

function chooseImage(item: SettingItem) {
  const input = document.createElement("input");
  input.type = "file";
  input.accept = imageTypes.join(",");
  input.onchange = () => {
    const file = input.files?.[0];
    if (!file) return;
    if (file.type && !imageTypes.includes(file.type)) {
      ElMessage.error("只支持 png、jpeg、gif、webp");
      return;
    }
    const limit = imageLimit(item);
    if (file.size > limit) {
      ElMessage.error(`图片不能超过 ${Math.ceil(limit / 1024)}KB`);
      return;
    }
    const reader = new FileReader();
    reader.onload = () => {
      const text = typeof reader.result === "string" ? reader.result : "";
      if (!/^data:image\/(png|jpeg|gif|webp);base64,/.test(text)) {
        ElMessage.error("只支持 png、jpeg、gif、webp");
        return;
      }
      item.value = text;
    };
    reader.readAsDataURL(file);
  };
  input.click();
}
</script>

<template>
  <div class="main">
    <div v-loading="loading" class="setting-shell">
      <el-empty v-if="!loading && treeData.length === 0" description="还没有可配置项" />

      <template v-else-if="treeData.length">
        <aside class="setting-nav" aria-label="配置分组">
          <section v-for="system in treeData" :key="system.id" class="setting-nav-group">
            <button type="button" class="setting-system" @click="onNode(system)">
              {{ system.label }}
            </button>
            <template v-for="group in systemNodes(system)" :key="group.id">
              <button
                type="button"
                class="setting-link"
                :class="{
                  'is-current': group.id === currentNodeKey,
                  'is-branch': !!group.children?.length
                }"
                :aria-current="group.id === currentNodeKey ? 'page' : undefined"
                @click="onNode(group)"
              >
                {{ group.label }}
              </button>
              <button
                v-for="child in group.children"
                :key="child.id"
                type="button"
                class="setting-link setting-link--nested"
                :class="{ 'is-current': child.id === currentNodeKey }"
                :aria-current="child.id === currentNodeKey ? 'page' : undefined"
                @click="onNode(child)"
              >
                {{ child.label }}
              </button>
            </template>
          </section>
        </aside>

        <section class="setting-main">
          <header v-if="currentGroup" class="setting-head">
            <p class="setting-kicker">
              {{ currentSubsystem?.title }}<template v-if="currentParentTitle"> / {{ currentParentTitle }}</template>
            </p>
            <h2 class="setting-title">{{ currentGroup.title }}</h2>
          </header>

          <div class="setting-body">
            <article
              v-for="item in items"
              :key="item.key"
              class="setting-row"
              :class="{ 'is-stacked': isStacked(item) }"
            >
              <div class="setting-copy">
                <h3 class="setting-name">
                  <span>{{ item.title }}</span>
                  <code class="setting-key">{{ item.key }}</code>
                </h3>
                <p v-if="item.hint" class="setting-hint">{{ item.hint }}</p>
              </div>
              <div class="setting-field">
                <el-switch
                  v-if="item.type === 'bool'"
                  :model-value="item.value === 'true'"
                  @change="onBool(item, $event)"
                />
                <el-input-number
                  v-else-if="item.type === 'int'"
                  :model-value="Number(item.value)"
                  :min="item.min || undefined"
                  :max="item.max || undefined"
                  :step="1"
                  step-strictly
                  @change="onInt(item, $event)"
                />
                <div v-else-if="item.type === 'image'" class="setting-image">
                  <img v-if="item.value" class="setting-preview" :src="item.value" alt="" />
                  <div class="setting-image-actions">
                    <el-button @click="chooseImage(item)">选择图片</el-button>
                    <el-button v-if="item.value" @click="item.value = ''">清除</el-button>
                  </div>
                </div>
                <el-input
                  v-else
                  v-model="item.value"
                  clearable
                  :maxlength="item.max || undefined"
                  show-word-limit
                />
              </div>
            </article>
          </div>

          <footer class="setting-actions">
            <el-button :disabled="saving || !currentGroup" @click="currentGroup && restore(currentGroup)">
              恢复默认
            </el-button>
            <el-button type="primary" :loading="saving" @click="save">保存</el-button>
          </footer>
        </section>
      </template>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.setting-shell {
  display: flex;
  align-items: stretch;
  min-height: calc(100vh - 154px);
  overflow: hidden;
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 12px;
}

.setting-nav {
  flex: 0 0 228px;
  padding: 18px 12px 20px;
  background: var(--el-fill-color-lighter);
  border-right: 1px solid var(--el-border-color-lighter);
}

.setting-nav-group + .setting-nav-group {
  margin-top: 18px;
}

.setting-system,
.setting-link {
  display: block;
  width: 100%;
  border: 0;
  text-align: left;
  cursor: pointer;
}

.setting-system {
  margin: 0 0 4px;
  padding: 6px 10px;
  background: transparent;
  color: var(--el-text-color-primary);
  font-size: 15px;
  font-weight: 600;
  line-height: 22px;
}

.setting-link {
  width: calc(100% - 18px);
  margin-left: 18px;
  padding: 8px 10px;
  border-radius: 8px;
  background: transparent;
  color: var(--el-text-color-regular);
  font-size: 14px;
  line-height: 22px;
}

.setting-link--nested {
  width: calc(100% - 36px);
  margin-left: 36px;
}

.setting-link.is-branch {
  color: var(--el-text-color-primary);
  font-weight: 600;
}

.setting-link:hover {
  background: var(--el-fill-color);
}

.setting-link.is-current {
  background: var(--el-color-primary-light-9);
  color: var(--el-color-primary);
  font-weight: 600;
}

.setting-system:focus-visible,
.setting-link:focus-visible {
  outline: 2px solid var(--el-color-primary);
  outline-offset: 2px;
}

.setting-main {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-width: 0;
  padding: 22px 28px 18px;
}

.setting-head {
  padding-bottom: 16px;
  border-bottom: 1px solid var(--el-border-color-extra-light);
}

.setting-kicker {
  margin: 0 0 4px;
  color: var(--el-text-color-secondary);
  font-size: 13px;
  line-height: 20px;
}

.setting-title {
  margin: 0;
  color: var(--el-text-color-primary);
  font-size: 22px;
  font-weight: 600;
  line-height: 30px;
}

.setting-body {
  flex: 1;
}

.setting-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 28px;
  align-items: center;
  padding: 18px 0;
  border-bottom: 1px solid var(--el-border-color-extra-light);
}

.setting-row.is-stacked {
  grid-template-columns: 1fr;
  gap: 12px;
  align-items: stretch;
}

.setting-name {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: baseline;
  margin: 0;
  color: var(--el-text-color-primary);
  font-size: 14px;
  font-weight: 600;
  line-height: 22px;
}

.setting-key {
  color: var(--el-text-color-secondary);
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 12px;
  font-weight: 400;
}

.setting-hint {
  max-width: 52em;
  margin: 4px 0 0;
  color: var(--el-text-color-secondary);
  font-size: 13px;
  line-height: 1.5;
}

.setting-field {
  display: flex;
  justify-content: flex-end;
}

.setting-row.is-stacked .setting-field {
  justify-content: stretch;
}

.setting-field :deep(.el-input) {
  width: 100%;
}

.setting-field :deep(.el-input-number) {
  width: 148px;
}

.setting-image {
  display: flex;
  gap: 16px;
  align-items: center;
  width: 100%;
  padding: 12px;
  background: var(--el-fill-color-lighter);
  border: 1px dashed var(--el-border-color);
  border-radius: 10px;
}

.setting-image-actions {
  display: flex;
  gap: 8px;
}

.setting-preview {
  width: 48px;
  height: 48px;
  object-fit: contain;
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 8px;
}

.setting-actions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
  padding-top: 16px;
}

@media (max-width: 768px) {
  .setting-shell {
    flex-direction: column;
    min-height: 0;
  }

  .setting-nav {
    flex: none;
    border-right: 0;
    border-bottom: 1px solid var(--el-border-color-lighter);
  }

  .setting-main {
    padding: 18px 16px 16px;
  }

  .setting-row {
    grid-template-columns: 1fr;
    gap: 10px;
  }

  .setting-field {
    justify-content: flex-start;
  }
}
</style>
