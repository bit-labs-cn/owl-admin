import { h } from "vue";
import { dialogService } from "@bit-labs.cn/owl-ui/components/ReDialog";
import { message } from "@bit-labs.cn/owl-ui/utils/message";
import { ElMessageBox } from "element-plus";
import { retryOutbox } from "@bit-labs.cn/owl-admin-ui/api/event";
import OutboxDetail from "./OutboxDetail.vue";
import InboxDetail from "./InboxDetail.vue";
import type { EventInboxRow, EventOutboxRow } from "./types";

/**
 * 打开发件箱事件详情。
 * 弹窗按 owl-dialog-scroll 约定：点遮罩、按 Esc 都不关闭。
 */
export function openOutboxDetail(row: EventOutboxRow) {
  dialogService.open({
    title: "发件箱事件详情",
    width: "820px",
    draggable: true,
    alignCenter: true,
    hideFooter: true,
    closeOnClickModal: false,
    closeOnPressEscape: false,
    contentRenderer: () => h(OutboxDetail, { row })
  });
}

/** 打开收件箱消费流水详情 */
export function openInboxDetail(row: EventInboxRow) {
  dialogService.open({
    title: "收件箱消费流水",
    width: "640px",
    draggable: true,
    alignCenter: true,
    hideFooter: true,
    closeOnClickModal: false,
    closeOnPressEscape: false,
    contentRenderer: () => h(InboxDetail, { row })
  });
}

/**
 * 二次确认后重投一条发件箱事件。
 * 只把行重置为待发送，真正投递仍由框架 Relay 完成。
 */
export async function confirmRetryOutbox(
  row: EventOutboxRow,
  reload?: () => void
) {
  try {
    await ElMessageBox.confirm(
      `确认重投事件「${row.eventType || row.eventId}」吗？确认后会重置为待发送，由事件中继在下一个轮询周期内重新投递。`,
      "重投事件",
      {
        confirmButtonText: "重投",
        cancelButtonText: "取消",
        type: "warning",
        draggable: true,
        closeOnClickModal: false,
        closeOnPressEscape: false
      }
    );
  } catch {
    return;
  }

  try {
    const res = await retryOutbox(row.id);
    message(res?.msg || "已重置为待发送", { type: "success" });
    reload?.();
  } catch (err) {
    message(err instanceof Error ? err.message : "重投失败", { type: "error" });
  }
}
