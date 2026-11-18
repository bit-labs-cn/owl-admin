import { formatTime } from "../../format";
import { kindOptions, labelOf, readStatusOptions, tagOf, todoStatusOptions } from "../../labels";
import type { InboxRow } from "./types";

export function createInboxColumns(): TableColumnList {
  return [
    { label: "标题", prop: "title", minWidth: 220 },
    {
      label: "消息类型",
      prop: "typeName",
      minWidth: 140,
      cellRenderer: ({ row }) => {
        const item = row as InboxRow;
        if (item.typeName) return item.typeName;
        if (item.typeCode) return `未知类型(${item.typeCode})`;
        return "未分类";
      }
    },
    {
      label: "种类",
      prop: "kind",
      width: 110,
      cellRenderer: ({ row }) => (
        <el-tag type={tagOf(kindOptions, (row as InboxRow).kind)}>{labelOf(kindOptions, (row as InboxRow).kind)}</el-tag>
      )
    },
    {
      label: "阅读状态",
      prop: "readStatus",
      width: 100,
      cellRenderer: ({ row }) => (
        <el-tag type={tagOf(readStatusOptions, (row as InboxRow).readStatus)}>
          {labelOf(readStatusOptions, (row as InboxRow).readStatus)}
        </el-tag>
      )
    },
    {
      label: "待办状态",
      prop: "todoStatus",
      width: 100,
      cellRenderer: ({ row }) => labelOf(todoStatusOptions, (row as InboxRow).todoStatus)
    },
    {
      label: "截止时间",
      prop: "dueAt",
      minWidth: 170,
      cellRenderer: ({ row }) => formatTime((row as InboxRow).dueAt)
    },
    {
      label: "接收时间",
      prop: "createdAt",
      minWidth: 170,
      cellRenderer: ({ row }) => formatTime((row as InboxRow).createdAt)
    },
    { label: "操作", fixed: "right", width: 220, slot: "operation" }
  ];
}
