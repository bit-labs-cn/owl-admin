# 路由、菜单与权限

## InitApi 与路由注册

所有 v1 API 在 `app/route/api.go` 的 `InitApi(app, appName)` 中注册。通过 `app.Invoke(func(..., engine *gin.Engine, ...) { ... })` 一次性注入所需 Handle、Gin Engine、Casbin Enforcer、JWT、LogService 等。

- 路由组：`gv1 := engine.Group("/api/v1", middleware.PermissionCheck(enforcer, jwtSvc))`，再挂 `OperationLog(logService)`。
- 每个业务模块用 `router.NewRouteInfoBuilder(appName, handle, gv1, router.MenuOption{...})` 创建 builder，再链式：
  - `.Get/.Post/.Put/.Delete(path, accessLevel, handleFunc).Name("中文").Build()`
  - 可选：`.Deps(...)` 声明依赖的其他接口（用于权限与前端按钮）、`.WithoutOperateLog()`、`.Description("...")`

## 访问级别

- `router.AccessPublic`：开放。
- `router.AccessAuthenticated`：仅需登录。
- `router.AccessAuthorized`：需授权（Casbin 校验权限标识）。
- `router.AccessSuperAdmin`：仅超管。

权限标识形如 `appName:moduleEn:方法名`，由框架根据 Handler 的 `ModuleName()` 与处理函数名生成，并写入全局路由表（`router.GetAllRoutes()`）。**权限校验依赖该运行时路由表，而非数据库中的“接口表”。**

## 菜单

- **顶层目录**：在 `InitMenu()` 中写死，如“用户权限”“日志管理”“系统管理”，每个目录的 `Children` 引用各模块的菜单变量（如 `userMenu`、`positionMenu`）。
- **模块菜单**：在 `InitApi` 里用 `NewRouteInfoBuilder(..., router.MenuOption{ComponentName, Path, Icon})` 注册路由后，用 `xxxMenu = r.GetMenu()` 得到该模块菜单，再在 `InitMenu()` 里挂到对应父级 `Children`。
- 若某组路由不需要在侧边栏展示，可传空 `MenuOption{}`（如 app upgrade、oauth）。

## Deps 与操作日志

- **Deps**：声明完成该操作所需的其他接口（如“更新用户”依赖“获取用户详情”“角色选项”），用于前端按钮权限与权限分配。
- **操作日志**：非 GET 且未标记 `WithoutOperateLog()` 的路由会由中间件根据 `router.GetAllRoutes()` 中的元数据记录操作日志。

## 完成定义

- 能说明在 `InitApi` 中注册一条新路由的完整链式写法及访问级别含义。
- 能说明如何让新模块出现在侧边栏（MenuOption + GetMenu + InitMenu 挂 Children）。
- 能说明权限依据的是运行时路由表而非数据库表。
