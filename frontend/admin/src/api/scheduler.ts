import { http } from "@bit-labs.cn/owl-ui/utils/http";

/** 定时任务列表 */
export const getSchedulerJobs = (params?: object) => {
  return http.request<ResultTable>(
    "get",
    "/api/v1/scheduler/jobs",
    { params },
    { silentMessage: true }
  );
};

/** 定时任务详情（含脚本配置，仅超管） */
export const getSchedulerJobDetail = (id: number | string) => {
  return http.request<Result>(
    "get",
    `/api/v1/scheduler/jobs/${id}`,
    {},
    { silentMessage: true }
  );
};

/** 新建自定义定时任务 */
export const createSchedulerJob = (data?: object) => {
  return http.request<Result>("post", "/api/v1/scheduler/jobs", { data });
};

/** 修改自定义定时任务 */
export const updateSchedulerJob = (
  id: number | string,
  data?: object
) => {
  return http.request<Result>("put", `/api/v1/scheduler/jobs/${id}`, { data });
};

/** 删除自定义定时任务 */
export const deleteSchedulerJob = (id: number | string) => {
  return http.request<Result>("delete", `/api/v1/scheduler/jobs/${id}`);
};

/** 启用或停用任务 */
export const changeSchedulerJobStatus = (
  id: number | string,
  status: number
) => {
  return http.request<Result>("put", `/api/v1/scheduler/jobs/${id}/status`, {
    data: { status }
  });
};

/** 手动触发任务 */
export const triggerSchedulerJob = (
  id: number | string,
  payload?: string
) => {
  return http.request<Result>("post", `/api/v1/scheduler/jobs/${id}/trigger`, {
    data: { payload: payload ?? "" }
  });
};

/** 执行记录列表 */
export const getSchedulerRuns = (params?: object) => {
  return http.request<ResultTable>(
    "get",
    "/api/v1/scheduler/runs",
    { params },
    { silentMessage: true }
  );
};

/** 执行记录详情 */
export const getSchedulerRunDetail = (id: number | string) => {
  return http.request<Result>(
    "get",
    `/api/v1/scheduler/runs/${id}`,
    {},
    { silentMessage: true }
  );
};

/** 重试执行记录 */
export const retrySchedulerRun = (id: number | string) => {
  return http.request<Result>("post", `/api/v1/scheduler/runs/${id}/retry`);
};

/** 取消执行记录 */
export const cancelSchedulerRun = (id: number | string) => {
  return http.request<Result>("post", `/api/v1/scheduler/runs/${id}/cancel`);
};

/** 清理指定天数之前的执行记录 */
export const cleanupSchedulerRuns = (olderThanDays: number) => {
  return http.request<Result>("post", "/api/v1/scheduler/runs/cleanup", {
    data: { olderThanDays }
  });
};

/** 预览 Cron 接下来的执行时间 */
export const previewSchedulerCron = (data: {
  cron: string;
  timezone?: string;
  count?: number;
}) => {
  return http.request<Result>(
    "post",
    "/api/v1/scheduler/cron/preview",
    { data },
    { silentMessage: true }
  );
};

/** 子系统登记的可调用方法目录（仅超管） */
export const getSchedulerActions = () => {
  return http.request<Result>(
    "get",
    "/api/v1/scheduler/actions",
    {},
    { silentMessage: true }
  );
};

/** 可调用方法的动态候选，例如先选开关再列它的命令 */
export const getSchedulerActionOptions = (
  code: string,
  params?: Record<string, string>
) => {
  return http.request<Result>(
    "post",
    `/api/v1/scheduler/actions/${encodeURIComponent(code)}/options`,
    { data: { params: params ?? {} } },
    { silentMessage: true }
  );
};

/** 定时任务汇总 */
export const getSchedulerStats = () => {
  return http.request<Result>(
    "get",
    "/api/v1/scheduler/stats",
    {},
    { silentMessage: true }
  );
};
