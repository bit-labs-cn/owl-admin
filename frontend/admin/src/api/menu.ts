import { http } from "@bit-labs.cn/owl-ui/utils/http";

class MenuAPI {
  /** 获取可用于分配的菜单，包含按钮 */
  getAssignableMenus = (): Promise<Result> => {
    return http.request<Result>("get", `/api/v1/menus/assignable`, {}, { silentMessage: true });
  };
  /** 获取所有菜单，不包含按钮*/
  getMenus = (): Promise<Result> => {
    return http.request<Result>("get", `/api/v1/menus`, {}, { silentMessage: true });
  };
}

export const menuApi = new MenuAPI();
