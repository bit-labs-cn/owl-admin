import { reactive, ref, onMounted, toRaw } from "vue";
import type { PaginationProps } from "@pureadmin/table";
import { legalAPI } from "@bit-labs.cn/owl-admin-ui/api/legal";
import type { LegalSceneRow } from "./types";

export function useSceneList() {
  const form = reactive({
    name: "",
    kind: ""
  });
  const dataList = ref<LegalSceneRow[]>([]);
  const loading = ref(true);
  const pagination = reactive<PaginationProps>({
    total: 0,
    pageSize: 10,
    currentPage: 1,
    background: true
  });

  async function onSearch() {
    loading.value = true;
    const payload: Record<string, unknown> = { ...toRaw(form) };
    payload.page = pagination.currentPage;
    payload.pageSize = pagination.pageSize;
    try {
      const { data } = await legalAPI.getScenes(payload);
      dataList.value = data.list ?? [];
      pagination.total = data.total ?? 0;
      pagination.pageSize = data.pageSize ?? pagination.pageSize;
      pagination.currentPage = data.currentPage ?? pagination.currentPage;
    } finally {
      loading.value = false;
    }
  }

  function resetForm(formEl) {
    if (!formEl) return;
    formEl.resetFields();
    pagination.currentPage = 1;
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

  return { form, loading, dataList, pagination, onSearch, resetForm, handleSizeChange, handleCurrentChange };
}
