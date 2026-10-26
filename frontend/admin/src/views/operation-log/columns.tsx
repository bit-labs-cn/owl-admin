import dayjs from "dayjs";
import { tableIndexColumn, type TablePaginationInput } from "@bit-labs.cn/owl-ui/utils/tableIndexColumn";

export function statusTagType(status: number) {
  if (status >= 200 && status < 300) return "success";
  if (status >= 300 && status < 400) return "primary";
  if (status >= 400 && status < 500) return "warning";
  if (status >= 500) return "danger";
  return "info";
}

export function formatOperateTime(value: unknown) {
  const n = typeof value === "number" ? value : Number(value);
  if (!Number.isFinite(n) || n <= 0) return "—";
  const ms = n < 1e12 ? n * 1000 : n;
  const formatted = dayjs(ms);
  return formatted.isValid() ? formatted.format("YYYY-MM-DD HH:mm:ss") : "—";
}

export function createColumns(pagination?: TablePaginationInput): TableColumnList {
  return [
    tableIndexColumn(pagination),
    { label: "ID", prop: "id", width: 90 },
    { label: "操作人", prop: "userName", minWidth: 120 },
    { label: "请求方式", prop: "method", minWidth: 100 },
    { label: "接口地址", prop: "path", minWidth: 160, align: "left" },
    { label: "接口名称", prop: "apiName", minWidth: 160, align: "left" },
    { label: "耗时(ms)", prop: "costMs", minWidth: 100 },
    { label: "UserAgent", prop: "userAgent", minWidth: 100 },
    { label: "Ip地址", prop: "ip", minWidth: 100 },
    {
      label: "状态",
      prop: "status",
      minWidth: 90,
      cellRenderer: ({ row }) => (
        <el-tag type={statusTagType(Number(row.status))}>{row.status}</el-tag>
      )
    },
    { label: "请求参数", prop: "reqBody", align: "left", minWidth: 160 },
    {
      label: "操作时间",
      prop: "createdAt",
      minWidth: 160,
      formatter: ({ createdAt }) => formatOperateTime(createdAt)
    },
    {
      label: "操作",
      fixed: "right",
      width: 120,
      align: "center",
      showOverflowTooltip: false,
      slot: "operation"
    }
  ];
}
