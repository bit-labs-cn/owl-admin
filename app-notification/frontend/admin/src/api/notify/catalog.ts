import { http } from "@bit-labs.cn/owl-ui/utils/http";

class TypeAPI {
  getList = (params?: object) => {
    return http.request<ResultTable>("get", "/api/v1/notify/types", { params }, { silentMessage: true });
  };
  options = () => {
    return http.request<Result>("get", "/api/v1/notify/type-options", {}, { silentMessage: true });
  };
  create = (data?: object) => {
    return http.request<Result>("post", "/api/v1/notify/types", { data });
  };
  update = (id: string, data?: object) => {
    return http.request<Result>("put", `/api/v1/notify/types/${id}`, { data });
  };
  remove = (id: string) => {
    return http.request<Result>("delete", `/api/v1/notify/types/${id}`);
  };
}

class TemplateAPI {
  getList = (params?: object) => {
    return http.request<ResultTable>("get", "/api/v1/notify/templates", { params }, { silentMessage: true });
  };
  create = (data?: object) => {
    return http.request<Result>("post", "/api/v1/notify/templates", { data });
  };
  update = (id: string, data?: object) => {
    return http.request<Result>("put", `/api/v1/notify/templates/${id}`, { data });
  };
  remove = (id: string) => {
    return http.request<Result>("delete", `/api/v1/notify/templates/${id}`);
  };
  preview = (data?: object) => {
    return http.request<Result>("post", "/api/v1/notify/templates/preview", { data });
  };
}

class DeliveryAPI {
  getList = (params?: object) => {
    return http.request<ResultTable>("get", "/api/v1/notify/deliveries", { params }, { silentMessage: true });
  };
  getDetail = (messageId: string) => {
    return http.request<Result>("get", `/api/v1/notify/deliveries/${messageId}`, {}, { silentMessage: true });
  };
  retry = (id: string) => {
    return http.request<Result>("post", `/api/v1/notify/deliveries/${id}/retry`);
  };
}

class MessageAPI {
  getList = (params?: object) => {
    return http.request<ResultTable>("get", "/api/v1/notify/messages", { params }, { silentMessage: true });
  };
  detail = (id: string) => {
    return http.request<Result>("get", `/api/v1/notify/messages/${id}`, {}, { silentMessage: true });
  };
}

class ScheduleAPI {
  getList = (params?: object) => {
    return http.request<ResultTable>("get", "/api/v1/notify/schedules", { params }, { silentMessage: true });
  };
  cancel = (id: string) => {
    return http.request<Result>("post", `/api/v1/notify/schedules/${id}/cancel`);
  };
  sendNow = (id: string) => {
    return http.request<Result>("post", `/api/v1/notify/schedules/${id}/send-now`);
  };
  retry = (id: string) => {
    return http.request<Result>("post", `/api/v1/notify/schedules/${id}/retry`);
  };
}

class SendAPI {
  recipients = (params?: object) => {
    return http.request<Result>("get", "/api/v1/notify/recipient-options", { params }, { silentMessage: true });
  };
  audience = () => {
    return http.request<Result>("get", "/api/v1/notify/audience-options", {}, { silentMessage: true });
  };
  estimate = (data?: object) => {
    return http.request<Result>("post", "/api/v1/notify/audience-estimate", { data }, { silentMessage: true });
  };
  send = (data?: object) => {
    return http.request<Result>("post", "/api/v1/notify/messages", { data });
  };
}

export const typeAPI = new TypeAPI();
export const templateAPI = new TemplateAPI();
export const deliveryAPI = new DeliveryAPI();
export const sendAPI = new SendAPI();
export const messageAPI = new MessageAPI();
export const scheduleAPI = new ScheduleAPI();
