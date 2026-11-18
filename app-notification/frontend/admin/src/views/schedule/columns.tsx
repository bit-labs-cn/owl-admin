import { formatTime } from "../../format";
import { channelOptions, kindOptions, labelOf, scheduleStatusOptions, tagOf } from "../../labels";
import { ScheduleStatusFailed, ScheduleStatusPending, ScheduleStatusSent } from "../../constants";
import type { ScheduleRow } from "./types";

export function createScheduleColumns(): TableColumnList {
  return [
    {
      label: "标题",
      prop: "title",
      minWidth: 220,
      cellRenderer: ({ row }) => {
        const item = row as ScheduleRow;
        return (
          <div class="schedule-title">
            <span>{item.title}</span>
            <span class="schedule-meta">{item.typeName || "未分类"} · {item.senderName}</span>
          </div>
        );
      }
    },
    {
      label: "种类",
      prop: "kind",
      width: 105,
      cellRenderer: ({ row }) => (
        <el-tag type={tagOf(kindOptions, (row as ScheduleRow).kind)}>
          {labelOf(kindOptions, (row as ScheduleRow).kind)}
        </el-tag>
      )
    },
    {
      label: "计划发送时间",
      prop: "planAt",
      minWidth: 170,
      cellRenderer: ({ row }) => {
        const item = row as ScheduleRow;
        return (
          <div class="schedule-title">
            <span>{formatTime(item.planAt)}</span>
            <span class="schedule-meta">
              {item.status === ScheduleStatusSent && item.sentAt
                ? `实际 ${formatTime(item.sentAt)}`
                : item.status === ScheduleStatusPending
                  ? "等待到点"
                  : ""}
            </span>
          </div>
        );
      }
    },
    {
      label: "接收人",
      prop: "audienceSummary",
      minWidth: 200,
      cellRenderer: ({ row }) => {
        const item = row as ScheduleRow;
        if (!item.audienceSummary) return <span class="muted">未知接收人</span>;
        return (
          <div class="schedule-title">
            <span>{item.audienceSummary}</span>
            <span class="schedule-meta">共 {item.recipientCount} 人</span>
          </div>
        );
      }
    },
    {
      label: "渠道",
      prop: "channels",
      minWidth: 150,
      cellRenderer: ({ row }) => {
        const channels = (row as ScheduleRow).channels;
        if (!channels || channels.length === 0) return <span class="muted">按类型默认</span>;
        return (
          <div class="schedule-tags">
            {channels.map(code => (
              <el-tag key={code} size="small" effect="plain">
                {labelOf(channelOptions, code)}
              </el-tag>
            ))}
          </div>
        );
      }
    },
    {
      label: "状态",
      prop: "status",
      width: 120,
      cellRenderer: ({ row }) => {
        const item = row as ScheduleRow;
        return (
          <div class="schedule-title">
            <el-tag type={tagOf(scheduleStatusOptions, item.status)}>
              {labelOf(scheduleStatusOptions, item.status)}
            </el-tag>
            {item.status === ScheduleStatusFailed && item.lastError ? (
              <span class="schedule-error" title={item.lastError}>
                {item.lastError}
              </span>
            ) : null}
          </div>
        );
      }
    },
    { label: "操作", fixed: "right", width: 190, slot: "operation" }
  ];
}
