<script setup lang="ts">
import { codecTools, useCodecTool } from "./useCodecTool";

defineOptions({ name: "SystemCodecTool" });

const {
  active,
  jsonInput,
  jsonOutput,
  jsonError,
  base64Input,
  base64Output,
  base64Error,
  urlInput,
  urlOutput,
  urlError,
  stamp,
  localTime,
  secondsText,
  millisText,
  timeError,
  digestInput,
  md5Text,
  shaText,
  digestError,
  uuids,
  formatJson,
  encodeBase64,
  decodeBase64,
  encodeUrl,
  decodeUrl,
  fromStamp,
  fromLocal,
  digest,
  makeUuids,
  copyText
} = useCodecTool();
</script>

<template>
  <div class="main">
    <div class="codec">
      <nav class="codec-nav" aria-label="编码工具">
        <button
          v-for="tool in codecTools"
          :key="tool.id"
          type="button"
          class="codec-link"
          :class="{ 'is-current': active === tool.id }"
          :aria-current="active === tool.id ? 'page' : undefined"
          @click="active = tool.id"
        >
          {{ tool.label }}
        </button>
      </nav>

      <section v-if="active === 'json'" class="codec-main">
        <header class="codec-head">
          <h2>JSON</h2>
          <p>格式化和压缩都在浏览器里完成。</p>
        </header>
        <el-input v-model="jsonInput" type="textarea" :rows="10" placeholder="粘贴 JSON" />
        <div class="codec-actions">
          <el-button type="primary" @click="formatJson(2)">格式化</el-button>
          <el-button @click="formatJson()">压缩</el-button>
          <el-button :disabled="!jsonOutput" @click="copyText(jsonOutput)">复制结果</el-button>
        </div>
        <p v-if="jsonError" class="codec-error">{{ jsonError }}</p>
        <el-input v-if="jsonOutput" :model-value="jsonOutput" type="textarea" :rows="10" readonly />
      </section>

      <section v-else-if="active === 'base64'" class="codec-main">
        <header class="codec-head">
          <h2>Base64</h2>
          <p>按 UTF-8 编解码，原文不会发到服务器。</p>
        </header>
        <el-input v-model="base64Input" type="textarea" :rows="8" placeholder="原文或 Base64" />
        <div class="codec-actions">
          <el-button type="primary" @click="encodeBase64">编码</el-button>
          <el-button @click="decodeBase64">解码</el-button>
          <el-button :disabled="!base64Output" @click="copyText(base64Output)">复制结果</el-button>
        </div>
        <p v-if="base64Error" class="codec-error">{{ base64Error }}</p>
        <el-input v-if="base64Output" :model-value="base64Output" type="textarea" :rows="8" readonly />
      </section>

      <section v-else-if="active === 'url'" class="codec-main">
        <header class="codec-head">
          <h2>URL</h2>
          <p>对查询参数里的一段文本做编解码。</p>
        </header>
        <el-input v-model="urlInput" type="textarea" :rows="8" placeholder="原文或已编码文本" />
        <div class="codec-actions">
          <el-button type="primary" @click="encodeUrl">编码</el-button>
          <el-button @click="decodeUrl">解码</el-button>
          <el-button :disabled="!urlOutput" @click="copyText(urlOutput)">复制结果</el-button>
        </div>
        <p v-if="urlError" class="codec-error">{{ urlError }}</p>
        <el-input v-if="urlOutput" :model-value="urlOutput" type="textarea" :rows="8" readonly />
      </section>

      <section v-else-if="active === 'time'" class="codec-main">
        <header class="codec-head">
          <h2>时间戳</h2>
          <p>10 位按秒、13 位按毫秒，结果用本地时间显示。</p>
        </header>
        <el-form label-width="88px">
          <el-form-item label="时间戳">
            <div class="hash-row">
              <el-input v-model="stamp" placeholder="秒或毫秒" />
              <el-button type="primary" @click="fromStamp">转成本地时间</el-button>
            </div>
          </el-form-item>
          <el-form-item label="本地时间">
            <div class="hash-row">
              <el-date-picker
                v-model="localTime"
                type="datetime"
                value-format="YYYY-MM-DD HH:mm:ss"
                placeholder="选择时间"
              />
              <el-button @click="fromLocal">转成时间戳</el-button>
            </div>
          </el-form-item>
        </el-form>
        <p v-if="timeError" class="codec-error">{{ timeError }}</p>
        <dl v-if="secondsText" class="time-result">
          <div>
            <dt>秒</dt>
            <dd>
              <code>{{ secondsText }}</code>
              <el-button link type="primary" @click="copyText(secondsText)">复制</el-button>
            </dd>
          </div>
          <div>
            <dt>毫秒</dt>
            <dd>
              <code>{{ millisText }}</code>
              <el-button link type="primary" @click="copyText(millisText)">复制</el-button>
            </dd>
          </div>
        </dl>
      </section>

      <section v-else-if="active === 'digest'" class="codec-main">
        <header class="codec-head">
          <h2>摘要</h2>
          <p>MD5 和 SHA-256 都在浏览器里计算。</p>
        </header>
        <el-input v-model="digestInput" type="textarea" :rows="8" placeholder="要计算的文本" />
        <div class="codec-actions">
          <el-button type="primary" @click="digest">计算</el-button>
        </div>
        <p v-if="digestError" class="codec-error">{{ digestError }}</p>
        <dl v-if="md5Text" class="time-result">
          <div>
            <dt>MD5</dt>
            <dd>
              <code>{{ md5Text }}</code>
              <el-button link type="primary" @click="copyText(md5Text)">复制</el-button>
            </dd>
          </div>
          <div v-if="shaText">
            <dt>SHA-256</dt>
            <dd>
              <code>{{ shaText }}</code>
              <el-button link type="primary" @click="copyText(shaText)">复制</el-button>
            </dd>
          </div>
        </dl>
      </section>

      <section v-else class="codec-main">
        <header class="codec-head">
          <h2>UUID</h2>
          <p>使用浏览器的随机数生成 UUID v4。</p>
        </header>
        <div class="codec-actions">
          <el-button type="primary" @click="makeUuids(1)">生成一个</el-button>
          <el-button @click="makeUuids(5)">生成五个</el-button>
          <el-button :disabled="!uuids.length" @click="copyText(uuids.join('\n'))">复制全部</el-button>
        </div>
        <ul class="uuid-list">
          <li v-for="item in uuids" :key="item">
            <code>{{ item }}</code>
            <el-button link type="primary" @click="copyText(item)">复制</el-button>
          </li>
        </ul>
        <p v-if="!uuids.length" class="codec-empty">还没有 UUID。</p>
      </section>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.codec {
  display: flex;
  align-items: stretch;
  min-height: calc(100vh - 154px);
  overflow: hidden;
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 12px;
}

.codec-nav {
  flex: 0 0 168px;
  padding: 18px 12px;
  background: var(--el-fill-color-lighter);
  border-right: 1px solid var(--el-border-color-lighter);
}

.codec-link {
  display: block;
  width: 100%;
  margin: 0 0 4px;
  padding: 8px 12px;
  color: var(--el-text-color-regular);
  text-align: left;
  background: transparent;
  border: 0;
  border-radius: 8px;
  cursor: pointer;
}

.codec-link.is-current {
  color: var(--el-color-primary);
  background: var(--el-bg-color);
}

.codec-main {
  flex: 1;
  min-width: 0;
  padding: 20px 22px;
}

.codec-head {
  margin-bottom: 16px;

  h2 {
    margin: 0;
    font-size: 18px;
    font-weight: 600;
  }

  p {
    margin: 6px 0 0;
    color: var(--el-text-color-secondary);
    font-size: 13px;
    line-height: 1.6;
  }
}

.codec-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin: 12px 0;
}

.codec-error {
  margin: 0 0 12px;
  color: var(--el-color-danger);
}

.codec-empty {
  margin: 0;
  color: var(--el-text-color-secondary);
}

.hash-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  width: 100%;
}

.time-result,
.uuid-list {
  margin: 0;
  padding: 0;
  list-style: none;
}

.time-result > div,
.uuid-list li {
  display: flex;
  gap: 12px;
  align-items: center;
  padding: 8px 0;
  border-top: 1px solid var(--el-border-color-lighter);
}

.time-result dt {
  flex: 0 0 72px;
  color: var(--el-text-color-secondary);
}

.time-result dd {
  display: flex;
  flex: 1;
  gap: 12px;
  align-items: center;
  justify-content: space-between;
  margin: 0;
  min-width: 0;
}

code {
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  word-break: break-all;
}

@media (max-width: 800px) {
  .codec {
    flex-direction: column;
  }

  .codec-nav {
    display: flex;
    flex: none;
    gap: 4px;
    overflow-x: auto;
    border-right: 0;
    border-bottom: 1px solid var(--el-border-color-lighter);
  }

  .codec-link {
    width: auto;
    white-space: nowrap;
  }
}
</style>
