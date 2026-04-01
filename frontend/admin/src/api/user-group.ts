import { http } from "@bit-labs.cn/owl-ui/utils/http";

class UserGroupAPI {
  /** 获取用户组列表 */
  getList = (params?: object) => {
    return http.request<ResultTable>("get", "/api/v1/user-groups", { params }, { silentMessage: true });
  };
  /** 创建用户组 */
  create = (data?: object) => {
    return http.request<Result>("post", "/api/v1/user-groups", { data });
  };
  /** 更新用户组 */
  update = (data?: object) => {
    return http.request<Result>("put", `/api/v1/user-groups/${data["id"]}`, { data });
  };
  /** 删除用户组 */
  remove = (id: string) => {
    return http.request<Result>("delete", `/api/v1/user-groups/${id}`);
  };
  /** 修改用户组状态 */
  changeStatus = (data?: object) => {
    return http.request<Result>("put", `/api/v1/user-groups/${data["id"]}/status`, { data });
  };
  /** 获取所有用户组选项（id+name） */
  getOptions = () => {
    return http.request<Result>("get", "/api/v1/user-groups-options", {}, { silentMessage: true });
  };
  /** 获取用户组下的用户列表（分页） */
  getUsersByGroupId = (id: string, params?: object) => {
    return http.request<ResultTable>("get", `/api/v1/user-groups/${id}/users`, { params }, { silentMessage: true });
  };
  /** 获取用户组关联的用户ID列表（轻量，用于表单回显） */
  getUserIdsByGroupId = (id: string) => {
    return http.request<Result>("get", `/api/v1/user-groups/${id}/user-ids`, {}, { silentMessage: true });
  };
  /** 为用户组分配用户 */
  assignUsersToGroup = (id: string, data?: object) => {
    return http.request<Result>("put", `/api/v1/user-groups/${id}/users`, { data });
  };
}

export const userGroupAPI = new UserGroupAPI();
