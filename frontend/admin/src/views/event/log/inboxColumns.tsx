import {
  tableIndexColumn,
  type TablePaginationInput
} from "@bit-labs.cn/owl-ui/utils/tableIndexColumn";
import { formatUnix } from "../format";
import { inboxStatusTag } from "../status";
import type { EventInboxRow } from "../types";

export function createInboxColumns(
  pagination?: TablePaginationInput
): TableColumnList {
  return [
    tableIndexColumn(pagination),
    { label: "事件ID", prop: "eventId", minWidth: 240, align: "left" },
    { label: "消费组", prop: "consumerGroup", minWidth: 220, align: "left" },
    {
      label: "状态",
      prop: "status",
      minWidth: 110,
      cellRenderer: ({ row }: { row: EventInboxRow }) => (
        <el-tag type={inboxStatusTag(row.status)}>{row.statusLabel}</el-tag>
      )
    },
    {
      label: "消费完成时间",
      prop: "processedAt",
      minWidth: 170,
      formatter: ({ processedAt }: EventInboxRow) => formatUnix(processedAt)
    },
    {
      label: "写入时间",
      prop: "createdAt",
      minWidth: 170,
      formatter: ({ createdAt }: EventInboxRow) => formatUnix(createdAt)
    },
    {
      label: "错误信息",
      prop: "errorMsg",
      minWidth: 200,
      align: "left",
      formatter: ({ errorMsg }: EventInboxRow) => errorMsg || "—"
    },
    {
      label: "操作",
      fixed: "right",
      width: 100,
      align: "center",
      showOverflowTooltip: false,
      slot: "operation"
    }
  ];
}
