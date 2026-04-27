import dayjs from "dayjs";
import type { Ref } from "vue";

export function createColumns({
  switchLoadMap,
  switchStyle,
  onChange
}: {
  switchLoadMap: Ref<Record<number, { loading: boolean }>>;
  switchStyle: Ref<Record<string, string>>;
  onChange: (scope: any) => void;
}): TableColumnList {
  return [
    { label: "角色名称", prop: "name" },
    { label: "角色标识", prop: "code" },
    {
      label: "状态",
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
          style={switchStyle.value}
          onChange={() => onChange(scope as any)}
        />
      ),
      minWidth: 90
    },
    { label: "描述", prop: "remark", minWidth: 160 },
    {
      label: "创建时间",
      prop: "createdAt",
      minWidth: 160,
      formatter: ({ createdAt }) =>
        dayjs(createdAt).format("YYYY-MM-DD HH:mm:ss")
    },
    { label: "操作", fixed: "right", minWidth: 300, slot: "operation" }
  ];
}
