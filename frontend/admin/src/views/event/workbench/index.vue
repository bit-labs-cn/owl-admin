<script setup lang="ts">
import { computed, h, ref, watch } from "vue";
import { useRouter } from "vue-router";
import { ElMessageBox } from "element-plus";
import { message } from "@bit-labs.cn/owl-ui/utils/message";
import { dialogService } from "@bit-labs.cn/owl-ui/components/ReDialog";
import ReCol from "@bit-labs.cn/owl-ui/components/ReCol";
import Segmented from "@bit-labs.cn/owl-ui/components/ReSegmented";
import { useRenderIcon } from "@bit-labs.cn/owl-ui/components/ReIcon/src/hooks";
import Refresh from "@iconify-icons/ep/refresh";
import Delete from "@iconify-icons/ep/delete";
import { cleanupConsumers } from "@bit-labs.cn/owl-admin-ui/api/event";
import { ADMIN_PERMS } from "../../../permissions";
import KpiCard from "./components/KpiCard.vue";
import TrendChart from "./components/TrendChart.vue";
import StatusDonut from "./components/StatusDonut.vue";
import RankBarChart from "./components/RankBarChart.vue";
import ConsumerGroupTable from "./components/ConsumerGroupTable.vue";
import RecentFailureList from "./components/RecentFailureList.vue";
import ConsumersCleanupForm from "./components/ConsumersCleanupForm.vue";
import { useWorkbench } from "./useWorkbench";
import { formatDuration, formatUnix } from "../format";
import { confirmRetryOutbox, openOutboxDetail } from "../useEventActions";
import type { EventOutboxRow } from "../types";

defineOptions({ name: "EventBusWorkbench" });

const router = useRouter();
const { days, loading, error, data, autoRefresh, load, toggleAutoRefresh } =
  useWorkbench();

const rangeOptions = [
  { label: "近 7 天", value: 7 },
  { label: "近 14 天", value: 14 }
];

// Segmented 通过 v-model 改 days，统一用 watch 触发重新取数，避免 change 与 modelValue 先后顺序不确定
watch(days, () => load());

const driverLabel = computed(() => {
  const driver = data.value?.driverName;
  if (!driver) return "未启用";
  if (driver === "redis") return "Redis Streams";
  if (driver === "rabbitmq") return "RabbitMQ";
  if (driver === "memory") return "内存（仅当前进程）";
  return driver;
});

const isMemoryDriver = computed(() => data.value?.driverName === "memory");

function percentOfTotal(value: number, total: number) {
  if (!total) return "0%";
  return `${((value / total) * 100).toFixed(1)}%`;
}

const kpiCards = computed(() => {
  const outbox = data.value?.outbox;
  const inbox = data.value?.inbox;
  const oldest = data.value?.oldestPending ?? null;

  return [
    {
      key: "outbox-pending",
      title: "待发送",
      value: outbox?.pending ?? 0,
      unit: "条",
      tone: "warning" as const,
      hint: oldest
        ? `最老一条已等待 ${formatDuration(oldest.ageSeconds)}`
        : "当前没有积压",
      query: { tab: "outbox", status: "0" }
    },
    {
      key: "outbox-sent",
      title: "已投递",
      value: outbox?.sent ?? 0,
      unit: "条",
      tone: "success" as const,
      hint: `占发件箱 ${percentOfTotal(outbox?.sent ?? 0, outbox?.total ?? 0)}`,
      query: { tab: "outbox", status: "1" }
    },
    {
      key: "outbox-failed",
      title: "投递失败",
      value: outbox?.failed ?? 0,
      unit: "条",
      tone: "danger" as const,
      hint: (outbox?.failed ?? 0) > 0 ? "可在事件记录中查看并重投" : "投递链路正常",
      query: { tab: "outbox", status: "2" }
    },
    {
      key: "inbox-failed",
      title: "消费失败",
      value: inbox?.failed ?? 0,
      unit: "条",
      tone: "info" as const,
      hint: `消费成功 ${inbox?.success ?? 0} 条 / 共 ${inbox?.total ?? 0} 条`,
      query: { tab: "inbox", status: "2" }
    }
  ];
});

function goLog(query: Record<string, string>) {
  router.push({ path: "/system/event/log/index", query });
}

function onRetry(row: EventOutboxRow) {
  confirmRetryOutbox(row, () => load(true));
}

const consumersCleanupFormRef = ref();

function openCleanupConsumers() {
  dialogService.open({
    title: "清理僵尸消费者",
    props: { formInline: { idleDays: 7 } },
    width: "560px",
    draggable: true,
    closeOnClickModal: false,
    closeOnPressEscape: false,
    contentRenderer: ({ options }) =>
      h(ConsumersCleanupForm, {
        ref: consumersCleanupFormRef,
        formInline: options.props.formInline
      }),
    beforeSure: done => {
      const formEl = consumersCleanupFormRef.value?.getRef();
      const data = consumersCleanupFormRef.value?.getFormData();
      formEl?.validate(async (valid: boolean) => {
        if (!valid) return;
        try {
          await ElMessageBox.confirm(
            `将删除空闲超过 ${data.idleDays} 天、且没有滞留消息的消费者记录，确认继续吗？`,
            "确认清理",
            {
              confirmButtonText: "清理",
              cancelButtonText: "取消",
              type: "warning",
              draggable: true,
              closeOnClickModal: false,
              closeOnPressEscape: false
            }
          );
        } catch {
          return;
        }

        try {
          const res = await cleanupConsumers({ idleDays: data.idleDays });
          const result = (res?.data ?? {}) as unknown as {
            removed?: number;
            skipped?: number;
          };
          done();
          message(
            `已删除 ${result.removed ?? 0} 个消费者记录，跳过有 pending 的 ${result.skipped ?? 0} 个`,
            { type: "success" }
          );
          load(true);
        } catch (err) {
          message(err instanceof Error ? err.message : "清理失败", {
            type: "error"
          });
        }
      });
    }
  });
}
</script>

<template>
  <div class="main">
    <el-card shadow="never" class="toolbar-card">
      <div class="toolbar">
        <div class="toolbar-title">
          <span class="title">事件总线工作台</span>
          <el-tag size="small" :type="isMemoryDriver ? 'warning' : 'success'">
            {{ driverLabel }}
          </el-tag>
          <span class="muted">
            统计近 {{ days }} 天 · 更新于 {{ formatUnix(data?.generatedAt) }}
          </span>
        </div>
        <div class="toolbar-actions">
          <Segmented v-model="days" :options="rangeOptions" />
          <el-switch
            :model-value="autoRefresh"
            active-text="自动刷新"
            @update:model-value="toggleAutoRefresh"
          />
          <el-button
            type="primary"
            :icon="useRenderIcon(Refresh)"
            :loading="loading"
            @click="load()"
          >
            刷新
          </el-button>
        </div>
      </div>
    </el-card>

    <el-alert
      v-if="error"
      :title="error"
      type="error"
      :closable="false"
      show-icon
      class="mb-4"
    />
    <el-alert
      v-else-if="isMemoryDriver"
      title="当前使用内存驱动：事件只在当前进程内存活，重启即丢，多实例之间也不会互通。生产环境请配置 event.driver 为 redis 或 rabbitmq。"
      type="warning"
      :closable="false"
      show-icon
      class="mb-4"
    />

    <el-row :gutter="16" v-loading="loading">
      <re-col
        v-for="card in kpiCards"
        :key="card.key"
        class="mb-4"
        :value="6"
        :md="12"
        :sm="12"
        :xs="24"
      >
        <KpiCard
          :title="card.title"
          :value="card.value"
          :unit="card.unit"
          :hint="card.hint"
          :tone="card.tone"
          clickable
          @click="goLog(card.query)"
        />
      </re-col>

      <re-col class="mb-4" :value="17" :md="24" :sm="24" :xs="24">
        <el-card shadow="never" class="panel">
          <template #header>
            <div class="panel-head">
              <span class="panel-title">事件量趋势</span>
              <span v-if="data?.trendTruncated" class="muted">
                数据量较大，仅统计最近的部分记录
              </span>
            </div>
          </template>
          <TrendChart :data="data?.trend ?? []" />
        </el-card>
      </re-col>

      <re-col class="mb-4" :value="7" :md="24" :sm="24" :xs="24">
        <el-card shadow="never" class="panel">
          <template #header>
            <div class="panel-head">
              <span class="panel-title">发件箱状态构成</span>
            </div>
          </template>
          <StatusDonut
            :summary="data?.outbox ?? { total: 0, pending: 0, sent: 0, failed: 0 }"
            :oldest="data?.oldestPending ?? null"
          />
        </el-card>
      </re-col>

      <re-col class="mb-4" :value="12" :md="24" :sm="24" :xs="24">
        <el-card shadow="never" class="panel">
          <template #header>
            <div class="panel-head">
              <span class="panel-title">事件类型 TOP 10</span>
            </div>
          </template>
          <RankBarChart :items="data?.topTypes ?? []" empty-text="统计范围内还没有事件" />
        </el-card>
      </re-col>

      <re-col class="mb-4" :value="12" :md="24" :sm="24" :xs="24">
        <el-card shadow="never" class="panel">
          <template #header>
            <div class="panel-head">
              <span class="panel-title">事件来源分布</span>
            </div>
          </template>
          <RankBarChart
            :items="data?.topSources ?? []"
            tone="success"
            empty-text="统计范围内还没有事件"
          />
        </el-card>
      </re-col>

      <re-col class="mb-4" :value="14" :md="24" :sm="24" :xs="24">
        <el-card shadow="never" class="panel">
          <template #header>
            <div class="panel-head">
              <span class="panel-title">消费组积压</span>
              <span class="head-actions">
                <span class="muted">
                  lag 为未投递条数，pending 为已投递未确认条数
                </span>
                <Perms :value="ADMIN_PERMS.event.cleanupConsumers">
                  <el-button
                    link
                    type="primary"
                    :icon="useRenderIcon(Delete)"
                    @click="openCleanupConsumers"
                  >
                    清理僵尸消费者
                  </el-button>
                </Perms>
              </span>
            </div>
          </template>
          <ConsumerGroupTable
            :groups="data?.consumerGroups ?? []"
            :stream="data?.stream ?? { key: '', exists: false, length: 0 }"
            :app-name="data?.appName ?? ''"
          />
        </el-card>
      </re-col>

      <re-col class="mb-4" :value="10" :md="24" :sm="24" :xs="24">
        <el-card shadow="never" class="panel">
          <template #header>
            <div class="panel-head">
              <span class="panel-title">最近投递失败</span>
              <span class="muted">最多 10 条</span>
            </div>
          </template>
          <RecentFailureList
            :rows="data?.recentFailures ?? []"
            @detail="openOutboxDetail"
            @retry="onRetry"
          />
        </el-card>
      </re-col>
    </el-row>
  </div>
</template>

<style lang="scss" scoped>
.main {
  margin: 16px 16px 0;
}

.toolbar-card {
  margin-bottom: 16px;
  border-color: var(--owl-border-color, rgba(0, 0, 0, 0.06));
}

.toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  align-items: center;
  justify-content: space-between;
}

.toolbar-title {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  align-items: center;
}

.title {
  font-size: 16px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.muted {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.toolbar-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  align-items: center;
}

.panel {
  height: 100%;
  border-color: var(--owl-border-color, rgba(0, 0, 0, 0.06));

  :deep(.el-card__header) {
    padding: 12px 16px;
  }

  :deep(.el-card__body) {
    padding: 12px 16px 16px;
  }
}

.panel-head {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
  justify-content: space-between;
}

.panel-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.head-actions {
  display: inline-flex;
  flex-wrap: wrap;
  gap: 10px;
  align-items: center;
}
</style>
