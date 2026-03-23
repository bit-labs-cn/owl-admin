# flex-admin

基于 Go + Gin 的后台管理服务（RBAC 权限体系），采用 `bit-labs.cn/owl` 应用框架组织子应用、依赖注入与配置管理。提供用户/角色/菜单/接口权限、组织架构、字典、地区、日志、文件上传、APP 升级检查与第三方 OAuth 登录等能力。

## 目录结构

项目采用典型的分层结构：`route(路由)` → `handle(HTTP 处理)` → `service(领域服务)` → `repository(数据访问)` → `model(数据模型)`，并通过 `middleware`、`event/listener` 等模块完成鉴权、操作审计、权限同步等横切能力。

```text
flex-admin/
├─ main.go                       # 入口：启动 owl 应用并运行 HTTP 服务
├─ go.mod / go.sum               # Go 模块与依赖（包含对 bit-labs.cn/owl 的本地 replace）
├─ README.md                     # 项目说明
├─ conf/                         # 运行配置目录（启动时会自动生成模板配置）
│  ├─ .gitignore                 # 默认忽略 *.yaml 配置，避免提交敏感信息
│  └─ oauth.example.yaml         # OAuth 示例配置（需复制为 oauth.yaml 才会生效）
└─ app/                          # 业务子应用（admin）
   ├─ app.go                     # 子应用注册：服务提供者、路由、绑定、启动钩子
   ├─ cmd/                       # CLI 命令（子应用级别；切换为 ConsoleShell 模式后可用）
   │  ├─ gen-password.go         # 生成 bcrypt 密码（用于配置中的 admin.password）
   │  └─ version.go              # 输出版本号
   ├─ route/                     # 路由注册与菜单定义
   │  └─ api.go                  # /api/v1 下的路由、访问级别、菜单元信息
   ├─ middleware/                # Gin 中间件（横切逻辑）
   │  ├─ permission_check.go     # JWT 鉴权 + Casbin 授权校验
   │  └─ operation_log.go        # 操作日志采集（非 GET 请求）
   ├─ handle/                    # HTTP Handler（参数绑定/返回结构/调用 service）
   │  ├─ oauth/                  # 第三方 OAuth 登录与回调处理
   │  └─ v1/                     # v1 API handlers（用户/角色/字典/部门等）
   ├─ service/                   # 业务服务（核心业务规则/事务/校验/事件）
   ├─ repository/                # 数据访问层（GORM 查询、分页、聚合）
   ├─ model/                     # 数据模型（GORM 表结构与领域方法）
   ├─ database/                  # 数据库相关
   │  ├─ auto_migrate_gen.go      # 自动迁移入口（AutoMigrate）
   │  └─ seeder/                 # 初始化数据（字典主表/字典项等）
   ├─ event/                     # 领域事件定义（如分配角色、分配菜单）
   ├─ listener/                  # 事件监听器（将事件同步到 Casbin/菜单仓库等）
   └─ provider/                  # 子应用扩展 Provider（如 JWT）
      └─ jwt/                    # JWT 服务与示例配置模板（jwt.yaml）
```

## 核心功能

以下能力均通过 `/api/v1` 前缀提供（详见 [api.go](file:///d:/project/bit-labs/flex-admin/app/route/api.go)）：

- 认证与鉴权（RBAC）
  - JWT 登录、解析与用户上下文注入（`Authorization: Bearer <token>`）
  - Casbin 授权校验：根据路由权限标识判断是否可访问
  - 访问级别：公开（Public）/ 仅登录（Authenticated）/ 需授权（Authorized）/ 超级管理员（SuperAdmin）
- 用户管理
  - 登录、获取当前用户信息、修改密码
  - 用户 CRUD、启用/禁用、重置密码
  - 给用户分配角色、查询用户角色、按角色生成用户菜单
- 角色与权限管理
  - 角色 CRUD、启用/禁用、角色选项
  - 给角色分配菜单：事件驱动同步到 Casbin policy
  - 菜单列表与可分配菜单查询（含按钮类菜单）
  - 接口管理：输出系统所有已注册接口及其权限信息
- 组织与基础数据
  - 部门管理：新增/删除/更新/列表
  - 岗位管理：CRUD、启用/禁用、岗位选项
  - 字典管理：字典与字典项 CRUD、按类型获取启用字典项
  - 地区管理：省市区数据查询（平铺）
- 日志与审计
  - 登录日志、操作日志分页查询
  - 自动操作审计：默认对非 GET 请求记录操作日志（可在路由级关闭）
- 文件与集成
  - 文件上传（由 owl storage provider 提供存储能力）
  - APP 升级检查：查询最新可升级版本
  - 第三方 OAuth 登录：GitHub / Google / Gitee（可配）

## 快速开始

### 环境要求

- Go：`1.24.x`（见 [go.mod](file:///d:/project/bit-labs/flex-admin/go.mod#L1-L6)）
- 数据库：PostgreSQL / MySQL / SQLite（三选一，取决于 `conf/database.yaml`）
- Redis：用于分布式锁与缓存（`conf/redis.yaml`）

> 重要：本项目 `go.mod` 中对 `bit-labs.cn/owl` 使用了本地 `replace ../owl`，需确保同级目录存在 `owl/`（例如：`d:/project/bit-labs/owl`），或自行调整 `replace` 指向。

### 安装依赖

```bash
go mod download
```

### 准备配置

配置目录为项目根目录下的 `conf/`。首次启动时，`owl` 会在 `conf/` 中自动生成缺失的配置模板文件（如 `app.yaml`、`router.yaml`、`database.yaml`、`redis.yaml`、`log.yaml`、`jwt.yaml`、`storage.yaml`、`captcha.yaml` 等）。

你需要重点检查/修改以下配置：

- `conf/app.yaml`
  - `admin.username` / `admin.password`：用于内置超管登录（密码为 bcrypt hash；示例模板默认对应明文 `123qwe`）
  - 说明：当登录用户名命中 `admin.username` 时，会返回内置超级管理员资料（ID/用户名等为系统内置固定值）
- `conf/router.yaml`
  - `server.host` / `server.port`：HTTP 监听地址与端口（默认 `0.0.0.0:8080`）
- `conf/database.yaml`
  - `driver`、`host`、`port`、`database`、`username`、`password` 等
- `conf/redis.yaml`
  - `mode`、`single.host`、`single.port`、`single.password` 等
- `conf/jwt.yaml`
  - `signing-key`、`issuer`、`expire`
- `conf/oauth.yaml`（可选）
  - 将 `conf/oauth.example.yaml` 复制为 `conf/oauth.yaml` 并填写各 provider 的 `client-id/client-secret/redirect-url`

另外，本项目支持 `.env` 文件加载（按优先级依次尝试：`.env.local`、`.env.<APP_ENV>`、`.env`），也支持通过环境变量覆盖配置（Viper 会将 `-` 与 `.` 映射为 `_`，并以配置文件名作为前缀，例如：`DATABASE_HOST`、`JWT_SIGNING_KEY` 等）。

### 启动服务

```bash
go run .
```

启动后：

- HTTP 服务会按照 `conf/router.yaml` 的 `server.host/port` 监听
- 启动过程中会自动执行数据库迁移（AutoMigrate）与字典初始化（Seeder）

## 使用示例

以下示例假设服务运行在 `http://127.0.0.1:8080`。

### 1）登录获取 Token

```bash
curl -X POST "http://127.0.0.1:8080/api/v1/users/login" ^
  -H "Content-Type: application/json" ^
  -d "{\"username\":\"admin\",\"password\":\"123qwe\"}"
```

成功后会返回 `accessToken`（字段名以实际响应为准）；后续请求使用：

```text
Authorization: Bearer <accessToken>
```

### 2）获取当前用户信息与菜单

```bash
curl "http://127.0.0.1:8080/api/v1/users/me" ^
  -H "Authorization: Bearer <accessToken>"

curl "http://127.0.0.1:8080/api/v1/users/me/menus" ^
  -H "Authorization: Bearer <accessToken>"
```

### 3）创建用户

```bash
curl -X POST "http://127.0.0.1:8080/api/v1/users" ^
  -H "Content-Type: application/json" ^
  -H "Authorization: Bearer <accessToken>" ^
  -d "{\"username\":\"demo\",\"nickName\":\"演示用户\",\"password\":\"P@ssw0rd\"}"
```

### 4）上传文件

```bash
curl -X POST "http://127.0.0.1:8080/api/v1/files/upload" ^
  -H "Authorization: Bearer <accessToken>" ^
  -F "file=@./test.png"
```

### 5）查询操作日志（分页）

```bash
curl -X POST "http://127.0.0.1:8080/api/v1/monitor/operation-logs" ^
  -H "Content-Type: application/json" ^
  -H "Authorization: Bearer <accessToken>" ^
  -d "{\"page\":1,\"pageSize\":20}"
```

### 6）第三方登录（可选）

浏览器访问：

```text
GET /api/v1/oauth/github/login
GET /api/v1/oauth/google/login
GET /api/v1/oauth/gitee/login
```

并确保 `conf/oauth.yaml` 中的 `redirect-url` 与前端/回调地址一致。

## 技术栈与依赖

- 语言与基础设施
  - Go 1.24.x
  - `bit-labs.cn/owl`：应用框架（DI/配置/日志/路由/基础 Provider）
- Web 与路由
  - Gin（`github.com/gin-gonic/gin`）
  - Swagger 注解（`github.com/swaggo/swag`，handlers 内可见 `@Summary/@Router` 等）
- 数据与缓存
  - GORM（`gorm.io/gorm`）
  - Redis（`github.com/redis/go-redis/v9`，通过 owl provider 注入）
- 认证与权限
  - JWT（`github.com/golang-jwt/jwt/v5`）
  - Casbin（`github.com/casbin/casbin/v2`）
- 其他
  - OAuth2（`golang.org/x/oauth2`）
  - Socket.IO（`github.com/googollee/go-socket.io`）
  - 配置管理（Viper）、命令行（Cobra）、参数校验（validator/v10）

## 开发文档

在本项目中**新增业务模块**的完整说明（架构、标准模板、路由/菜单/权限、迁移/Seeder/事件、操作清单与常见坑）见 **[docs/](docs/)** 目录。适合开发人员与 AI 开发工具按顺序阅读后直接开发新模块。

## 开发扩展指南（新增模块）

新增一个业务模块通常遵循以下路径（与现有 `user/role/dict/...` 保持一致）：

1. `app/model/`：定义表结构与领域方法
2. `app/repository/`：封装查询与分页（GORM）
3. `app/service/`：实现业务规则、校验与事务
4. `app/handle/v1/`：实现 HTTP Handler（绑定参数/调用 service/返回统一响应）
5. `app/route/api.go`：注册路由、权限级别与菜单元信息

更详细的步骤、模板引用与自检清单见 [docs/](docs/)。
