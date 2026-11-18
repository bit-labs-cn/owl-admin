import { formatTime } from "../../format";
import { kindOptions, labelOf, tagOf } from "../../labels";
import type { DeliveryBatchRow } from "./types";

export function createDeliveryColumns(): TableColumnList {
  return [
    {
      label: "标题",
      prop: "title",
      minWidth: 200,
      cellRenderer: ({ row }) => {
        const item = row as DeliveryBatchRow;
        return (
          <div class="batch-title">
            <span>{item.title}</span>
            <span class="batch-meta">
              {item.typeName || "未分类"} · {formatTime(item.createdAt)}
            </span>
          </div>
        );
      }
    },
    {
      label: "种类",
      prop: "kind",
      width: 105,
      cellRenderer: ({ row }) => (
        <el-tag type={tagOf(kindOptions, (row as DeliveryBatchRow).kind)}>
          {labelOf(kindOptions, (row as DeliveryBatchRow).kind)}
        </el-tag>
      )
    },
    {
      label: "投递",
      prop: "successCount",
      width: 190,
      cellRenderer: ({ row }) => {
        const item = row as DeliveryBatchRow;
        const percent = item.total > 0 ? Math.round((item.successCount / item.total) * 100) : 0;
        return (
          <div class="batch-progress">
            <el-progress percentage={percent} stroke-width={6} show-text={false} />
            <span class="batch-progress-text">
              成功 {item.successCount} / 共 {item.total}
            </span>
          </div>
        );
      }
    },
    {
      label: "异常",
      prop: "failedCount",
      width: 150,
      cellRenderer: ({ row }) => {
        const item = row as DeliveryBatchRow;
        if (item.failedCount === 0 && item.skippedCount === 0) {
          return <span class="muted">无</span>;
        }
        return (
          <div class="batch-tags">
            {item.failedCount > 0 ? <el-tag type="danger" size="small">失败 {item.failedCount}</el-tag> : null}
            {item.skippedCount > 0 ? <el-tag type="info" size="small">跳过 {item.skippedCount}</el-tag> : null}
            {item.pendingCount > 0 ? <el-tag type="warning" size="small">待发送 {item.pendingCount}</el-tag> : null}
          </div>
        );
      }
    },
    { label: "渠道数", prop: "channelCount", width: 90 },
    { label: "发送人", prop: "senderName", width: 120 },
    { label: "操作", fixed: "right", width: 100, slot: "operation" }
  ];
}
