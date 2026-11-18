import { formatTime } from "../../format";
import { KindTodo } from "../../constants";
import { kindOptions, labelOf, tagOf } from "../../labels";
import type { MessageRow } from "./types";

export function createMessageColumns(): TableColumnList {
  return [
    { label: "标题", prop: "title", minWidth: 220 },
    {
      label: "种类",
      prop: "kind",
      width: 110,
      cellRenderer: ({ row }) => (
        <el-tag type={tagOf(kindOptions, (row as MessageRow).kind)}>{labelOf(kindOptions, (row as MessageRow).kind)}</el-tag>
      )
    },
    {
      label: "消息类型",
      prop: "typeName",
      minWidth: 140,
      cellRenderer: ({ row }) => {
        const item = row as MessageRow;
        if (item.typeName) return item.typeName;
        if (item.typeCode) return `未知类型(${item.typeCode})`;
        return "未分类";
      }
    },
    {
      label: "接收 / 已读",
      prop: "readCount",
      width: 150,
      cellRenderer: ({ row }) => {
        const item = row as MessageRow;
        const percent = item.recipientCount > 0 ? Math.round((item.readCount / item.recipientCount) * 100) : 0;
        return (
          <div class="read-cell">
            <el-progress percentage={percent} stroke-width={6} show-text={false} />
            <span class="read-text">
              {item.readCount}/{item.recipientCount}
              {item.unreadCount > 0 ? <span class="unread">未读 {item.unreadCount}</span> : <span class="all-read">全部已读</span>}
            </span>
          </div>
        );
      }
    },
    {
      label: "待办",
      prop: "todoPending",
      width: 130,
      cellRenderer: ({ row }) => {
        const item = row as MessageRow;
        if (item.kind !== KindTodo) return "—";
        return `待处理 ${item.todoPending} / 已完成 ${item.todoDone}`;
      }
    },
    {
      label: "投递失败",
      prop: "deliveryFailed",
      width: 110,
      cellRenderer: ({ row }) => {
        const failed = (row as MessageRow).deliveryFailed;
        return failed > 0 ? <el-tag type="danger">{failed}</el-tag> : <span class="muted">无</span>;
      }
    },
    { label: "发送人", prop: "senderName", width: 130 },
    {
      label: "发送时间",
      prop: "createdAt",
      minWidth: 170,
      cellRenderer: ({ row }) => formatTime((row as MessageRow).createdAt)
    },
    { label: "操作", fixed: "right", width: 110, slot: "operation" }
  ];
}
