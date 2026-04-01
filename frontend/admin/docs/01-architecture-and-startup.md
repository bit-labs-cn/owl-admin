# 架构与启动

本篇说明 Owl-Admin-UI 的**包定位**、**自动注册方式**以及宿主如何接入。

## 概念说明

- **Owl-Admin-UI 是子系统包**：不是独立 SPA，而是通过 `defineSubsystem()` 导出一个子系统定义，由宿主在 `createFlexAdmin({ subsystems: [adminSubsystem] })` 时注册。
- **视图自动发现**：`viewModules: import.meta.glob("./views/**/*.{vue,tsx}")` 会收集 `src/views/` 下所有页面，键经 `viewModulesPathPrefix: "/system"` 重写后，与后端返回的菜单 component/path（如 `/system/position/index`）匹配，用于动态路由组件解析。
- **本包不定义业务静态路由**：`src/routes/index.ts` 当前为空数组，所有业务路由由后端菜单 + 动态路由注入，页面通过 viewModules 自动挂载。

## 代码入口

- 子系统定义与导出：`src/index.ts`（`defineSubsystem({ name: "admin", viewModulesPathPrefix: "/system", viewModules: import.meta.glob(...) })`）
- 静态路由（当前为空）：`src/routes/index.ts`
- 视图目录：`src/views/`（user、role、dept、position、menu、dict、api、login-log、operation-log 等）

## 宿主接入

- 宿主安装 `@bit-labs.cn/owl-admin-ui`，在入口中：
  - `import adminSubsystem from "@bit-labs.cn/owl-admin-ui"`
  - `createFlexAdmin({ subsystems: [adminSubsystem] })`
- 后端菜单接口返回的菜单树中，每条需要对应本包页面的项，其 `component` 或 path 需为 `/system/xxx/...` 形式（与 glob 键经 prefix 重写后的路径一致），且 `name` 与对应页面组件的 `defineOptions({ name })` 一致。

## 注意事项

- 新页面不需要在本包 `routes/index.ts` 里手工注册；只需在 `views/` 下新增文件，并保证后端菜单的 component/name 与之一致。
- 本包无 Pinia store，状态均为页面级（ref/reactive/computed、useXList 等 composable）。

## 完成定义

- 读者能说清本包是“子系统包 + 视图自动发现”，以及宿主如何接入、后端菜单如何对应到视图。
