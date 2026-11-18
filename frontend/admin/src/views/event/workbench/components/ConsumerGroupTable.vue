<script setup lang="ts">
import { computed } from "vue";
import { formatDuration } from "../../format";
import type { EventConsumerGroup, EventStreamSnapshot } from "../../types";

const props = defineProps<{
  groups: EventConsumerGroup[];
  stream: EventStreamSnapshot;
  appName?: string;
}>();

const rows = computed(() => props.groups ?? []);
const error = computed(() => props.stream?.error || "");

function lagTone(lag: number) {
  return lag > 0 ? "warning" : "success";
}

function pendingTone(pending: number) {
  return pending > 0 ? "danger" : "success";
}

/** 没有活跃消费者且还有积压，说明这组当前没人读，消息会一直堆着 */
function stalled(row: EventConsumerGroup) {
  return row.activeConsumers === 0 && row.lag > 0;
}
</script>

<template>
  <div class="consumer-groups">
    <el-alert
      v-if="error"
      :title="`Redis Stream 读取失败：${error}`"
      type="warning"
      :closable="false"
      show-icon
      class="mb-3"
    />
    <el-descriptions :column="4" border size="small" class="stream-meta">
      <el-descriptions-item label="应用">
        <el-tag type="primary" size="small">{{ appName || "—" }}</el-tag>
      </el-descriptions-item>
      <el-descriptions-item label="Stream">
        {{ stream?.key || "—" }}
      </el-descriptions-item>
      <el-descriptions-item label="状态">
        <el-tag :type="stream?.exists ? 'success' : 'info'" size="small">
          {{ stream?.exists ? "已创建" : "未创建" }}
        </el-tag>
      </el-descriptions-item>
      <el-descriptions-item label="消息总量">
        {{ stream?.length ?? 0 }}
      </el-descriptions-item>
    </el-descriptions>

    <el-empty
      v-if="rows.length === 0"
      description="当前没有消费组，或尚未配置 Redis 驱动"
      :image-size="80"
    />
    <el-table v-else :data="rows" border size="small" row-key="name">
      <el-table-column
        prop="name"
        label="消费组"
        min-width="220"
        show-overflow-tooltip
      />
      <el-table-column label="消费者" width="130" align="center">
        <template #default="{ row }">
          <el-tooltip
            placement="top"
            content="活跃 = 最近 2 分钟内在读消息的消费者。累计包含进程重启后留在 Redis 里的旧记录。"
          >
            <span class="consumers">
              <el-tag
                :type="row.activeConsumers > 0 ? 'success' : 'info'"
                size="small"
              >
                {{ row.activeConsumers }} 活跃
              </el-tag>
              <span class="muted">/ {{ row.totalConsumers }}</span>
            </span>
          </el-tooltip>
        </template>
      </el-table-column>
      <el-table-column label="最近活跃" width="130" align="center">
        <template #default="{ row }">
          {{
            row.lastActiveSeconds < 0
              ? "从未读取"
              : `${formatDuration(row.lastActiveSeconds)}前`
          }}
        </template>
      </el-table-column>
      <el-table-column label="pending" width="110" align="center">
        <template #default="{ row }">
          <el-tag :type="pendingTone(row.pending)" size="small">
            {{ row.pending }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="本主题未投递" width="140" align="center">
        <template #default="{ row }">
          <el-tooltip
            placement="top"
            content="该主题在发件箱里还没投递成功的条数（Relay 侧积压）。和 lag 是两个维度：lag 是消费端还没读到的条数，这里算的是还没发出去的量。"
          >
            <el-tag
              v-if="row.topicKnown && row.topicPending > 0"
              type="danger"
              size="small"
            >
              {{ row.topicPending }}
            </el-tag>
            <el-tag
              v-else-if="row.topicKnown"
              type="success"
              size="small"
            >
              0
            </el-tag>
            <el-tag v-else type="info" size="small">—</el-tag>
          </el-tooltip>
        </template>
      </el-table-column>
      <el-table-column label="lag" width="100" align="center">
        <template #default="{ row }">
          <!-- Redis 无法确定 lag 时返回 -1，此时显示占位而不是误导性的 0；
               lag 是整条 stream 的未读量，不是本主题单独的积压 -->
          <el-tooltip
            placement="top"
            content="这个消费组还没读到的条数，是整个 stream 的口径（所有主题共用一条 stream），不是本主题单独的积压。"
          >
            <el-tag v-if="row.lag < 0" type="info" size="small">—</el-tag>
            <el-tag v-else :type="lagTone(row.lag)" size="small">{{ row.lag }}</el-tag>
          </el-tooltip>
        </template>
      </el-table-column>
      <el-table-column prop="entriesRead" label="已读条数" width="110" align="center" />
      <el-table-column
        prop="lastDeliveredId"
        label="最后投递 ID"
        min-width="170"
        show-overflow-tooltip
      />
      <el-table-column label="状态" width="150" align="center">
        <template #default="{ row }">
          <el-tag v-if="stalled(row)" type="warning" size="small">
            无活跃消费者
          </el-tag>
          <el-tag v-else-if="row.pending > 0" type="danger" size="small">
            处理中/待重试
          </el-tag>
          <el-tag v-else type="success" size="small">正常</el-tag>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<style lang="scss" scoped>
.stream-meta {
  margin-bottom: 12px;
}

.consumers {
  display: inline-flex;
  gap: 4px;
  align-items: center;
}

.muted {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
</style>
