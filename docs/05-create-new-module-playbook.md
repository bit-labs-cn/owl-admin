# 新增业务模块操作清单

按本清单在 owl-admin 中新增一个常规业务模块（以单表 CRUD 为例），可避免漏步骤。

如果你希望直接照着一个完整案例实现，请同时阅读：

- [07-example-notice-module.md](07-example-notice-module.md)
- [08-startup-and-verification.md](08-startup-and-verification.md)

## 1. Model

- [ ] 在 `app/model/` 新建文件，定义结构体（嵌入 `db.BaseModel`）、实现 `TableName()`。
- [ ] 若有状态字段，用常量或方法封装，避免魔法数字。

## 2. Repository

- [ ] 在 `app/repository/` 新建文件，定义 `XxxRepositoryInterface`（含 `WithContext`、Save、Detail、Retrieve 等）。
- [ ] 实现体带 `WithContext(ctx)`，构造函数返回接口类型。

## 3. Service

- [ ] 在 `app/service/` 新建文件，定义 `CreateXxxReq`、`UpdateXxxReq`、`RetrieveXxxReq`（带 `validate` 标签）。
- [ ] 实现 Service 结构体及构造函数（依赖 repo、locker、validate 等）。
- [ ] 写操作中：`validate.Struct`、必要时加锁、`copier.Copy` 到 model、调 `repo.WithContext(ctx).Save/...`。

## 4. Handle

- [ ] 在 `app/handle/v1/` 新建文件，实现 `router.Handler`（`ModuleName()`）。
- [ ] 各方法：Bind 参数 → 调 Service → `router.Success` / `router.Fail` / `router.PageSuccess`。

## 5. Binds

- [ ] 在 `app/app.go` 的 `Binds()` 中追加 `NewXxxHandle`、`NewXxxService`、`NewXxxRepository`。

## 6. Route

- [ ] 在 `app/route/api.go` 的 `InitApi` 的 `Invoke` 参数中增加 `xxxHandle *v1.XxxHandle`。
- [ ] 在 `Invoke` 内用 `router.NewRouteInfoBuilder(appName, xxxHandle, gv1, router.MenuOption{...})` 注册所有路由，设置访问级别与 `.Name(...).Build()`。
- [ ] 若需在侧边栏展示：`xxxMenu = r.GetMenu()`，并在 `InitMenu()` 中将 `xxxMenu` 挂到对应父菜单的 `Children`。

### 推荐的插入位置

- `Binds()`：放在同类型模块附近，便于后续维护
- `InitApi()`：若是普通管理模块，建议放在 `position` 块附近
- `InitMenu()`：普通系统管理模块通常挂在 `System` 下

## 7. 迁移

- [ ] 在 `app/database/auto_migrate_gen.go` 的 `Migrate(db)` 的 `AutoMigrate` 列表中追加 `&Xxx{}`。

## 8. 可选

- [ ] **Seeder**：在 `app/database/seeder/` 新增函数，并在 `Bootstrap()` 中显式调用。
- [ ] **事件**：在 `app/event/` 定义常量，在 Service 中 Publish，在 `listener.Init` 中 Subscribe。
- [ ] **Provider**：在 `app/provider/` 实现 Provider，在 SubApp 的 `ServiceProviders()` 中注册。

## 完成定义

- 能按清单独立完成一个单表 CRUD 模块从 model 到可访问接口的全流程。
- 能说明最容易漏掉的步骤（通常是 Binds、Migrate、InitApi 中未注册路由或未注入 Handle）。
- 能按 [08-startup-and-verification.md](08-startup-and-verification.md) 完成一次登录、调接口和看菜单的完整验收。
