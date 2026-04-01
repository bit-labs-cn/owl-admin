import type {
  Dict,
  DictItem,
  DictRepositoryInterface
} from "@bit-labs.cn/owl-admin-ui/views/dict/dict_repository";
import { http } from "@bit-labs.cn/owl-ui/utils/http";

class DictAPI implements DictRepositoryInterface {
  /**
   * 新增字典
   * @param data
   * @returns Result
   */
  createDict(data?: Dict): Promise<Result> {
    return http.request<Result>("post", "/api/v1/dict", { data });
  }
  /**
   * 删除字典
   * @param id
   * @returns Result
   */
  deleteDict(id: number): Promise<Result> {
    return http.request<Result>("delete", `/api/v1/dict/${id}`);
  }
  /**
   * 更新字典
   * @param data
   * @returns Result
   */
  updateDict(data?: Dict): Promise<Result> {
    return http.request<Result>("put", `/api/v1/dict/${data.id}`, { data });
  }

  /**
   * 获取字典列表（支持 name/type/status 查询）
   * @param params 可选查询参数
   * @returns Dict[]
   */
  getDicts(params?: { name?: string; type?: string; status?: string }): Promise<Dict[]> {
    return new Promise((resolve, reject) => {
      http
        .request<Result>("get", "/api/v1/dict", { params: params ?? {} }, { silentMessage: true })
        .then(res => {
          if (res.success) {
            resolve(res.data);
          } else {
            reject(res.msg);
          }
        });
    });
  }
  /**
   * 新增字典项
   * @param dictID
   * @param data
   */
  createDictItem(dictID: string, data?: DictItem): Promise<Result> {
    return http.request<Result>("post", `/api/v1/dict/${dictID}/item`, {
      data
    });
  }

  /**
   * 删除字典项
   * @param dictID
   * @param id
   * @returns Result
   */
  deleteDictItem(dictID: string, id: string): Promise<Result> {
    return http.request<Result>("delete", `/api/v1/dict/${dictID}/item/${id}`);
  }

  /**
   * 获取字典项列表
   * @param dictID
   */
  getDictItems(dictID: string): Promise<DictItem[]> {
    return new Promise((resolve, reject) => {
      http
        .request<Result>("get", `/api/v1/dict/${dictID}/item`, {}, { silentMessage: true })
        .then(res => {
          if (res.success) {
            resolve(res.data);
          } else {
            reject(res.msg);
          }
        });
    });
  }

  /**
   * 更新字典项
   * @param dictID
   * @param data
   */
  updateDictItem(dictID: string, data?: DictItem): Promise<Result> {
    return http.request<Result>(
      "put",
      `/api/v1/dict/${dictID}/item/${data.id}`,
      { data }
    );
  }
}

export { DictAPI };
