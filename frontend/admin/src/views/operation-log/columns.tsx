import dayjs from "dayjs";

export const columns: TableColumnList = [
  { label: "ID", prop: "id", width: 90 },
  { label: "操作人", prop: "userName", minWidth: 120 },
  { label: "请求方式", prop: "method", minWidth: 100 },
  { label: "接口地址", prop: "path", minWidth: 160, align: "left" },
  { label: "接口名称", prop: "apiName", minWidth: 160, align: "left" },
  { label: "耗时(ms)", prop: "costMs", minWidth: 100 },
  { label: "UserAgent", prop: "userAgent", minWidth: 100 },
  { label: "Ip地址", prop: "ip", minWidth: 100 },
  { label: "状态", prop: "status", minWidth: 90 },
  { label: "请求参数", prop: "reqBody", align: "left", minWidth: 160 },
  {
    label: "操作时间",
    prop: "createTime",
    minWidth: 160,
    formatter: ({ createTime }) =>
      dayjs(createTime).format("YYYY-MM-DD HH:mm:ss")
  }
];
