import dayjs from "dayjs";
import type { Ref } from "vue";

export function createColumns({
  switchLoadMap,
  onChange
}: {
  switchLoadMap: Ref<Record<number, { loading: boolean }>>;
  onChange: (scope: any) => void;
}): TableColumnList {
  return [
    { label: "岗位编号", prop: "id", width: 100 },
    { label: "岗位名称", prop: "name", minWidth: 160 },
    { label: "备注", prop: "remark", align: "left", minWidth: 200 },
    {
      label: "状态",
      prop: "status",
      minWidth: 100,
      cellRenderer: scope => (
        <el-switch
          size={scope.props.size === "small" ? "small" : "default"}
          loading={switchLoadMap.value[scope.index]?.loading}
          v-model={scope.row.status}
          active-value={1}
          inactive-value={0}
          active-text="已启用"
          inactive-text="已停用"
          inline-prompt
          onChange={() => onChange(scope as any)}
        />
      )
    },
    {
      label: "创建时间",
      prop: "createdAt",
      minWidth: 160,
      formatter: ({ createdAt }) =>
        dayjs(createdAt).format("YYYY-MM-DD HH:mm:ss")
    },
    { label: "操作", fixed: "right", width: 220, slot: "operation" }
  ];
}
