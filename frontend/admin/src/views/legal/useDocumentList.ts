import { reactive, ref, onMounted, toRaw } from "vue";
import type { PaginationProps } from "@pureadmin/table";
import { legalAPI } from "@bit-labs.cn/owl-admin-ui/api/legal";
import type { LegalDocumentRow } from "./types";

export function useDocumentList() {
  const form = reactive({
    name: "",
    kind: "",
    status: ""
  });
  const dataList = ref<LegalDocumentRow[]>([]);
  const loading = ref(true);
  const switchLoadMap = ref<Record<number, { loading: boolean }>>({});
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
      const { data } = await legalAPI.getDocuments(payload);
      dataList.value = data.list ?? [];
      pagination.total = data.total ?? 0;
      pagination.pageSize = data.pageSize ?? pagination.pageSize;
      pagination.currentPage = data.currentPage ?? pagination.currentPage;
    } finally {
      loading.value = false;
    }
  }

  function onChange({ row, index }: { row: LegalDocumentRow; index: number }) {
    switchLoadMap.value[index] = { loading: true };
    legalAPI
      .updateDocument({
        id: row.id,
        name: row.name,
        kind: row.kind,
        status: row.status
      })
      .then(() => onSearch())
      .finally(() => {
        switchLoadMap.value[index] = { loading: false };
      });
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
