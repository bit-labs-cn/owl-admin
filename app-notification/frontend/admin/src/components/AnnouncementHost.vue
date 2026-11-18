<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { DisplayModeBanner, DisplayModePopup, LevelDanger, LevelWarning } from "../constants";
import { dismissAnnouncement, useMyAnnouncements } from "./useMyAnnouncements";

const { list } = useMyAnnouncements();
const banners = computed(() => list.value.filter(item => item.displayMode === DisplayModeBanner));
const popups = computed(() => list.value.filter(item => item.displayMode === DisplayModePopup));
const popupOpen = ref(false);
const currentPopup = computed(() => popups.value[0]);
let popupBusy = false;

watch(
  popups,
  items => {
    popupOpen.value = items.length > 0;
  },
  { immediate: true }
);

function levelClass(level: number) {
  if (level === LevelDanger) return "is-danger";
  if (level === LevelWarning) return "is-warning";
  return "is-info";
}

async function closeBanner(id: string) {
  try {
    await dismissAnnouncement(id);
  } catch {
    // 关闭失败时保留横幅
  }
}

async function acknowledgePopup() {
  const current = currentPopup.value;
  if (!current || popupBusy) return;
  popupBusy = true;
  try {
    await dismissAnnouncement(current.id);
  } catch {
    popupOpen.value = true;
  } finally {
    popupBusy = false;
  }
}

async function onPopupBeforeClose(done: () => void) {
  await acknowledgePopup();
  done();
}
</script>

<template>
  <div class="announce-host">
    <div
      v-for="item in banners"
      :key="item.id"
      class="announce-banner"
      :class="levelClass(item.level)"
    >
      <span class="announce-banner-title">{{ item.title || "无标题" }}</span>
      <span v-if="item.content" class="announce-banner-content">{{ item.content }}</span>
      <button type="button" class="announce-banner-close" @click="closeBanner(item.id)">关闭</button>
    </div>
    <el-dialog
      :model-value="popupOpen"
      :title="currentPopup?.title || '公告'"
      width="520px"
      align-center
      append-to-body
      :close-on-click-modal="false"
      :close-on-press-escape="false"
      :before-close="onPopupBeforeClose"
    >
      <div class="announce-popup-body">{{ currentPopup?.content || "无正文" }}</div>
      <template #footer>
        <el-button type="primary" @click="acknowledgePopup">知道了</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style lang="scss" scoped>
.announce-banner {
  display: flex;
  gap: 12px;
  align-items: center;
  min-height: 36px;
  padding: 6px 16px;
  font-size: 13px;
  line-height: 1.4;
}

.announce-banner-title {
  flex: none;
  font-weight: 600;
}

.announce-banner-content {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.announce-banner-close {
  flex: none;
  padding: 0;
  font-size: 12px;
  color: inherit;
  cursor: pointer;
  background: transparent;
  border: 0;
}

.is-info {
  color: var(--el-color-primary);
  background: var(--el-color-primary-light-9);
}

.is-warning {
  color: var(--el-color-warning);
  background: var(--el-color-warning-light-9);
}

.is-danger {
  color: var(--el-color-danger);
  background: var(--el-color-danger-light-9);
}

.announce-popup-body {
  max-height: 360px;
  overflow: auto;
  line-height: 1.6;
  white-space: pre-wrap;
}
</style>
