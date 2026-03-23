# 进阶模板与常见坑

## 进阶模板（按场景选用）

- **主表 + 子表**：参考 **dict**（`model/dict.go`、`repository/dict.go`、`service/dict_service.go`、`handle/v1/dict_handle.go`）。主表与字典项分表，有 CreateItem、UpdateItem、RetrieveItems、GetDictByType 等扩展 API。
- **树形/层级**：参考 **dept**（`model/dept.go`、`service/dept_service.go`）。parent_id + 唯一性校验、Children 自关联。
- **权限与多对多**：参考 **role**、**user** 与 **listener**。角色分配菜单、用户分配角色会发布事件，由 listener 同步到 Casbin；涉及菜单、角色、用户关系时逻辑较多，不宜作为第一个模板。

**建议**：第一个模块用 **position**；需要主子表再用 **dict**；需要树再用 **dept**；需要权限联动再看 **role/user** 与 **listener**。

## 常见坑

1. **漏注册 Binds**  
   新增了 Handle/Service/Repository 但未在 `app.go` 的 `Binds()` 中追加，会导致 `InitApi` 的 `Invoke` 无法解析依赖，启动失败。

2. **漏注册路由**  
   只写了 Handle 和 Service 但未在 `route/api.go` 的 `InitApi` 里用 `NewRouteInfoBuilder` 注册路由，接口不可访问。

3. **漏加迁移**  
   新 model 未加入 `database/auto_migrate_gen.go` 的 `Migrate()`，表不会自动创建，运行时报错。

4. **权限依据**  
   权限校验和“接口管理”读取的是 `router.GetAllRoutes()` 的运行时数据，不是数据库里的 `admin_api` 表，不要误以为改表即可改权限。

5. **事件未成对**  
   只在 `event/` 下定义了常量没有用；必须在某处 `Publish` 且在 `listener.Init` 里 `Subscribe` 才会生效。

6. **空实现勿照抄**  
   部分 Handle 的 `Detail()` 等在当前代码中为空实现，不要当作标准模板照抄；以 position 的完整实现为准。

7. **超管账号**  
   超级管理员登录来自配置（如 `conf/app.yaml` 的 `admin.username` / `admin.password`），不是数据库用户表，测试权限时需注意。

8. **Swagger 与路由一致**  
   若使用 Swagger 注解，`@Router` 必须与真实注册路径一致，否则文档与行为不符。

## 完成定义

- 能根据业务形态（单表 / 主子表 / 树形 / 权限联动）选择对应模板。
- 能列出至少 5 条本项目内常见坑并说明原因。
- 能区分“框架约定”与“当前业务实现细节”，不把历史包袱当规范。
