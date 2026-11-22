import {
  tableIndexColumn,
  type TablePaginationInput
} from "@bit-labs.cn/owl-ui/utils/tableIndexColumn";
import {
  formatDuration,
  formatSchedulerTime,
  runStatusTag
} from "./constants";
import type { SchedulerRunRow } from "./types";

export function createRunColumns(
  pagination?: TablePaginationInput
): TableColumnList {
  return [
    tableIndexColumn(pagination),
    {
      label: "任务",
      prop: "jobName",
      minWidth: 220,
      align: "left",
      cellRenderer: ({ row }: { row: SchedulerRunRow }) => (
        <div class="flex flex-col gap-1">
          <span class="font-medium">{row.jobName || "—"}</span>
          <span class="text-xs text-[var(--el-text-color-secondary)]">{row.jobCode}</span>
        </div>
      )
    },
    {
      label: "触发方式",
      prop: "triggerType",
      minWidth: 100,
      formatter: ({ triggerTypeText, triggerType }: SchedulerRunRow) =>
        triggerTypeText || `未知(${triggerType})`
    },
    {
      label: "状态",
      prop: "status",
      minWidth: 100,
      cellRenderer: ({ row }: { row: SchedulerRunRow }) => (
        <el-tag size="small" type={runStatusTag(row.status)}>
          {row.statusText || "—"}
        </el-tag>
      )
    },
    {
      label: "计划时间",
      prop: "plannedAt",
      minWidth: 170,
      formatter: ({ plannedAt }: SchedulerRunRow) => formatSchedulerTime(plannedAt)
    },
    {
      label: "开始时间",
      prop: "startedAt",
      minWidth: 170,
      formatter: ({ startedAt }: SchedulerRunRow) => formatSchedulerTime(startedAt)
    },
    {
      label: "耗时",
      prop: "durationMs",
      minWidth: 100,
      formatter: ({ durationMs }: SchedulerRunRow) => formatDuration(durationMs)
    },
    {
      label: "尝试",
      prop: "attempt",
      minWidth: 90,
      formatter: ({ attemptText, attempt, maxAttempts }: SchedulerRunRow) =>
        attemptText || `${attempt}/${maxAttempts}`
    },
    {
      label: "错误摘要",
      prop: "errorMsg",
      minWidth: 200,
      align: "left",
      showOverflowTooltip: true,
      formatter: ({ errorMsg }: SchedulerRunRow) => errorMsg || "—"
    },
    {
      label: "操作",
      fixed: "right",
      width: 230,
      align: "center",
      showOverflowTooltip: false,
      slot: "operation"
    }
  ];
}
