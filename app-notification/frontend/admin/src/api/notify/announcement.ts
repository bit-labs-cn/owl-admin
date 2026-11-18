import { http } from "@bit-labs.cn/owl-ui/utils/http";

class AnnouncementAPI {
  getList = (params?: object) => {
    return http.request<ResultTable>("get", "/api/v1/notify/announcements", { params }, { silentMessage: true });
  };
  detail = (id: string) => {
    return http.request<Result>("get", `/api/v1/notify/announcements/${id}`, {}, { silentMessage: true });
  };
  create = (data?: object) => {
    return http.request<Result>("post", "/api/v1/notify/announcements", { data });
  };
  update = (id: string, data?: object) => {
    return http.request<Result>("put", `/api/v1/notify/announcements/${id}`, { data });
  };
  remove = (id: string) => {
    return http.request<Result>("delete", `/api/v1/notify/announcements/${id}`);
  };
  publish = (id: string) => {
    return http.request<Result>("post", `/api/v1/notify/announcements/${id}/publish`);
  };
  offline = (id: string) => {
    return http.request<Result>("post", `/api/v1/notify/announcements/${id}/offline`);
  };
  mine = (params?: object) => {
    return http.request<ResultTable>("get", "/api/v1/notify/my/announcements", { params }, { silentMessage: true });
  };
  dismiss = (id: string) => {
    return http.request<Result>("post", `/api/v1/notify/my/announcements/${id}/dismiss`);
  };
  readState = (id: string) => {
    return http.request<Result>("get", `/api/v1/notify/announcements/${id}/read-state`, {}, { silentMessage: true });
  };
}

export const announcementAPI = new AnnouncementAPI();
