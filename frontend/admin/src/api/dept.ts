import { http } from "@bit-labs.cn/owl-ui/utils/http";

class DeptAPI {
  createDept = (data?: object) => {
    return http.request<Result>("post", "/api/v1/dept", { data });
  };

  deleteDept = (id: string) => {
    return http.request<Result>("delete", `/api/v1/dept/${id}`);
  };

  /** GET 使用 params，与后端 ShouldBindQuery 一致；部门树需较大 pageSize */
  getDepts = (params?: object) => {
    return http.request<Result>(
      "get",
      "/api/v1/dept",
      { params: { page: 1, pageSize: 1000, ...params } },
      { silentMessage: true }
    );
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
