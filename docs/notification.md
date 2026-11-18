# 通知中心

`owl-admin` 仓库里的通知子系统，目录 `app-notification/`。它作为独立子应用注册，不走 ServiceProvider。

告诉用户发生了什么（通知、消息），以及需要做什么（待办）。站内信用 SSE 即时推到浏览器，站外走邮件、短信和 App 推送。

短信要真正发出去，需要同进程里装了 `owl-sms`；它在启动时把自己的发送服务注册给通知中心。其它渠道不依赖任何子系统。

## 代码位置

- 子应用入口：`app-notification/app.go`
- 模型/仓储/服务/接口：`app-notification/{model,repository,service,handle}`
- 路由与菜单：`app-notification/route/api.go`
- 对外契约：`app-notification/contract/`（只依赖标准库）
- 迁移与空值回填：`app-notification/database/`

前端子系统包已撤下，待重做。后端接口、表、契约保持可用。

宿主只需要注册后端子应用：

```go
owl.NewApp(
    &admin.SubAppAdmin{},
    &notification.SubAppNotification{},
).WebShell()
```

## 业务方怎么发

只依赖 `bit-labs.cn/owl-admin/app-notification/contract`。未安装通知中心时 `Get()` 返回 nil，调用方跳过即可。

```go
import "bit-labs.cn/owl-admin/app-notification/contract"

if n := contract.Get(); n != nil {
    _, err := n.Send(ctx, &contract.Message{
        EventCode:   "inspection.task_pending",
        Source:      "inspection",
        Kind:        contract.KindTodo,
        Category:    contract.CategoryBiz,
        Title:       "请在 18:00 前完成 3 号机房巡检",
        ContentHTML: "<p>请在 <strong>18:00</strong> 前完成 <strong>3号机房</strong> 巡检。</p>",
        BizKey:      "inspection_task:" + id,
        Audience:    contract.Audience{UserIDs: contract.UintIDs{userID}},
        Channels:    []string{contract.ChannelInApp, contract.ChannelSMS, contract.ChannelApp},
        Link: &contract.DeepLink{
            App:   "inspection",
            Path:  "/inspection/task/index",
            Query: map[string]string{"id": id},
        },
    })
}
```

处理完成后勾掉待办：

```go
_ = contract.Get().CompleteByBizKey(ctx, "inspection_task:"+id)
```

事件可以事先登记，后台就能按事件选模板和默认渠道：

```go
_ = n.RegisterEvents(ctx, []contract.EventDescriptor{{
    Code: "inspection.task_pending", Name: "待巡检", Source: "inspection",
    Kind: contract.KindTodo, Category: contract.CategoryBiz,
    DefaultChannels: []string{contract.ChannelInApp, contract.ChannelSMS, contract.ChannelApp},
}})
```

## 渠道

| 编码 | 行为 |
| --- | --- |
| `inapp` | 写入站内信并立刻 SSE 推送 |
| `email` | 框架 `mailer.MailerManager`，不需要在通知里另填 SMTP 账号 |
| `sms` | 由 `owl-sms` 启动时注册的 `TextSender` 发送；任务上要选短信模板编码 |
| `app` | APNs（token）与 FCM（HTTP v1），按设备 provider 分发 |

渠道账号的 `configJson` 由前端按渠道生成，后台不再让用户直接编辑 JSON：

- APNs：`teamId`、`keyId`、`bundleId`、`privateKey`（p8 PEM）、`production`
- FCM：`projectId`、`clientEmail`、`privateKey`
- 短信：`templateCode`，模板编码从 `owl-sms` 的 `sms_template` 读取
- 邮件：使用系统 `mailer`，配置不承载账号密码

账号列表接口会把配置遮罩；编辑弹窗先调 `GET /api/v1/notify/channel-accounts/:id` 取回完整配置，避免把遮罩值当成新配置保存。

## 公告怎么让用户看到

公告有三种展示方式，后台和顶栏都必须按同一份约定渲染：

| displayMode | 行为 |
| --- | --- |
| 1 顶部横幅 | 由 `owl-ui` 的 `AnnounceBar` 展示，用户关闭后写 `notify_announcement_state` |
| 2 弹窗 | 由 `AnnouncementPopup` 展示；只对弹窗生效 `frequency`（一次 / 每会话 / 每次登录） |
| 3 仅公告中心 | 不主动打扰，用户从顶栏铃铛的「公告」页签查看 |

顶栏铃铛固定有「公告」页签，列出生效中的公告，所以公告即使不在横幅或弹窗里，也一定能被用户找到。

`GET /api/v1/notify/my/announcements` 返回当前用户可见的公告，并已按 `audienceJson` 过滤：

- `all=true` 或受众为空：所有启用用户
- 指定用户、角色、部门（含下级部门）、用户组：取并集
- 已关闭且允许关闭的公告会被跳过

发布、下线、删除、改回草稿都会通过 Redis 广播 `announcement` 事件，前端收到后重新拉取，不会继续显示已经下线的公告。

## 受众与跳转选项

后台表单不允许让人手填用户 ID、角色 ID 或 JSON。通知子系统提供三个只要求登录的选项接口：

- `GET /api/v1/notify/audience-options`：用户前 200 个、角色、部门树、用户组、外部名单
- `GET /api/v1/notify/audience-users?keyword=&ids=`：按关键字搜索用户，或按 ID 回显已选用户
- `GET /api/v1/notify/link-options`：可跳转的站内页面（来自 `admin_menu`）
- `GET /api/v1/notify/sms-templates`：`owl-sms` 已启用的短信模板；未安装时返回空数组

前端对应组件：`AudiencePicker.vue`、`DeepLinkPicker.vue`、`ChannelConfigFields.vue`。发送任务时还会调 `POST /api/v1/notify/tasks-estimate` 显示实际收件人数和缺邮箱/手机号的人数。

## 实时

浏览器先 `POST /api/v1/notify/my/stream-ticket` 拿 60 秒一次性票据，再 `GET /api/v1/notify/stream?ticket=`。票据用后即删，JWT 不进 URL。

多实例通过 Redis 频道 `notify:push` 扇出。断线后客户端重新取票据，并重新拉未读数和列表，服务端不保存重放日志。

反向代理要关闭响应缓冲，例如 Nginx `proxy_buffering off`，并保留响应头 `X-Accel-Buffering: no`。读超时要大于 20 秒心跳。

## HTML

正文用白名单消毒，保留排版、链接、图片、表格，去掉脚本、iframe 和 `javascript:`。同时存纯文本，给短信和推送用。跳转不解析 HTML，走 `DeepLink`。

## 表

`notify_template`、`notify_task`、`notify_delivery`、`notify_inbox`、`notify_device`、`notify_announcement`、`notify_announcement_state`、`notify_subscription`、`notify_setting`、`notify_contact_list`、`notify_contact`、`notify_channel_account`、`notify_event`。

主体用 `principal_type + principal_id`。1 是后台用户，2 预留给门户用户。

关联列表必须显示可读名称，不允许把裸 ID 丢给用户：

- 投递记录：`taskTitle` 由任务表补全，地址仍脱敏
- 推送设备：`principalName` 由 `admin_user` 补全，token 只显示遮罩
- 我的消息：支持按标题、阅读状态和类型筛选；详情直接渲染消毒后的 HTML，待办可在详情里完成或重开

## 前端

前端子系统包已撤下，待重做。撤下的原因和现状：

- 原实现依赖 `owl-ui` 的 `notification` 扩展点、`core/router/deepLink` 与 `core/store/modules/notice`，这些在 `owl-ui` 中已全部撤销，`owl-ui` 不再包含任何通知相关代码。
- 后端菜单根节点暂时 `ShowLink: false`，侧栏不显示「通知中心」，避免指向不存在的前端组件。接口、权限与按钮菜单都保留。
- 前端恢复时，把 [app-notification/route/api.go](../app-notification/route/api.go) 里 `menuRoot` 的 `ShowLink` 改回 `true`，并让前端 `viewModulesPathPrefix` 与后端 `/system/notify` 对齐。
