import { http } from "@bit-labs.cn/owl-ui/utils/http";
import { applyPublicSite } from "@bit-labs.cn/owl-ui/config/siteBrand";
import type { SettingView } from "./views/setting/types";

/** 读取公开站点配置并作用到侧栏、登录页和页脚。失败时保持现有外观。 */
export async function loadPublicSite() {
  try {
    const res = await http.request<{ success: boolean; data: SettingView }>(
      "get",
      "/api/v1/setting/public",
      {},
      { silentMessage: true, silentError: true }
    );
    applyPublicSite(res.data);
  } catch {
    // 公开配置读不到时沿用平台配置和内置标志。
  }
}
