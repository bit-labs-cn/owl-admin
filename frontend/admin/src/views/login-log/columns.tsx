import dayjs from "dayjs";

export const columns: TableColumnList = [
  { label: "ID", prop: "id", width: 90 },
  { label: "用户名", prop: "userName", minWidth: 120 },
  { label: "登录IP", prop: "ip", minWidth: 120 },
  { label: "地点", prop: "location", minWidth: 140 },
  { label: "设备UA", prop: "userAgent", minWidth: 160, align: "left" },
  {
    label: "登录时间",
    prop: "createTime",
    minWidth: 160,
    formatter: ({ createTime }) =>
      dayjs(createTime).format("YYYY-MM-DD HH:mm:ss")
  }
];
