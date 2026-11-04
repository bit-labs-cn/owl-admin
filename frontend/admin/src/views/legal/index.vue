<script setup lang="ts">
import { provide, ref } from "vue";
import DocumentPane from "./DocumentPane.vue";
import VersionPane from "./VersionPane.vue";
import ScenePane from "./ScenePane.vue";
import EventPane from "./EventPane.vue";

defineOptions({ name: "SystemLegal" });

const tab = ref("document");
provide("legalTab", tab);
</script>

<template>
  <div class="main">
    <DocumentPane v-if="tab === 'document'" />
    <VersionPane v-else-if="tab === 'version'" />
    <ScenePane v-else-if="tab === 'scene'" />
    <EventPane v-else />
  </div>
</template>

<style lang="scss" scoped>
.main {
  margin: 16px 16px 0;
}
</style>
