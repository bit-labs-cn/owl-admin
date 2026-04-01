# 权限、角色与菜单模型

本篇说明本包在**角色-菜单分配**与**按钮权限（auths）**上的职责，以及实际校验由谁执行。

## 概念说明

- **本包职责**：提供角色管理、菜单管理、用户管理等**配置界面**，即维护“角色拥有哪些菜单/按钮”“菜单的 auths 是什么”等元数据；不负责在运行时拦截路由或隐藏按钮。
- **运行时执行**：路由与菜单的访问控制、按钮的显隐（v-auth/auths）由 **owl-ui** 完成：菜单按用户 roles 过滤，按钮按路由 meta.auths 或用户 permissions 判断。
- **auths**：菜单项上的 `auths` 表示该菜单下按钮级权限标识，后端返回给前端的菜单数据中带 auths，前端用其做 v-auth 等控制；本包内的菜单编辑表单（如 MenuForm）会编辑这些字段并提交给后端。

## 代码入口

- 角色-菜单分配界面：`src/views/role/`（角色列表、分配菜单）
- 菜单编辑与 auths：`src/views/menu/MenuForm.vue`、`rules.ts`、`README.md`（字段含义）
- 实际权限校验：在 owl-ui 的 router、directives/auth、directives/perms 中

## 注意事项

- 本包只做“配置端”；权限的“执行端”在 owl-ui，新增页面时只需保证后端菜单/角色数据正确，前端按 meta.auths/roles 和 v-auth 使用即可。
- 不要在本包内重复实现一套权限校验逻辑，应依赖 owl-ui 的 store 与指令。

## 完成定义

- 读者能区分本包（配置）与 owl-ui（执行），并知道 auths 的编辑与使用流程。
