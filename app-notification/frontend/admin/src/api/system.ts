import { http } from "@bit-labs.cn/owl-ui/utils/http";

/**
 * 系统用户接口。接收人选择弹窗直接复用系统用户列表，不另开通知专用接口。
 * 支持的查询条件见 `app/service/user_service.go` 的 `RetrieveUserReq`。
 */
class SystemUserAPI {
  getList = (params?: object) => {
    return http.request<ResultTable>("get", "/api/v1/users", { params }, { silentMessage: true });
  };
}

export const systemUserAPI = new SystemUserAPI();
