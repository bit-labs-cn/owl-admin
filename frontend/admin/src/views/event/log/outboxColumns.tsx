import {
  tableIndexColumn,
  type TablePaginationInput
} from "@bit-labs.cn/owl-ui/utils/tableIndexColumn";
import { formatUnix } from "../format";
import { eventPriorityLabel, eventPriorityTag, outboxStatusTag } from "../status";
import type { EventOutboxRow } from "../types";

export function createOutboxColumns(
  pagination?: TablePaginationInput
): TableColumnList {
  return [
    tableIndexColumn(pagination),
    { label: "事件类型", prop: "eventType", minWidth: 200, align: "left" },
    { label: "事件ID", prop: "eventId", minWidth: 220, align: "left" },
    { label: "来源", prop: "source", minWidth: 120 },
    { label: "业务实体", prop: "subject", minWidth: 140, align: "left" },
    {
      label: "优先级",
      prop: "priority",
      minWidth: 100,
      cellRenderer: ({ row }: { row: EventOutboxRow }) => (
        <el-tag type={eventPriorityTag(row.priority)}>
          {eventPriorityLabel(row.priority)}
        </el-tag>
      )
    },
    {
      label: "状态",
      prop: "status",
      minWidth: 100,
      cellRenderer: ({ row }: { row: EventOutboxRow }) => (
        <el-tag type={outboxStatusTag(row.status)}>{row.statusLabel}</el-tag>
      )
    },
    { label: "重试次数", prop: "retryCount", minWidth: 90 },
    {
      label: "创建时间",
      prop: "createdAt",
      minWidth: 170,
      formatter: ({ createdAt }: EventOutboxRow) => formatUnix(createdAt)
    },
    {
      label: "投递时间",
      prop: "publishedAt",
      minWidth: 170,
      formatter: ({ publishedAt }: EventOutboxRow) => formatUnix(publishedAt)
    },
    {
      label: "操作",
      fixed: "right",
      width: 150,
      align: "center",
      showOverflowTooltip: false,
      slot: "operation"
    }
  ];
}
