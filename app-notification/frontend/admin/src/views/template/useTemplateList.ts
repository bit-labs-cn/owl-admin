import { onMounted, reactive, ref } from "vue";
import type { PaginationProps } from "@pureadmin/table";
import { templateAPI, typeAPI } from "../../api/notify";
import { applyPage } from "../../format";
import type { TemplateRow, TypeOption } from "./types";

export function useTemplateList() {
  const form = reactive<{ typeCode: string }>({ typeCode: "" });
  const dataList = ref<TemplateRow[]>([]);
  const typeOptions = ref<TypeOption[]>([]);
  const loading = ref(true);
  const pagination = reactive<PaginationProps>({ total: 0, pageSize: 10, currentPage: 1, background: true });

  async function loadTypes() {
    const res = await typeAPI.options();
    typeOptions.value = (res?.data ?? []) as TypeOption[];
  }

  async function onSearch() {
    loading.value = true;
    try {
      const res = await templateAPI.getList({
        page: pagination.currentPage,
        pageSize: pagination.pageSize,
        typeCode: form.typeCode || undefined
      });
      dataList.value = (res?.data?.list ?? []) as TemplateRow[];
      applyPage(pagination, res);
    } finally {
      loading.value = false;
    }
  }

  function resetForm(formEl?: { resetFields: () => void }) {
    formEl?.resetFields();
    form.typeCode = "";
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
    loadTypes();
    onSearch();
  });

  return {
    form,
    dataList,
    typeOptions,
    loading,
    pagination,
    onSearch,
    resetForm,
    handleSizeChange,
    handleCurrentChange
  };
}
