import { http } from "@bit-labs.cn/owl-ui/utils/http";

class InboxAPI {
  counts = () => {
    return http.request<Result>("get", "/api/v1/notify/my/counts", {}, { silentMessage: true });
  };
  getList = (params?: object) => {
    return http.request<ResultTable>("get", "/api/v1/notify/my/inbox", { params }, { silentMessage: true });
  };
  read = (data?: object) => {
    return http.request<Result>("post", "/api/v1/notify/my/inbox/read", { data });
  };
  readAll = (kind?: number) => {
    return http.request<Result>("post", "/api/v1/notify/my/inbox/read-all", {
      params: kind ? { kind } : {}
    });
  };
  remove = (id: string) => {
    return http.request<Result>("delete", `/api/v1/notify/my/inbox/${id}`);
  };
  done = (id: string) => {
    return http.request<Result>("post", `/api/v1/notify/my/todos/${id}/done`);
  };
  reopen = (id: string) => {
    return http.request<Result>("post", `/api/v1/notify/my/todos/${id}/reopen`);
  };
}

export const inboxAPI = new InboxAPI();
