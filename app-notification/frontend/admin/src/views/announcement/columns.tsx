import { formatTime } from "../../format";
import {
  announcementStatusOptions,
  displayModeOptions,
  labelOf,
  levelOptions,
  scopeOptions,
  tagOf
} from "../../labels";
import type { AnnouncementRow } from "./types";

export function createAnnouncementColumns(): TableColumnList {
  return [
    { label: "标题", prop: "title", minWidth: 200 },
    {
      label: "级别",
      prop: "level",
      width: 100,
      cellRenderer: ({ row }) => (
        <el-tag type={tagOf(levelOptions, (row as AnnouncementRow).level)}>
          {labelOf(levelOptions, (row as AnnouncementRow).level)}
        </el-tag>
      )
    },
    {
      label: "展示方式",
      prop: "displayMode",
      width: 110,
      cellRenderer: ({ row }) => labelOf(displayModeOptions, (row as AnnouncementRow).displayMode)
    },
    {
      label: "受众范围",
      prop: "targetScope",
      width: 110,
      cellRenderer: ({ row }) => labelOf(scopeOptions, (row as AnnouncementRow).targetScope)
    },
    {
      label: "状态",
      prop: "status",
      width: 100,
      cellRenderer: ({ row }) => (
        <el-tag type={tagOf(announcementStatusOptions, (row as AnnouncementRow).status)}>
          {labelOf(announcementStatusOptions, (row as AnnouncementRow).status)}
        </el-tag>
      )
    },
    {
      label: "生效时间",
      prop: "startAt",
      minWidth: 160,
      cellRenderer: ({ row }) => formatTime((row as AnnouncementRow).startAt) || "立即"
    },
    {
      label: "结束时间",
      prop: "endAt",
      minWidth: 160,
      cellRenderer: ({ row }) => formatTime((row as AnnouncementRow).endAt) || "永久"
    },
    { label: "操作", fixed: "right", width: 240, slot: "operation" }
  ];
}
