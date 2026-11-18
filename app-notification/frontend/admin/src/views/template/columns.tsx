import type { TemplateRow, TypeOption } from "./types";

export function createTemplateColumns(typeOptions: () => TypeOption[]): TableColumnList {
  return [
    {
      label: "消息类型",
      prop: "typeCode",
      minWidth: 160,
      cellRenderer: ({ row }) => {
        const code = (row as TemplateRow).typeCode;
        const found = typeOptions().find(item => item.code === code);
        return found?.name || `未知类型(${code})`;
      }
    },
    { label: "标题模板", prop: "title", minWidth: 220 },
    { label: "跳转链接模板", prop: "link", minWidth: 200 },
    { label: "操作", fixed: "right", width: 160, slot: "operation" }
  ];
}
