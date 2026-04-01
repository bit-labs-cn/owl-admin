import { ref, type Ref } from "vue";
import type {
  Dict,
  DictItem,
  DictRepositoryInterface
} from "./dict_repository";
import SmartTable from "./smart_table";


export class DictTable extends SmartTable {
  private itemTable: DictItemTable;
  private readonly repo: DictRepositoryInterface;
  choseDict: Ref<Dict> = ref();
  columns: TableColumnList = [
    {
      label: "",
      width: 40,
      cellRenderer: () => this.sortable("#dictTable")
    },
    {
      label: "字典名（中）",
      prop: "name",
      cellRenderer: ({ row, index }) => this.editable(row, index, "name")
    },
    {
      label: "字典名（英）",
      prop: "type",
      cellRenderer: ({ row, index }) => this.editable(row, index, "type")
    },
    {
      label: "状态",
      prop: "status",
      cellRenderer: ({ row, index }) => this.changeStatus(row, index)
    },
    {
      label: "描述",
      prop: "desc",
      cellRenderer: ({ row, index }) => this.editable(row, index, "desc")
    },
    {
      label: "排序",
      prop: "sort",
      width: 60,
      cellRenderer: ({ row, index }) => this.editable(row, index, "sort")
    },
    {
      label: "操作",
      fixed: "right",
      slot: "operation"
    }
  ];

  constructor(repo: DictRepositoryInterface) {
    super();

    this.repo = repo;
    this.itemTable = new DictItemTable("0", repo);
    this.loadData();
  }
  /**
   * 加载字典项数据（当前选中字典的项列表）
   */
  private getDictItems() {
    this.loading.value = true;
    this.repo
      .getDictItems(this.choseDict.value.id)
      .then(res => {
        this.itemTable.dataList.value = res;
      })
      .catch(reason => {
        console.log("获取字典列表失败：", reason);
      })
      .finally(() => {
        this.loading.value = false;
      });
  }
  data() {
    return this.dataList.value;
  }

  itemData() {
    return this.itemTable.dataList.value;
  }

  itemIsLoading() {
    return this.itemTable.isLoading();
  }

  itemColumns() {
    return this.itemTable.columns;
  }

  createDict() {
    let newDict: Dict = {
      id: "0",
      name: "新字典",
      type: "new",
      desc: "新字典",
      status: "1",
      sort: "0"
    };
    const editDict = Object.assign({ ...newDict, editing: true });
    this.addRow(editDict);
  }

  deleteDict(index: number) {
    this.repo.deleteDict(this.dataList.value[index].id).then(() => {
      this.removeRow(index);
      this.getDictItems();
    });
  }

  createDictItem() {
    let newDictItem: DictItem = {
      sort: "0",
      id: "0",
      label: "",
      value: "",
      extend: "",
      status: "1",
      dictID: this.choseDict.value.id
    };
    const edit = Object.assign({ ...newDictItem, editing: true });
    this.itemTable.addRow(edit);
  }

  deleteDictItem(index: number) {
    const id = this.itemTable.dataList.value[index].id;
    this.repo.deleteDictItem(this.choseDict.value.id, id).then(() => {
      this.itemTable.removeRow(index);
    });
  }

  chooseDict = (row: any) => {
    this.choseDict.value = row;
    this.itemTable = new DictItemTable(row.id, this.repo);
  };

  rowClass = ({ row }) => {
    if (row === this.choseDict.value) {
      return "add-bg";
    }
    return "";
  };
  /**
   * 行内编辑提交保存
   */
  submitSave(dict: Dict): Promise<any> {
    if (dict.id === "0") {
      return this.repo.createDict(dict);
    } else {
      return this.repo.updateDict(dict);
    }
  }
  /**
   * 加载字典列表（支持搜索参数）
   * @param params 可选，name/type/status 等查询参数
   */
  loadData(params?: { name?: string; type?: string; status?: string }) {
    this.loading.value = true;
    this.repo
      .getDicts(params)
      .then(res => {
        this.dataList.value = res;
        this.choseDict.value = res[0];
        if (res.length > 0) {
          const dictID = this.choseDict.value.id;
          this.itemTable = new DictItemTable(dictID, this.repo);
        }
      })
      .catch(reason => {
        console.log("获取字典失败：", reason);
      })
      .finally(() => {
        this.loading.value = false;
      });
  }
}

class DictItemTable extends SmartTable {
  private repo: DictRepositoryInterface;
  private dictID: string;
  constructor(dictID: string, repo: DictRepositoryInterface) {
    super();
    this.repo = repo;
    this.loading.value = true;
    this.dictID = dictID;
    this.loadData();
  }

  columns: TableColumnList = [
    {
      label: "",
      width: 40,
      cellRenderer: () => this.sortable("#dictItemTable")
    },
    {
      label: "名称",
      prop: "label",
      cellRenderer: ({ row, index }) => this.editable(row, index, "label")
    },
    {
      label: "字典值",
      prop: "value",
      cellRenderer: ({ row, index }) => this.editable(row, index, "value")
    },
    {
      label: "扩展值",
      prop: "extend",
      cellRenderer: ({ row, index }) => this.editable(row, index, "extend")
    },
    {
      label: "启用状态",
      prop: "status",
      cellRenderer: ({ row, index }) => this.changeStatus(row, index)
    },
    {
      label: "排序",
      prop: "sort",
      width: 60,
      cellRenderer: ({ row, index }) => this.editable(row, index, "sort")
    },
    {
      label: "操作",
      fixed: "right",
      slot: "operation"
    }
  ];

  submitSave(dictItem: DictItem): Promise<Result> {
    if (dictItem.id === "0") {
      return this.repo.createDictItem(dictItem.dictID, dictItem);
    } else {
      return this.repo.updateDictItem(dictItem.dictID, dictItem);
    }
  }
  loadData() {
    this.repo
      .getDictItems(this.dictID)
      .then(res => {
        this.dataList.value = res;
      })
      .catch(reason => {
        console.log("重新加载字典项失败：", reason);
      })
      .finally(() => {
        this.loading.value = false;
      });
  }
}
