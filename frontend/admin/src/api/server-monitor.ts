import { http } from "@bit-labs.cn/owl-ui/utils/http";
import type { ServerSnapshot } from "../views/tools/server/types";

class ServerMonitorAPI {
  snapshot = () => {
    return http.request<{ success: boolean; data: ServerSnapshot }>(
      "get",
      "/api/v1/tools/server",
      {},
      { silentMessage: true }
    );
  };
}

export const serverMonitorAPI = new ServerMonitorAPI();
