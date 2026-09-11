import { http } from "@bit-labs.cn/owl-ui/utils/http";
import { formatToken, getToken } from "@bit-labs.cn/owl-ui/utils/auth";

class AppVersionAPI {
  uploadPackage = async (file: File, apkType: number) => {
    const fd = new FormData();
    fd.append("file", file);
    fd.append("apkType", String(apkType));

    const headers: HeadersInit = {};
    try {
      const access = String(getToken()?.accessToken ?? "").trim();
      if (access) headers.Authorization = formatToken(access);
    } catch {
      /* 未登录时由接口返回鉴权错误 */
    }

    const resp = await fetch("/api/v1/app-versions/upload", {
      method: "POST",
      headers,
      body: fd
    });
    const res = await resp.json();
    const ok = res?.success === true || res?.code === 0 || res?.code === "0";
    if (!resp.ok || !ok) {
      throw new Error(res?.msg || "上传失败");
    }
    const data = (res?.data ?? {}) as { url?: string; originalName?: string };
    if (!data.url) throw new Error("上传接口未返回文件地址");
    return { url: data.url, originalName: data.originalName ?? file.name };
  };

  create = (data?: object) => {
    return http.request<Result>("post", "/api/v1/app-versions", { data });
  };

  update = (data?: object) => {
    return http.request<Result>("put", `/api/v1/app-versions/${data["id"]}`, {
      data
    });
  };

  remove = (id: string | number) => {
    return http.request<Result>("delete", `/api/v1/app-versions/${id}`);
  };

  changeStatus = (id: string | number, status: number) => {
    return http.request<Result>("put", `/api/v1/app-versions/${id}/status`, {
      data: { id, status }
    });
  };

  getList = (params?: object) => {
    return http.request<ResultTable>("get", "/api/v1/app-versions", { params }, { silentMessage: true });
  };
}

export const appVersionAPI = new AppVersionAPI();
