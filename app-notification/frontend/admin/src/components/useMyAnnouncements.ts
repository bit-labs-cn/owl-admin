import { onMounted, ref } from "vue";
import { announcementAPI } from "../api/notify";

export interface MineAnnouncement {
  id: string;
  title: string;
  content: string;
  level: number;
  displayMode: number;
}

const list = ref<MineAnnouncement[]>([]);
let pending: Promise<void> | null = null;
let bootstrapped = false;

export async function refreshAnnouncements() {
  if (pending) return pending;
  pending = (async () => {
    try {
      const res = await announcementAPI.mine({ page: 1, pageSize: 20 });
      list.value = (res?.data?.list ?? []) as MineAnnouncement[];
    } catch {
      // 公告接口失败时不打断当前页面
    } finally {
      pending = null;
    }
  })();
  return pending;
}

export async function dismissAnnouncement(id: string) {
  await announcementAPI.dismiss(id);
  list.value = list.value.filter(item => item.id !== id);
}

export function useMyAnnouncements() {
  onMounted(() => {
    if (bootstrapped) return;
    bootstrapped = true;
    void refreshAnnouncements();
  });

  return { list, refreshAnnouncements, dismissAnnouncement };
}
