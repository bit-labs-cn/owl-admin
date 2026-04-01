# 路由、菜单与视图契约

本篇说明本包的 **/system 前缀**、后端菜单字段含义，以及页面 **defineOptions({ name })** 与菜单 **name** 必须一致的约定。

## 概念说明

- **路径前缀**：本包 `viewModulesPathPrefix` 为 `/system`。后端返回的菜单中，对应本包页面的项，其 `component` 或 `path` 应为 `/system/xxx/...`（如 `/system/position/index`），以便框架的 `addAsyncRoutes` 能匹配到本包 viewModules 重写后的键。
- **菜单 name**：后端菜单项的 `name` 必须与对应 Vue 页面的 `defineOptions({ name })` **完全一致**（如 `SystemPosition`），否则多标签、keep-alive 与路由高亮可能异常。
- **字段含义**：菜单类型、parentId、title、path、component、rank、auths、keepAlive、showLink 等见本包内 `src/views/menu/README.md`，与 owl-ui 的 CustomizeRouteMeta 对齐。

## 代码入口

- 菜单字段说明：`src/views/menu/README.md`
- 本包子系统定义：`src/index.ts`（viewModulesPathPrefix: "/system"）
- 视图与路由解析：由 owl-ui 的 `addAsyncRoutes` 与 `getModulesRoutes()` 完成，本包只提供 viewModules

## 标准约定

- 新增页面时，先在 `views/xxx/index.vue` 中写死 `defineOptions({ name: "SystemXxx" })`，再在后端菜单中配置同名 `name`、path/component 为 `/system/xxx/index`。
- 若不传 `component`，框架会按 `path` 匹配 viewModules；传了 `component` 则按 component 字符串匹配。两者都需与 prefix 后的视图路径一致。

## 注意事项

- **name 不一致**是隐蔽问题：页面能打开，但多标签或缓存可能错乱，AI 生成新页时务必与后端菜单 name 对齐。
- 本包不在此定义静态路由表，所有业务路由来自后端 + 动态注入。

## 完成定义

- 读者能说清 /system 前缀、菜单 name 与 defineOptions name 的一致性要求，以及菜单字段的用途。
