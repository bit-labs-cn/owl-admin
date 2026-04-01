import { reactive, ref, onMounted } from "vue";
import { deptAPI } from "@bit-labs.cn/owl-admin-ui/api/dept";
import { isAllEmpty } from "@pureadmin/utils";
import { handleTree } from "@bit-labs.cn/owl-ui/utils/tree";

export function useDeptList() {
  const form = reactive({ name: "", status: null });
  const dataList = ref([]);
  const loading = ref(true);

  async function onSearch() {
    loading.value = true;
    const { data } = await deptAPI.getDepts();
    let newData = data;
    if (!isAllEmpty(form.name)) {
      newData = newData.filter(item => item.name.includes(form.name));
    }
    if (!isAllEmpty(form.status)) {
      newData = newData.filter(item => item.status === form.status);
    }
    dataList.value = handleTree(newData);
    setTimeout(() => {
      loading.value = false;
    }, 500);
  }

  function resetForm(formEl) {
    if (!formEl) return;
    formEl.resetFields();
    onSearch();
  }

  function formatHigherDeptOptions(treeList) {
    if (!treeList || !treeList.length) return;
    const newTreeList = [];
    for (let i = 0; i < treeList.length; i++) {
      treeList[i].disabled = treeList[i].status === 0;
      formatHigherDeptOptions(treeList[i].children);
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
    formatHigherDeptOptions,
    handleSelectionChange
  };
}
