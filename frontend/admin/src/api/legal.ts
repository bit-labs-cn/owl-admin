import { http } from "@bit-labs.cn/owl-ui/utils/http";

class LegalAPI {
  getDocuments = (params?: object) => {
    return http.request<ResultTable>("get", "/api/v1/legal/documents", { params }, { silentMessage: true });
  };
  getDocumentOptions = () => {
    return http.request<Result>("get", "/api/v1/legal/document-options", {}, { silentMessage: true });
  };
  createDocument = (data?: object) => {
    return http.request<Result>("post", "/api/v1/legal/documents", { data });
  };
  updateDocument = (data?: object) => {
    return http.request<Result>("put", `/api/v1/legal/documents/${data["id"]}`, { data });
  };
  deleteDocument = (id: string) => {
    return http.request<Result>("delete", `/api/v1/legal/documents/${id}`);
  };

  getVersions = (params?: object) => {
    return http.request<ResultTable>("get", "/api/v1/legal/versions", { params }, { silentMessage: true });
  };
  createVersion = (data?: object) => {
    return http.request<Result>("post", "/api/v1/legal/versions", { data });
  };
  updateVersion = (data?: object) => {
    return http.request<Result>("put", `/api/v1/legal/versions/${data["id"]}`, { data });
  };
  deleteVersion = (id: string) => {
    return http.request<Result>("delete", `/api/v1/legal/versions/${id}`);
  };
  publishVersion = (id: string, changeClass: string) => {
    return http.request<Result>("post", `/api/v1/legal/versions/${id}/publish`, {
      data: { changeClass }
    });
  };

  getScenes = (params?: object) => {
    return http.request<ResultTable>("get", "/api/v1/legal/scenes", { params }, { silentMessage: true });
  };
  getSceneOptions = () => {
    return http.request<Result>("get", "/api/v1/legal/scene-options", {}, { silentMessage: true });
  };
  createScene = (data?: object) => {
    return http.request<Result>("post", "/api/v1/legal/scenes", { data });
  };
  updateScene = (data?: object) => {
    return http.request<Result>("put", `/api/v1/legal/scenes/${data["id"]}`, { data });
  };
  deleteScene = (id: string) => {
    return http.request<Result>("delete", `/api/v1/legal/scenes/${id}`);
  };

  getEvents = (params?: object) => {
    return http.request<ResultTable>("get", "/api/v1/legal/events", { params }, { silentMessage: true });
  };
}

export const legalAPI = new LegalAPI();
