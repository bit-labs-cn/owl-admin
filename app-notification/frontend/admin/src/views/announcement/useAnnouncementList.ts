import { onMounted, reactive, ref } from "vue";
import type { PaginationProps } from "@pureadmin/table";
import { announcementAPI } from "../../api/notify";
import { applyPage } from "../../format";
import type { AnnouncementRow } from "./types";

export function useAnnouncementList() {
  const form = reactive<{ title: string; status?: number }>({ title: "", status: undefined });
  const dataList = ref<AnnouncementRow[]>([]);
  const loading = ref(true);
  const pagination = reactive<PaginationProps>({ total: 0, pageSize: 10, currentPage: 1, background: true });

  async function onSearch() {
    loading.value = true;
    try {
      const res = await announcementAPI.getList({
        page: pagination.currentPage,
        pageSize: pagination.pageSize,
        title: form.title || undefined,
        status: form.status || undefined
      });
      dataList.value = (res?.data?.list ?? []) as AnnouncementRow[];
      applyPage(pagination, res);
    } finally {
      loading.value = false;
    }
  }

  function resetForm(formEl?: any) {
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
