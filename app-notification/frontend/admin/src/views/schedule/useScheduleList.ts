import { onMounted, reactive, ref } from "vue";
import type { PaginationProps } from "@pureadmin/table";
import { message } from "@bit-labs.cn/owl-ui/utils/message";
import { scheduleAPI } from "../../api/notify";
import { applyPage } from "../../format";
import type { ScheduleRow } from "./types";

export function useScheduleList() {
  const form = reactive<{ keyword: string; status?: number }>({
    keyword: "",
    status: undefined
  });
  const dataList = ref<ScheduleRow[]>([]);
  const loading = ref(true);
  const pagination = reactive<PaginationProps>({ total: 0, pageSize: 10, currentPage: 1, background: true });

  async function onSearch() {
    loading.value = true;
    try {
      const res = await scheduleAPI.getList({
        page: pagination.currentPage,
        pageSize: pagination.pageSize,
        keyword: form.keyword || undefined,
        status: form.status ?? undefined
      });
      dataList.value = (res?.data?.list ?? []) as ScheduleRow[];
      applyPage(pagination, res);
    } catch {
      // 接口失败时不能装作「没有任务」，要让人看到是加载失败
      dataList.value = [];
      message("定时发送列表加载失败，请稍后重试", { type: "error" });
    } finally {
      loading.value = false;
    }
  }

  function resetForm(formEl?: { resetFields: () => void }) {
    formEl?.resetFields();
    form.keyword = "";
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

  return {
    form,
    dataList,
    loading,
    pagination,
    onSearch,
    resetForm,
    handleSizeChange,
    handleCurrentChange
  };
}
