import dayjs from "dayjs";
import type { Ref } from "vue";
import { tableIndexColumn, type TablePaginationInput } from "@bit-labs.cn/owl-ui/utils/tableIndexColumn";

export function createColumns({
  switchLoadMap,
  switchStyle,
  onBeforeStatusChange,
  pagination
}: {
  switchLoadMap: Ref<Record<number, { loading: boolean }>>;
  switchStyle: Ref<Record<string, string>>;
  onBeforeStatusChange: (scope: any) => Promise<boolean>;
  pagination?: TablePaginationInput;
}): TableColumnList {
  return [
    tableIndexColumn(pagination),
    { label: "用户组名称", prop: "name", minWidth: 120 },
    { label: "用户组编码", prop: "code", minWidth: 120 },
    { label: "排序", prop: "sort", minWidth: 70 },
    {
      label: "状态",
      cellRenderer: scope => (
        <el-switch
          size={scope.props.size === "small" ? "small" : "default"}
          loading={switchLoadMap.value[scope.index]?.loading}
          v-model={scope.row.status}
          active-value={1}
          inactive-value={2}
          active-text="已启用"
          inactive-text="已停用"
          inline-prompt
          style={switchStyle.value}
          before-change={() => onBeforeStatusChange(scope as any)}
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
    { label: "操作", fixed: "right", width: 250, slot: "operation" }
  ];
}
