<script setup lang="ts">
import { provide, ref, watch } from "vue";
import { useRoute } from "vue-router";
import OutboxPane from "./OutboxPane.vue";
import InboxPane from "./InboxPane.vue";

defineOptions({ name: "EventBusLog" });

const route = useRoute();
const tab = ref(route.query.tab === "inbox" ? "inbox" : "outbox");
// 工作台 KPI 点击会带上 status，切到对应 Tab 时作为初始筛选条件
const initialStatus = ref(
  typeof route.query.status === "string" ? route.query.status : ""
);

// 页面可能被 keep-alive 缓存，从工作台再次跳进来时 query 变了要跟上
watch(
  () => route.query,
  query => {
    tab.value = query.tab === "inbox" ? "inbox" : "outbox";
    initialStatus.value = typeof query.status === "string" ? query.status : "";
  }
);

provide("eventLogTab", tab);
</script>

<template>
  <div class="main">
    <OutboxPane
      v-if="tab === 'outbox'"
      :key="`outbox-${initialStatus}`"
      :initial-status="initialStatus"
    />
    <InboxPane
      v-else
      :key="`inbox-${initialStatus}`"
      :initial-status="initialStatus"
    />
  </div>
</template>

<style lang="scss" scoped>
.main {
  margin: 16px 16px 0;
}
</style>
