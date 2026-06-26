import { tableIndexColumn, type TablePaginationInput } from "@bit-labs.cn/owl-ui/utils/tableIndexColumn";

export function createColumns(pagination?: TablePaginationInput): TableColumnList {
  return [
    tableIndexColumn(pagination),
    { label: "ID", prop: "id", width: 90 },
    { label: "用户名", prop: "userName", minWidth: 120 },
    { label: "登录IP", prop: "ip", minWidth: 120 },
    { label: "地点", prop: "location", minWidth: 140 },
    { label: "设备UA", prop: "userAgent", minWidth: 160, align: "left" },
    {
      label: "登录时间",
      prop: "loginTime",
      minWidth: 160
    }
  ];
}
