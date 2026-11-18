<script setup lang="ts">
import { computed, ref } from "vue";
import { channelOptions, kindOptions, labelOf, tagOf } from "../../labels";
import { ChannelEmail, ChannelInApp, ChannelSMS, ChannelDingTalk, ChannelWeCom, KindTodo } from "../../constants";
import { hasRichContent, richToText } from "../../format";
import RichContentView from "../../components/RichContentView.vue";

interface PreviewForm {
  title: string;
  content: string;
  link: string;
  dueAt: string;
  sendMode?: string;
  planAt?: string;
}

const props = defineProps<{
  form: PreviewForm;
  /** 最终投递渠道：类型自带 + 额外勾选，由发送页算好传进来。 */
  channels: string[];
  /** 种类由所选消息类型决定，发送页算好后传进来。 */
  kind: number;
  typeName?: string;
  /** 预计接收人数，来自发送页的人数预估。 */
  recipientCount: number;
}>();

const activeChannel = ref<string>(ChannelInApp);
const viewMode = ref<"card" | "detail">("card");

const selectedChannels = computed(() => {
  const list = channelOptions.map(item => item.value as string).filter(code => props.channels.includes(code));
  return list.length > 0 ? list : [ChannelInApp];
});

const currentChannel = computed(() =>
  selectedChannels.value.includes(activeChannel.value) ? activeChannel.value : selectedChannels.value[0]
);

const notConnected = computed(() => currentChannel.value === ChannelSMS || currentChannel.value === ChannelDingTalk);

const kindLabel = computed(() => labelOf(kindOptions, props.kind));
const kindTag = computed(() => tagOf(kindOptions, props.kind));

const titleText = computed(() => props.form.title || "未填写标题");
/** 富文本正文直接按 HTML 预览；企业微信和短信只吃文字，这里转成纯文本。 */
const hasContent = computed(() => hasRichContent(props.form.content));
const contentText = computed(() => richToText(props.form.content));
const recipientText = computed(() => (props.recipientCount > 0 ? `发给 ${props.recipientCount} 位接收人` : "尚未选择接收人"));

const channelNote = computed(() => {
  if (notConnected.value) {
    return `${labelOf(channelOptions, currentChannel.value)}暂未接入发送器，实际发送记录会标记为跳过。`;
  }
  if (currentChannel.value === ChannelWeCom) {
    return "发到企业微信群机器人，整个群共用一条消息。";
  }
  if (currentChannel.value === ChannelEmail) {
    return "站内用户使用账号邮箱；外部人员使用填写的邮箱。";
  }
  return "写入收件箱并实时推送到在线浏览器。";
});
</script>

<template>
  <div class="preview-body">
    <div class="channel-tabs">
      <button
        v-for="code in selectedChannels"
        :key="code"
        type="button"
        class="channel-tab"
        :class="{ 'is-active': code === currentChannel }"
        @click="activeChannel = code"
      >
        {{ labelOf(channelOptions, code) }}
      </button>
    </div>

    <div v-if="currentChannel === ChannelInApp" class="view-toggle">
      <button type="button" :class="{ 'is-active': viewMode === 'card' }" @click="viewMode = 'card'">小卡片</button>
      <button type="button" :class="{ 'is-active': viewMode === 'detail' }" @click="viewMode = 'detail'">详情</button>
    </div>

    <div v-if="notConnected" class="device-note">{{ channelNote }}</div>

    <div class="preview-stage">
      <template v-if="currentChannel === ChannelInApp">
        <div v-if="viewMode === 'card'" class="inapp-card">
          <span class="bell">铃</span>
          <div class="inapp-card-main">
            <div class="inapp-title">{{ titleText }}</div>
            <div class="meta">
              <span>{{ kindLabel }}</span>
              <span>{{ typeName || "未分类" }}</span>
              <span>刚刚</span>
            </div>
          </div>
        </div>

        <div v-else class="inapp-detail">
          <div class="inapp-head">
            <span class="inapp-title">{{ titleText }}</span>
            <el-tag size="small" :type="kindTag">{{ kindLabel }}</el-tag>
          </div>
          <div class="meta">
            <span>{{ typeName || "未分类" }}</span>
            <span v-if="form.sendMode === 'scheduled'">预约 {{ form.planAt || "未选时间" }}</span>
            <span v-else-if="kind === KindTodo && form.dueAt">截止 {{ form.dueAt }}</span>
            <span v-else-if="kind === KindTodo">截止时间未填</span>
            <span v-else>刚刚</span>
          </div>
          <RichContentView
            v-if="hasContent"
            class="inapp-body"
            :content="form.content"
            :keep-line-break="false"
          />
          <div v-else class="inapp-body is-empty">正文会在这里实时预览，左侧输入后立即更新。</div>
          <div v-if="form.link" class="inapp-link">查看详情 &rsaquo;</div>
        </div>
      </template>

      <div v-else-if="currentChannel === ChannelEmail" class="mail-window">
        <div class="mail-bar">
          <span class="mail-dot" />
          <span class="mail-dot" />
          <span class="mail-dot" />
          <span class="mail-subject">{{ titleText }}</span>
        </div>
        <div class="mail-head">
          <span class="mail-from">通知中心</span>
          <span class="mail-to">{{ recipientText }}</span>
        </div>
        <div class="mail-body">
          <div class="mail-h1">{{ titleText }}</div>
          <RichContentView
            v-if="hasContent"
            class="mail-text"
            :content="form.content"
            :keep-line-break="false"
          />
          <div v-else class="mail-text is-empty">正文会在这里实时预览，左侧输入后立即更新。</div>
          <div v-if="form.link" class="mail-button">前往处理</div>
        </div>
      </div>

      <div v-else-if="currentChannel === ChannelWeCom" class="wecom-card">
        <div class="wecom-head">
          <span class="wecom-badge">企业微信</span>
          <span class="wecom-chat">群机器人</span>
        </div>
        <div class="wecom-body">
          <div class="wecom-h1">{{ titleText }}</div>
          <div class="wecom-text">{{ contentText || "正文会在这里实时预览。" }}</div>
          <div v-if="form.link" class="wecom-link">点击前往处理</div>
        </div>
      </div>

      <div v-else class="sms-bubble">
        <div class="sms-text">{{ titleText }}</div>
        <div class="sms-text">{{ contentText || "正文会在这里实时预览。" }}</div>
        <div v-if="form.link" class="sms-link">{{ form.link }}</div>
      </div>
    </div>

    <div class="safety">
      <span class="safety-icon">!</span>
      <span>
        <b>发送前请再核对</b><br />
        标题、链接和接收范围确认无误后再发送。站内信写入后无法撤回；外部渠道失败可在「发送记录」里重试。
      </span>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.preview-body {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.channel-tabs {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.channel-tab {
  padding: 4px 12px;
  border: 1px solid var(--el-border-color);
  border-radius: 999px;
  background: transparent;
  color: var(--el-text-color-regular);
  font-size: 12px;
  cursor: pointer;
  transition: all 0.15s ease;

  &:hover {
    border-color: var(--el-color-primary-light-5);
    color: var(--el-color-primary);
  }

  &.is-active {
    border-color: var(--el-color-primary);
    background: var(--el-color-primary-light-9);
    color: var(--el-color-primary);
    font-weight: 600;
  }
}

.view-toggle {
  display: inline-flex;
  gap: 4px;
  padding: 3px;
  border-radius: 8px;
  background: var(--el-fill-color-light);
  width: max-content;

  button {
    height: 26px;
    padding: 0 12px;
    border: 0;
    border-radius: 6px;
    background: transparent;
    color: var(--el-text-color-secondary);
    font-size: 12px;
    cursor: pointer;

    &.is-active {
      background: var(--el-bg-color);
      color: var(--el-color-primary);
      font-weight: 600;
      box-shadow: 0 1px 4px rgb(16 24 40 / 10%);
    }
  }
}

.device-note {
  padding: 6px 10px;
  border-radius: 6px;
  background: var(--el-color-warning-light-9);
  color: var(--el-color-warning-dark-2);
  font-size: 12px;
  line-height: 1.5;
}

.preview-stage {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.inapp-card {
  display: flex;
  align-items: flex-start;
  gap: 11px;
  padding: 14px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 10px;
  background: var(--el-fill-color-blank);
}

.bell {
  display: grid;
  place-items: center;
  flex: 0 0 auto;
  width: 34px;
  height: 34px;
  border-radius: 9px;
  background: var(--el-color-primary-light-9);
  color: var(--el-color-primary);
  font-size: 13px;
  font-weight: 700;
}

.inapp-card-main {
  min-width: 0;
}

.inapp-detail {
  padding: 16px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 10px;
  background: var(--el-fill-color-blank);
}

.inapp-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 10px;
}

.inapp-title {
  font-size: 15px;
  font-weight: 600;
  line-height: 1.4;
  word-break: break-word;
}

.meta {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 5px;
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.inapp-body {
  margin-top: 12px;
  padding: 12px;
  border-radius: 8px;
  background: var(--el-fill-color-light);
  color: var(--el-text-color-primary);
  font-size: 13px;
  line-height: 1.7;
  word-break: break-word;

  &.is-empty {
    color: var(--el-text-color-secondary);
  }
}

.inapp-link {
  margin-top: 12px;
  padding-top: 10px;
  border-top: 1px dashed var(--el-border-color-lighter);
  color: var(--el-color-primary);
  font-size: 13px;
  font-weight: 600;
}

.mail-window {
  overflow: hidden;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 10px;
  background: var(--el-fill-color-blank);
}

.mail-bar {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 12px;
  border-bottom: 1px solid var(--el-border-color-lighter);
  background: var(--el-fill-color-light);
}

.mail-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--el-border-color);
}

.mail-subject {
  margin-left: 8px;
  overflow: hidden;
  color: var(--el-text-color-regular);
  font-size: 12px;
  font-weight: 600;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.mail-head {
  display: flex;
  justify-content: space-between;
  padding: 10px 14px;
  border-bottom: 1px solid var(--el-border-color-lighter);
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.mail-from {
  color: var(--el-text-color-primary);
  font-weight: 600;
}

.mail-body {
  padding: 16px 14px;
}

.mail-h1 {
  margin-bottom: 10px;
  font-size: 16px;
  font-weight: 700;
}

.mail-text {
  color: var(--el-text-color-regular);
  font-size: 13px;
  line-height: 1.75;
  word-break: break-word;

  &.is-empty {
    color: var(--el-text-color-secondary);
  }
}

.mail-button {
  display: inline-block;
  margin-top: 14px;
  padding: 6px 14px;
  border-radius: 6px;
  background: var(--el-color-primary);
  color: var(--el-color-white);
  font-size: 12px;
}

.wecom-card {
  overflow: hidden;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 10px;
  background: var(--el-fill-color-blank);
}

.wecom-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  background: var(--el-color-primary);
  color: var(--el-color-white);
  font-size: 12px;
}

.wecom-badge {
  font-weight: 600;
}

.wecom-chat {
  opacity: 0.85;
}

.wecom-body {
  padding: 14px;
}

.wecom-h1 {
  margin-bottom: 8px;
  font-size: 15px;
  font-weight: 700;
}

.wecom-text {
  color: var(--el-text-color-regular);
  font-size: 13px;
  line-height: 1.7;
  white-space: pre-wrap;
  word-break: break-word;
}

.wecom-link {
  margin-top: 10px;
  color: var(--el-color-primary);
  font-size: 13px;
}

.sms-bubble {
  max-width: 86%;
  padding: 12px 14px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 12px 12px 12px 2px;
  background: var(--el-fill-color-light);
}

.sms-text {
  color: var(--el-text-color-primary);
  font-size: 13px;
  line-height: 1.7;
  white-space: pre-wrap;
  word-break: break-word;

  & + & {
    margin-top: 6px;
  }
}

.sms-link {
  margin-top: 8px;
  color: var(--el-color-primary);
  font-size: 12px;
  word-break: break-all;
}

.safety {
  display: flex;
  gap: 9px;
  padding: 11px 12px;
  border: 1px solid var(--el-color-warning-light-7);
  border-radius: 9px;
  background: var(--el-color-warning-light-9);
  color: var(--el-color-warning-dark-2);
  font-size: 12px;
  line-height: 1.55;
}

.safety-icon {
  display: grid;
  place-items: center;
  flex: 0 0 auto;
  width: 18px;
  height: 18px;
  border-radius: 50%;
  background: var(--el-color-warning);
  color: var(--el-color-white);
  font-size: 11px;
  font-weight: 700;
}
</style>
