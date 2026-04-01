# Owl-Admin-UI 开发文档

本目录为 **@bit-labs.cn/owl-admin-ui** 的开发者手册，面向 AI 开发工具与人工开发者。Owl-Admin-UI 是**后台前端子系统包**，依赖 `@bit-labs.cn/owl-ui`，通过 `defineSubsystem()` 注册，提供后台管理相关视图与 API；路由与菜单由宿主 + 后端驱动，本包不定义静态业务路由。

**适用范围**：本文档适用于**在 owl-admin-ui 包内扩展**（新增/修改后台管理页面、API、目录规范）。若要新建**独立前端子系统包**（与 owl-admin-ui 平级的另一个 npm 包，由宿主通过 `createFlexAdmin({ subsystems: [yourSubsystem] })` 注册），请阅读 **owl-ui** 的 `docs/03-subsystem-contract.md` 与 `docs/08-minimal-subsystem-template.md`。

## 阅读顺序

建议按以下顺序阅读：

| 顺序 | 文档 | 说明 |
|------|------|------|
| 1 | [01-architecture-and-startup.md](01-architecture-and-startup.md) | 包定位、自动注册与宿主接入 |
| 2 | [02-feature-folder-pattern.md](02-feature-folder-pattern.md) | 标准目录：index.vue + useXList + columns + Form + types/rules |
| 3 | [03-routing-menu-view-contract.md](03-routing-menu-view-contract.md) | /system 前缀、菜单字段、defineOptions name 与菜单 name 一致 |
| 4 | [04-api-and-data-fetching.md](04-api-and-data-fetching.md) | API 分层、分页与请求风格 |
| 5 | [05-ui-patterns-tables-dialogs-forms.md](05-ui-patterns-tables-dialogs-forms.md) | PureTableBar、addDialog、contentRenderer、getRef 校验闭环 |
| 6 | [06-permission-role-menu-model.md](06-permission-role-menu-model.md) | 角色-菜单、auths、本包维护 vs owl-ui 执行 |
| 7 | [07-canonical-examples-and-ai-guardrails.md](07-canonical-examples-and-ai-guardrails.md) | 标准/复杂模板、勿照抄实现、AI 自检清单 |

## 术语表

| 术语 | 含义 |
|------|------|
| **viewModulesPathPrefix** | 本包为 `/system`，后端菜单 component/path 需与该前缀下的视图路径一致 |
| **defineOptions({ name })** | 页面组件 name 必须与后端菜单项 `name` 一致，否则 keep-alive/多标签可能异常 |
| **getRef** | 表单子组件暴露的方法，供父页在 addDialog 的 beforeSure 中做校验与提交 |

## 相关项目与文档索引

- **owl-ui**：本包依赖的前端框架。路由、权限、布局、http 等见 owl-ui 的 `docs/`。
- **owl-admin**：后台 API 与菜单/角色数据来源。新增后端模块后需同步配置菜单与 component 路径。
- **owl**：后端应用框架。从零创建后端子系统时见 `owl/docs/`。

## 四仓阅读路径（全栈新子系统）

后端独立子系统：owl/docs。后台后端新模块：owl-admin/docs。前端框架与子系统：owl-ui/docs。后台前端新页面：本目录（01/02/03/07 优先）。
