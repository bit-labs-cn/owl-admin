<script setup lang="ts">
import { onMounted, ref, computed } from "vue";
import { useRouter } from "vue-router";
import BellIcon from "@iconify-icons/ep/bell";
import { inboxAPI, typeAPI } from "../api/notify";
import { DisplayModeInbox, ReadStatusUnread } from "../constants";
import { formatTime } from "../format";
import type { InboxCounts, InboxRow } from "../views/inbox/types";
import { dismissAnnouncement, refreshAnnouncements, useMyAnnouncements } from "./useMyAnnouncements";
import type { MineAnnouncement } from "./useMyAnnouncements";

const InboxPath = "/system/notify/inbox/index";
const RecentPageSize = 8;

const router = useRouter();
const { list } = useMyAnnouncements();
const inboxAnnouncements = computed(() => list.value.filter(item => item.displayMode === DisplayModeInbox));
const unread = ref(0);
const badgeCount = computed(() => unread.value + inboxAnnouncements.value.length);
const items = ref<InboxRow[]>([]);
const typeOptions = ref<{ code: string; name: string }[]>([]);
const filterType = ref("");
const detailOpen = ref(false);
const detail = ref<MineAnnouncement>();

async function refresh() {
  try {
    const [countRes, listRes] = await Promise.all([
      inboxAPI.counts(),
      inboxAPI.getList({
        page: 1,
        pageSize: RecentPageSize,
        typeCode: filterType.value || undefined,
        readStatus: ReadStatusUnread
      })
    ]);
    const counts = (countRes?.data ?? {}) as unknown as Partial<InboxCounts>;
    unread.value =
      counts.total !== undefined
        ? Number(counts.total)
        : Number(counts.unreadNotice ?? counts.notice ?? 0) + Number(counts.pendingTodo ?? counts.todo ?? 0);
    const listBody = listRes?.data as unknown as { list?: InboxRow[] } | undefined;
    // 接口已按未读过滤，这里再兜一层，避免任何来源的已读消息出现在提醒里
    items.value = (listBody?.list ?? []).filter(item => item.readStatus === ReadStatusUnread);
  } catch {
    // 通知接口失败时保持顶栏可用
  }
}

function onVisible(visible: boolean) {
  if (!visible) return;
  void refresh();
  void refreshAnnouncements();
}

function openAnnouncement(row: MineAnnouncement) {
  detail.value = row;
  detailOpen.value = true;
}

async function acknowledgeAnnouncement() {
  const current = detail.value;
  detailOpen.value = false;
  if (!current) return;
  try {
    await dismissAnnouncement(current.id);
  } catch {
    detailOpen.value = true;
  }
}

async function openItem(row: InboxRow) {
  if (row.readStatus === ReadStatusUnread) {
    try {
      await inboxAPI.read({ ids: [row.id] });
    } catch {
      // 标记失败仍进入收件箱
    }
  }
  await router.push(InboxPath);
}

function openInbox() {
  void router.push(InboxPath);
}

onMounted(async () => {
  try {
    const res = await typeAPI.options();
    typeOptions.value = (res?.data ?? []) as { code: string; name: string }[];
  } catch {
    typeOptions.value = [];
  }
  void refresh();
});
</script>

<template>
  <el-dropdown trigger="click" placement="bottom-end" @visible-change="onVisible">
    <span class="dropdown-badge navbar-bg-hover select-none">
      <el-badge :value="badgeCount === 0 ? '' : badgeCount" :max="99">
        <span class="header-notice-icon">
          <IconifyIconOffline :icon="BellIcon" />
        </span>
      </el-badge>
    </span>
    <template #dropdown>
      <div class="header-notice-panel">
        <div class="header-notice-filter" @click.stop>
          <el-select
            v-model="filterType"
            size="small"
            clearable
            placeholder="全部类型"
            @change="refresh"
          >
            <el-option v-for="item in typeOptions" :key="item.code" :label="item.name" :value="item.code" />
          </el-select>
        </div>
        <el-dropdown-menu class="header-notice-menu">
          <el-dropdown-item v-if="inboxAnnouncements.length === 0 && items.length === 0" disabled>
            暂无未读消息
          </el-dropdown-item>
          <el-dropdown-item
            v-for="item in inboxAnnouncements"
            :key="`announcement-${item.id}`"
            @click="openAnnouncement(item)"
          >
            <span class="header-notice-item">
              <span class="header-notice-item-title">{{ item.title || "无标题" }}</span>
              <span class="header-notice-item-time">公告</span>
            </span>
          </el-dropdown-item>
          <el-dropdown-item
            v-for="item in items"
            :key="`message-${item.id}`"
            @click="openItem(item)"
          >
            <span
              class="header-notice-item"
              :class="{ 'is-unread': item.readStatus === ReadStatusUnread }"
            >
              <span class="header-notice-item-title">{{ item.title || "无标题" }}</span>
              <span class="header-notice-item-time">{{
                [item.typeName || "未分类", formatTime(item.createdAt)].filter(Boolean).join(" · ")
              }}</span>
            </span>
          </el-dropdown-item>
          <el-dropdown-item divided @click="openInbox">查看全部</el-dropdown-item>
        </el-dropdown-menu>
      </div>
    </template>
  </el-dropdown>
  <el-dialog
    v-model="detailOpen"
    :title="detail?.title || '公告'"
    width="520px"
    align-center
    append-to-body
    :close-on-click-modal="false"
    :close-on-press-escape="false"
  >
    <div class="announce-detail">{{ detail?.content || "无正文" }}</div>
    <template #footer>
      <el-button type="primary" @click="acknowledgeAnnouncement">知道了</el-button>
    </template>
  </el-dialog>
</template>

<style lang="scss" scoped>
.dropdown-badge {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: var(--owl-header-height);
  cursor: pointer;

  .header-notice-icon {
    font-size: 18px;
  }
}

.header-notice-panel {
  padding-top: 4px;
}

.header-notice-filter {
  padding: 8px 12px 4px;
}

.header-notice-menu {
  width: 320px;

  :deep(.el-dropdown-menu__item) {
    height: auto;
    padding-top: 8px;
    padding-bottom: 8px;
    line-height: 1.4;
    white-space: normal;
  }
}

.header-notice-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
  width: 100%;
  line-height: 1.4;
  white-space: normal;
}

.header-notice-item-title {
  font-size: 14px;
}

.header-notice-item-time {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.is-unread .header-notice-item-title {
  font-weight: 600;
}

.announce-detail {
  max-height: 360px;
  overflow: auto;
  line-height: 1.6;
  white-space: pre-wrap;
}
</style>
