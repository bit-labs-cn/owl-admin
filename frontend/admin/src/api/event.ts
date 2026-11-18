import { http } from "@bit-labs.cn/owl-ui/utils/http";

/** 事件总线工作台聚合数据 */
export const getEventWorkbench = (params?: object) => {
  return http.request<Result>(
    "get",
    "/api/v1/events/workbench",
    { params },
    { silentMessage: true }
  );
};

/** 发件箱列表 */
export const getOutboxList = (params?: object) => {
  return http.request<ResultTable>(
    "get",
    "/api/v1/events/outbox",
    { params },
    { silentMessage: true }
  );
};

/** 发件箱详情 */
export const getOutboxDetail = (id: number | string) => {
  return http.request<Result>(
    "get",
    `/api/v1/events/outbox/${id}`,
    {},
    { silentMessage: true }
  );
};

/** 重投发件箱事件 */
export const retryOutbox = (id: number | string) => {
  return http.request<Result>("post", `/api/v1/events/outbox/${id}/retry`);
};

/** 清理事件历史 */
export const cleanupEvents = (data?: object) => {
  return http.request<Result>("post", "/api/v1/events/cleanup", { data });
};

/** 收件箱列表 */
export const getInboxList = (params?: object) => {
  return http.request<ResultTable>(
    "get",
    "/api/v1/events/inbox",
    { params },
    { silentMessage: true }
  );
};

/** 收件箱详情 */
export const getInboxDetail = (id: number | string) => {
  return http.request<Result>(
    "get",
    `/api/v1/events/inbox/${id}`,
    {},
    { silentMessage: true }
  );
};

/** 清理僵尸消费者（Redis 里退出进程留下的消费者记录） */
export const cleanupConsumers = (data?: object) => {
  return http.request<Result>("post", "/api/v1/events/consumers/cleanup", { data });
};
