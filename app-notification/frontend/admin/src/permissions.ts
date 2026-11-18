/**
 * app-notification 按钮权限标识。
 * 与后端 route.InitAPI 中 RouteInfoBuilder 生成的 permission 严格一致：
 * `<appName>:<moduleEn>:<处理函数名>`。
 */
export const NOTIFY_PERMS = {
  announcement: {
    readState: "notify:announcement:ReadState",
    create: "notify:announcement:Create",
    update: "notify:announcement:Update",
    delete: "notify:announcement:Delete",
    publish: "notify:announcement:Publish",
    offline: "notify:announcement:Offline"
  },
  type: {
    create: "notify:notify-type:Create",
    update: "notify:notify-type:Update",
    delete: "notify:notify-type:Delete"
  },
  template: {
    create: "notify:notify-template:Create",
    update: "notify:notify-template:Update",
    delete: "notify:notify-template:Delete",
    preview: "notify:notify-template:Preview"
  },
  delivery: {
    detail: "notify:notify-delivery:Detail",
    retry: "notify:notify-delivery:Retry"
  },
  message: {
    detail: "notify:notify-message:Detail"
  },
  schedule: {
    cancel: "notify:notify-schedule:Cancel",
    sendNow: "notify:notify-schedule:SendNow",
    retry: "notify:notify-schedule:Retry"
  },
  send: {
    send: "notify:notify-send:Send"
  }
} as const;
