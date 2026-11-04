import type { Ref } from "vue";
import { tableIndexColumn, type TablePaginationInput } from "@bit-labs.cn/owl-ui/utils/tableIndexColumn";
import type { LegalDocumentRow } from "./types";

export function createDocumentColumns({
  switchLoadMap,
  onChange,
  pagination
}: {
  switchLoadMap: Ref<Record<number, { loading: boolean }>>;
  onChange: (scope: { row: LegalDocumentRow; index: number }) => void;
  pagination?: TablePaginationInput;
}): TableColumnList {
  return [
    tableIndexColumn(pagination),
    { label: "名称", prop: "name", minWidth: 160, align: "left" },
    { label: "标识", prop: "key", minWidth: 140 },
    { label: "性质", prop: "kindLabel", minWidth: 100 },
    { label: "当前版本", prop: "currentVersion", minWidth: 110, formatter: ({ currentVersion }) => currentVersion || "未发布" },
    {
      label: "所在场景",
      prop: "sceneNames",
      minWidth: 180,
      align: "left",
      formatter: ({ sceneNames }) => (sceneNames?.length ? sceneNames.join("、") : "未挂场景")
    },
    {
      label: "状态",
      prop: "status",
      minWidth: 100,
      cellRenderer: scope => (
        <el-switch
          loading={switchLoadMap.value[scope.index]?.loading}
          v-model={scope.row.status}
          active-value="enabled"
          inactive-value="disabled"
          active-text="启用"
          inactive-text="停用"
          inline-prompt
          onChange={() => onChange(scope as any)}
        />
      )
    },
    { label: "操作", fixed: "right", width: 160, slot: "operation" }
  ];
}
