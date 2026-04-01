import { http } from "@bit-labs.cn/owl-ui/utils/http";
/** 获取系统管理-部门管理列表 */
export const getApis = () => {
  return http.request<Result>("get", "/api/v1/api", {}, { silentMessage: true });
};

/** 获取系统监控-在线用户列表 */
export const getOnlineLogsList = (data?: object) => {
  return http.request<ResultTable>("post", "/online-logs", { data }, { silentMessage: true });
};

/** 获取系统监控-登录日志列表 */
export const getLoginLogsList = (data?: object) => {
  return http.request<ResultTable>(
    "post",
    "/api/v1/monitor/login-logs",
    { data },
    { silentMessage: true }
  );
};

/** 获取系统监控-操作日志列表 */
export const getOperationLogsList = (data?: object) => {
  return http.request<ResultTable>(
    "post",
    "/api/v1/monitor/operation-logs",
    { data },
    { silentMessage: true }
  );
};

/** 获取系统监控-系统日志列表 */
export const getSystemLogsList = (data?: object) => {
  return http.request<ResultTable>("post", "/system-logs", { data }, { silentMessage: true });
};

/** 获取系统监控-系统日志-根据 id 查日志详情 */
export const getSystemLogsDetail = (data?: object) => {
  return http.request<Result>("post", "/system-logs-detail", { data }, { silentMessage: true });
};
