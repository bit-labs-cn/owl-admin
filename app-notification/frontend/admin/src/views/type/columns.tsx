import { parseJson } from "../../format";
import { channelOptions, kindOptions, labelOf, tagOf, typeStatusOptions } from "../../labels";
import type { TypeRow } from "./types";

export function createTypeColumns(): TableColumnList {
  return [
    { label: "名称", prop: "name", minWidth: 160 },
    { label: "编码", prop: "code", minWidth: 160 },
    {
      label: "种类",
      prop: "kind",
      width: 110,
      cellRenderer: ({ row }) => (
        <el-tag type={tagOf(kindOptions, (row as TypeRow).kind)}>{labelOf(kindOptions, (row as TypeRow).kind)}</el-tag>
      )
    },
    {
      label: "默认渠道",
      prop: "defaultChannels",
      minWidth: 180,
      cellRenderer: ({ row }) =>
        parseJson<string[]>((row as TypeRow).defaultChannels, [])
          .map(code => labelOf(channelOptions, code))
          .join("、") || "无"
    },
    {
      label: "状态",
      prop: "status",
      width: 90,
      cellRenderer: ({ row }) => (
        <el-tag type={tagOf(typeStatusOptions, (row as TypeRow).status)}>
          {labelOf(typeStatusOptions, (row as TypeRow).status)}
        </el-tag>
      )
    },
    { label: "操作", fixed: "right", width: 160, slot: "operation" }
  ];
}
