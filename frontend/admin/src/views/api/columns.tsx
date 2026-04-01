export const columns: TableColumnList = [
  { label: "模块", prop: "module" },
  { label: "请求方式", prop: "method" },
  { label: "路由分组", prop: "group" },
  { label: "接口地址", prop: "path" },
  { label: "接口名称", prop: "name", align: "left" },
  { label: "接口描述", prop: "description", align: "left" },
  { label: "授权码", prop: "permission" },
  {
    label: "是否需要授权",
    prop: "accessLevel",
    cellRenderer: ({ row }) => (
      <>
        <el-tag v-show={row.accessLevel === "仅超管"} type="danger">
          {row.accessLevel}
        </el-tag>
        <el-tag v-show={row.accessLevel === "需要授权"} type="warning">
          {row.accessLevel}
        </el-tag>
        <el-tag v-show={row.accessLevel === "需要登录"} type="primary">
          {row.accessLevel}
        </el-tag>
        <el-tag v-show={row.accessLevel === "开放接口"} type="success">
          {row.accessLevel}
        </el-tag>
      </>
    )
  }
];
