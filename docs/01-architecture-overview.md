# 项目架构总览

## 目录结构

owl-admin 采用分层结构：**route → handle → service → repository → model**，并辅以 middleware、event/listener、provider、database。

```
owl-admin/
├── main.go                 # 入口：owl.NewApp(&admin.SubAppAdmin{}).WebShell()
├── go.mod / go.sum
├── conf/                   # 运行配置（框架与 Provider 自动生成模板）
└── app/                    # 业务子应用（admin）
    ├── app.go              # SubApp 定义：ServiceProviders、Binds、RegisterRouters、Menu、Bootstrap
    ├── cmd/                # 子应用命令（如 gen-password、version）
    ├── route/
    │   └── api.go          # /api/v1 路由注册与菜单树 InitApi / InitMenu
    ├── middleware/         # 鉴权、操作日志等
    ├── handle/
    │   ├── oauth/          # 第三方登录
    │   └── v1/             # v1 API 的 HTTP 处理
    ├── service/            # 业务服务（校验、锁、事务、事件）
    ├── repository/         # 数据访问（GORM）
    ├── model/              # 数据模型
    ├── database/
    │   ├── auto_migrate_gen.go   # 迁移入口 Migrate(db)
    │   └── seeder/               # 初始化数据（如字典）
    ├── event/              # 领域事件常量
    ├── listener/            # 事件订阅（如 Casbin 同步）
    └── provider/           # 子应用扩展 Provider（如 JWT）
```

## 启动链路

1. **main** 调用 `owl.NewApp(&admin.SubAppAdmin{}).WebShell()`。
2. 框架对 `SubAppAdmin` 执行 `injectAppInstance`、注册 `Binds()`、收集 `ServiceProviders()`，再统一执行所有 Provider 的 `Register()` 和配置生成。
3. **RegisterRouters()** 调用 `route.InitApi(i.app, i.Name())`，在 `InitApi` 中通过 `app.Invoke(...)` 取得 `*gin.Engine`、各 Handle、鉴权依赖等，用 `router.NewRouteInfoBuilder` 注册路由并挂中间件（PermissionCheck、OperationLog）。
4. **Menu()** 返回 `route.InitMenu()`，将各模块菜单挂到顶层目录下。
5. **Bootstrap()** 中执行数据库迁移（`database.Migrate`）、字典 Seeder（`seeder.InitAllDictData`）、事件监听器注册（`listener.Init(i.app)`）。
6. 所有 Provider 的 `Boot()` 执行完毕后，框架启动 Gin HTTP 服务。

## 请求处理链路

请求进入后：**Gin 路由 → 中间件（鉴权、操作日志）→ Handle（参数绑定、调 Service）→ Service（校验、锁、调 Repository）→ Repository（GORM）→ Model**。

- **Handle**：薄层，只做 Bind、调 Service、统一返回 `router.Success` / `router.Fail` / `router.PageSuccess`。
- **Service**：业务规则、`validate`、分布式锁、事件发布、调用 `repository.WithContext(ctx)`。
- **Repository**：接口 + 实现，实现带 `WithContext`，封装 GORM 查询与保存。

## 完成定义

- 能说出 route、handle、service、repository、model 的职责与调用顺序。
- 能说明 SubApp 的 RegisterRouters、Menu、Bootstrap 在启动中的位置。
- 能说明一次 HTTP 请求从进入到返回经过哪些层。
