# 启动与验证流程

本篇用于回答一个最实际的问题：**模块写完后，怎么确认它真的接好了？**

如果你刚按 [07-example-notice-module.md](07-example-notice-module.md) 新增了模块，建议按本篇逐项验证。

## 启动前检查

### 1. 代码接线

- `app/model/` 已新增 model
- `app/repository/` 已新增 repository
- `app/service/` 已新增 service
- `app/handle/v1/` 已新增 handle
- `app/app.go` 的 `Binds()` 已追加 `NewXxxHandle`、`NewXxxService`、`NewXxxRepository`
- `app/route/api.go` 的 `InitApi()` 已新增 handle 注入和路由块
- `app/route/api.go` 的 `InitMenu()` 已挂载菜单
- `app/database/auto_migrate_gen.go` 已把新 model 加入 `AutoMigrate`

### 2. 配置

启动前至少确认：

- `conf/database.yaml`
- `conf/redis.yaml`
- `conf/router.yaml`
- `conf/jwt.yaml`
- `conf/app.yaml`

当前仓库默认端口通常是：

```yaml
server:
  port: 8086
```

默认超管账号来自 `conf/app.yaml`：

```yaml
admin:
  username: admin
```

默认说明里提到的初始密码明文是：

```text
123qwe
```

## 启动命令

```bash
go run .
```

若依赖尚未安装：

```bash
go mod tidy
go run .
```

## 首次启动后应看到什么

### 1. 服务启动

- HTTP 服务监听 `conf/router.yaml` 中配置的端口，默认通常是 `8086`
- 若启动直接 panic，优先检查 `Binds()` 和 `InitApi()` 的依赖注入

### 2. 数据迁移执行

`Bootstrap()` 会调用：

- `database.Migrate(migDB)`
- `seeder.InitAllDictData(migDB)`
- `listener.Init(i.app)`

因此你的新表应该在启动后自动创建。

### 3. 新菜单出现在菜单树里

若你把新模块挂在 `InitMenu()` 中，登录后获取菜单接口时应能看到它。

## Smoke Test

下面以 `notice` 模块为例。

### 1. 登录拿 token

```bash
curl -X POST "http://127.0.0.1:8086/api/v1/users/login" ^
  -H "Content-Type: application/json" ^
  -d "{\"username\":\"admin\",\"password\":\"123qwe\"}"
```

期望：

- 返回 `accessToken`

### 2. 创建公告

```bash
curl -X POST "http://127.0.0.1:8086/api/v1/notices" ^
  -H "Content-Type: application/json" ^
  -H "Authorization: Bearer <accessToken>" ^
  -d "{\"title\":\"系统公告\",\"content\":\"这是一条测试公告\",\"status\":1}"
```

期望：

- 返回 `success: true`
- 数据库中出现一条 `admin_notice` 记录

### 3. 查询公告列表

```bash
curl "http://127.0.0.1:8086/api/v1/notices?page=1&pageSize=10" ^
  -H "Authorization: Bearer <accessToken>"
```

期望：

- 返回 `success: true`
- `data.list` 中包含刚创建的记录

### 4. 修改状态

```bash
curl -X PUT "http://127.0.0.1:8086/api/v1/notices/1/status" ^
  -H "Content-Type: application/json" ^
  -H "Authorization: Bearer <accessToken>" ^
  -d "{\"status\":2}"
```

期望：

- 返回 `success: true`
- 目标记录的 `status` 变更成功

### 5. 查看菜单

```bash
curl "http://127.0.0.1:8086/api/v1/users/me/menus" ^
  -H "Authorization: Bearer <accessToken>"
```

期望：

- 若角色拥有新模块对应权限，则返回菜单树中包含你的新菜单

## 常见失败与定位

### 启动时报 `Invoke` 或容器解析错误

优先检查：

1. `app/app.go` 的 `Binds()` 是否漏注册
2. `InitApi()` 的参数列表是否新增了 `xxxHandle`
3. 构造函数签名是否与依赖一致

### 接口 404

优先检查：

1. `route/api.go` 中是否真的注册了该路由
2. 路由路径是否和调用路径一致
3. 新路由块是否执行了 `.Build()`

### 接口 403 或“无权限”

优先检查：

1. 路由访问级别是否用了 `router.AccessAuthorized`
2. 当前登录用户是否为超管或已有对应角色权限
3. 菜单/角色分配后是否需要重新登录刷新 token 与菜单

### 数据库报表不存在

优先检查：

1. `database/auto_migrate_gen.go` 是否追加了新 model
2. `Bootstrap()` 是否真的执行了 `database.Migrate`
3. 当前连接的数据库是不是你以为的那个实例

### 菜单不显示

优先检查：

1. `xxxMenu = r.GetMenu()` 是否执行
2. `InitMenu()` 是否把 `xxxMenu` 挂到父菜单
3. 当前用户是否拥有对应按钮/菜单权限

## 最终验收清单

- [ ] 服务能成功启动
- [ ] 新表已自动创建
- [ ] 创建接口可用
- [ ] 列表接口可用
- [ ] 状态切换接口可用
- [ ] 菜单能在 `/users/me/menus` 中看到
- [ ] 权限与操作日志行为符合预期

## 完成定义

- 看完本文后，开发者或 AI 能对一个新模块完成“启动、建表、登录、调接口、看菜单”的完整验收。
- 即使失败，也能快速定位是 DI、路由、权限、迁移还是菜单接线的问题。
