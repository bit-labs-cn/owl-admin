<script setup lang="ts">
import { usePasswordTool } from "./usePasswordTool";

defineOptions({ name: "SystemPasswordTool" });

const {
  form,
  passwords,
  generating,
  plain,
  hash,
  hashing,
  charsetSelected,
  generate,
  hashPassword,
  copyText,
  copyAll
} = usePasswordTool();
</script>

<template>
  <div class="main">
    <div class="tool-grid">
      <section class="tool-panel">
        <header class="tool-head">
          <div>
            <h2>随机密码</h2>
            <p>生成结果只在这一页显示，不会保存。</p>
          </div>
          <el-button v-if="passwords.length" @click="copyAll">复制全部</el-button>
        </header>

        <el-form label-width="88px">
          <el-form-item label="长度">
            <el-input-number v-model="form.length" :min="8" :max="64" :step="1" step-strictly />
          </el-form-item>
          <el-form-item label="条数">
            <el-input-number v-model="form.count" :min="1" :max="20" :step="1" step-strictly />
          </el-form-item>
          <el-form-item label="字符">
            <div class="charset">
              <el-checkbox v-model="form.upper">大写字母</el-checkbox>
              <el-checkbox v-model="form.lower">小写字母</el-checkbox>
              <el-checkbox v-model="form.digit">数字</el-checkbox>
              <el-checkbox v-model="form.symbol">符号</el-checkbox>
              <el-checkbox v-model="form.excludeAmbiguous">排除易混字符</el-checkbox>
            </div>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" :disabled="!charsetSelected" :loading="generating" @click="generate">
              生成
            </el-button>
          </el-form-item>
        </el-form>

        <p v-if="!passwords.length" class="tool-empty">点生成后，密码会出现在这里。</p>
        <ul v-else class="password-list">
          <li v-for="(item, index) in passwords" :key="index">
            <code>{{ item }}</code>
            <el-button link type="primary" @click="copyText(item)">复制</el-button>
          </li>
        </ul>
      </section>

      <section class="tool-panel">
        <header class="tool-head">
          <div>
            <h2>计算密文</h2>
            <p>用 bcrypt 算出可写入用户密码的密文。明文和密文都不会记入操作日志。</p>
          </div>
        </header>
        <el-form label-width="88px" @submit.prevent="hashPassword">
          <el-form-item label="明文">
            <el-input v-model="plain" type="password" show-password maxlength="72" placeholder="要计算的密码" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" :loading="hashing" @click="hashPassword">计算</el-button>
          </el-form-item>
          <el-form-item v-if="hash" label="密文">
            <div class="hash-row">
              <el-input :model-value="hash" readonly />
              <el-button @click="copyText(hash)">复制</el-button>
            </div>
          </el-form-item>
        </el-form>
      </section>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.tool-grid {
  display: grid;
  grid-template-columns: minmax(0, 1.2fr) minmax(320px, 0.8fr);
  gap: 12px;
}

.tool-panel {
  padding: 20px 22px 8px;
  background: var(--el-bg-color);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 12px;
}

.tool-head {
  display: flex;
  gap: 16px;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 18px;

  h2 {
    margin: 0;
    font-size: 18px;
    font-weight: 600;
    line-height: 1.4;
  }

  p {
    margin: 6px 0 0;
    color: var(--el-text-color-secondary);
    font-size: 13px;
    line-height: 1.6;
  }
}

.charset {
  display: flex;
  flex-wrap: wrap;
  gap: 4px 16px;
}

.tool-empty {
  margin: 0 0 16px;
  color: var(--el-text-color-secondary);
}

.password-list {
  margin: 0 0 16px;
  padding: 0;
  list-style: none;

  li {
    display: flex;
    gap: 12px;
    align-items: center;
    justify-content: space-between;
    padding: 8px 0;
    border-top: 1px solid var(--el-border-color-lighter);
  }

  code {
    font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
    font-size: 14px;
    line-height: 1.5;
    word-break: break-all;
  }
}

.hash-row {
  display: flex;
  gap: 8px;
  width: 100%;
}

@media (max-width: 960px) {
  .tool-grid {
    grid-template-columns: 1fr;
  }
}
</style>
