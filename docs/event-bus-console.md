# 事件总线工作台与事件记录

`owl` 的可靠事件总线把待发事件写进发件箱 `owl_outbox_events`，由 Relay 投递到 Redis Streams / RabbitMQ；消费端用收件箱 `owl_inbox_events` 做幂等去重。这两张表由框架在 `runAutoMigrate` 里自动迁移，子系统不用注册。

本页是它的可视化管理入口，代码在 `owl-admin`：

- 后端：`app/repository/event.go`、`app/service/event_service.go`、`app/handle/v1/event_handle.go`、`app/route/api.go`
- 前端：`frontend/admin/src/views/event/`、`frontend/admin/src/api/event.ts`

## 菜单与权限

事件总线按应用隔离：每个部署用自己的 `conf/app.yaml` 的 `app-name` 决定 `Source`、消费组前缀与默认 Stream key（`owl:stream:events:<app-name>`）。工作台顶部会显示当前应用标识与实际 Stream key，便于确认「看的是哪条通道」。

侧栏顶级目录「事件总线」（路由 `/system/event`，name `EventBus`），下有两个菜单：

| 菜单 | 路由 | 组件 name | 接口前缀 |
| --- | --- | --- | --- |
| 工作台 | `/system/event/workbench/index` | `EventBusWorkbench` | `/api/v1/events/workbench`、`/api/v1/events/consumers/cleanup` |
| 事件记录 | `/system/event/log/index` | `EventBusLog` | `/api/v1/events/outbox`、`/api/v1/events/inbox`、`/api/v1/events/cleanup` |

接口都是「需要授权」。给角色分配菜单时，除了勾选两个菜单，还要勾选各自下面的按钮节点，否则页面能看到但接口会 403。

## 应用隔离与升级

事件通道按应用隔离，标识取自 `conf/app.yaml` 的 `app-name`（必填）：

| 项 | 取值 |
| --- | --- |
| 事件 `Source` | `app-name` |
| 消费组前缀 | `app-name` |
| Redis Stream key | `owl:stream:events:<app-name>` |
| RabbitMQ 交换机 | `owl.topic.events:<app-name>` |

- 多个部署共用同一个 Redis / RabbitMQ 时，各自只读写自己那条通道，不会看到对方的事件，也不会抢对方的消费组。
- 同一应用的多实例共用一个通道（集群分摊消费），所以 **`app-name` 必须一部署一值**。
- 需要对接历史 key，或临时把多个部署并到一条通道时，在 `conf/event.yaml` 写 `channel` 显式覆盖。
- `app-name` 为空会直接启动失败并提示具体缺什么，不会静默回退到公共名称——静默回退正是「所有应用混在一条 stream」的成因。

升级到按应用隔离后，历史 `owl:stream:events` 不再被读取：发件箱里待发送的事件会由 Relay 投到新通道；确认旧 stream 无待处理数据后，可用 `DEL owl:stream:events` 清理。

## 工作台

数据来自 `GET /api/v1/events/workbench?days=7`，`days` 取值 1～30，默认 7。

| 区块 | 数据来源 | 说明 |
| --- | --- | --- |
| 应用 / Stream | `*event.Bus` 的 `AppName()` 与 `Target()` | 展示当前应用标识与实际通道名，避免和 Redis 里看到的 key 对不上 |
| 待发送 / 已投递 / 投递失败 | 发件箱 `GROUP BY status` | 点击卡片跳到事件记录对应 Tab 与状态 |
| 消费失败 | 收件箱 `GROUP BY status` | 点击跳到收件箱的消费失败筛选 |
| 事件量趋势 | 发件箱 `created_at`、收件箱 `processed_at`（仅成功） | 按本地时区分天，取数上限 20000 行/侧，超出时返回 `trendTruncated` |
| 发件箱状态构成 | 同状态计数 | 环图 + 最老待发送事件的等待时长 |
| 事件类型 TOP 10 / 来源分布 | 发件箱 `GROUP BY event_type` / `source` | 只统计当前范围 |
| 消费组积压 | Redis `XLEN` + `XINFO GROUPS` + `XINFO CONSUMERS`；发件箱按 `event_type` 分组的待发送量 | Redis 不可用时只显示错误原因，其它区块照常 |

消费组表里的两列积压是两个维度，不要混读：

| 列 | 来源 | 含义 |
| --- | --- | --- |
| lag | Redis `XINFO GROUPS` | **整个 stream** 还没被这个组读到的条数（所有主题共用一条 stream，包含别的主题） |
| 本主题未投递 | 发件箱 `status = 待发送` 按 `event_type` 分组 | 这个主题还有多少条**还没被 Relay 投出去**；查不到就是 0 |
| pending | Redis `XINFO GROUPS` | 已投递但未 ACK 的条数 |

组合起来判断：

- `lag 大 + 本主题未投递 0 + 活跃消费者 0`：消息都发出去了，是消费端没人读 → 查订阅方是否启动。
- `本主题未投递 大`：Relay 没把事件投出去 → 查发件箱里的 `error_msg` 与中继日志。
- `pending 大`：投出去了但 handler 一直失败 → 查 `owl_inbox_events` 的失败流水与 `error_msg`。

「本主题未投递」靠组名反推主题：组名是 `app-name.主题`，去掉前缀即可。组名超长时 `Bus.Subscribe` 会截断成 `前缀.摘要`，此时前端显示 `—`。
| 最近投递失败 | 发件箱 `status = 投递失败` 最近 10 条 | 可直接打开详情或重投 |

驱动名取自 `*event.Bus` 的 `DriverName()`。当驱动是 `memory` 时页面顶部会提示：事件只在当前进程内存活，重启即丢、多实例不互通。

趋势按天分桶放在 Go 里做（`buildTrend`），不用 SQL 的 `DATE()` / `date_trunc`，避免引入数据库方言。

## 事件记录

### 发件箱

`GET /api/v1/events/outbox`，分页参数 `page` / `pageSize`，其余筛选：

| 参数 | 匹配方式 |
| --- | --- |
| `eventId` | 精确 |
| `eventType` | 模糊 |
| `source` | 精确 |
| `subject` | 模糊 |
| `status` | 精确，取值与 `OutboxStatus*` 一致 |
| `createdAt` | `start,end`（Unix 秒） |

行内可「详情 / 重投」。重投调用 `POST /api/v1/events/outbox/{id}/retry`：

- 只对未投递成功的行生效，SQL 条件带 `status <> 已发送`，已发送的事件不会重复投递。
- 成功后把 `status` 重置为待发送、`retry_count` 置 0、`next_retry_at` 清空、`error_msg` 清空。
- 真正投递仍由框架 Relay 完成（默认 2 秒轮询，`PublishTx` 后还会立即 `Notify`），页面在下一个刷新周期就能看到状态变化。

### 收件箱

`GET /api/v1/events/inbox`，筛选参数 `eventId`（精确）、`consumerGroup`（模糊）、`status`（1 消费成功、2 消费失败）、`processedAt`（`start,end` Unix 秒）。

收件箱只读：消息已经 ACK，单独把一条流水改成成功并不能补回缺失的投递上下文，重试应该由发布方重新发起。

### 清理历史

`POST /api/v1/events/cleanup`：

```json
{ "olderThanDays": 30, "targets": ["outbox", "inbox"] }
```

- `olderThanDays`：1～365。
- 发件箱只删 `status = 已发送 AND created_at < cutoff`。
- 收件箱只删 `status = 消费成功 AND processed_at < cutoff`。
- 待发送、投递失败、消费失败的流水一律保留，方便继续排查和重投。

## Redis Stream 字段读法

| 字段 | 含义 |
| --- | --- |
| `lag` | 组内还没投递给任何消费者的条数；持续增长说明消费者没起来或都卡住了 |
| `pending` | 已投递但未 ACK 的条数；长期不降说明 handler 一直返回错误（超过 5 次才 ACK 丢弃） |
| `entries-read` | 组已读到的位置，和 `lag` 一起看消费者是否推进 |
| `last-delivered-id` | 最后投递的消息 ID，用于和 stream 最新 ID 对比 |

`lag` 的实际算法就是 `stream 已写入总数 - 本组 entries-read`，所以它统计的是**整条 stream**，不是单个主题：所有消费组共用一条 stream，各自读全量再按主题过滤。要看某个主题自己的积压，用表里的「本主题未投递」列。

两个数字反映的是不同阶段的问题，别混着看：

- `pending > 0`：消息已经投出去了，卡在处理中或处理失败，30 秒后由 `XAUTOCLAIM` 重新认领重试。
- `lag > 0` 且 `pending = 0`：消息还在 stream 里没人读。说明这一组当前**没有活跃消费者**，页面会在「状态」列直接标出来。

### 消费者数与「僵尸消费者」

页面上的「消费者」列是「活跃 / 累计」两个数：

- **活跃**：最近 2 分钟内在读消息的消费者。活着的消费者一直在阻塞读（2 秒一次），空闲时间是秒级，所以这个数基本就是当前真实工作单元数。
- **累计**：`XINFO GROUPS` 里记录的消费者个数。Redis 不会自动回收消费者，而 [driver_redis.go](../owl/provider/event/driver_redis.go) 的消费者名是 `主机名-进程号-序号`，**进程每次重启都会生成新名字，旧名字永久留在组里**。所以累计数会随重启次数线性增长，不表示并发。

消费组面板右上角有「清理僵尸消费者」，调用 `POST /api/v1/events/consumers/cleanup`：

```json
{ "idleDays": 7 }
```

- 只删「空闲超过 `idleDays` 天」且「`pending = 0`」的消费者记录（`XGROUP DELCONSUMER`）。
- 有未确认消息的消费者一律跳过，避免丢掉待重试的 PEL。
- 删除消费者记录不影响正在运行的消费者，也不会丢消息。

排查「某组 lag 一直涨、活跃消费者为 0」时，先确认订阅该主题的子系统是否真的启动了（`subscribe.Init` 有没有执行、进程是否在跑），而不是先清消费者。

消费组名是「组前缀.主题」，前缀取 `conf/app.yaml` 的 `app-name`（为空回退 `owl`）。这是刻意的：同组是竞争消费语义，不同主题共用组会导致消息被投给不匹配的订阅者并被 ACK 掉，监听者静默收不到事件。

## 相关约定

- 页面只读发件箱 / 收件箱，写操作只有重投、清理历史与清理僵尸消费者，都不改变投递语义。
- 不提供「手工发布事件」入口：发布必须和业务写入放在同一个 `gorm.DB` 事务里（`PublishTx`），脱离业务上下文手工造事件会产生无法追溯的数据。
- 状态取值沿用 `owl/provider/event/model.go` 的 `OutboxStatus*` / `InboxStatus*`，前端文案集中在 `views/event/status.ts`，不散落裸数字。
