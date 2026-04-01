import { reactive, ref, onMounted, toRaw } from "vue";
import type { PaginationProps } from "@pureadmin/table";
import { positionAPI } from "@bit-labs.cn/owl-admin-ui/api/position";


export function usePositionList() {
  const form = reactive({
    name: "",
    status: ""
  });

  const dataList = ref([]);
  const loading = ref(true);
  const switchLoadMap = ref<Record<number, { loading: boolean }>>({});
  const pagination = reactive<PaginationProps>({
    total: 0,
    pageSize: 10,
    currentPage: 1,
    background: true
  });

  function onChange({ row, index }) {
    switchLoadMap.value[index] = Object.assign({}, switchLoadMap.value[index], {
      loading: true
    });
    setTimeout(() => {
      switchLoadMap.value[index] = Object.assign(
        {},
        switchLoadMap.value[index],
        { loading: false }
      );
      positionAPI.changeStatus(row.id, row.status).then(() => {
        onSearch();
      });
    }, 300);
  }

  async function onSearch() {
    loading.value = true;
    const payload: any = toRaw(form);
    payload.page = pagination.currentPage;
    payload.pageSize = pagination.pageSize;
    const { data } = await positionAPI.getPositions(payload);
    dataList.value = data.list ?? [];
    pagination.total = data.total ?? 0;
    pagination.pageSize = data.pageSize ?? pagination.pageSize;
    pagination.currentPage = data.currentPage ?? pagination.currentPage;
    setTimeout(() => (loading.value = false), 300);
  }

  function resetForm(formEl) {
    if (!formEl) return;
    formEl.resetFields();
    onSearch();
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

  onMounted(() => {
    onSearch();
  });

  return {
    form,
    loading,
    dataList,
    pagination,
    switchLoadMap,
    onSearch,
    resetForm,
    onChange,
    handleSizeChange,
    handleCurrentChange
  };
}
