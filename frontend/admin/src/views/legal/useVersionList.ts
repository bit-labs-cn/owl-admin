import { reactive, ref, onMounted, toRaw, watch } from "vue";
import type { PaginationProps } from "@pureadmin/table";
import { legalAPI } from "@bit-labs.cn/owl-admin-ui/api/legal";
import type { LegalOption, LegalVersionRow } from "./types";

export function useVersionList() {
  const documents = ref<LegalOption[]>([]);
  const form = reactive({
    documentId: ""
  });
  const dataList = ref<LegalVersionRow[]>([]);
  const loading = ref(false);
  const pagination = reactive<PaginationProps>({
    total: 0,
    pageSize: 10,
    currentPage: 1,
    background: true
  });

  async function loadDocuments() {
    const { data } = await legalAPI.getDocumentOptions();
    documents.value = Array.isArray(data) ? data : [];
    if (!form.documentId && documents.value[0]?.id) {
      form.documentId = documents.value[0].id;
    }
  }

  async function onSearch() {
    if (!form.documentId) {
      dataList.value = [];
      pagination.total = 0;
      return;
    }
    loading.value = true;
    const payload: Record<string, unknown> = { ...toRaw(form) };
    payload.page = pagination.currentPage;
    payload.pageSize = pagination.pageSize;
    try {
      const { data } = await legalAPI.getVersions(payload);
      dataList.value = data.list ?? [];
      pagination.total = data.total ?? 0;
      pagination.pageSize = data.pageSize ?? pagination.pageSize;
      pagination.currentPage = data.currentPage ?? pagination.currentPage;
    } finally {
      loading.value = false;
    }
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

  watch(
    () => form.documentId,
    () => {
      pagination.currentPage = 1;
      onSearch();
    }
  );

  onMounted(async () => {
    await loadDocuments();
  });

  return {
    documents,
    form,
    loading,
    dataList,
    pagination,
    loadDocuments,
    onSearch,
    handleSizeChange,
    handleCurrentChange
  };
}
