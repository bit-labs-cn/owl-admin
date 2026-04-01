import { http } from "@bit-labs.cn/owl-ui/utils/http";

class RoleAPI {
  createRole = (data?: object) => {
    return http.request<Result>("post", "/api/v1/roles", { data });
  };

  deleteRole = (id: string) => {
    return http.request<Result>("delete", `/api/v1/roles/${id}`);
  };

  updateRole = (data?: object) => {
    return http.request<Result>("put", `/api/v1/roles/${data["id"]}`, {
      data
    });
  };
  /**根据角色id获取角色菜单id列表*/
  getRoleMenuIds = (id?: object) => {
    return http.request<Result>("get", `/api/v1/roles/${id}/menu-ids`, {}, { silentMessage: true });
  };

  changeStatus = (data?: object) => {
    return http.request<Result>("put", `/api/v1/roles/${data["id"]}/status`, {
      data
    });
  };

  /** 获取系统管理-角色管理列表 */
  getRoles = (params?: object) => {
    return http.request<ResultTable>("get", "/api/v1/roles", { params }, { silentMessage: true });
  };

  /** 根据角色id获取该角色下的用户列表（分页） */
  getUsersByRoleId = (id: string, params?: object) => {
    return http.request<ResultTable>(
      "get",
      `/api/v1/roles/${id}/users`,
      { params },
      { silentMessage: true }
    );
  };

  /** 系统管理-用户管理-获取所有角色列表 */
  getRolesOption = (data?: object) => {
    return http.request<Result>("get", "/api/v1/roles-options", { data }, { silentMessage: true });
  };

  /** 分配菜单给角色 */
  assignMenusToRole = (data?: object) => {
    return http.request<Result>("put", `/api/v1/roles/${data["id"]}/menus`, {
      data
    });
  };
}

export const roleAPI = new RoleAPI();
