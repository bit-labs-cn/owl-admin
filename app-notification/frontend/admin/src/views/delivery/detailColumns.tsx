import { formatTime } from "../../format";
import { RecipientTypeBroadcast, RecipientTypeExternal } from "../../constants";
import { channelOptions, deliveryStatusOptions, labelOf, tagOf } from "../../labels";
import type { DeliveryRow } from "./types";

export function createDeliveryDetailColumns(): TableColumnList {
  return [
    {
      label: "渠道",
      prop: "channel",
      width: 110,
      cellRenderer: ({ row }) => labelOf(channelOptions, (row as DeliveryRow).channel)
    },
    { label: "接收人", prop: "userName", minWidth: 150 },
    {
      label: "接收人类型",
      prop: "recipientType",
      width: 110,
      cellRenderer: ({ row }) => {
        const item = row as DeliveryRow;
        if (item.recipientType === RecipientTypeExternal) {
          return <el-tag type="warning">外部人员</el-tag>;
        }
        if (item.recipientType === RecipientTypeBroadcast) {
          return <el-tag type="info">群机器人</el-tag>;
        }
        return <el-tag>站内用户</el-tag>;
      }
    },
    {
      label: "状态",
      prop: "status",
      width: 90,
      cellRenderer: ({ row }) => (
        <el-tag type={tagOf(deliveryStatusOptions, (row as DeliveryRow).status)}>
          {labelOf(deliveryStatusOptions, (row as DeliveryRow).status)}
        </el-tag>
      )
    },
    { label: "次数", prop: "attempts", width: 70 },
    { label: "最后错误", prop: "lastError", minWidth: 180 },
    {
      label: "时间",
      prop: "createdAt",
      minWidth: 165,
      cellRenderer: ({ row }) => formatTime((row as DeliveryRow).createdAt)
    },
    { label: "操作", fixed: "right", width: 90, slot: "operation" }
  ];
}
