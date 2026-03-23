# 迁移、Seeder、事件与 Provider

## 迁移

- **入口**：`app/database/auto_migrate_gen.go` 的 `Migrate(db *gorm.DB)`，内部对所需 model 执行 `db.Migrator().AutoMigrate(&Xxx{}, ...)`。
- **调用时机**：在 SubApp 的 `Bootstrap()` 中通过 `app.Invoke(func(gdb *gorm.DB) { ... })` 取得 DB，使用降日志的 Session（如 `gdb.Session(&gorm.Session{Logger: ...})`）后调用 `database.Migrate(migDB)`，通常以 goroutine 执行避免阻塞启动。
- **新增表**：在 `Migrate` 的 `AutoMigrate` 参数列表中追加 `&YourModel{}`。

## Seeder

- **位置**：`app/database/seeder/`，如 `dictionary.go`。
- **调用**：在 `Bootstrap()` 中显式调用，例如 `seeder.InitAllDictData(migDB)`。不会自动扫描，需在代码里写死调用。
- **写法**：建议幂等（先查再决定是否插入），可放在事务中。参考 `InitAllDictData` 与 `initDictDataWithTx`。

## 事件与监听器

- **事件常量**：在 `app/event/` 下定义（如 `AssignRoleToUser`、`AssignMenuToRole`）。
- **发布**：在 Service 中通过注入的 `EventBus.Bus` 调用 `bus.Publish(event.Topic, payload)`。
- **订阅**：在 `app/listener/listener.go` 的 `Init(app)` 中通过 `app.Invoke(func(bus EventBus.Bus, ...) { bus.Subscribe(event.XXX, handler) })` 注册。**仅定义事件常量不会生效，必须有一处 Publish 和一处 Subscribe。**
- 参考：`event/assign_role_to_user.go`、`service` 中的 `AssignMenusToRole`/`AssignRoleToUser`、`listener.Init`。

## Provider

- 子应用专属能力（如 JWT）在 `app/provider/` 下实现 `foundation.ServiceProvider`（Register、Boot、Conf、Description），结构体需含 `app foundation.Application` 字段。
- 在 SubApp 的 `ServiceProviders()` 中返回该 Provider，框架会注入 `app`、执行 `Register()` 和 `Conf()` 生成配置、最后执行 `Boot()`。
- 参考：`app/provider/jwt/jwt_service_provider.go`。

## 完成定义

- 能说明新增一张表需要在哪处修改、迁移在何时被调用。
- 能说明 Seeder 的添加方式与调用位置。
- 能说明事件从定义到发布、订阅的完整链路。
- 能说明如何新增一个子应用级 Provider 并在 SubApp 中注册。
