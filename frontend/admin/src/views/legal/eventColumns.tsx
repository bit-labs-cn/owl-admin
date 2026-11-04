import dayjs from "dayjs";
import { tableIndexColumn, type TablePaginationInput } from "@bit-labs.cn/owl-ui/utils/tableIndexColumn";
import type { LegalEventRow } from "./types";

export function createEventColumns(
  pagination: TablePaginationInput | undefined,
  openUser: (row: LegalEventRow) => void
): TableColumnList {
  return [
    tableIndexColumn(pagination),
    {
      label: "用户",
      prop: "userLabel",
      minWidth: 140,
      cellRenderer: ({ row }) => (
        <el-button link type="primary" onClick={() => openUser(row)}>
          {row.userLabel || "未知用户"}
        </el-button>
      )
    },
    { label: "条款", prop: "documentName", minWidth: 160, align: "left" },
    { label: "版本", prop: "version", minWidth: 100 },
    { label: "动作", prop: "actionLabel", minWidth: 90 },
    { label: "场景", prop: "sceneName", minWidth: 120 },
    {
      label: "时间",
      prop: "createdAt",
      minWidth: 170,
      formatter: ({ createdAt }) => (createdAt ? dayjs(createdAt).format("YYYY-MM-DD HH:mm:ss") : "—")
    }
  ];
}
