import { http } from "@bit-labs.cn/owl-ui/utils/http";
import type { GeneratePasswordResult, HashPasswordResult } from "../views/tools/password/types";

export interface GeneratePasswordPayload {
  length: number;
  count: number;
  upper: boolean;
  lower: boolean;
  digit: boolean;
  symbol: boolean;
  excludeAmbiguous: boolean;
}

class PasswordToolAPI {
  generate = (data: GeneratePasswordPayload) => {
    return http.request<{ success: boolean; data: GeneratePasswordResult }>(
      "post",
      "/api/v1/tools/passwords",
      { data }
    );
  };

  hash = (data: { password: string }) => {
    return http.request<{ success: boolean; data: HashPasswordResult }>(
      "post",
      "/api/v1/tools/password-hashes",
      { data }
    );
  };
}

export const passwordToolAPI = new PasswordToolAPI();
