<script setup lang="ts">
import { ref } from "vue";
import JobPane from "./JobPane.vue";
import RunPane from "./RunPane.vue";
import SchedulerSummary from "./SchedulerSummary.vue";
import { useSchedulerStats } from "./useStats";
import type { SchedulerJobRow } from "./types";

defineOptions({ name: "SystemScheduler" });

const tab = ref<"jobs" | "runs">("jobs");
const runFilter = ref<{ jobId: string; jobCode: string } | null>(null);
const { stats, load } = useSchedulerStats();

function openRuns(row: SchedulerJobRow) {
  runFilter.value = { jobId: row.id, jobCode: row.code };
  tab.value = "runs";
}

function refreshStats() {
  load();
}
</script>

<template>
  <div class="main">
    <SchedulerSummary :stats="stats" />
    <JobPane
      v-if="tab === 'jobs'"
      :tab="tab"
      @update:tab="tab = $event"
      @open-runs="openRuns"
      @changed="refreshStats"
    />
    <RunPane
      v-else
      :key="runFilter?.jobId ?? 'all'"
      :tab="tab"
      :job-filter="runFilter"
      @update:tab="tab = $event"
      @changed="refreshStats"
    />
  </div>
</template>

<style lang="scss" scoped>
.main {
  margin: 16px 16px 0;
}
</style>
