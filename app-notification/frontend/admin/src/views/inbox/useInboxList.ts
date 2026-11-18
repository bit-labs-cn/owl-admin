import { onMounted, reactive, ref } from "vue";
import type { PaginationProps } from "@pureadmin/table";
import { inboxAPI, typeAPI } from "../../api/notify";
import { applyPage } from "../../format";
import type { InboxCounts, InboxRow } from "./types";

export function useInboxList() {
  const form = reactive<{ title: string; readStatus?: number; typeCode: string }>({
    title: "",
    readStatus: undefined,
    typeCode: ""
  });
  const typeOptions = ref<{ code: string; name: string }[]>([]);
  const dataList = ref<InboxRow[]>([]);
  const loading = ref(true);
  const counts = ref<InboxCounts>({
    unreadNotice: 0,
    pendingTodo: 0,
    total: 0,
    notice: 0,
    todo: 0,
    message: 0
  });
  const pagination = reactive<PaginationProps>({ total: 0, pageSize: 10, currentPage: 1, background: true });

  async function loadCounts() {
    const res = await inboxAPI.counts();
    counts.value = (res?.data ?? counts.value) as InboxCounts;
  }

  async function onSearch() {
    loading.value = true;
    try {
      const res = await inboxAPI.getList({
        page: pagination.currentPage,
        pageSize: pagination.pageSize,
        title: form.title || undefined,
        // 未读取值是 0，不能用 || 判断，否则会被当成「未选」
        readStatus: form.readStatus ?? undefined,
        typeCode: form.typeCode || undefined
      });
      dataList.value = (res?.data?.list ?? []) as InboxRow[];
      applyPage(pagination, res);
      await loadCounts();
    } finally {
      loading.value = false;
    }
  }

  function resetForm(formEl?: any) {
    formEl?.resetFields();
    form.readStatus = undefined;
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

  onMounted(async () => {
    try {
      const res = await typeAPI.options();
      typeOptions.value = (res?.data ?? []) as { code: string; name: string }[];
    } catch {
      typeOptions.value = [];
    }
    onSearch();
  });

  return {
    form,
    typeOptions,
    dataList,
    loading,
    counts,
    pagination,
    onSearch,
    resetForm,
    handleSizeChange,
    handleCurrentChange
  };
}
