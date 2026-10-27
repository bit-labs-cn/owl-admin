import { reactive, ref, onMounted } from "vue";
import { getApis } from "@bit-labs.cn/owl-admin-ui/api/system";
import { isAllEmpty } from "@pureadmin/utils";

export function useApiList() {
  const form = reactive({
    keyword: "",
    method: "",
    operateLog: ""
  });
  const dataList = ref([]);
  const loading = ref(true);

  function resetForm(formEl) {
    if (!formEl) return;
    formEl.resetFields();
    onSearch();
  }

  async function onSearch() {
    loading.value = true;
    const { data } = await getApis();
    let newData = data;
    if (!isAllEmpty(form.keyword)) {
      newData = newData.filter(
        item =>
          item.name?.includes(form.keyword) ||
          item.path?.includes(form.keyword)
      );
    }
    if (!isAllEmpty(form.method)) {
      newData = newData.filter(item => item.method === form.method);
    }
    if (!isAllEmpty(form.operateLog)) {
      const record = form.operateLog === "1";
      newData = newData.filter(item => Boolean(item.operateLog) === record);
    }
    dataList.value = newData;
    setTimeout(() => {
      loading.value = false;
    }, 500);
  }

  function handleSelectionChange(val) {
    console.log("handleSelectionChange", val);
  }

  onMounted(() => {
    onSearch();
  });

  return { form, loading, dataList, onSearch, resetForm, handleSelectionChange };
}
