import { h } from "vue";
import { transformI18n } from "@bit-labs.cn/owl-ui/plugins/i18n";
import { useRenderIcon } from "@bit-labs.cn/owl-ui/components/ReIcon/src/hooks";

export const columns: TableColumnList = [
  {
    label: "菜单名称",
    prop: "title",
    align: "left",
    cellRenderer: ({ row }) => (
      <>
        <span class="inline-block mr-1">
          {h(useRenderIcon(row.icon), {
            style: { paddingTop: "1px" }
          })}
        </span>
        <span>{transformI18n(row.meta.title)}</span>
      </>
    )
  },
  { label: "菜单图标", prop: "meta.icon" },
  {
    label: "菜单类型",
    prop: "menuType",
    cellRenderer: ({ row }) => (
      <>
        <el-tag type="success" v-show={row.menuType == "目录"}>
          {row.menuType}
        </el-tag>
        <el-tag type="warning" v-show={row.menuType == "菜单"}>
          {row.menuType}
        </el-tag>
        <el-tag type="danger" v-show={row.menuType == "按钮"}>
          {row.menuType}
        </el-tag>
      </>
    )
  },
  { label: "前端组件", prop: "name" },
  { label: "前端路由", prop: "path" },
  {
    label: "依赖权限",
    cellRenderer: ({ row }) => (
      <div>
        {row.dependentsPermission != undefined &&
          row.dependentsPermission.map((action: any, index: number) => (
            <span key={index}>
              {action}
              <br />
            </span>
          ))}
      </div>
    )
  },
  { label: "排序", prop: "rank", width: 100 }
];
