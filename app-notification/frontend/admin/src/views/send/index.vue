<script setup lang="ts">
import { computed, h, ref } from "vue";
import { message } from "@bit-labs.cn/owl-ui/utils/message";
import { useRenderIcon } from "@bit-labs.cn/owl-ui/components/ReIcon/src/hooks";
import {
  channelOptions,
  labelOf
} from "../../labels";
import {
  ChannelDingTalk,
  ChannelEmail,
  ChannelInApp,
  ChannelSMS,
  ChannelWeCom,
  KindTodo
} from "../../constants";
import { NOTIFY_PERMS } from "../../permissions";
import { dialogService } from "@bit-labs.cn/owl-ui/components/ReDialog";
import RichTextEditor from "../../components/RichTextEditor.vue";
import PreviewPanel from "./PreviewPanel.vue";
import RecipientPicker from "./RecipientPicker.vue";
import { delayPresets, useManualSend, type AudienceMode, type AudienceOption } from "./useManualSend";

defineOptions({ name: "NotifySend" });

const {
  formRef,
  form,
  rules,
  kind,
  typeChannels,
  effectiveChannels,
  updateChannels,
  estimate,
  estimateLoading,
  estimateError,
  sending,
  recipients,
  typeOptions,
  audience,
  audienceMode,
  recipientLabel,
  applyTargets,
  removeExternal,
  audienceCount,
  applyDelay,
  applyTomorrowMorning,
  disablePastDate,
  reset,
  submit
} = useManualSend();

const isTodo = computed(() => kind.value === KindTodo);
const isScheduled = computed(() => form.sendMode === "scheduled");
const selectedTypeName = computed(() => typeOptions.value.find(item => item.code === form.typeCode)?.name);

/** 种类和默认渠道都由消息类型决定，这里把结果说清楚。 */
const typeHint = computed(() => {
  if (isTodo.value) {
    return `「${selectedTypeName.value}」是待办类型，需要填写截止时间；右侧渠道按该类型的默认渠道带入，不能取消。`;
  }
  if (selectedTypeName.value) {
    return `「${selectedTypeName.value}」决定右侧的默认渠道（不能取消）；正文以本页填写为准，不套用类型模板。`;
  }
  return "不选消息类型时按普通通知发送，渠道全部由本页勾选；正文以本页填写为准。";
});

const audienceModes: { value: AudienceMode; label: string }[] = [
  { value: "user", label: "指定用户" },
  { value: "role", label: "按角色" },
  { value: "dept", label: "按部门" },
  { value: "group", label: "按用户组" }
];

const channelRows = [
  {
    value: ChannelInApp,
    label: "站内信",
    desc: "写入收件箱，并实时推送到在线浏览器",
    tag: "已接入",
    connected: true
  },
  {
    value: ChannelEmail,
    label: "邮件",
    desc: "站内用户用账号邮箱，外部人员用填写邮箱",
    tag: "已接入",
    connected: true
  },
  {
    value: ChannelWeCom,
    label: "企业微信",
    desc: "群机器人推送，整群共用一条消息",
    tag: "已接入",
    connected: true
  },
  {
    value: ChannelSMS,
    label: "短信",
    desc: "暂未接入发送器，选中后会记为跳过",
    tag: "未接入",
    connected: false
  },
  {
    value: ChannelDingTalk,
    label: "钉钉",
    desc: "暂未接入发送器，选中后会记为跳过",
    tag: "未接入",
    connected: false
  }
];

const totalText = computed(() => estimate.value.total.toLocaleString("zh-CN"));

const lockedChannels = computed(() => new Set(typeChannels.value));

/** 渠道最终用哪几个：类型自带的不可取消，再加上额外勾选的。 */
const channelSummary = computed(() => {
  const count = effectiveChannels.value.length;
  if (count === 0) return "站内信";
  if (typeChannels.value.length > 0 && effectiveChannels.value.length === typeChannels.value.length) {
    return `类型默认 ${count} 个`;
  }
  return `${count} 个`;
});

const audienceReady = computed(() => estimate.value.total > 0);
const contentReady = computed(() => form.title.trim() !== "");
// 渠道可以只靠类型自带的默认值，没有任何渠道时才把这项算作没完成
const strategyReady = computed(() => effectiveChannels.value.length > 0);
/** 三项都齐了才允许发送，这里折成百分比给投递策略卡看。 */
const readyPercent = computed(() => {
  const done = [audienceReady.value, contentReady.value, strategyReady.value].filter(Boolean).length;
  return Math.round((done / 3) * 100);
});

const audienceHintText = computed(() => {
  if (estimateError.value) return estimateError.value;
  if (estimateLoading.value) return "正在估算接收人数…";
  if (estimate.value.total > 0) return "多条件已按人合并去重，发送当刻再按组织关系解析";
  if (form.externals.some(item => !item.email && item.phone)) return "外部人员要填邮箱才会实际投递，只有手机号会被跳过";
  if (audienceCount() > 0) return "当前条件没有匹配到启用用户，换个条件试试";
  return "点「选择接收人」筛选并勾选站内用户，外部人员也在同一个弹窗里导入";
});

interface AudienceChip {
  key: string;
  label: string;
  remove: () => void;
}

function flattenDepts(nodes: AudienceOption[], out: AudienceOption[] = []) {
  for (const node of nodes) {
    out.push(node);
    if (node.children?.length) flattenDepts(node.children, out);
  }
  return out;
}

const deptFlat = computed(() => flattenDepts(audience.depts));

/** 接收人目标：站内用户 + 外部人员，显示在选择框里。 */
const targetChips = computed<AudienceChip[]>(() => {
  const chips: AudienceChip[] = [];

  for (const id of form.userIds) {
    chips.push({
      key: `user-${id}`,
      label: recipientLabel(id),
      remove: () => {
        form.userIds = form.userIds.filter(item => item !== id);
      }
    });
  }

  form.externals.forEach((item, index) => {
    if (!item.email && !item.phone) return;
    chips.push({
      key: `external-${index}`,
      label: `外部：${item.name || item.email || item.phone}`,
      remove: () => removeExternal(index)
    });
  });

  return chips;
});

/** 圈选条件：角色、部门、用户组，显示在汇总区。 */
const conditionChips = computed<AudienceChip[]>(() => {
  const chips: AudienceChip[] = [];

  for (const id of form.roleIds) {
    const name = audience.roles.find(item => item.id === id)?.name;
    chips.push({
      key: `role-${id}`,
      label: `角色：${name || `未知角色(${id})`}`,
      remove: () => {
        form.roleIds = form.roleIds.filter(item => item !== id);
      }
    });
  }

  for (const id of form.deptIds) {
    const name = deptFlat.value.find(item => item.id === id)?.name;
    chips.push({
      key: `dept-${id}`,
      label: `部门：${name || `未知部门(${id})`}`,
      remove: () => {
        form.deptIds = form.deptIds.filter(item => item !== id);
      }
    });
  }

  if (form.deptIds.length > 0 && form.includeDeptChild) {
    chips.push({
      key: "dept-child",
      label: "含下级部门",
      remove: () => {
        form.includeDeptChild = false;
      }
    });
  }

  for (const id of form.groupIds) {
    const name = audience.groups.find(item => item.id === id)?.name;
    chips.push({
      key: `group-${id}`,
      label: `用户组：${name || `未知用户组(${id})`}`,
      remove: () => {
        form.groupIds = form.groupIds.filter(item => item !== id);
      }
    });
  }

  return chips;
});

/** 定时发送时间离现在多久，帮人确认自己选的是不是想要的那一刻。 */
const planHint = computed(() => {
  if (!form.planAt) return "";
  const target = new Date(form.planAt.replace(" ", "T")).getTime();
  if (Number.isNaN(target)) return "";
  const diff = target - Date.now();
  if (diff <= 0) return "这个时间已经过去了，请重新选择";
  const minutes = Math.round(diff / 60000);
  if (minutes < 60) return `将在约 ${minutes} 分钟后发送`;
  const hours = Math.round(minutes / 60);
  if (hours < 24) return `将在约 ${hours} 小时后发送`;
  return `将在约 ${Math.round(hours / 24)} 天后发送`;
});

const recipientPickerRef = ref();

/**
 * 接收人选择走弹窗：弹窗里能筛选、翻页勾选、导入外部人员，一套动作拿到最终目标。
 * 弹窗自己不做发送，确认后把选择结果回填到表单。
 */
function openRecipientPicker() {
  dialogService.open({
    title: "选择接收人",
    width: "960px",
    draggable: true,
    closeOnClickModal: false,
    closeOnPressEscape: false,
    alignCenter: true,
    appendToBody: true,
    props: {
      selectedUserIds: form.userIds,
      externals: form.externals,
      depts: audience.depts,
      knownUsers: recipients.value
    },
    contentRenderer: ({ options }) =>
      h(RecipientPicker, {
        ref: recipientPickerRef,
        selectedUserIds: options.props.selectedUserIds,
        externals: options.props.externals,
        depts: options.props.depts,
        knownUsers: options.props.knownUsers
      }),
    beforeSure: done => {
      const picker = recipientPickerRef.value;
      const error = picker?.validate?.() ?? "";
      if (error) {
        message(error, { type: "warning" });
        return;
      }
      applyTargets(picker.getSelection());
      done();
    }
  });
}
</script>

<template>
  <div class="main send-workbench">
    <div class="workbench-grid">
      <section class="stack">
        <el-card shadow="never" class="panel">
          <template #header>
            <div class="panel-head">
              <div class="panel-title"><span class="panel-icon">人</span>接收范围</div>
              <div class="hint"><span class="status-dot" />人数为实时预估</div>
            </div>
          </template>

          <div class="mode-tabs">
            <button
              v-for="item in audienceModes"
              :key="item.value"
              type="button"
              class="mode-tab"
              :class="{ 'is-active': audienceMode === item.value }"
              @click="audienceMode = item.value"
            >
              {{ item.label }}
            </button>
          </div>

          <div class="audience-body">
            <template v-if="audienceMode === 'user'">
              <div class="picker-box">
                <div class="picker-box-head">
                  <div class="picker-box-title">
                    <span class="picker-count">{{ form.userIds.length }}</span>
                    <span class="picker-unit">位站内用户</span>
                    <span v-if="form.externals.length > 0" class="picker-extra">+ {{ form.externals.length }} 位外部人员</span>
                  </div>
                  <div class="picker-box-actions">
                    <el-button type="primary" plain :icon="useRenderIcon('ri:user-add-line')" @click="openRecipientPicker">
                      选择接收人
                    </el-button>
                  </div>
                </div>
                <div v-if="targetChips.length > 0" class="chips picker-chips">
                  <span v-for="chip in targetChips" :key="chip.key" class="chip">
                    {{ chip.label }}
                    <button type="button" :aria-label="`移除 ${chip.label}`" @click="chip.remove()">×</button>
                  </span>
                </div>
                <p v-else class="picker-empty">还没有选择接收人，点「选择接收人」筛选、勾选或导入外部人员。</p>
              </div>
              <p class="form-note">已禁用的账号不可选；同一个人被多个条件命中时只发一条。</p>
            </template>

            <template v-else-if="audienceMode === 'role'">
              <el-select v-model="form.roleIds" multiple filterable collapse-tags class="w-full" placeholder="选择角色">
                <el-option
                  v-for="item in audience.roles"
                  :key="item.id"
                  :label="item.name"
                  :value="item.id"
                  :disabled="item.disabled"
                />
              </el-select>
              <p class="form-note">按角色圈人时，只发给角色下处于启用状态的用户。</p>
            </template>

            <template v-else-if="audienceMode === 'dept'">
              <el-tree-select
                v-model="form.deptIds"
                :data="audience.depts"
                node-key="id"
                multiple
                show-checkbox
                check-strictly
                collapse-tags
                collapse-tags-tooltip
                :props="{ label: 'name', children: 'children', disabled: 'disabled' }"
                class="w-full"
                placeholder="选择部门"
              />
              <el-checkbox v-model="form.includeDeptChild" class="include-child">同时发送给下级部门</el-checkbox>
            </template>

            <template v-else-if="audienceMode === 'group'">
              <el-select v-model="form.groupIds" multiple filterable collapse-tags class="w-full" placeholder="选择用户组">
                <el-option
                  v-for="item in audience.groups"
                  :key="item.id"
                  :label="item.name"
                  :value="item.id"
                  :disabled="item.disabled"
                />
              </el-select>
              <p class="form-note">用户组成员按发送当刻的名单解析，成员变动会影响实际收件人。</p>
            </template>
          </div>

          <div class="audience-summary">
            <div class="summary-left">
              <div class="count">{{ totalText }}</div>
              <div class="count-label">预计有效接收人（已自动去重）</div>
              <div v-if="conditionChips.length > 0" class="chips">
                <span v-for="chip in conditionChips" :key="chip.key" class="chip">
                  {{ chip.label }}
                  <button type="button" :aria-label="`移除 ${chip.label}`" @click="chip.remove()">×</button>
                </span>
              </div>
              <p class="form-note summary-note">{{ audienceHintText }}</p>
            </div>
            <div class="summary-meta">
              <span>站内用户 {{ estimate.userCount }} 人</span>
              <span>外部人员 {{ estimate.externalCount }} 人</span>
              <span>已自动排除停用账号</span>
              <b v-if="estimateLoading">正在估算…</b>
              <b v-else-if="estimate.total > 0">可以发送</b>
              <b v-else class="is-muted">尚未选择接收人</b>
            </div>
          </div>
        </el-card>

        <el-card shadow="never" class="panel">
          <template #header>
            <div class="panel-head">
              <div class="panel-title"><span class="panel-icon">文</span>通知内容</div>
              <div class="hint">内容将在右侧实时预览</div>
            </div>
          </template>

          <el-form ref="formRef" :model="form" :rules="rules" label-width="96px" class="content-form">
            <el-form-item label="消息类型" prop="typeCode">
              <el-select v-model="form.typeCode" clearable filterable placeholder="不选则按普通通知发送" class="w-full">
                <el-option v-for="item in typeOptions" :key="item.code" :label="item.name" :value="item.code">
                  <span>{{ item.name }}</span>
                  <span v-if="item.kind === KindTodo" class="option-side">待办</span>
                  <span v-else-if="item.defaultChannels.length > 0" class="option-side">
                    {{ item.defaultChannels.map(code => labelOf(channelOptions, code)).join("、") }}
                  </span>
                </el-option>
              </el-select>
              <p class="form-note">{{ typeHint }}</p>
            </el-form-item>

            <el-form-item label="标题" prop="title">
              <el-input v-model="form.title" maxlength="255" show-word-limit clearable placeholder="请输入标题" />
              <p class="form-note">接收人先在收件箱里看到标题，一句话说清要做什么。</p>
            </el-form-item>

            <el-form-item label="正文" prop="content">
              <RichTextEditor v-model="form.content" :height="260" />
              <p class="form-note">
                正文支持富文本，站内信和邮件按 HTML 显示；企业微信和短信会自动转成纯文本。图片走系统文件上传。
              </p>
            </el-form-item>

            <el-form-item label="跳转链接" prop="link">
              <el-input v-model="form.link" clearable placeholder="例如 /system/notify/inbox/index" />
              <p class="form-note">接收人点击通知后跳转到这个路径，留空则不显示入口。</p>
            </el-form-item>

            <el-form-item v-if="isTodo" label="截止时间" prop="dueAt">
              <el-date-picker
                v-model="form.dueAt"
                type="datetime"
                value-format="YYYY-MM-DD HH:mm:ss"
                placeholder="待办需要填写截止时间"
                class="w-full"
              />
              <p class="form-note">超过截止时间未完成的待办会被标记为已过期。</p>
            </el-form-item>
          </el-form>
        </el-card>
      </section>

      <aside class="stack">
        <el-card shadow="never" class="panel">
          <template #header>
            <div class="panel-head">
              <div class="panel-title"><span class="panel-icon">渠</span>发送渠道</div>
              <div class="hint">
                {{ typeChannels.length > 0 ? `类型自带 ${typeChannels.length} 个，可再补充` : "可多选" }}
              </div>
            </div>
          </template>

          <el-checkbox-group
            :model-value="effectiveChannels"
            class="channel-list"
            @update:model-value="updateChannels"
          >
            <el-checkbox
              v-for="item in channelRows"
              :key="item.value"
              :value="item.value"
              :disabled="lockedChannels.has(item.value)"
              class="channel"
            >
              <span class="channel-main">
                <span class="channel-name">{{ item.label }}</span>
                <span class="channel-desc">{{ item.desc }}</span>
              </span>
              <span v-if="lockedChannels.has(item.value)" class="channel-tag is-locked">类型默认</span>
              <span v-else class="channel-tag" :class="{ 'is-muted': !item.connected }">{{ item.tag }}</span>
            </el-checkbox>
          </el-checkbox-group>

          <p class="form-note">
            {{
              typeChannels.length > 0
                ? `「${selectedTypeName}」自带的渠道不能取消，可以在上面再勾选额外的渠道；都不勾也按该类型的渠道发送。`
                : "未选消息类型时可以勾选渠道；都不勾只发站内信。"
            }}
            短信与钉钉暂未接入发送器，选中后会记为跳过。
          </p>
        </el-card>

        <el-card shadow="never" class="panel">
          <template #header>
            <div class="panel-head">
              <div class="panel-title"><span class="panel-icon">时</span>投递策略</div>
            </div>
          </template>

          <el-radio-group v-model="form.sendMode" class="radio-row">
            <el-radio value="now" border class="radio-card">
              <strong>立即发送</strong>
              <small>确认后立刻进入投递流程</small>
            </el-radio>
            <el-radio value="scheduled" border class="radio-card">
              <strong>定时发送</strong>
              <small>到点由后台调度发送</small>
            </el-radio>
          </el-radio-group>

          <div v-if="isScheduled" class="plan-box">
            <el-date-picker
              v-model="form.planAt"
              type="datetime"
              value-format="YYYY-MM-DD HH:mm:ss"
              placeholder="选择到点发送的时间"
              :disabled-date="disablePastDate"
              class="w-full"
            />
            <div class="preset-row">
              <el-button
                v-for="item in delayPresets"
                :key="item.minutes"
                link
                type="primary"
                size="small"
                @click="applyDelay(item.minutes)"
              >
                {{ item.label }}
              </el-button>
              <el-button link type="primary" size="small" @click="applyTomorrowMorning">明天 9:00</el-button>
            </div>
            <p class="form-note">{{ planHint || "到点由后台调度发送，可在「定时发送」里查看、取消或立即发送。" }}</p>
          </div>

          <div class="batch-box">
            <div class="batch-row"><span>投递方式</span><b>{{ isScheduled ? "后台定时调度" : "站内信即时 + 外部渠道异步" }}</b></div>
            <div class="batch-row"><span>站内用户</span><b>{{ estimate.userCount }} 人</b></div>
            <div class="batch-row"><span>外部人员</span><b>{{ estimate.externalCount }} 人</b></div>
            <div class="batch-row"><span>失败处理</span><b>发送记录里可重试</b></div>
            <div class="ready-line">
              <span>准备完成度</span>
              <b>{{ readyPercent }}%</b>
            </div>
            <div class="progress"><span :style="{ width: `${readyPercent}%` }" /></div>
          </div>
        </el-card>

        <el-card shadow="never" class="panel">
          <template #header>
            <div class="panel-head">
              <div class="panel-title"><span class="panel-icon">预</span>实时预览</div>
              <div class="hint">随输入更新</div>
            </div>
          </template>
          <PreviewPanel
            :form="form"
            :channels="effectiveChannels"
            :kind="kind"
            :type-name="selectedTypeName"
            :recipient-count="estimate.total"
          />
        </el-card>
      </aside>
    </div>

    <footer class="send-footer">
      <div class="footer-left">
        <span>预计接收 <b>{{ totalText }} 人</b></span>
        <span>渠道 <b>{{ channelSummary }}</b></span>
        <span>方式 <b>{{ isScheduled ? `定时 · ${form.planAt || "待选时间"}` : "立即发送" }}</b></span>
      </div>
      <div class="footer-actions">
        <el-button :disabled="sending" @click="reset">清空</el-button>
        <Perms :value="NOTIFY_PERMS.send.send">
          <el-button
            type="primary"
            :loading="sending"
            :icon="useRenderIcon(isScheduled ? 'ri:timer-line' : 'ri:send-plane-line')"
            @click="submit"
          >
            {{ isScheduled ? "定时发送" : "立即发送" }}
          </el-button>
        </Perms>
      </div>
    </footer>
  </div>
</template>

<style lang="scss" scoped>
.send-workbench {
  // 底部发送条用 fixed 贴住视口，留出空间避免最后一屏内容被盖住
  padding-bottom: 92px;
}

.workbench-grid {
  display: grid;
  grid-template-columns: minmax(0, 1.65fr) minmax(340px, 0.75fr);
  gap: 16px;
  align-items: start;
}

.stack {
  display: grid;
  gap: 16px;
  min-width: 0;
}

.panel {
  border-color: var(--owl-border-color, rgba(0, 0, 0, 0.06));

  :deep(.el-card__header) {
    padding: 14px 18px;
  }

  :deep(.el-card__body) {
    padding: 18px;
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
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 15px;
  font-weight: 700;
  color: var(--el-text-color-primary);
}

.panel-icon {
  display: grid;
  place-items: center;
  width: 28px;
  height: 28px;
  border-radius: 8px;
  background: var(--el-color-primary-light-9);
  color: var(--el-color-primary);
  font-size: 13px;
  font-weight: 700;
}

.hint {
  display: inline-flex;
  align-items: center;
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.status-dot {
  width: 8px;
  height: 8px;
  margin-right: 6px;
  border-radius: 50%;
  background: var(--el-color-warning);
}

.mode-tabs {
  display: flex;
  gap: 6px;
  width: max-content;
  max-width: 100%;
  margin-bottom: 16px;
  padding: 4px;
  overflow: auto;
  border-radius: 9px;
  background: var(--el-fill-color-light);
}

.mode-tab {
  height: 32px;
  padding: 0 14px;
  border: 0;
  border-radius: 7px;
  background: transparent;
  color: var(--el-text-color-regular);
  font-size: 13px;
  white-space: nowrap;
  cursor: pointer;
  transition: all 0.15s ease;

  &:hover {
    color: var(--el-color-primary);
  }

  &.is-active {
    background: var(--el-bg-color);
    color: var(--el-color-primary);
    font-weight: 700;
    box-shadow: 0 1px 4px rgba(16, 24, 40, 0.1);
  }
}

.audience-body {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.include-child {
  margin-left: 0;
}

.picker-box {
  padding: 12px 14px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 9px;
  background: var(--el-fill-color-lighter);
}

.picker-box-head {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  align-items: center;
  justify-content: space-between;
}

.picker-box-title {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  align-items: baseline;
}

.picker-count {
  color: var(--el-color-primary);
  font-size: 20px;
  font-weight: 800;
  line-height: 1;
}

.picker-unit {
  color: var(--el-text-color-regular);
  font-size: 12px;
}

.picker-extra {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.picker-chips {
  margin-top: 10px;
  max-height: 96px;
  overflow-y: auto;
}

.picker-empty {
  margin: 8px 0 0;
  color: var(--el-text-color-secondary);
  font-size: 12px;
  line-height: 1.5;
}

.external-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1.2fr) minmax(0, 1fr) auto;
  gap: 8px;
  align-items: center;
}

.audience-summary {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 18px;
  align-items: center;
  margin-top: 16px;
  padding: 14px 16px;
  border: 1px solid var(--el-color-primary-light-8);
  border-radius: 10px;
  background: var(--el-color-primary-light-9);
}

.summary-left {
  min-width: 0;
}

.count {
  color: var(--el-color-primary);
  font-size: 25px;
  font-weight: 800;
  letter-spacing: -0.5px;
  line-height: 1.1;
}

.count-label {
  margin-top: 2px;
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.chips {
  display: flex;
  flex-wrap: wrap;
  gap: 7px;
  margin-top: 10px;
}

.chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 9px;
  border: 1px solid var(--el-color-primary-light-7);
  border-radius: 999px;
  background: var(--el-bg-color);
  color: var(--el-text-color-regular);
  font-size: 12px;

  button {
    padding: 0;
    border: 0;
    background: none;
    color: var(--el-text-color-secondary);
    font-size: 13px;
    line-height: 1;
    cursor: pointer;

    &:hover {
      color: var(--el-color-danger);
    }
  }
}

.summary-note {
  margin: 8px 0 0;
}

.summary-meta {
  display: grid;
  gap: 6px;
  color: var(--el-text-color-regular);
  font-size: 12px;
  text-align: right;

  b {
    color: var(--el-color-success);
  }

  b.is-muted {
    color: var(--el-text-color-secondary);
  }
}

.content-form {
  :deep(.el-form-item) {
    margin-bottom: 18px;
  }

  :deep(.el-form-item__label) {
    color: var(--el-text-color-primary);
    font-weight: 600;
  }
}

.form-note {
  margin: 6px 0 0;
  color: var(--el-text-color-secondary);
  font-size: 12px;
  line-height: 1.5;
}

.option-side {
  float: right;
  margin-left: 12px;
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.channel-list {
  display: grid;
  gap: 10px;
}

.channel {
  display: flex;
  align-items: flex-start;
  width: 100%;
  height: auto;
  margin-right: 0;
  padding: 11px 12px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 9px;
  transition: all 0.15s ease;

  &:hover {
    border-color: var(--el-color-primary-light-5);
  }

  // 类型自带的渠道勾选框被禁用，但要看起来是「已锁定」而不是「不可用」
  &.is-disabled {
    border-color: var(--el-color-primary-light-8);
    background: var(--el-color-primary-light-9);

    &:hover {
      border-color: var(--el-color-primary-light-8);
    }

    .channel-name {
      color: var(--el-text-color-primary);
    }

    .channel-desc {
      color: var(--el-text-color-secondary);
    }
  }

  :deep(.el-checkbox__input) {
    margin-top: 3px;
  }

  :deep(.el-checkbox__label) {
    display: flex;
    flex: 1;
    align-items: flex-start;
    gap: 8px;
    min-width: 0;
    padding-left: 10px;
    white-space: normal;
  }
}

.channel-main {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-width: 0;
}

.channel-name {
  color: var(--el-text-color-primary);
  font-size: 13px;
  font-weight: 600;
}

.channel-desc {
  margin-top: 2px;
  color: var(--el-text-color-secondary);
  font-size: 12px;
  line-height: 1.5;
}

.channel-tag {
  flex: 0 0 auto;
  padding: 2px 7px;
  border-radius: 5px;
  background: var(--el-color-success-light-9);
  color: var(--el-color-success);
  font-size: 11px;
  white-space: nowrap;

  &.is-locked {
    background: var(--el-color-primary-light-8);
    color: var(--el-color-primary);
  }

  &.is-muted {
    background: var(--el-fill-color-light);
    color: var(--el-text-color-secondary);
  }
}

.radio-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 9px;
  margin-bottom: 12px;
}

.radio-card {
  height: auto;
  margin-right: 0;
  padding: 11px 12px;
  align-items: flex-start;

  :deep(.el-radio__label) {
    display: flex;
    flex-direction: column;
    gap: 2px;
    font-size: 13px;
    line-height: 1.4;
  }

  strong {
    font-size: 13px;
    font-weight: 600;
  }

  small {
    color: var(--el-text-color-secondary);
    font-size: 12px;
  }
}

.plan-box {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-bottom: 12px;
}

.preset-row {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.batch-box {
  display: grid;
  gap: 8px;
  padding: 13px;
  border-radius: 9px;
  background: var(--el-fill-color-light);
}

.batch-row,
.ready-line {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  color: var(--el-text-color-secondary);
  font-size: 12px;

  b {
    color: var(--el-text-color-primary);
    text-align: right;
  }
}

.progress {
  height: 6px;
  overflow: hidden;
  border-radius: 99px;
  background: var(--el-border-color-lighter);

  span {
    display: block;
    height: 100%;
    border-radius: 99px;
    background: linear-gradient(90deg, var(--el-color-primary), var(--el-color-primary-light-3));
    transition: width 0.6s;
  }
}

// 发送条固定贴住视口底部，不随内容滚动。
// 页面滚在框架的 .el-scrollbar 里，sticky 会被其中 overflow:hidden 的祖先限制成随内容走，
// 所以这里用 fixed；底部不留缝隙，左侧按侧边栏宽度让出空间，避免盖住菜单。
.send-footer {
  position: fixed;
  right: 0;
  bottom: 0;
  left: 210px;
  z-index: 900;
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  align-items: center;
  justify-content: space-between;
  padding: 12px 24px;
  border-top: 1px solid var(--owl-border-color, rgba(0, 0, 0, 0.06));
  background: var(--el-bg-color);
  box-shadow: 0 -6px 24px -18px rgba(16, 24, 40, 0.4);
  backdrop-filter: blur(8px);
}

// 侧边栏折叠后只剩 54px 轨道
:global(.app-wrapper.hideSidebar) .send-footer {
  left: 54px;
}

// 横向布局没有左侧栏，底部条贴满宽度
:global(body[layout="horizontal"]) .send-footer {
  left: 0;
}

@media (max-width: 760px) {
  .send-footer {
    right: 0;
    bottom: 0;
    left: 0;
    padding: 12px 16px;
  }
}

.footer-left {
  display: flex;
  flex-wrap: wrap;
  gap: 22px;
  color: var(--el-text-color-secondary);
  font-size: 12px;

  b {
    color: var(--el-text-color-primary);
  }
}

.footer-actions {
  display: flex;
  gap: 9px;
}

@media (max-width: 1180px) {
  .workbench-grid {
    grid-template-columns: minmax(0, 1fr);
  }
}

@media (max-width: 700px) {
  .footer-actions {
    width: 100%;
  }

  .footer-actions :deep(.el-button) {
    flex: 1;
  }

  .audience-summary {
    grid-template-columns: minmax(0, 1fr);
  }

  .summary-meta {
    text-align: left;
  }

  .radio-row {
    grid-template-columns: minmax(0, 1fr);
  }

  .send-footer {
    flex-direction: column;
    align-items: stretch;
  }

  .send-workbench {
    // 窄屏底部条会折成两行，留更高的空间
    padding-bottom: 152px;
  }

  .footer-left {
    justify-content: space-between;
    gap: 12px;
  }
}
</style>
