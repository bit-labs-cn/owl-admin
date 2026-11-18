# 通知中心

`owl-admin` 的通知子系统，目录 `app-notification/`。它作为独立子应用注册，不走 ServiceProvider。

用两条线承载消息：知会通知（发生了什么）和行动待办（该做什么）。站内信写库并走 SSE 即时推送；邮件和企业微信群机器人通过事件总线异步发出去。

## 代码位置

- 子应用入口：`app-notification/app.go`
- 事件与订阅：`app-notification/{event,subscribe}`
- 模型 / 仓储 / 服务 / 接口：`app-notification/{model,repository,service,handle}`
- 路由与菜单：`app-notification/route/api.go`
- 对外契约：`app-notification/contract/`（只依赖标准库）
- 迁移：`app-notification/database/`
- 前端子系统包：`app-notification/frontend/admin/`

## 业务方怎么发

只依赖 `bit-labs.cn/owl-admin/app-notification/contract`。未注册时 `Get()` 返回 nil。

```go
if n := contract.Get(); n != nil {
    _ = n.SendDirect(ctx, &contract.DirectMessage{
        Kind:     contract.KindTodo,
        TypeCode: "approval.todo",
        Title:    "审批待处理",
        Content:  "单号 {{.OrderNo}} 需要你审批",
        Vars:     map[string]any{"OrderNo": "A1001"},
        BizKey:   "approval:" + id,
        UserIDs:  []uint{userID},
        Channels: []string{contract.ChannelSMS, contract.ChannelWeCom, contract.ChannelDingTalk},
    })
}

_ = contract.Get().CompleteByBizKey(ctx, "approval:"+id)
```

- `Kind` 只区分通知（`KindNotice`）和待办（`KindTodo`）。业务调用用它指定；管理端人工发送以所选消息类型的 `kind` 为准（`service.resolveKind`）。
- `TypeCode` 可空。有值时套用该类型的默认渠道和模板；空值则用调用方给的标题正文。
- 模板**不带渠道**：`notification_template` 一个类型只有一条，标题、正文、链接在发送时按目标渠道转换。
- `Channels` 可空。传了就按这条消息指定，空则取类型默认渠道，再空则只有站内信。
- `BizKey` 非空时，同一业务键只落一条消息，重复发送直接返回，用于业务重复触发去重。
- `Vars` 是模板变量，写成 `{{.OrderNo}}`。缺变量会渲染失败并返回业务错误，不会发出半截内容。
- `SkipTemplate` 为 true 时跳过类型模板，直接用 `Title`、`Content`、`Link`；人工发送使用它。
- `Content` 可以传富文本 HTML，也可以传纯文本：服务端 `ContentHTML` 会把纯文本转义并把换行转成 `<br>`，富文本按白名单清洗。站内信和邮件按 HTML 投递，企业微信、短信、钉钉自动转纯文本。
- `Payload` 是附加业务元数据，会随发送记录保存并传给渠道插件。
- 标题可以由类型模板提供；模板和 `Title` 都为空时返回业务错误。

## 手动发送

管理端「通知中心 → 发送通知」可以人工发一条通知或待办，走的是和业务代码同一条发送管线：
- 接口：`POST /api/v1/notify/messages`（需要权限）
- 收发件人候选：`GET /api/v1/notify/recipient-options?keyword=&ids=1,2`（只要求登录）
- 受众选项：`GET /api/v1/notify/audience-options`（只要求登录）
- 接收人预估：`POST /api/v1/notify/audience-estimate`（只要求登录，入参与发送的受众字段一致，返回 `userCount` / `externalCount` / `total`）

接收人预估只做统计、不写库：调用 `Repository.ResolveAudience` 解析当前条件，外部人员只统计填了邮箱的。发送页在选择变化后合并成一次请求（300ms 防抖），并用序号召回的旧响应直接丢弃。

接收人支持五种来源，可以同时选，最终合并去重：

| 来源 | 说明 |
| --- | --- |
| 指定用户 | **弹窗选择**：调用系统用户列表接口 `GET /api/v1/users` 按关键词、部门筛选后分页勾选，已禁用用户不可选 |
| 角色 | 选中角色下的启用用户 |
| 部门 | 选中部门下的启用用户，可勾选「同时发送给下级部门」 |
| 用户组 | 选中用户组下的启用用户 |
| 外部人员 | 在选择弹窗里逐条添加，或用 Excel 导入（下载模板 → 选 .xlsx）；没有系统账号，只按邮箱投递 |

受众解析在 `repository/audience_repository.go`：`ResolveAudience` 把用户、角色、部门（可递归下级）和用户组合并、过滤启用用户并去重。

**指定用户走弹窗，不用下拉**：人多时下拉装不下，`views/send/RecipientPicker.vue` 把「筛选 → 勾选 → 导入外部人员 → 得到最终目标」收在一套动作里。左侧是分页用户表，数据**直接取系统用户列表** `GET /api/v1/users`（`keyword` 匹配登录名/昵称、`deptId` 自带下级部门，见 `app/service/user_service.go` 的 `RetrieveUserReq`），通知模块不另开用户接口、也不直接查 `admin_user`；表头可全选本页，禁用账号置灰不可选。右侧是已选站内用户标签和外部人员编辑区：逐条添加，或「导入 Excel」——先下载模板（列：称呼 / 邮箱 / 手机号，邮箱必填），再选 `.xlsx` 导入。翻页、改筛选都不会丢掉已选；确认后弹窗把 `{ userIds, users, externals }` 回填到表单，发送页只展示目标标签，不再逐个选人。

外部人员的 Excel 解析在浏览器里做（`utils/recipientExcel.ts`，exceljs 按需动态加载）：第一张工作表、第一行是表头，中英文列名都认，列顺序不限；缺邮箱、邮箱格式错误、表内重复的行会被跳过并列出前 5 条原因，已在其列表里的邮箱按重复跳过；单次最多 2000 行。这块数据只是表单里的收件人，不入库，所以不需要后端接口。

外部人员没有站内信，只按填写的邮箱发送。若没有勾选邮件渠道、邮件渠道已关闭或没填邮箱，都会在发送记录里留下「跳过」并写明原因。

页面填消息类型、标题、正文、跳转链接、渠道、接收人；待办类型还会要求填截止时间。**种类不在发送页单独选**：它由所选消息类型自带的 `kind` 决定（`service.resolveKind`），选了待办类型就按待办发送并校验截止时间，不选类型按普通通知发送。正文以页面填写为准，不套用类型模板（`DirectMessage.SkipTemplate`）。

**渠道以消息类型为准**：`GET /api/v1/notify/type-options` 的类型项带 `defaultChannels`，发送页把它当作不可取消的渠道直接带入（勾选框禁用，标「类型默认」），用户只能在其上**追加**没有的渠道；提交的 `channels` 是「类型自带 + 额外勾选」的并集（`typeChannels` ∪ `form.channels`）。没选类型时渠道全部由页面勾选，都不勾则由服务端回落到站内信。类型下拉在每个选项右侧标出待办或该类型的默认渠道，选待办类型时提示需要截止时间。

发送时间可以选「立即发送」或「定时发送」。选定时发送时填一个时间点，页面提供 10 分钟 / 30 分钟 / 1 小时 / 1 天和明天 9:00 的快捷填充；接口传 `planAt` 就走定时，不传就是立即发送。两种模式共用同一套表单和同一套接收人选择，不另开编辑页。

**正文是富文本**：发送页用 `components/RichTextEditor.vue`（wangEditor 5），只开标题、加粗、斜体、下划线、删除线、颜色、列表、引用、链接、图片、撤销重做。图片走框架的 `POST /api/v1/files/upload`，`api/notify/file.ts` 会把返回的相对地址补成绝对地址再插入，保证邮件客户端能显示。正文存库前由后端 `ContentHTML` 归一与清洗；站内信和邮件按 HTML 显示，企业微信、短信、钉钉自动转纯文本。收件箱详情、消息接收情况、右侧预览统一用 `components/RichContentView.vue` 渲染（`format.ts` 的 `isRichContent` 判断，早期纯文本仍按纯文本展示）。

页面是两栏发送工作台，没有页头标题条和步骤条，打开即是表单：

- 左栏宽一些（1.65fr），放「接收范围」和「通知内容」两张卡；右栏（1.35fr，最小 340px）放「发送渠道」「投递策略」「实时预览」
- 接收范围卡用分段页签切换四种来源（指定用户 / 按角色 / 按部门 / 按用户组）。「指定用户」是一个带「选择接收人」按钮的框：显示已选用户数和目标标签，点开 `RecipientPicker` 弹窗完成筛选、勾选和外部人员导入（外部人员不再单独一个页签）；下方汇总区显示实时预估人数，圈选条件（角色/部门/用户组）列成可单独移除的标签；条件命中不到启用用户时明确提示
- 发送渠道卡每个渠道一行，写明投递方式和是否已接入（短信、钉钉标「未接入」）。选了消息类型时，该类型自带的渠道显示为已勾选且不可取消，标「类型默认」；其余渠道可继续勾选，最终提交两者的并集
- 投递策略卡用两张单选卡切换立即 / 定时；定时才有时间选择和快捷填充。下方投递信息显示站内用户数、外部人员数和「准备完成度」，没有虚构的批次与耗时
- 实时预览卡（`views/send/PreviewPanel.vue`）可切换已选渠道，站内信再分「小卡片 / 详情」两种形态，底部固定提示发送前核对内容
- 底部有一条 `position: fixed` 的发送条，紧贴视口底部、不随内容滚动、下面不留缝：显示预计接收人数、渠道数、发送方式，以及「清空」和主发送按钮（`submit`）；左侧按侧边栏宽度让位（展开 210px、折叠 54px，横向布局贴满），页面另有 `padding-bottom` 避免内容被它盖住
- 窗口窄于 1180px 时两栏堆叠；窄于 700px 时表单项、单选卡变成单列，发送条折成两行

## 定时发送

「通知中心 → 定时发送」（菜单 `NotifySchedule`，`/system/notify/schedule/index`）管理还没到点的任务。数据落在 `notification_schedule`，一条任务一行。

- 创建：发送页选「定时发送」并填时间，即 `POST /api/v1/notify/messages` 带 `planAt`；不传 `planAt` 是立即发送
- 列表：`GET /api/v1/notify/schedules?keyword=&status=&page=&pageSize=`（`kind` 参数仍兼容，页面不再提供种类筛选）
- 取消：`POST /api/v1/notify/schedules/:id/cancel`（只有待发送、失败的能取消）
- 立即发送：`POST /api/v1/notify/schedules/:id/send-now`（跳过计划时间）
- 重试：`POST /api/v1/notify/schedules/:id/retry`（只有失败的能重试，会清零尝试次数）

任务保存的是**接收人选择条件**（用户、角色、部门、用户组、外部人员），不是当时解析出的用户 ID 列表。到点发送时才重新解析，所以计划期间入职、调岗的人按发送当刻的组织关系收件。列表里另存一份计划时的接收人摘要（`audience_summary` + `recipient_count`），只用于展示，不参与发送。

调度在 `service/worker.go`，每 10 秒扫一次到点任务（`claim_due_schedules`）。领取用带状态条件的更新抢占：只有 `status=1 待发送` 且 `plan_at <= now` 的行会被改成 `2 发送中` 并 `attempts+1`，多实例同时轮询也只有一个实例领到同一条。发送走 `SendDirect`，业务键固定为 `notify-schedule:<任务ID>`，即使重复领取也不会发出第二条消息。

失败时未满 3 次按尝试次数退避（第 n 次失败后等 n 分钟）再退回待发送，满 3 次标为发送失败，可在列表里手动重试。发送成功后回填 `message_id`，能直接对到「消息记录」里的那一条。

状态：1 待发送、2 发送中、3 已发送、4 已取消、5 发送失败。

## 消息记录与已读情况

「通知中心 → 消息记录」（菜单 `NotifyMessage`，`/system/notify/message/index`）是管理端视角，列出所有发出去的消息：

- 每行显示接收人数、已读进度、未读数、待办待处理 / 已完成、投递失败数、发送人和发送时间
- 「接收情况」打开明细：消息正文、未读人员标签，以及每个接收人的阅读状态、阅读时间、待办状态和完成时间
- 支持按标题、消息类型筛选（种类由消息类型决定，不再单独提供种类筛选，只在行内展示）

接口：

- `GET /api/v1/notify/messages`：消息列表与统计
- `GET /api/v1/notify/messages/:id`：消息正文与接收人明细

统计用子查询在数据库侧聚合，列表页不需要额外请求逐条统计。

「发送记录」（`NotifyDelivery`）按发送批次展示，一条消息一行，点击「明细」查看该次发送的全部渠道记录：

- 批次行显示投递进度（成功 / 总计）、失败 / 跳过 / 待发送数量、渠道数、发送人
- 筛选条件（状态、渠道、接收人类型）作用于明细，批次合计也只算命中的明细
- 明细里，站内信写入后立即记「成功」；邮件、企业微信先记「待发送」，由事件总线消费者投递并回写结果；被系统设置关闭的渠道记「跳过」并写明原因；短信、钉钉记「跳过」并注明未接入
- 失败行可以在明细里单独重试

接口：

- `GET /api/v1/notify/deliveries`：批次列表（`messageId,title,total,successCount,failedCount,skippedCount,pendingCount,channelCount,...`）
- `GET /api/v1/notify/deliveries/:messageId`：该次发送的全部投递明细
- `POST /api/v1/notify/deliveries/:id/retry`：重试某一条失败的投递

批次查询用 `GROUP BY` 聚合。MySQL 默认开启 `ONLY_FULL_GROUP_BY`，所以 `SELECT` 里出现的非聚合列（包括取发送人的子查询用到的 `m.sender_id`）都必须列进 `GROUP BY`，否则整条列表接口会 500、页面显示为空。

发送记录只反映渠道投递，不包含阅读状态；某个接收人看没看、待办办没办，看「消息记录」。

## 渠道

| 编码 | 本轮行为 |
| --- | --- |
| `inapp` | 写入站内信并立刻 SSE 推送 |
| `email` | 框架 `mailer.MailerManager` 发送，邮箱取自 `admin_user.email` |
| `wecom` | 企业微信群机器人 Webhook，一条消息只发一次 |
| `sms` | 未接入发送器，发送记录标为跳过 |
| `dingtalk` | 未接入发送器，发送记录标为跳过 |

渠道决定顺序：

1. 调用方传了 `Channels` 就只用这些渠道。
2. 没传则用消息类型的默认渠道。
3. 类型也没有默认渠道时，只发站内信。
4. 再与系统设置里的渠道开关求交：邮件、企业微信被关掉时记「跳过」，不发给任何人。
5. 没有发送器的渠道记一条「跳过」记录，写明未接入，不发布事件。

## 发送管线

1. 解析类型、默认渠道和模板（模板只按类型取一条）。
2. 正文统一归一成 HTML（`service.ContentHTML`）：`IsHTML` 认得出的富文本按 `SanitizeHTML` 白名单清洗；纯文本按 HTML 转义并把换行转成 `<br>`，兼容早期只存纯文本的调用方。
3. 渲染模板（`renderTemplate`）：标题、链接用 `text/template`，正文用 `html/template`，变量按 HTML 转义；没有模板时回退到调用方文案。渲染后的正文再过一次 `ContentHTML` 归一。
4. 按渠道取文案：站内信和邮件直接用 HTML；企业微信、短信、钉钉由 `renderedSet.forChannel` 调 `HTMLToText` 转成纯文本。
5. 站内信写 `notification_message` + `notification_recipient`，然后走 SSE；同时写一条「成功」的发送记录。
6. 需要走外部渠道的记录写 `notification_delivery`，待发送的在同一事务里 `PublishTx` 主题 `notify.channel.send`，载荷只有发送记录 ID。
7. `subscribe/subscribe.go` 的消费者按渠道调用插件并回写发送记录。邮件的 `ChannelRequest` 带 `Content`（HTML）和 `ContentText`（纯文本），分别作为 `text/html` 与 `text/plain` 两个部分投递。
8. 被系统渠道开关关闭的渠道不发布事件，只留「跳过」记录。

重试：暂时失败时次数加一，未满 5 次返回错误交给总线重投；满 5 次标为失败并正常返回，避免 RabbitMQ 重新入队后无限重试。已成功或已跳过的记录直接返回，靠总线 Inbox 与记录状态双重防重复发送。

发送记录是给人看的投递结果，管理端可查看状态并在失败时手动重试；它不承担排队职责。

## 渠道配置

`SubAppNotification` 实现 `setting.Registrar`，配置进系统设置页的「通知中心」：

- `emailEnabled`：邮件渠道开关，默认开
- `wecomEnabled`：企业微信渠道开关，默认关
- `wecomWebhook`：群机器人地址，不公开

## 管理端与用户端

菜单根：`/system/notify`（name `NotifyCenter`）。页面组件名必须与菜单 name 一致。

| 菜单 name | 路径 | 用途 |
| --- | --- | --- |
| `NotifyInbox` | `/system/notify/inbox/index` | 我的消息：按标题、消息类型、阅读状态筛选，可标已读、完成或重开待办、删除 |
| `NotifySend` | `/system/notify/send/index` | 发送通知：人工发送通知或待办 |
| `NotifyMessage` | `/system/notify/message/index` | 消息记录：已发消息的接收、已读与待办情况 |
| `NotifyType` | `/system/notify/type/index` | 消息类型：编码、种类、默认渠道、启用状态 |
| `NotifyTemplate` | `/system/notify/template/index` | 消息模板：一个类型一条，正文富文本，支持示例变量预览 |
| `NotifyDelivery` | `/system/notify/delivery/index` | 发送记录：按发送批次聚合，点开看该次发送的全部渠道明细，失败可重试 |
| `NotifySchedule` | `/system/notify/schedule/index` | 定时发送：还没到点或失败的预约任务，可取消、立即发送、重试 |
| `NotifyAnnouncement` | `/system/notify/announcement/index` | 站点公告 |

顶栏铃铛只展示未读消息和待关闭的公告；已读消息不再出现，点击某条后会标为已读并从列表消失。公告横幅由布局扩展位 `banner` 承担。

收件箱与发送记录列表都用可读名称：消息类型显示类型名，找不到时显示 `未知类型(编码)`；发送记录显示接收人快照，外部人员显示填写的称呼，站内用户显示昵称，已删除或查不到时显示 `未知用户(id)`，群机器人显示「群机器人」。发送记录还能按接收人类型筛选。

### 消息模板

「通知中心 → 消息模板」（`NotifyTemplate`）维护每个消息类型的一份文案：标题、正文、跳转链接，变量写成 `{{.OrderNo}}`。

- 接口：`GET /api/v1/notify/templates?typeCode=&page=&pageSize=`、`POST /api/v1/notify/templates`、`PUT /api/v1/notify/templates/:id`、`DELETE /api/v1/notify/templates/:id`
- 预览：`POST /api/v1/notify/templates/preview`，入参 `title` / `body` / `link` / `vars`，返回渲染后的 `title` / `body`（HTML）/ `link`
- **模板不选渠道**：一个类型只有一条模板（`type_code` 唯一），发往哪些渠道由消息类型默认值和发送时勾选决定；同一份文案在投递时按渠道转换（邮件、站内信 HTML，企业微信、短信、钉钉纯文本）
- 正文用 `components/RichTextEditor.vue` 富文本编辑，服务端按 `html/template` 校验语法；列表不再有渠道筛选与渠道列，新增时如果该类型已有模板会提示直接编辑

### 按钮权限

后端 `AccessAuthorized` 的接口会由 `RouterInfoBuilder` 生成 `notify:<moduleEn>:<处理函数名>` 权限标识。前端把这些标识集中放在 `frontend/admin/src/permissions.ts` 的 `NOTIFY_PERMS`，页面用全局 `<Perms>` 包裹对应按钮；标识变了只改这一处。

| 按钮 | 权限标识 |
| --- | --- |
| 公告：已读情况 / 新增 / 编辑 / 删除 / 发布 / 下线 | `notify:announcement:ReadState`、`Create`、`Update`、`Delete`、`Publish`、`Offline` |
| 消息类型：新增 / 编辑 / 删除 | `notify:notify-type:Create`、`Update`、`Delete` |
| 消息模板：新增 / 编辑 / 删除 / 预览 | `notify:notify-template:Create`、`Update`、`Delete`、`Preview` |
| 发送记录：明细 / 重试 | `notify:notify-delivery:Detail`、`Retry` |
| 消息记录：接收情况 | `notify:notify-message:Detail` |
| 发送通知：发送（含定时发送） | `notify:notify-send:Send` |
| 定时发送：立即发送 / 重试 / 取消 | `notify:notify-schedule:SendNow`、`Retry`、`Cancel` |

「我的消息」「发送通知」页里的收件人、受众选项只要求登录（`AccessAuthenticated`），对应按钮不配权限。搜索、重置、清空等不产生授权动作的按钮同样不配。

## 公告

`notification_announcement` 用读扩散：只存一条公告，用户已读和关闭状态写 `notification_announcement_state`。

| displayMode | 行为 |
| --- | --- |
| 1 仅公告中心 | 只在收件箱 / 公告列表出现 |
| 2 顶部横幅 | 由布局 `banner` 扩展位展示，用户关闭后写状态 |
| 3 强提醒弹窗 | 前端弹出提示；同样记录关闭状态 |

`GET /api/v1/notify/my/announcements` 只返回已发布、在生效时间内、且当前用户尚未关闭的公告。

### 已读情况

公告列表每行有「已读情况」，接口 `GET /api/v1/notify/announcements/:id/read-state`：

- 先把公告受众解析成具体用户：`all` 取全部启用用户，`dept` 取部门及其下级，`role` 取角色成员
- 再与 `notification_announcement_state` 比对，没有已读记录的人算未读
- 返回接收人总数、已读、未读、已关闭横幅数，以及每个接收人的阅读状态、阅读时间、是否关闭过横幅

注意当前语义：公告的「已读」是在用户点「知道了」或关闭横幅 / 弹窗时写入 `read_at` 的。只是把公告列表打开看，不会产生已读记录，所以「仅公告中心」的公告在没人点关闭前，已读数会是 0。

## 实时

浏览器先 `POST /api/v1/notify/my/stream-ticket` 拿 60 秒一次性票据，再 `GET /api/v1/notify/stream?ticket=`。票据用后即删，JWT 不进 URL。

多实例通过 Redis 频道 `notify:push` 扇出。断线后客户端重新取票据，并重新拉未读数和列表。

## 表

| 表 | 说明 |
| --- | --- |
| `notification_message` | 消息本体，含 `type_code` 和业务键 `biz_key` |
| `notification_recipient` | 接收流水与阅读 / 待办状态 |
| `notification_announcement` | 公告本体 |
| `notification_announcement_state` | 公告的用户已读 / 关闭状态 |
| `notification_type` | 消息类型，编码唯一 |
| `notification_template` | 文案模板，`type_code` 唯一，正文是富文本 |
| `notification_delivery` | 全渠道投递结果，含渲染文案、接收人快照与附加业务元数据 |
| `notification_schedule` | 定时发送任务，含计划时间、接收人条件、状态与发送结果 |

启动迁移后会补齐两条内置类型：`system.notice`（系统通知）和 `system.todo`（待办），默认渠道都是站内信。不预写模板，缺模板时回退到调用方标题正文。

表名前缀是 `notification_`，表名由各模型的 `TableName()` 给出，新增表沿用同一前缀。迁移只按当前模型建表，不兼容早期「一个类型一个渠道一条模板」的结构。

## 不做

- 周期重复（每天 / 每周 / 每月）、合并摘要、节流、多步骤工作流；定时发送目前只支持一次性时间点
- 短信、钉钉、个人微信、飞书
- 公告模型的读扩散之外改造
