# Owl-Admin 开发文档

本目录为 **owl-admin**（基于 Owl 框架的后台管理子应用）的开发者手册，面向 AI 开发工具与人工开发者。阅读后可在本项目中新增业务模块，并正确接入路由、菜单、权限、迁移、Seeder、事件与 Provider。

## 阅读顺序

建议按以下顺序阅读：

| 顺序 | 文档 | 说明 |
|------|------|------|
| 1 | [01-architecture-overview.md](01-architecture-overview.md) | 项目目录与启动链路 |
| 2 | [02-standard-module-template.md](02-standard-module-template.md) | 标准 CRUD 模块模板（position） |
| 3 | [03-routing-menu-permission.md](03-routing-menu-permission.md) | 路由、菜单与权限 |
| 4 | [04-migrate-seeder-event-provider.md](04-migrate-seeder-event-provider.md) | 迁移、Seeder、事件与 Provider |
| 5 | [05-create-new-module-playbook.md](05-create-new-module-playbook.md) | 新增业务模块操作清单 |
| 6 | [06-advanced-patterns-and-pitfalls.md](06-advanced-patterns-and-pitfalls.md) | 进阶模板与常见坑 |
| 7 | [07-example-notice-module.md](07-example-notice-module.md) | 完整可照抄的实战模块示例 |
| 8 | [08-startup-and-verification.md](08-startup-and-verification.md) | 登录、建表、接口、菜单验证闭环 |

## 相关项目与文档索引

- **owl**：本项目的应用框架。  
  **若要基于框架从零创建新的子系统（独立项目）**，请阅读 owl 的 `docs/`：应用生命周期 → SubApp 契约 → Provider 与 DI → 路由/配置/运行目录 → 创建新子系统 playbook → 常见坑与自检清单。  
  同组织下 owl 仓库路径：`owl/docs/`。
- **owl-ui**：前端框架包。路由、权限、子系统契约见 `owl-ui/docs/`。
- **owl-admin-ui**：后台前端子系统包。在后台中新增业务页面见 `owl-admin-ui/docs/`。

## 推荐工作流

如果你的目标是“直接在本项目里做出一个新模块”，建议按这个顺序：

1. 先看 [02-standard-module-template.md](02-standard-module-template.md) 理解标准分层。
2. 再按 [07-example-notice-module.md](07-example-notice-module.md) 直接照着改出一个完整模块。
3. 最后按 [08-startup-and-verification.md](08-startup-and-verification.md) 完成登录、接口与菜单验证。
