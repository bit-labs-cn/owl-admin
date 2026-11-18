import { http } from "@bit-labs.cn/owl-ui/utils/http";

class FileAPI {
  /** 单文件上传；FormData 必须放在 Axios 的 data 字段，不能直接传给 http.post */
  upload = (data: FormData) => {
    return http.request<Result>(
      "post",
      "/api/v1/files/upload",
      { data },
      {
        headers: { "Content-Type": "multipart/form-data" },
        silentMessage: true
      }
    );
  };
}

export const fileAPI = new FileAPI();

type UploadPayload = { url?: string; originalName?: string };

function pickUploadUrl(res: unknown): string {
  const r = res as { data?: UploadPayload } & UploadPayload;
  const url = r?.data?.url ?? r?.url;
  if (!url || typeof url !== "string") {
    throw new Error("上传未返回文件地址");
  }
  return url;
}

/** 上传一张图片，返回可以直接嵌进富文本的绝对地址。 */
export async function uploadNotifyImage(file: File): Promise<string> {
  const fd = new FormData();
  fd.append("file", file, file.name);
  const url = pickUploadUrl(await fileAPI.upload(fd));
  // 存储返回的常是 /storage/... 相对地址，邮件客户端不认，这里补成绝对地址
  return new URL(url, window.location.origin).toString();
}
