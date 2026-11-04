import dayjs from "dayjs";
import { tableIndexColumn, type TablePaginationInput } from "@bit-labs.cn/owl-ui/utils/tableIndexColumn";

export function createVersionColumns(pagination?: TablePaginationInput): TableColumnList {
  return [
    tableIndexColumn(pagination),
    { label: "版本号", prop: "version", minWidth: 110 },
    { label: "标题", prop: "title", minWidth: 160, align: "left" },
    { label: "状态", prop: "statusLabel", minWidth: 90 },
    {
      label: "变更",
      prop: "changeLabel",
      minWidth: 100,
      formatter: ({ changeLabel }) => changeLabel || "—"
    },
    {
      label: "发布时间",
      prop: "publishedAt",
      minWidth: 170,
      formatter: ({ publishedAt }) => (publishedAt ? dayjs(publishedAt).format("YYYY-MM-DD HH:mm:ss") : "—")
    },
    { label: "操作", fixed: "right", width: 280, slot: "operation" }
  ];
}
