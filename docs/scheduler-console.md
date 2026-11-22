# 定时任务管理台

`owl` 的定时任务引擎在 `provider/scheduler`，本页是它的可视化管理入口，代码在 `owl-admin`：

- 后端：`app/repository/scheduler.go`、`app/service/scheduler_service.go`、`app/handle/v1/scheduler_handle.go`、`app/route/api.go`
- 前端：`frontend/admin/src/views/scheduler/`、`frontend/admin/src/api/scheduler.ts`

引擎本身的行为（排期、抢占、租约、重试）见 `owl/docs/09-scheduler.md`。

## 菜单与权限

「系统管理 → 定时任务」，路由 `/system/scheduler/index`，组件 name `SystemScheduler`，页内用 Tab 切换「任务」与「执行记录」。

权限标识由后端 `RouteInfoBuilder` 生成，统一是 `admin:scheduler:<处理函数名>`，前端集中在 `src/permissions.ts` 的 `ADMIN_PERMS.scheduler`：

| 权限 | 接口 | 级别 |
| --- | --- | --- |
| `admin:scheduler:RetrieveJobs` | `GET /api/v1/scheduler/jobs` | 需授权 |
| `admin:scheduler:JobDetail` | `GET /api/v1/scheduler/jobs/:id` | 仅超管 |
| `admin:scheduler:CreateJob` | `POST /api/v1/scheduler/jobs` | 仅超管 |
| `admin:scheduler:UpdateJob` | `PUT /api/v1/scheduler/jobs/:id` | 仅超管 |
| `admin:scheduler:DeleteJob` | `DELETE /api/v1/scheduler/jobs/:id` | 仅超管 |
| `admin:scheduler:ChangeStatus` | `PUT /api/v1/scheduler/jobs/:id/status` | 需授权 |
| `admin:scheduler:TriggerJob` | `POST /api/v1/scheduler/jobs/:id/trigger` | 需授权 |
| `admin:scheduler:RetrieveRuns` | `GET /api/v1/scheduler/runs` | 需授权 |
| `admin:scheduler:RunDetail` | `GET /api/v1/scheduler/runs/:id` | 需授权 |
| `admin:scheduler:RetryRun` | `POST /api/v1/scheduler/runs/:id/retry` | 需授权 |
| `admin:scheduler:CancelRun` | `POST /api/v1/scheduler/runs/:id/cancel` | 需授权 |
| `admin:scheduler:CleanupRuns` | `POST /api/v1/scheduler/runs/cleanup` | 需授权 |
| `admin:scheduler:PreviewCron` | `POST /api/v1/scheduler/cron/preview` | 需授权 |
| `admin:scheduler:Stats` | `GET /api/v1/scheduler/stats` | 需授权 |
| `admin:scheduler:RetrieveActions` | `GET /api/v1/scheduler/actions` | 仅超管 |
| `admin:scheduler:ActionOptions` | `POST /api/v1/scheduler/actions/:code/options` | 仅超管 |

给角色分配菜单时，除了勾选「定时任务」菜单，还要勾选需要使用的按钮节点。自定义任务的增删改与详情只给超管：脚本可能包含命令与 Token，权限收得更紧。**触发 / 重试这类运行操作按按钮权限放行**，方法本身不再声明「仅超管」，登记了什么方法前端就全量列出。

## 任务列表

- 展示内容：名称与来源（内置 / 自定义）、任务编码、所属子系统、可读执行规则与下次执行、最近执行结果与时间、成功 / 失败统计、启停开关。
- **内置任务**：不出现编辑、删除按钮；可以启停、手动触发、查看详情与执行记录。
- **自定义任务**：可以编辑、删除；删除前会检查是否仍有待执行 / 执行中的记录。
- 列表接口不返回脚本配置；只有超管点「详情」时才请求 `JobDetail` 并展示脚本。
- 启停开关调用 `ChangeStatus`，失败时会把开关拨回原状态，避免界面与后端不一致。

### 新建 / 修改自定义任务

表单字段：

| 字段 | 说明 |
| --- | --- |
| 任务名称 / 说明 | 名称必填，说明会显示在详情里 |
| 执行时间 | 点「设置时间」在弹窗里选频率与时刻，由界面生成表达式；不用手写 Cron |
| 时区 | 留空用 `scheduler.yaml` 的默认时区 |
| 超时 / 最大尝试 / 重试间隔 | 单次执行超时、包含首次在内的尝试上限、失败后的退避秒数 |
| 并发策略 | 跳过本次（默认）或允许并发 |
| 执行方式 | HTTP 请求、Shell 命令，或调用子系统注册的方法 |

### 执行时间弹窗

表单只显示「每天 08:00 执行」这类可读规则和最终表达式，表达式由弹窗生成：

| 模式 | 需要填的内容 | 生成的表达式示例 |
| --- | --- | --- |
| 按分钟重复 | 每隔多少分钟（1 表示每分钟） | `*/5 * * * *` |
| 按小时重复 | 每隔多少小时、第几分钟 | `0 */2 * * *` |
| 每天 | 时、分 | `0 8 * * *` |
| 每周 | 周几（可多选）、时、分 | `30 9 * * 1,3` |
| 每月 | 几号、时、分 | `0 8 1 * *` |
| 自定义表达式 | 直接写表达式 | `0 8 * * 1-5` |

弹窗下方实时显示可读规则、最终表达式与接下来 5 次执行时间（通过 `POST /scheduler/cron/preview` 计算）。编辑已有任务时会尽量把表达式反解成对应模式；反解不了（如 6 段表达式、范围写法）会落到「自定义表达式」并原样保留。

### 选择调用方法

「调用方法」不设下拉：点击后打开选择弹窗，**左侧是子系统**（通知中心、vlink 协调器等，括注方法数量），**右侧是该系统登记的方法**，逐条展示方法名称、所属分组、访问级别、说明与参数清单；顶部可按名称或说明跨系统搜索。选中后回到表单，方法名称与说明直接显示在字段里。

- **调用子系统方法**：表单按动作声明的参数动态渲染；下拉候选由子系统运行时提供，例如先选开关再列该开关的命令。保存时后端会校验必填、类型、范围与候选，并生成「开关：前台；命令：开机」这样的可读摘要。
- HTTP 任务按表单填写的字段组装 `{"method","url","headers","body","contentType","successStatus"}`；`headers` 必须是 JSON 对象，`successStatus` 是逗号分隔的状态码（留空表示 2xx 成功）。
- Shell 任务组装 `{"command","workingDir"}`；服务端 `shell-enabled=false` 时不允许创建。
- 保存前会校验执行时间、时区与脚本配置，错误直接以业务错误提示。

## 手动触发

「触发」会立即写入一条待执行记录，调度循环在秒级内领取执行。内置处理函数任务可填写一段 JSON 参数，在 `Handler` 的 `Input.Payload` 里拿到它。

调用方法、HTTP、Shell 任务按任务里保存的配置执行，手动触发不会改变参数，因此弹窗不显示「执行参数」输入框。

触发被拒绝的常见原因：

| 提示 | 原因 |
| --- | --- |
| 定时任务已停用，请先启用 | 任务状态为停用 |
| 定时任务已下线，不能再执行 | 代码里已移除该内置任务 |
| 任务仍在执行，本次触发已跳过 | 并发策略为「跳过」，已有待执行 / 执行中记录 |
| 定时任务引擎未启用 | `scheduler.enabled=false`，或没有可用数据库 |
| 任务绑定的方法已不可用 | 子系统没有加载该动作（例如未部署对应子系统） |

## 执行记录

- 支持按任务编码、状态、触发方式筛选；从任务列表点「记录」会自动带上该任务过滤。
- 「自动刷新」每 5 秒静默刷新一次，适合观察执行中状态；页面被 keep-alive 缓存时自动停表。
- 操作：
  - **详情**：计划 / 开始 / 结束时间、耗时、尝试次数、执行参数、输出、错误信息与 panic 堆栈，支持复制。
  - **重试**：仅失败、超时、已取消的记录可重试，会生成一条新的执行记录并保留 `retry_of_run_id`。
  - **取消**：待执行记录直接置为已取消；本实例执行中的记录会中断任务，其它实例上运行的记录会提示无法取消。
  - **清理历史**：删除指定天数之前的终态记录（成功 / 失败 / 取消 / 超时 / 跳过），未完成与最近的历史保留。

## 状态取值

| 对象 | 取值 |
| --- | --- |
| 任务来源 | 1 内置，2 自定义 |
| 任务状态 | 1 启用，2 停用，3 已下线 |
| 执行方式 | 1 内置处理函数，2 HTTP，3 Shell，4 调用子系统方法 |
| 执行状态 | 1 待执行，2 执行中，3 成功，4 失败，5 已取消，6 超时，7 已跳过 |
| 触发方式 | 1 定时，2 手动，3 重试 |
| 并发策略 | 1 跳过，2 允许并发 |

前端常量在 `src/views/scheduler/constants.ts`，与后端 `provider/scheduler/model.go` 保持一一对应。

## 业务错误码

| 错误码 | 含义 |
| --- | --- |
| `SCHEDULER_JOB_NOT_FOUND` | 任务不存在 |
| `SCHEDULER_BUILTIN_IMMUTABLE` | 内置任务不可修改或删除 |
| `SCHEDULER_JOB_NOT_ALLOWED` | 任务状态不允许当前操作 |
| `SCHEDULER_JOB_RUNNING` | 任务仍在执行 |
| `SCHEDULER_RUN_NOT_FOUND` | 执行记录不存在 |
| `SCHEDULER_RUN_NOT_RETRYABLE` | 该记录不能重试 |
| `SCHEDULER_RUN_NOT_CANCELABLE` | 该记录不能取消 |
| `SCHEDULER_RUN_RUNNING` | 记录在其它实例上运行，无法取消 |
| `SCHEDULER_CRON_INVALID` | Cron 或时区无效 |
| `SCHEDULER_SCRIPT_INVALID` | 脚本配置无效 |
| `SCHEDULER_PAYLOAD_INVALID` | 手动触发参数不是合法 JSON |
| `SCHEDULER_ENGINE_UNAVAILABLE` | 调度引擎未启用或数据库不可用 |
| `SCHEDULER_ACTION_FORBIDDEN` | 当前用户无权触发任务绑定的方法 |
