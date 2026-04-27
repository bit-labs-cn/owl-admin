import { type Ref, reactive, ref, onMounted, watch, computed, toRaw } from "vue";
import type { PaginationProps } from "@pureadmin/table";
import { ElMessage, ElMessageBox } from "element-plus";
import { roleAPI } from "@bit-labs.cn/owl-admin-ui/api/role";
import { menuApi } from "@bit-labs.cn/owl-admin-ui/api/menu";

import { transformI18n } from "@bit-labs.cn/owl-ui/plugins/i18n";
import { getKeyList } from "@pureadmin/utils";

export function useRoleList(menuTreeRef: Ref) {
  const form = reactive({ name: "", code: "", status: "" });
  const curRow = ref();
  const dataList = ref([]);
  const treeIds = ref([]);
  const treeData = ref([]);
  const isShow = ref(false);
  const loading = ref(true);
  const treeSearchValue = ref();
  const switchLoadMap = ref({});
  const isExpandAll = ref(false);
  const isSelectAll = ref(false);

  const treeProps = { value: "id", label: "meta.title", children: "children" };
  const pagination = reactive<PaginationProps>({
    total: 0,
    pageSize: 10,
    currentPage: 1,
    background: true
  });

  const buttonClass = computed(() => [
    "!h-[20px]",
    "reset-margin",
    "!text-gray-500",
    "dark:!text-white",
    "dark:hover:!text-primary"
  ]);

  function onChange({ row, index }) {
    ElMessageBox.confirm(
      `确认要<strong>${row.status === 0 ? "停用" : "启用"}</strong><strong style='color:var(--el-color-primary)'>${row.name}</strong>吗?`,
      "系统提示",
      {
        confirmButtonText: "确定",
        cancelButtonText: "取消",
        type: "warning",
        dangerouslyUseHTMLString: true,
        draggable: true
      }
    )
      .then(() => {
        switchLoadMap.value[index] = Object.assign(
          {},
          switchLoadMap.value[index],
          { loading: true }
        );
        setTimeout(() => {
          switchLoadMap.value[index] = Object.assign(
            {},
            switchLoadMap.value[index],
            { loading: false }
          );
          roleAPI.changeStatus(row).then(() => {});
        }, 300);
      })
      .catch(() => {
        row.status === 0 ? (row.status = 1) : (row.status = 0);
      });
  }

  function handleDelete(row) {
    roleAPI.deleteRole(row.id).then(() => {
      onSearch();
    });
  }

  function handleSizeChange(val: number) {
    pagination.pageSize = val;
    pagination.currentPage = 1;
    onSearch();
  }

  function handleCurrentChange(val: number) {
    pagination.currentPage = val;
    onSearch();
  }

  function handleSelectionChange(val) {
    console.log("handleSelectionChange", val);
  }

  async function onSearch() {
    loading.value = true;
    const res = await roleAPI.getRoles({
      ...toRaw(form),
      page: pagination.currentPage,
      pageSize: pagination.pageSize
    });
    // 后端返回 { data: { list }, total, currentPage, pageSize }，total 在顶层
    dataList.value = res?.data?.list ?? [];
    pagination.total = res?.total ?? 0;
    pagination.pageSize = res?.pageSize ?? pagination.pageSize;
    pagination.currentPage = res?.currentPage ?? pagination.currentPage;
    setTimeout(() => {
      loading.value = false;
    }, 500);
  }

  function resetForm(formEl) {
    if (!formEl) return;
    formEl.resetFields();
    onSearch();
  }

  async function handleMenu(row?: any) {
    const { id } = row ?? {};
    if (id) {
      curRow.value = row;
      isShow.value = true;
      const menuRes = await roleAPI.getRoleMenuIds(id);
      menuTreeRef.value?.setCheckedKeys(menuRes?.data ?? []);
    } else {
      curRow.value = null;
      isShow.value = false;
    }
  }

  function rowStyle({ row: { id } }) {
    return {
      cursor: "pointer",
      background: id === curRow.value?.id ? "var(--el-fill-color-light)" : ""
    };
  }

  async function handleSave() {
    const { id } = curRow.value;
    if (!id) return;
    await roleAPI.assignMenusToRole({
      id,
      menuIds: menuTreeRef.value.getCheckedKeys()
    });
    ElMessage.success("已保存");
  }

  const onQueryChanged = (query: string) => {
    menuTreeRef.value!.filter(query);
  };

  const filterMethod = (query: string, node) => {
    return transformI18n(node.title)!.includes(query);
  };

  onMounted(async () => {
    onSearch();
    const { data: menuData } = await menuApi.getAssignableMenus();
    treeIds.value = getKeyList(menuData, "id");
    treeData.value = menuData;
  });

  watch(isExpandAll, val => {
    val
      ? menuTreeRef.value.setExpandedKeys(treeIds.value)
      : menuTreeRef.value.setExpandedKeys([]);
  });

  watch(isSelectAll, val => {
    val
      ? menuTreeRef.value.setCheckedKeys(treeIds.value)
      : menuTreeRef.value.setCheckedKeys([]);
  });

  return {
    form,
    isShow,
    curRow,
    loading,
    rowStyle,
    dataList,
    treeData,
    treeProps,
    pagination,
    isExpandAll,
    isSelectAll,
    treeSearchValue,
    switchLoadMap,
    buttonClass,
    onSearch,
    resetForm,
    onChange,
    handleMenu,
    handleSave,
    handleDelete,
    filterMethod,
    transformI18n,
    onQueryChanged,
    handleSizeChange,
    handleCurrentChange,
    handleSelectionChange
  };
}
