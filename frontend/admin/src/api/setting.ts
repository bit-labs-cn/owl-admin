import { http } from "@bit-labs.cn/owl-ui/utils/http";
import type { SettingView } from "../views/setting/types";

class SettingAPI {
  get = () => {
    return http.request<{ success: boolean; data: SettingView }>(
      "get",
      "/api/v1/setting",
      {},
      { silentMessage: true }
    );
  };
  update = (data: { subsystem: string; values: { key: string; value: string }[] }) => {
    return http.request<{ success: boolean }>("put", "/api/v1/setting", { data });
  };
}

export const settingAPI = new SettingAPI();
