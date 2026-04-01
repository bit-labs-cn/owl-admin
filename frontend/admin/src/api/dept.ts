import { http } from "@bit-labs.cn/owl-ui/utils/http";

class DeptAPI {
  createDept = (data?: object) => {
    return http.request<Result>("post", "/api/v1/dept", { data });
  };

  deleteDept = (id: string) => {
    return http.request<Result>("delete", `/api/v1/dept/${id}`);
  };

  getDepts = (data?: object) => {
    return http.request<Result>("get", "/api/v1/dept", { data }, { silentMessage: true });
  };

  updateDept = (data?: object) => {
    return http.request<Result>("put", `/api/v1/dept/${data["id"]}`, {
      data
    });
  };

  changeStatus = (data?: object) => {
    return http.request<Result>("put", `/api/v1/dept/${data["id"]}/status`, {
      data
    });
  };
}

export const deptAPI = new DeptAPI();
