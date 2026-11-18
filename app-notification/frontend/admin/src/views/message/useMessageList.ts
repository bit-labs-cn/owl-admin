import { onMounted, reactive, ref } from "vue";
import type { PaginationProps } from "@pureadmin/table";
import { messageAPI, typeAPI } from "../../api/notify";
import { applyPage } from "../../format";
import type { MessageDetail, MessageRow } from "./types";

export function useMessageList() {
  const form = reactive<{ title: string; typeCode: string }>({
    title: "",
    typeCode: ""
  });
  const typeOptions = ref<{ code: string; name: string }[]>([]);
  const dataList = ref<MessageRow[]>([]);
  const loading = ref(true);
  const detail = ref<MessageDetail>();
  const pagination = reactive<PaginationProps>({ total: 0, pageSize: 10, currentPage: 1, background: true });

  async function loadTypes() {
    try {
      const res = await typeAPI.options();
      typeOptions.value = (res?.data ?? []) as { code: string; name: string }[];
    } catch {
      typeOptions.value = [];
    }
  }

  async function onSearch() {
    loading.value = true;
    try {
      const res = await messageAPI.getList({
        page: pagination.currentPage,
        pageSize: pagination.pageSize,
        title: form.title || undefined,
        typeCode: form.typeCode || undefined
      });
      dataList.value = (res?.data?.list ?? []) as MessageRow[];
      applyPage(pagination, res);
    } finally {
      loading.value = false;
    }
  }

  async function loadDetail(row: MessageRow) {
    detail.value = undefined;
    const res = await messageAPI.detail(row.id);
    detail.value = (res?.data ?? undefined) as MessageDetail | undefined;
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
    typeOptions,
    dataList,
    loading,
    detail,
    pagination,
    onSearch,
    loadDetail,
    resetForm,
    handleSizeChange,
    handleCurrentChange
  };
}
