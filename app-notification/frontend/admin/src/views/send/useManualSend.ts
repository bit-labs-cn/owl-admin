import { computed, onBeforeUnmount, onMounted, reactive, ref, watch } from "vue";
import { message } from "@bit-labs.cn/owl-ui/utils/message";
import { sendAPI, typeAPI } from "../../api/notify";
import { KindNotice, KindTodo } from "../../constants";
import { hasRichContent, toRFC3339 } from "../../format";

export interface RecipientOption {
  id: string;
  name: string;
  username: string;
  email: string;
  phone: string;
  disabled: boolean;
}

export interface TypeOption {
  code: string;
  name: string;
  /** 类型自带的种类，发送页据此决定发通知还是待办。 */
  kind: number;
  /** 类型的默认渠道，发送页当作不可更改的渠道直接带上。 */
  defaultChannels: string[];
}

/** 发送页的接收人预估，来自 POST /notify/audience-estimate。 */
export interface AudienceEstimate {
  userCount: number;
  externalCount: number;
  total: number;
}

export interface AudienceOption {
  id: string;
  name: string;
  disabled: boolean;
  children?: AudienceOption[];
}

export interface ExternalRecipient {
  name: string;
  email: string;
  phone: string;
}

export type AudienceMode = "user" | "role" | "dept" | "group";

/** 立即发送，还是登记一条定时任务。 */
export type SendMode = "now" | "scheduled";

/** 延迟发送的快捷选项，值是从现在起算的分钟数。 */
export const delayPresets = [
  { label: "10 分钟后", minutes: 10 },
  { label: "30 分钟后", minutes: 30 },
  { label: "1 小时后", minutes: 60 },
  { label: "1 天后", minutes: 60 * 24 }
];

function pad(value: number) {
  return String(value).padStart(2, "0");
}

/** 把 Date 转成日期选择器用的 `YYYY-MM-DD HH:mm:ss` */
export function toDateTimeText(date: Date) {
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(
    date.getHours()
  )}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`;
}

export function useManualSend() {
  const formRef = ref();
  const sending = ref(false);
  const recipients = ref<RecipientOption[]>([]);
  const recipientLoading = ref(false);
  const typeOptions = ref<TypeOption[]>([]);
  const audienceMode = ref<AudienceMode>("user");

  const audience = reactive({
    roles: [] as AudienceOption[],
    depts: [] as AudienceOption[],
    groups: [] as AudienceOption[]
  });

  const form = reactive({
    typeCode: "",
    title: "",
    content: "",
    link: "",
    // 这里只存发送页额外勾选的渠道；消息类型自带的渠道由 typeChannels 只读带入
    channels: [] as string[],
    userIds: [] as string[],
    roleIds: [] as string[],
    deptIds: [] as string[],
    groupIds: [] as string[],
    includeDeptChild: true,
    externals: [] as ExternalRecipient[],
    dueAt: "",
    sendMode: "now" as SendMode,
    planAt: ""
  });

  const selectedType = computed(() => typeOptions.value.find(item => item.code === form.typeCode));

  /** 消息类型自带的渠道，发送页只读带入，用户不能取消。 */
  const typeChannels = computed(() => selectedType.value?.defaultChannels ?? []);

  /** 最终投递渠道：类型自带 + 额外勾选，去重后一起提交。 */
  const effectiveChannels = computed(() => [
    ...new Set([...typeChannels.value, ...form.channels])
  ]);

  /** 勾选变化时只保留额外渠道，类型自带的那些不允许被取消。 */
  function updateChannels(values: string[]) {
    const locked = new Set(typeChannels.value);
    form.channels = values.filter(value => !locked.has(value));
  }

  /** 种类由所选消息类型决定：待办类型就发待办，不选类型就发普通通知。 */
  const kind = computed(() => (selectedType.value?.kind === KindTodo ? KindTodo : KindNotice));

  // 换成非待办类型时清掉截止时间，避免留下不会再发的旧值。
  watch(kind, value => {
    if (value !== KindTodo) form.dueAt = "";
  });

  const rules = {
    title: [{ required: true, message: "请输入标题", trigger: "blur" }]
  };

  const estimate = ref<AudienceEstimate>({ userCount: 0, externalCount: 0, total: 0 });
  const estimateLoading = ref(false);
  const estimateError = ref("");

  /** 发送和预估共用同一份受众条件，避免两处拼参数拼出差异。 */
  function audiencePayload() {
    return {
      userIds: form.userIds,
      roleIds: form.roleIds,
      deptIds: form.deptIds,
      groupIds: form.groupIds,
      includeDeptChild: form.includeDeptChild,
      externals: form.externals.filter(item => item.email || item.phone)
    };
  }

  // 选人、选部门、改外部人员邮箱都会影响人数，合并成一次 300ms 的请求
  let estimateSeq = 0;
  let estimateTimer: ReturnType<typeof setTimeout> | undefined;

  async function refreshEstimate() {
    const seq = ++estimateSeq;
    estimateLoading.value = true;
    try {
      const res = await sendAPI.estimate(audiencePayload());
      // 选择变化很快时，只认最后一次请求的结果，避免旧结果覆盖新人数
      if (seq !== estimateSeq) return;
      estimate.value = (res?.data ?? { userCount: 0, externalCount: 0, total: 0 }) as AudienceEstimate;
      estimateError.value = "";
    } catch {
      // 预估失败不影响发送，但要让人知道这个数字没更新
      if (seq === estimateSeq) estimateError.value = "接收人预估失败，发送时会按实际条件解析";
    } finally {
      if (seq === estimateSeq) estimateLoading.value = false;
    }
  }

  watch(
    () => [
      form.userIds.join(","),
      form.roleIds.join(","),
      form.deptIds.join(","),
      form.groupIds.join(","),
      form.includeDeptChild,
      form.externals.map(item => `${item.name}|${item.email}|${item.phone}`).join(",")
    ],
    () => {
      if (estimateTimer) clearTimeout(estimateTimer);
      estimateTimer = setTimeout(refreshEstimate, 300);
    },
    { immediate: true }
  );

  onBeforeUnmount(() => {
    if (estimateTimer) clearTimeout(estimateTimer);
  });

  async function searchRecipients(keyword: string) {
    recipientLoading.value = true;
    try {
      const res = await sendAPI.recipients({ keyword: keyword || undefined });
      const list = (res?.data ?? []) as RecipientOption[];
      const selected = recipients.value.filter(item => form.userIds.includes(item.id));
      const merged = [...selected];
      for (const item of list) {
        if (!merged.some(exist => exist.id === item.id)) merged.push(item);
      }
      recipients.value = merged;
    } finally {
      recipientLoading.value = false;
    }
  }

  async function onRecipientSearch(keyword: string) {
    await searchRecipients(keyword);
  }

  function recipientLabel(id: string) {
    const found = recipients.value.find(item => item.id === id);
    if (!found) return `未知用户(${id})`;
    return found.username && found.name !== found.username ? `${found.name}（${found.username}）` : found.name;
  }

  function addExternal() {
    form.externals.push({ name: "", email: "", phone: "" });
  }

  function removeExternal(index: number) {
    form.externals.splice(index, 1);
  }

  /** 接收人选择弹窗确认后回填：站内用户 ID、外部人员，并把认到的用户并进名称缓存。 */
  function applyTargets(payload: {
    userIds: string[];
    users: RecipientOption[];
    externals: ExternalRecipient[];
  }) {
    form.userIds = payload.userIds;
    form.externals = payload.externals.map(item => ({ ...item }));
    for (const user of payload.users) {
      if (!recipients.value.some(item => item.id === user.id)) {
        recipients.value.push(user);
      }
    }
    for (const external of form.externals) {
      if (!external.name) external.name = external.email || external.phone;
    }
  }

  function audienceCount() {
    return (
      form.userIds.length +
      form.roleIds.length +
      form.deptIds.length +
      form.groupIds.length +
      form.externals.filter(item => item.email || item.phone).length
    );
  }

  /** 按「从现在起 N 分钟」填计划时间，给延迟发送用。 */
  function applyDelay(minutes: number) {
    form.planAt = toDateTimeText(new Date(Date.now() + minutes * 60 * 1000));
  }

  /** 明天上午 9 点，常见的预约场景。 */
  function applyTomorrowMorning() {
    const date = new Date();
    date.setDate(date.getDate() + 1);
    date.setHours(9, 0, 0, 0);
    form.planAt = toDateTimeText(date);
  }

  /** 计划时间不能选过去。 */
  function disablePastDate(date: Date) {
    return date.getTime() < Date.now() - 24 * 60 * 60 * 1000;
  }

  async function reset() {
    form.typeCode = "";
    form.title = "";
    form.content = "";
    form.link = "";
    form.channels = [];
    form.userIds = [];
    form.roleIds = [];
    form.deptIds = [];
    form.groupIds = [];
    form.includeDeptChild = true;
    form.externals = [];
    form.dueAt = "";
    form.sendMode = "now";
    form.planAt = "";
    audienceMode.value = "user";
    formRef.value?.clearValidate();
  }

  async function submit() {
    const valid = await formRef.value?.validate().catch(() => false);
    if (!valid) return;
    if (audienceCount() === 0) {
      message("请选择接收人或填写外部人员", { type: "warning" });
      return;
    }
    if (kind.value === KindTodo && !form.dueAt) {
      message("待办需要填写截止时间", { type: "warning" });
      return;
    }
    if (form.sendMode === "scheduled") {
      if (!form.planAt) {
        message("请填写定时发送时间", { type: "warning" });
        return;
      }
      if (new Date(form.planAt.replace(" ", "T")).getTime() <= Date.now()) {
        message("定时发送时间必须晚于当前时间", { type: "warning" });
        return;
      }
    }
    sending.value = true;
    try {
      await sendAPI.send({
        kind: kind.value,
        typeCode: form.typeCode || undefined,
        title: form.title,
        // 编辑器空内容会留下 <p><br></p>，这里当没有正文处理，避免存一段空 HTML
        content: hasRichContent(form.content) ? form.content : "",
        link: form.link,
        channels: effectiveChannels.value,
        ...audiencePayload(),
        dueAt: kind.value === KindTodo && form.dueAt ? toRFC3339(form.dueAt) : undefined,
        planAt: form.sendMode === "scheduled" && form.planAt ? toRFC3339(form.planAt) : undefined
      });
      message(form.sendMode === "scheduled" ? `已排入定时发送：${form.planAt}` : "已发送", { type: "success" });
      await reset();
    } finally {
      sending.value = false;
    }
  }

  onMounted(async () => {
    const [typeRes, audienceRes] = await Promise.all([typeAPI.options(), sendAPI.audience(), searchRecipients("")]);
    typeOptions.value = (typeRes?.data ?? []) as TypeOption[];
    const data = (audienceRes?.data ?? {}) as { roles?: AudienceOption[]; depts?: AudienceOption[]; groups?: AudienceOption[] };
    audience.roles = data.roles ?? [];
    audience.depts = data.depts ?? [];
    audience.groups = data.groups ?? [];
  });

  return {
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
    recipientLoading,
    typeOptions,
    audience,
    audienceMode,
    onRecipientSearch,
    recipientLabel,
    applyTargets,
    addExternal,
    removeExternal,
    audienceCount,
    applyDelay,
    applyTomorrowMorning,
    disablePastDate,
    reset,
    submit
  };
}
