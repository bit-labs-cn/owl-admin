import { onMounted, reactive, ref } from "vue";
import type { PaginationProps } from "@pureadmin/table";
import { message } from "@bit-labs.cn/owl-ui/utils/message";
import { deliveryAPI } from "../../api/notify";
import { applyPage } from "../../format";
import type { DeliveryBatchRow, DeliveryRow } from "./types";

export function useDeliveryList() {
  const form = reactive<{ channel: string; status?: number; recipientType?: number }>({
    channel: "",
    status: undefined,
    recipientType: undefined
  });
  const dataList = ref<DeliveryBatchRow[]>([]);
  const loading = ref(true);
  const detailList = ref<DeliveryRow[]>([]);
  const detailLoading = ref(false);
  const pagination = reactive<PaginationProps>({ total: 0, pageSize: 10, currentPage: 1, background: true });

  async function onSearch() {
    loading.value = true;
    try {
      const res = await deliveryAPI.getList({
        page: pagination.currentPage,
        pageSize: pagination.pageSize,
        channel: form.channel || undefined,
        status: form.status ?? undefined,
        recipientType: form.recipientType ?? undefined
      });
      dataList.value = (res?.data?.list ?? []) as DeliveryBatchRow[];
      applyPage(pagination, res);
    } catch {
      // 接口失败时不能装作「没有记录」，要让人看到是加载失败
      dataList.value = [];
      message("发送记录加载失败，请稍后重试", { type: "error" });
    } finally {
      loading.value = false;
    }
  }

  async function loadDetail(row: DeliveryBatchRow) {
    detailLoading.value = true;
    detailList.value = [];
    try {
      const res = await deliveryAPI.getDetail(row.messageId);
      detailList.value = (res?.data ?? []) as DeliveryRow[];
    } finally {
      detailLoading.value = false;
    }
  }

  function resetForm(formEl?: { resetFields: () => void }) {
    formEl?.resetFields();
    form.channel = "";
    form.status = undefined;
    form.recipientType = undefined;
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
    detailList,
    detailLoading,
    pagination,
    onSearch,
    loadDetail,
    resetForm,
    handleSizeChange,
    handleCurrentChange
  };
}
