export interface Dict {
  id: string;
  name: string;
  type: string;
  status: string;
  desc: string;
  sort: string;
}

export interface DictItem {
  id: string;
  label: string;
  value: string;
  extend: string;
  status: string;
  sort: string;
  dictID: string;
}

export interface DictRepositoryInterface {
  /**
   * 新增字典
   * @param dict
   * @returns boolean 新增是否成功
   */
  createDict(dict: Dict): Promise<Result>;

  /**
   * 更新字典
   * @param dict
   * @returns boolean 更新是否成功
   */
  updateDict(dict: Dict): Promise<Result>;

  /**
   * 删除字典
   * @param id
   * @returns boolean 删除是否成功
   */
  deleteDict(id: number): Promise<Result>;

  /**
   * 获取字典列表
   * @param params 可选，name/type/status 等查询参数
   * @returns Dict[]
   */
  getDicts(params?: { name?: string; type?: string; status?: string }): Promise<Dict[]>;

  /**
   * 新增字典项
   * @param dictID
   * @param dictItem
   * @returns boolean 新增是否成功
   */
  createDictItem(dictID: string, dictItem: DictItem): Promise<Result>;

  /**
   * 更新字典项
   * @param dictID
   * @param dictItem
   * @returns boolean 更新是否成功
   */
  updateDictItem(dictID: string, dictItem: DictItem): Promise<Result>;

  /**
   * 删除字典项
   * @param dictID
   * @param id
   * @returns Result 删除是否成功
   */
  deleteDictItem(dictID: string, id: string): Promise<Result>;

  /**
   * 获取字典项列表
   * @returns DictItem[] 字典项列表
   * @param dictID
   */
  getDictItems(dictID: string): Promise<DictItem[]>;
}
