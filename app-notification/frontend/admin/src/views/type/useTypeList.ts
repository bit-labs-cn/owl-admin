import { onMounted, reactive, ref } from "vue";
import type { PaginationProps } from "@pureadmin/table";
import { typeAPI } from "../../api/notify";
import { applyPage } from "../../format";
import type { TypeRow } from "./types";

export function useTypeList() {
  const form = reactive<{ name: string; status?: number }>({ name: "", status: undefined });
  const dataList = ref<TypeRow[]>([]);
  const loading = ref(true);
  const pagination = reactive<PaginationProps>({ total: 0, pageSize: 10, currentPage: 1, background: true });

  async function onSearch() {
    loading.value = true;
    try {
      const res = await typeAPI.getList({
        page: pagination.currentPage,
        pageSize: pagination.pageSize,
        name: form.name || undefined,
        status: form.status || undefined
      });
      dataList.value = (res?.data?.list ?? []) as TypeRow[];
      applyPage(pagination, res);
    } finally {
      loading.value = false;
    }
  }

  function resetForm(formEl?: { resetFields: () => void }) {
    formEl?.resetFields();
    form.status = undefined;
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

  onMounted(() => onSearch());

  return { form, dataList, loading, pagination, onSearch, resetForm, handleSizeChange, handleCurrentChange };
}
