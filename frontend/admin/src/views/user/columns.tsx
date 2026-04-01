import dayjs from "dayjs";
import userAvatar from "@bit-labs.cn/owl-ui/assets/user.jpg";
import type { Ref } from "vue";

export function createColumns({
  switchLoadMap,
  switchStyle,
  onChange
}: {
  switchLoadMap: Ref<Record<number, { loading: boolean }>>;
  switchStyle: Ref<Record<string, string>>;
  onChange: (scope: any) => void;
}): TableColumnList {
  return [
    {
      label: "勾选列",
      type: "selection",
      fixed: "left",
      reserveSelection: true
    },
    {
      label: "用户头像",
      prop: "avatar",
      cellRenderer: ({ row }) => (
        <el-image
          fit="cover"
          preview-teleported={true}
          src={row.avatar || userAvatar}
          preview-src-list={Array.of(row.avatar || userAvatar)}
          class="w-[24px] h-[24px] rounded-full align-middle"
        />
      ),
      width: 90
    },
    { label: "用户昵称", prop: "nickname", minWidth: 130 },
    { label: "账号", prop: "username", minWidth: 130 },
    {
      label: "角色",
      prop: "roles",
      minWidth: 130,
      cellRenderer: ({ row }) => {
        const roles = row.roles;
        if (roles.length === 0) {
          return <el-tag type="info">未分配角色</el-tag>;
        }
        return (
          <>
            {roles.map((role, index) => (
              <span>
                <el-tag>{role.name}</el-tag>
                {index !== roles.length - 1 && " "}
              </span>
            ))}
          </>
        );
      }
    },
    {
      label: "部门",
      prop: "depts",
      minWidth: 130,
      cellRenderer: ({ row }) => {
        const depts = row.depts ?? [];
        if (depts.length === 0) return <span class="text-gray-400">—</span>;
        return depts.map(d => d.name).join("、");
      }
    },
    {
      label: "性别",
      prop: "sex",
      minWidth: 90,
      cellRenderer: ({ row, props }) => (
        <el-tag
          size={props.size}
          type={row.sex === 1 ? "danger" : null}
          effect="plain"
        >
          {row.sex === 1 ? "女" : "男"}
        </el-tag>
      )
    },
    { label: "手机号码", prop: "phone", minWidth: 90 },
    {
      label: "状态",
      prop: "status",
      minWidth: 90,
      cellRenderer: scope => (
        <el-switch
          size={scope.props.size === "small" ? "small" : "default"}
          loading={switchLoadMap.value[scope.index]?.loading}
          v-model={scope.row.status}
          active-value={1}
          inactive-value={0}
          active-text="已启用"
          inactive-text="已停用"
          inline-prompt
          style={switchStyle.value}
          onChange={() => onChange(scope as any)}
        />
      )
    },
    {
      label: "创建时间",
      minWidth: 90,
      prop: "createTime",
      formatter: ({ createTime }) =>
        dayjs(createTime).format("YYYY-MM-DD HH:mm:ss")
    },
    { label: "操作", fixed: "right", width: 280, slot: "operation" }
  ];
}
