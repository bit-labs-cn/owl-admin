import { http } from "@bit-labs.cn/owl-ui/utils/http";

class PositionAPI {
  /** 创建岗位 */
  createPosition = (data?: object) => {
    return http.request<Result>("post", "/api/v1/position", { data });
  };

  /** 更新岗位 */
  updatePosition = (data?: object) => {
    return http.request<Result>("put", `/api/v1/position/${data["id"]}`, {
      data
    });
  };

  /** 删除岗位 */
  deletePosition = (id: string | number) => {
    return http.request<Result>("delete", `/api/v1/position/${id}`);
  };

  /** 修改岗位状态 */
  changeStatus = (id: number, status: number) => {
    return http.request<Result>("put", `/api/v1/position/${id}/status`, {
      data: { id, status }
    });
  };

  /** 获取岗位列表（分页，GET 使用 params 传参） */
  getPositions = (params?: object) => {
    return http.request<ResultTable>("get", "/api/v1/position", { params }, { silentMessage: true });
  };

  /** 获取岗位选项（id,name） */
  getPositionOptions = (params?: object) => {
    return http.request<Result>("get", "/api/v1/position-options", { params }, { silentMessage: true });
  };
}

export const positionAPI = new PositionAPI();
