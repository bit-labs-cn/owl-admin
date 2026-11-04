<script setup lang="ts">
import { computed } from "vue";
import HealthGauge from "./HealthGauge.vue";
import MonitorChart from "./MonitorChart.vue";
import { useServerMonitor } from "./useServerMonitor";

defineOptions({ name: "SystemServerMonitor" });

const {
  snapshot,
  history,
  labels,
  latest,
  earlier,
  healthScore,
  loading,
  loadError,
  autoRefresh,
  load,
  formatBytes,
  formatRate,
  formatDuration,
  percent,
  shiftText,
  collectedText
} = useServerMonitor();

const cpuSeries = computed(() => [
  { name: "用户态", color: "#5b8ff9", data: history.value.map(point => point.cpuUser || 0) },
  { name: "内核态", color: "#f6bd16", data: history.value.map(point => point.cpuSystem || 0) },
  { name: "IO 等待", color: "#5ad8a6", data: history.value.map(point => point.cpuIOWait || 0) }
]);

const netSeries = computed(() => [
  { name: "入网", color: "#5b8ff9", area: true, data: history.value.map(point => point.netIn) },
  { name: "出网", color: "#f6bd16", area: true, data: history.value.map(point => point.netOut) }
]);

const diskSeries = computed(() => [
  { name: "读", color: "#5ad8a6", data: history.value.map(point => point.diskRead) },
  { name: "写", color: "#e8684a", data: history.value.map(point => point.diskWrite) }
]);

const diskIORate = computed(() => {
  if (latest.value?.diskRead == null || latest.value.diskWrite == null) return null;
  return latest.value.diskRead + latest.value.diskWrite;
});

function rateAxis(value: number) {
  return formatRate(value);
}

function percentAxis(value: number) {
  return `${percent(value)}%`;
}
</script>

<template>
  <div v-loading="loading" class="main">
    <div class="monitor">
      <header class="monitor-head">
        <div>
          <h2>{{ snapshot?.host.hostname || "运行概况" }}</h2>
          <p v-if="snapshot">
            {{ snapshot.host.platform }} {{ snapshot.host.platformVersion }} · {{ snapshot.host.os }} /
            {{ snapshot.host.arch }} · {{ snapshot.host.cpuCores }} 核 · 已运行
            {{ formatDuration(snapshot.host.uptimeSeconds) }}
          </p>
          <p v-else>正在读取这台机器的运行情况</p>
          <p>近 15 分钟的曲线留在这个页面里，每 3 秒一个点，离开后重新累计。</p>
        </div>
        <div class="monitor-actions">
          <span v-if="collectedText()" class="collected">采集于 {{ collectedText() }}</span>
          <el-switch v-model="autoRefresh" active-text="自动刷新" />
          <el-button @click="load">刷新</el-button>
        </div>
      </header>

      <el-alert v-if="loadError" :title="loadError" type="error" show-icon :closable="false" />

      <template v-if="snapshot">
        <div class="kpi-row">
          <section class="metric">
            <h3>CPU</h3>
            <p class="metric-value">
              {{ percent(snapshot.host.cpuPercent) }}%
              <span
                v-if="shiftText(snapshot.host.cpuPercent, earlier?.cpu)"
                class="delta"
                :class="{
                  'is-up': earlier && snapshot.host.cpuPercent > earlier.cpu,
                  'is-down': earlier && snapshot.host.cpuPercent < earlier.cpu
                }"
              >
                {{ shiftText(snapshot.host.cpuPercent, earlier?.cpu) }}
              </span>
            </p>
            <p class="metric-note">{{ snapshot.host.cpuModel || "型号没有读到" }}</p>
            <p v-if="snapshot.host.cpuError" class="metric-warn">{{ snapshot.host.cpuError }}</p>
          </section>

          <section class="metric">
            <h3>内存</h3>
            <p class="metric-value">
              {{ percent(snapshot.host.memoryPercent) }}%
              <span
                v-if="shiftText(snapshot.host.memoryPercent, earlier?.memory)"
                class="delta"
                :class="{
                  'is-up': earlier && snapshot.host.memoryPercent > earlier.memory,
                  'is-down': earlier && snapshot.host.memoryPercent < earlier.memory
                }"
              >
                {{ shiftText(snapshot.host.memoryPercent, earlier?.memory) }}
              </span>
            </p>
            <p class="metric-note">
              {{ formatBytes(snapshot.host.memoryUsed) }} / {{ formatBytes(snapshot.host.memoryTotal) }}
            </p>
            <p v-if="snapshot.host.memoryError" class="metric-warn">{{ snapshot.host.memoryError }}</p>
          </section>

          <section class="metric">
            <h3>网络</h3>
            <p class="metric-value">{{ formatRate(latest?.netIn ?? null) }}</p>
            <p class="metric-note">出 {{ formatRate(latest?.netOut ?? null) }}</p>
            <p v-if="snapshot.host.netError" class="metric-warn">{{ snapshot.host.netError }}</p>
          </section>

          <section class="metric">
            <h3>磁盘 IO</h3>
            <p class="metric-value">{{ formatRate(diskIORate) }}</p>
            <p class="metric-note">读 {{ formatRate(latest?.diskRead ?? null) }} · 写 {{ formatRate(latest?.diskWrite ?? null) }}</p>
            <p v-if="snapshot.host.diskIOError" class="metric-warn">{{ snapshot.host.diskIOError }}</p>
          </section>
        </div>

        <div class="chart-row">
          <section class="metric">
            <header class="chart-head">
              <h3>CPU 走势</h3>
              <span>每 3 秒</span>
            </header>
            <MonitorChart :labels="labels" :series="cpuSeries" :y-max="100" :format-value="percentAxis" />
          </section>
          <section class="metric health">
            <header class="chart-head">
              <h3>主机健康</h3>
              <span>综合 CPU / 内存 / 磁盘占用</span>
            </header>
            <HealthGauge :score="healthScore" />
            <dl class="health-facts">
              <div>
                <dt>负载</dt>
                <dd>{{ snapshot.host.loadKnown ? snapshot.host.load1 : "没有读到" }}</dd>
              </div>
              <div>
                <dt>温度</dt>
                <dd>{{ snapshot.host.temperatureKnown ? `${snapshot.host.temperatureC} °C` : "没有读到" }}</dd>
              </div>
              <div>
                <dt>进程</dt>
                <dd>{{ snapshot.host.processCount > 0 ? snapshot.host.processCount : "没有读到" }}</dd>
              </div>
            </dl>
          </section>
        </div>

        <div class="chart-row chart-row--even">
          <section class="metric">
            <header class="chart-head">
              <h3>网络吞吐</h3>
              <span>每 3 秒</span>
            </header>
            <MonitorChart :labels="labels" :series="netSeries" :format-value="rateAxis" />
          </section>
          <section class="metric">
            <header class="chart-head">
              <h3>磁盘读写</h3>
              <span>每 3 秒</span>
            </header>
            <MonitorChart :labels="labels" :series="diskSeries" :format-value="rateAxis" />
          </section>
        </div>

        <section class="metric">
          <h3>磁盘</h3>
          <el-alert v-if="snapshot.host.diskError" :title="snapshot.host.diskError" type="warning" show-icon :closable="false" />
          <p v-if="!snapshot.host.disks.length && !snapshot.host.diskError" class="metric-note">没有读到磁盘</p>
          <ul v-else class="disk-list">
            <li v-for="disk in snapshot.host.disks" :key="disk.mount">
              <div class="disk-name">
                <span>{{ disk.mount }}</span>
                <span>{{ disk.fsType }} · {{ percent(disk.percent) }}%</span>
              </div>
              <el-progress :percentage="percent(disk.percent)" :stroke-width="8" />
              <p>{{ formatBytes(disk.used) }} / {{ formatBytes(disk.total) }}</p>
            </li>
          </ul>
        </section>

        <div class="detail-row">
          <section class="metric">
            <h3>进程</h3>
            <dl>
              <div>
                <dt>Go</dt>
                <dd>{{ snapshot.runtime.goVersion }}</dd>
              </div>
              <div>
                <dt>已运行</dt>
                <dd>{{ formatDuration(snapshot.runtime.uptimeSeconds) }}</dd>
              </div>
              <div>
                <dt>Goroutine</dt>
                <dd>{{ snapshot.runtime.goroutines }}</dd>
              </div>
              <div>
                <dt>堆内存</dt>
                <dd>{{ formatBytes(snapshot.runtime.heapAlloc) }} / {{ formatBytes(snapshot.runtime.heapSys) }}</dd>
              </div>
              <div>
                <dt>GC</dt>
                <dd>{{ snapshot.runtime.numGC }} 次</dd>
              </div>
            </dl>
          </section>

          <section class="metric">
            <h3>数据库</h3>
            <el-alert v-if="snapshot.database.error" :title="snapshot.database.error" type="warning" show-icon :closable="false" />
            <dl>
              <div>
                <dt>Ping</dt>
                <dd>{{ snapshot.database.ok ? `${snapshot.database.pingMilliseconds} ms` : "没有连通" }}</dd>
              </div>
              <div>
                <dt>打开</dt>
                <dd>{{ snapshot.database.open }}</dd>
              </div>
              <div>
                <dt>使用中</dt>
                <dd>{{ snapshot.database.inUse }}</dd>
              </div>
              <div>
                <dt>空闲</dt>
                <dd>{{ snapshot.database.idle }}</dd>
              </div>
              <div>
                <dt>等待次数</dt>
                <dd>{{ snapshot.database.waitCount }}</dd>
              </div>
            </dl>
          </section>

          <section class="metric">
            <h3>Redis</h3>
            <el-alert
              v-if="!snapshot.redis.ok"
              :title="snapshot.redis.error || 'Redis 不可用'"
              type="warning"
              show-icon
              :closable="false"
            />
            <dl v-else>
              <div>
                <dt>版本</dt>
                <dd>{{ snapshot.redis.version || "没有读到" }}</dd>
              </div>
              <div>
                <dt>已运行</dt>
                <dd>{{ formatDuration(snapshot.redis.uptimeSeconds) }}</dd>
              </div>
              <div>
                <dt>内存</dt>
                <dd>{{ formatBytes(snapshot.redis.usedMemory) }}</dd>
              </div>
              <div>
                <dt>连接数</dt>
                <dd>{{ snapshot.redis.connectedClients }}</dd>
              </div>
              <div>
                <dt>命中率</dt>
                <dd>{{ snapshot.redis.hitRateKnown ? `${snapshot.redis.hitRate}%` : "暂无访问" }}</dd>
              </div>
            </dl>
          </section>
        </div>
      </template>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.monitor {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 20px 22px 22px;
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 12px;
}

.monitor-head {
  display: flex;
  gap: 16px;
  align-items: flex-start;
  justify-content: space-between;

  h2 {
    margin: 0;
    font-size: 20px;
    font-weight: 600;
  }

  p {
    margin: 6px 0 0;
    color: var(--el-text-color-secondary);
    font-size: 13px;
    line-height: 1.5;
  }
}

.monitor-actions {
  display: flex;
  gap: 12px;
  align-items: center;
}

.collected {
  color: var(--el-text-color-secondary);
  font-size: 13px;
  white-space: nowrap;
}

.kpi-row,
.detail-row,
.chart-row {
  display: grid;
  gap: 12px;
}

.kpi-row,
.detail-row {
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.detail-row,
.chart-row {
  grid-template-columns: minmax(0, 1.7fr) minmax(280px, 0.9fr);
}

.chart-row--even,
.detail-row {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.chart-row--even {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.metric {
  padding: 16px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 10px;

  h3 {
    margin: 0;
    font-size: 14px;
    font-weight: 600;
  }

  .el-alert {
    margin: 12px 0;
  }
}

.chart-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  margin-bottom: 4px;

  span {
    color: var(--el-text-color-secondary);
    font-size: 12px;
  }
}

.metric-value {
  display: flex;
  gap: 8px;
  align-items: baseline;
  margin: 10px 0 0;
  font-size: 28px;
  font-weight: 600;
  line-height: 1.2;
  font-variant-numeric: tabular-nums;
}

.delta {
  font-size: 13px;
  font-weight: 600;
}

.delta.is-up {
  color: var(--el-color-danger);
}

.delta.is-down {
  color: var(--el-color-success);
}

.metric-note,
.metric-warn {
  margin: 8px 0 0;
  font-size: 13px;
  line-height: 1.5;
}

.metric-note {
  color: var(--el-text-color-secondary);
}

.metric-warn {
  color: var(--el-color-warning);
}

.health-facts {
  margin-top: 4px;
}

dl {
  margin: 0;
}

dl > div {
  display: grid;
  grid-template-columns: 72px minmax(0, 1fr);
  gap: 8px;
  padding: 7px 0;
  border-top: 1px solid var(--el-border-color-extra-light);
}

dt {
  color: var(--el-text-color-secondary);
}

dd {
  margin: 0;
  font-variant-numeric: tabular-nums;
  word-break: break-word;
}

.disk-list {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 16px;
  margin: 14px 0 0;
  padding: 0;
  list-style: none;

  p {
    margin: 6px 0 0;
    color: var(--el-text-color-secondary);
    font-size: 13px;
  }
}

.disk-name {
  display: flex;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 6px;
  font-size: 13px;
}

@media (max-width: 1100px) {
  .monitor-head,
  .monitor-actions,
  .kpi-row,
  .detail-row,
  .chart-row,
  .chart-row--even {
    grid-template-columns: 1fr;
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
