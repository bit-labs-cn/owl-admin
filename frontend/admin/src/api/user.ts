import { http } from "@bit-labs.cn/owl-ui/utils/http";

export interface ImportUserError {
  row: number;
  username: string;
  reason: string;
}

export interface ImportUserResult {
  total: number;
  created: number;
  failed: number;
  errors?: ImportUserError[];
}

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
  /** 下载用户导入模板 */
  importTemplate = () => {
    return http.request(
      "get",
      "/api/v1/users/import-template",
      { responseType: "blob" },
      { silentMessage: true }
    ) as Promise<Blob>;
  };
  /** 批量导入用户 */
  importUsers = (file: File) => {
    const data = new FormData();
    data.append("file", file);
    return http.request<Result & { data?: ImportUserResult }>(
      "post",
      "/api/v1/users/import",
      {
        data,
        timeout: 180000,
        headers: { "Content-Type": "multipart/form-data" }
      }
    );
  };
  /** 导出用户导入失败明细 */
  exportImportErrors = (errors: ImportUserError[]) => {
    return http.request(
      "post",
      "/api/v1/users/import-errors/export",
      {
        data: { errors },
        responseType: "blob"
      },
      { silentMessage: true }
    ) as Promise<Blob>;
  };
  /** 批量删除用户 */
  batchDeleteUsers = (ids: Array<string | number>) => {
    return http.request<Result>("post", "/api/v1/users/batch-delete", {
      data: { ids: ids.map(String) }
    });
  };
  /** 批量修改用户状态 */
  batchChangeStatus = (ids: Array<string | number>, status: number) => {
    return http.request<Result>("put", "/api/v1/users/batch-status", {
      data: { ids: ids.map(String), status }
    });
  };
  /** 批量分配用户角色（覆盖） */
  batchAssignRoles = (ids: Array<string | number>, roleIDs: Array<string | number>) => {
    return http.request<Result>("put", "/api/v1/users/batch-roles", {
      data: { ids: ids.map(String), roleIDs: roleIDs.map(String) }
    });
  };
  /** 批量设置用户部门 */
  batchAssignDept = (ids: Array<string | number>, parentId?: string | number | "") => {
    return http.request<Result>("put", "/api/v1/users/batch-dept", {
      data: {
        ids: ids.map(String),
        parentId: parentId === undefined || parentId === null ? "" : String(parentId)
      }
    });
  };
}

export const userManageAPI = new UserManageAPI();
