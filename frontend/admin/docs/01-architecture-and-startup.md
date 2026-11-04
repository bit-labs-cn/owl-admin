# 架构与启动

本篇说明 Owl-Admin-UI 的**包定位**、**自动注册方式**以及宿主如何接入。

## 概念说明

- **Owl-Admin-UI 是子系统包**：通过 `defineSubsystem()` 导出子系统定义。其它业务宿主安装后，在 `createFlexAdmin({ subsystems: [adminSubsystem] })` 时注册。
- **本仓库也可以单独启动**：`frontend/admin/main.ts` 是宿主入口，只注册本包。在 `frontend/admin` 执行 `pnpm dev`（构建用 `pnpm build`）。脚本走同级 `owl-ui-builder`。默认地址 `http://localhost:8848/`，接口代理到 `VITE_BASE_URL`（默认 `http://localhost:8080/`）。子系统本体仍只从 `src/index.ts` 导出。
- **视图自动发现**：`viewModules: import.meta.glob("./views/**/*.{vue,tsx}")` 会收集 `src/views/` 下所有页面，键经 `viewModulesPathPrefix: "/system"` 重写后，与后端返回的菜单 component/path（如 `/system/position/index`）匹配，用于动态路由组件解析。
- **本包不定义业务静态路由**：`src/routes/index.ts` 当前为空数组，所有业务路由由后端菜单 + 动态路由注入，页面通过 viewModules 自动挂载。

## 代码入口

- 子系统定义与导出：`src/index.ts`（`defineSubsystem({ name: "admin", viewModulesPathPrefix: "/system", viewModules: import.meta.glob(...) })`）
- 本仓库宿主入口：`main.ts`（`createFlexAdmin({ subsystems: [adminSubsystem] })`，来自 `@bit-labs.cn/owl-ui/bootstrap`）
- 独立启动配置：`.env` / `.env.development` / `.env.production`、`public/platform-config.json`
- 静态路由（当前为空）：`src/routes/index.ts`
- 视图目录：`src/views/`（user、role、dept、position、menu、dict、api、login-log、operation-log、tools/password、tools/server、tools/codec 等）

## 宿主接入

- 宿主安装 `@bit-labs.cn/owl-admin-ui`，在入口中：
  - `import adminSubsystem from "@bit-labs.cn/owl-admin-ui"`
  - `createFlexAdmin({ subsystems: [adminSubsystem] })`
- 后端菜单接口返回的菜单树中，每条需要对应本包页面的项，其 `component` 或 path 需为 `/system/xxx/...` 形式（与 glob 键经 prefix 重写后的路径一致），且 `name` 与对应页面组件的 `defineOptions({ name })` 一致。

## 注意事项

- 新页面不需要在本包 `routes/index.ts` 里手工注册；只需在 `views/` 下新增文件，并保证后端菜单的 component/name 与之一致。
- 本包无 Pinia store。页面状态用 ref/reactive/computed、useXList。公开配置由 owl-ui 在启动时请求 `GET /api/v1/setting/public` 写入 `config/settings`。用 `settingValue("admin.site.siteName")` 读取。设置页保存后会再请求一次，再用完整配置覆盖缓存。

## 完成定义

- 读者能说清本包是“子系统包 + 视图自动发现”，以及宿主如何接入、后端菜单如何对应到视图。
- 读者能在 `frontend/admin` 用 `pnpm dev` 单独启动本仓库前端。
