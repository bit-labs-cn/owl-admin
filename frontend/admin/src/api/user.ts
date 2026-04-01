import { http } from "@bit-labs.cn/owl-ui/utils/http";

class UserManageAPI {
  /** 获取用户管理列表（GET 使用 params 传参，便于后端 query 绑定与分页） */
  getUserList = (params?: object) => {
    return http.request<ResultTable>("get", "/api/v1/users", { params }, { silentMessage: true });
  };
  /** 创建用户 */
  createUser = (data?: object) => {
    return http.request<Result>("post", "/api/v1/users", { data });
  };
  /** 更新用户 */
  updateUser = (data?: object) => {
    return http.request<Result>("put", `/api/v1/users/${data["id"]}`, {
      data
    });
  };
  /** 删除用户 */
  deleteUser = (id: string) => {
    return http.request<Result>("delete", `/api/v1/users/${id}`);
  };
  /** 修改用户状态 */
  changeStatus = (id: number, status: number) => {
    return http.request<Result>("put", `/api/v1/users/${id}/status`, {
      data: { id, status }
    });
  };
  /** 根据 userId 获取对应角色 id 列表 */
  getRoleIds = (userID: number) => {
    return http.request<Result>("get", `/api/v1/users/${userID}/roles`, {}, { silentMessage: true });
  };
  /** 分配角色给用户 */
  assignRoleToUser = (data?: object) => {
    return http.request<Result>("post", `/api/v1/users/${data["id"]}/roles`, {
      data
    });
  };
  /** 重置用户密码 */
  resetPassword = (id: number, newPassword: string) => {
    return http.request<Result>("put", `/api/v1/users/${id}/reset`, {
      data: {
        userId: String(id),
        newPassword
      }
    });
  };
  /** 修改用户头像 */
  changeAvatar = (id: number | string, avatar: string) => {
    return http.request<Result>("put", `/api/v1/users/${id}/avatar`, {
      data: { avatar }
    });
  };
}

export const userManageAPI = new UserManageAPI();
