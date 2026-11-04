import { tableIndexColumn, type TablePaginationInput } from "@bit-labs.cn/owl-ui/utils/tableIndexColumn";

export function createSceneColumns(pagination?: TablePaginationInput): TableColumnList {
  return [
    tableIndexColumn(pagination),
    { label: "名称", prop: "name", minWidth: 140, align: "left" },
    { label: "标识", prop: "key", minWidth: 140 },
    { label: "用途", prop: "kindLabel", minWidth: 90 },
    {
      label: "条款",
      prop: "itemSummary",
      minWidth: 220,
      align: "left",
      formatter: ({ itemSummary }) => itemSummary || "还没有条款"
    },
    { label: "操作", fixed: "right", width: 160, slot: "operation" }
  ];
}
