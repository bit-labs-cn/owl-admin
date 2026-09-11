import dayjs from "dayjs";
import type { Ref } from "vue";
import { tableIndexColumn, type TablePaginationInput } from "@bit-labs.cn/owl-ui/utils/tableIndexColumn";

function apkTypeLabel(apkType: number) {
  if (apkType === 1) return "iOS";
  if (apkType === 2) return "Android";
  return "-";
}

export function createColumns({
  switchLoadMap,
  onChange,
  pagination
}: {
  switchLoadMap: Ref<Record<number, { loading: boolean }>>;
  onChange: (scope: any) => void;
  pagination?: TablePaginationInput;
}): TableColumnList {
  return [
    tableIndexColumn(pagination),
    { label: "版本号", prop: "version", minWidth: 120 },
    { label: "版本名称", prop: "versionName", minWidth: 140 },
    {
      label: "平台",
      prop: "apkType",
      width: 100,
      cellRenderer: ({ row }) => (
        <el-tag type={row.apkType === 1 ? "primary" : "success"}>
          {apkTypeLabel(row.apkType)}
        </el-tag>
      )
    },
    {
      label: "安装包",
      prop: "apkUrl",
      minWidth: 120,
      cellRenderer: ({ row }) =>
        row.apkUrl ? (
          <a href={row.apkUrl} target="_blank" rel="noreferrer">
            下载
          </a>
        ) : (
          "-"
        )
    },
    { label: "更新内容", prop: "content", minWidth: 200, showOverflowTooltip: true },
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
      prop: "createTime",
      minWidth: 160,
      formatter: ({ createTime }) =>
        createTime ? dayjs(createTime).format("YYYY-MM-DD HH:mm:ss") : "-"
    },
    { label: "操作", fixed: "right", width: 180, slot: "operation" }
  ];
}
