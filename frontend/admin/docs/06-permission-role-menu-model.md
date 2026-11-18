# 权限、角色与菜单模型

本篇说明本包在**角色-菜单分配**与**按钮权限**上的职责，以及实际校验由谁执行。

## 概念说明

- **本包职责**：提供角色管理、菜单管理、用户管理等**配置界面**，即维护「角色拥有哪些菜单/按钮」「菜单的 auths 是什么」等元数据；同时本包自己的页面按钮也要按权限显隐。
- **运行时执行**：路由与菜单的访问控制、按钮的显隐由 **owl-ui** 完成：菜单按用户 roles 过滤，按钮按 `v-perms` / `<Perms>` 比对 `/api/v1/users/me/permissions` 返回的权限码。
- **auths**：菜单项上的 `auths` 是 pure-admin 风格的按钮权限标识，配置在菜单数据里；owl 后端实际生成并下发的是 `permission`（见下）。

## 按钮权限标识从哪来

后端 `owl/provider/router.RouterInfoBuilder.Build()` 为每个路由生成：

```
permission = <appName>:<moduleEn>:<处理函数名>
```

例如 `admin:user:ResetPassword`、`admin:legal:CreateVersion`。`appName` 是子应用名（`admin`），`moduleEn` 来自 handle 的 `ModuleName()`。`AccessAuthorized` 与 `AccessSuperAdmin` 的接口才会被前端按钮使用；`AccessAuthenticated` 只要求登录。

超管的权限列表是 `*:*:*`，`hasPerms` 会直接放行。

## 代码入口

- 权限常量：`src/permissions.ts` 的 `ADMIN_PERMS`，页面只引用它，不散落裸字符串
- 角色-菜单分配界面：`src/views/role/`（角色列表、分配菜单）
- 菜单编辑与 auths：`src/views/menu/MenuForm.vue`、`rules.ts`、`README.md`
- 实际权限判断：owl-ui 的 `router/utils.ts`（`hasAuth`）、`utils/auth.ts`（`hasPerms`）、`directives/auth`、`directives/perms`

## 用法

按钮或操作按钮用全局 `<Perms>` 包裹：

```vue
<Perms :value="ADMIN_PERMS.user.create">
  <el-button type="primary" @click="openDialog()">新增用户</el-button>
</Perms>
```

表格列里的开关这类非按钮控件，用 `hasPerms` 判断后 `disabled`：

```tsx
disabled={!hasPerms(ADMIN_PERMS.user.changeStatus)}
```

- 只读动作（详情、明细、已读情况）同样要绑权限。
- 弹窗里的按钮（导入、导出、上传、预览、重试）同样要绑。
- 不要用页面权限或 `*:*:*` 代替具体动作权限。

## 注意事项

- 改了后端处理函数名或 `ModuleName()`，必须同步改 `src/permissions.ts`，否则按钮会永久隐藏。
- 不要在本包内重复实现一套权限校验逻辑，应依赖 owl-ui 的 store 指令与 `hasPerms`。
- `AccessSuperAdmin` 的接口（字典、系统配置）不会生成按钮菜单项，非超管本就拿不到权限码，按钮会自然隐藏。

## 完成定义

- 读者能区分本包（配置 + 页面按钮）与 owl-ui（执行），知道权限码怎么生成、写在哪、怎么用。
