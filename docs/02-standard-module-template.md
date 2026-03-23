# 标准模块模板（position）

新增一个**单表 CRUD** 模块时，建议以 **position（岗位）** 为模板，它结构完整、无多对多与事件噪音。

如果你想直接复制一套完整实现，请继续看 [07-example-notice-module.md](07-example-notice-module.md)。

## 模板文件清单

| 层级 | 文件 | 说明 |
|------|------|------|
| model | `app/model/position.go` | 表结构、`TableName()`、状态方法 |
| repository | `app/repository/position.go` | 接口 + 实现、`WithContext`、Save/Detail/Retrieve/Options |
| service | `app/service/position_service.go` | Create/Update/Delete/Retrieve/ChangeStatus/Options、校验与锁 |
| handle | `app/handle/v1/position_handle.go` | 实现 `router.Handler`，各方法 Bind → Service → router 响应 |
| route | `app/route/api.go` | 在 `InitApi` 中为 positionHandle 建 `NewRouteInfoBuilder` 并注册路由 |
| app | `app/app.go` | 在 `Binds()` 中注册 NewPositionHandle、NewPositionService、NewPositionRepository |
| database | `app/database/auto_migrate_gen.go` | 在 `Migrate(db)` 中 `AutoMigrate(&Position{})` |

## 最小文件树

```text
app/
├── model/position.go
├── repository/position.go
├── service/position_service.go
├── handle/v1/position_handle.go
├── route/api.go
├── app.go
└── database/auto_migrate_gen.go
```

## Model

- 嵌入 `db.BaseModel`，实现 `TableName()`（如 `admin_position`）。
- 业务状态用字段 + 常量（如 Status），必要时提供 `Enable()`/`Disable()` 等小方法。
- 参考：`app/model/position.go`。

## Repository

- 定义 `XxxRepositoryInterface`，包含 `Save`、`Detail`、`Retrieve`、以及本模块特有的 `Options` 等，并嵌入 `contract.WithContext[XxxRepositoryInterface]`。
- 实现体持有一个 `*gorm.DB` 和 `db.BaseRepository[model.Xxx]`，构造函数返回接口类型。
- `WithContext(ctx)` 返回新实例，将 `db` 换为 `db.WithContext(ctx)`。
- 参考：`app/repository/position.go`。

## Service

- 定义请求 DTO：`CreateXxxReq`、`UpdateXxxReq`、`RetrieveXxxReq`，使用 `validate` 标签。
- Service 依赖：Repository、`redis.LockerFactory`、`*validator.Validate`，可选 `*gorm.DB`（用于 BaseRepository）。
- 写操作前 `validate.Struct(req)`，必要时 `locker.New().Lock(key)`，再用 `copier.Copy` 映射到 model，调用 `repo.WithContext(ctx).Save(...)`。
- 参考：`app/service/position_service.go`。

## Handle

- 实现 `router.Handler`：`ModuleName() (string, string)` 返回英文模块名与中文名（如 `"position", "岗位管理"`）。
- 每个 HTTP 方法：`ctx.ShouldBindJSON/ShouldBindQuery` → 调用 Service → `router.Success` / `router.Fail` / `router.PageSuccess`。
- 路径参数用 `ctx.Param("id")` 并转成所需类型（如 `cast.ToUint`）。
- 参考：`app/handle/v1/position_handle.go`。

## Route 与 Binds

- 在 `InitApi` 的 `Invoke` 中增加对 `positionHandle` 的注入，新建一组路由：
  - `r := router.NewRouteInfoBuilder(appName, positionHandle, gv1, router.MenuOption{ComponentName, Path, Icon})`
  - 然后 `r.Post("/position", ...).Name("创建岗位").Build()` 等，最后 `positionMenu = r.GetMenu()`。
- 在 `InitMenu()` 中将 `positionMenu` 挂到对应父菜单的 `Children`。
- 在 `app.go` 的 `Binds()` 中追加 `v1.NewPositionHandle`、`service.NewPositionService`、`repository.NewPositionRepository`。

## 迁移

- 在 `app/database/auto_migrate_gen.go` 的 `Migrate(db)` 中增加 `&Position{}`（若为新 model）。

## 完成定义

- 能按 position 模板列出新增一个单表 CRUD 模块需要的 7 处改动及对应文件。
- 能写出符合项目风格的 Repository 接口（含 WithContext）与 Service 的校验/锁/调用 repo 的写法。
