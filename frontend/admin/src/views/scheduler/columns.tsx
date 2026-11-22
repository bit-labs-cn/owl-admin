import type { Ref } from "vue";
import {
  tableIndexColumn,
  type TablePaginationInput
} from "@bit-labs.cn/owl-ui/utils/tableIndexColumn";
import { hasPerms } from "@bit-labs.cn/owl-ui/utils/auth";
import { ADMIN_PERMS } from "../../permissions";
import {
  SchedulerJobStatus,
  formatSchedulerTime,
  jobSourceTag,
  jobStatusTag,
  runStatusTag
} from "./constants";
import { describeCron } from "./cronPreset";
import type { SchedulerJobRow } from "./types";

export function createJobColumns({
  pagination,
  switchLoadMap,
  onChangeStatus
}: {
  pagination?: TablePaginationInput;
  switchLoadMap: Ref<Record<number, { loading: boolean }>>;
  onChangeStatus: (scope: any) => void;
}): TableColumnList {
  return [
    tableIndexColumn(pagination),
    {
      label: "任务名称",
      prop: "name",
      minWidth: 240,
      align: "left",
      cellRenderer: ({ row }: { row: SchedulerJobRow }) => (
        <div class="flex flex-col gap-1">
          <div class="flex items-center gap-2">
            <span class="font-medium">{row.name}</span>
            <el-tag size="small" type={jobSourceTag(row.source)}>
              {row.sourceText || (row.source === 1 ? "内置" : "自定义")}
            </el-tag>
          </div>
          <span class="text-xs text-[var(--el-text-color-secondary)]">{row.code}</span>
        </div>
      )
    },
    {
      label: "所属子系统",
      prop: "ownerApp",
      minWidth: 120,
      formatter: ({ ownerApp }: SchedulerJobRow) => ownerApp || "自定义"
    },
    {
      label: "调度",
      prop: "cron",
      minWidth: 200,
      align: "left",
      cellRenderer: ({ row }: { row: SchedulerJobRow }) => (
        <div class="flex flex-col gap-1">
          <span class="text-xs">{describeCron(row.cron)}</span>
          {row.scriptType === 4 && row.actionName ? (
            <span class="text-xs text-[var(--el-text-color-secondary)]">
              调用：{row.actionName}
              {row.scriptSummary ? `（${row.scriptSummary}）` : ""}
            </span>
          ) : null}
          <span class="text-xs text-[var(--el-text-color-secondary)]">
            下次：{formatSchedulerTime(row.nextRunAt)}
          </span>
        </div>
      )
    },
    {
      label: "最近执行",
      prop: "lastRunAt",
      minWidth: 170,
      cellRenderer: ({ row }: { row: SchedulerJobRow }) => (
        <div class="flex flex-col gap-1">
          {row.lastStatus == null ? (
            <span class="text-xs text-[var(--el-text-color-secondary)]">从未执行</span>
          ) : (
            <el-tag size="small" type={runStatusTag(row.lastStatus)}>
              {row.lastStatusText || "—"}
            </el-tag>
          )}
          <span class="text-xs text-[var(--el-text-color-secondary)]">
            {formatSchedulerTime(row.lastRunAt)}
          </span>
        </div>
      )
    },
    {
      label: "执行统计",
      prop: "runCount",
      minWidth: 130,
      cellRenderer: ({ row }: { row: SchedulerJobRow }) => {
        const failed = row.failCount ?? 0;
        const success = Math.max((row.runCount ?? 0) - failed, 0);
        return (
          <span class="text-xs">
            成功 {success} · 失败 {failed}
          </span>
        );
      }
    },
    {
      label: "状态",
      prop: "status",
      minWidth: 100,
      cellRenderer: ({ row, index, props }: any) => (
        <el-switch
          size={props.size === "small" ? "small" : "default"}
          loading={switchLoadMap.value[index]?.loading}
          v-model={row.status}
          active-value={SchedulerJobStatus.Enabled}
          inactive-value={SchedulerJobStatus.Disabled}
          active-text="启用"
          inactive-text="停用"
          inline-prompt
          disabled={
            row.status === SchedulerJobStatus.Retired ||
            !hasPerms(ADMIN_PERMS.scheduler.changeStatus)
          }
          onChange={() => onChangeStatus({ row, index })}
        />
      )
    },
    {
      label: "操作",
      fixed: "right",
      width: 300,
      align: "center",
      showOverflowTooltip: false,
      slot: "operation"
    }
  ];
}
