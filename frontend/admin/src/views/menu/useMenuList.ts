import { reactive, ref, onMounted } from "vue";
import { menuApi } from "@bit-labs.cn/owl-admin-ui/api/menu";
import { transformI18n } from "@bit-labs.cn/owl-ui/plugins/i18n";
import { isAllEmpty } from "@pureadmin/utils";

export function useMenuList() {
  const form = reactive({ title: "" });
  const dataList = ref([]);
  const loading = ref(true);

  function resetForm(formEl) {
    if (!formEl) return;
    formEl.resetFields();
    onSearch();
  }

  async function onSearch() {
    loading.value = true;
    const { data } = await menuApi.getMenus();
    let newData = data;
    if (!isAllEmpty(form.title)) {
      newData = newData.filter(item =>
        transformI18n(item.title).includes(form.title)
      );
    }
    dataList.value = newData;
    setTimeout(() => {
      loading.value = false;
    }, 500);
  }

  function formatHigherMenuOptions(treeList) {
    if (!treeList || !treeList.length) return;
    const newTreeList = [];
    for (let i = 0; i < treeList.length; i++) {
      treeList[i].title = transformI18n(treeList[i].title);
      formatHigherMenuOptions(treeList[i].children);
      newTreeList.push(treeList[i]);
    }
    return newTreeList;
  }

  function handleSelectionChange(val) {
    console.log("handleSelectionChange", val);
  }

  onMounted(() => {
    onSearch();
  });

  return {
    form,
    loading,
    dataList,
    onSearch,
    resetForm,
    formatHigherMenuOptions,
    handleSelectionChange
  };
}
