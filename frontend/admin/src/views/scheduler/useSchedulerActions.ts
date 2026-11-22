import { h, ref } from "vue";
import { ElMessageBox } from "element-plus";
import { dialogService } from "@bit-labs.cn/owl-ui/components/ReDialog";
import { message } from "@bit-labs.cn/owl-ui/utils/message";
import {
  cleanupSchedulerRuns,
  createSchedulerJob,
  deleteSchedulerJob,
  getSchedulerJobDetail,
  retrySchedulerRun,
  cancelSchedulerRun,
  triggerSchedulerJob,
  updateSchedulerJob
} from "@bit-labs.cn/owl-admin-ui/api/scheduler";
import JobForm from "./JobForm.vue";
import JobDetail from "./JobDetail.vue";
import RunDetail from "./RunDetail.vue";
import TriggerDialog from "./TriggerDialog.vue";
import CleanupForm from "./CleanupForm.vue";
import {
  SchedulerOverlap,
  SchedulerRunStatus,
  SchedulerScriptType
} from "./constants";
import type {
  SchedulerJobDetail,
  SchedulerJobFormData,
  SchedulerJobRow,
  SchedulerRunRow
} from "./types";

const DIALOG_OPTIONS = {
  draggable: true,
  alignCenter: true,
  closeOnClickModal: false,
  closeOnPressEscape: false
};

function warnText(err: unknown, fallback: string): string {
  return err instanceof Error ? err.message : fallback;
}

function safeParse(raw?: string): Record<string, any> {
  if (!raw) return {};
  try {
    const parsed = JSON.parse(raw);
    return parsed && typeof parsed === "object" ? parsed : {};
  } catch {
    return {};
  }
}

function emptyForm(): SchedulerJobFormData {
  return {
    name: "",
    description: "",
    cron: "",
    timezone: "",
    timeoutSeconds: 60,
    maxAttempts: 1,
    retryBackoffSeconds: 60,
    overlapPolicy: SchedulerOverlap.Skip,
    scriptType: SchedulerScriptType.Action,
    actionCode: "",
    actionParams: {},
    httpMethod: "POST",
    httpURL: "",
    httpHeaders: "",
    httpBody: "",
    httpContentType: "application/json",
    httpSuccessStatus: "",
    shellCommand: "",
    shellWorkingDir: ""
  };
}

/** 把详情里的脚本配置还原成表单字段。 */
function formFromDetail(detail: SchedulerJobDetail): SchedulerJobFormData {
  const cfg = safeParse(detail.scriptConfig);
  const base = emptyForm();
  const form: SchedulerJobFormData = {
    ...base,
    id: detail.id,
    name: detail.name ?? "",
    description: detail.description ?? "",
    cron: detail.cron ?? "",
    timezone: detail.timezone ?? "",
    timeoutSeconds: detail.timeoutSeconds || 60,
    maxAttempts: detail.maxAttempts || 1,
    retryBackoffSeconds: detail.retryBackoffSeconds ?? 60,
    overlapPolicy: detail.overlapPolicy || SchedulerOverlap.Skip,
    scriptType: detail.scriptType || SchedulerScriptType.Action
  };

  if (detail.scriptType === SchedulerScriptType.Action) {
    form.actionCode = detail.actionCode || cfg.action || "";
    form.actionParams = {};
    const stored = cfg.params;
    if (stored && typeof stored === "object") {
      for (const [key, value] of Object.entries(stored)) {
        form.actionParams[key] = value == null ? "" : String(value);
      }
    }
    return form;
  }

  if (detail.scriptType === SchedulerScriptType.Shell) {
    form.shellCommand = cfg.command ?? "";
    form.shellWorkingDir = cfg.workingDir ?? "";
  } else {
    form.httpMethod = (cfg.method ?? "POST").toUpperCase();
    form.httpURL = cfg.url ?? "";
    form.httpBody = cfg.body ?? "";
    form.httpContentType = cfg.contentType ?? "application/json";
    form.httpHeaders =
      cfg.headers && Object.keys(cfg.headers).length
        ? JSON.stringify(cfg.headers, null, 2)
        : "";
    form.httpSuccessStatus = Array.isArray(cfg.successStatus)
      ? cfg.successStatus.join(",")
      : "";
  }
  return form;
}

/** 打开任务详情。 */
export function openJobDetail(row: SchedulerJobRow) {
  dialogService.open({
    ...DIALOG_OPTIONS,
    title: "定时任务详情",
    width: "760px",
    hideFooter: true,
    contentRenderer: () => h(JobDetail, { row })
  });
}

/**
 * 新建或修改自定义任务。
 * 修改时先拉取详情（含脚本配置），再回填表单。
 */
export async function openJobForm(row: SchedulerJobRow | null, onDone: () => void) {
  let formInline = emptyForm();

  if (row) {
    try {
      const res = await getSchedulerJobDetail(row.id);
      const detail = (res?.data ?? null) as unknown as SchedulerJobDetail | null;
      if (!detail) {
        message("任务详情为空，请刷新后重试", { type: "error" });
        return;
      }
      formInline = formFromDetail(detail);
    } catch (err) {
      message(warnText(err, "任务详情加载失败"), { type: "error" });
      return;
    }
  }

  const jobFormRef = ref();
  dialogService.open({
    ...DIALOG_OPTIONS,
    title: row ? "修改定时任务" : "新建定时任务",
    width: "720px",
    props: { formInline },
    contentRenderer: ({ options }) =>
      h(JobForm, { ref: jobFormRef, formInline: options.props.formInline }),
    beforeSure: done => {
      const formEl = jobFormRef.value?.getRef();
      formEl?.validate(async (valid: boolean) => {
        if (!valid) return;
        const payload = jobFormRef.value?.getPayload();
        if (!payload.ok) {
          message(payload.message, { type: "error" });
          return;
        }
        try {
          if (row) {
            await updateSchedulerJob(row.id, payload.data);
            message("任务已更新", { type: "success" });
          } else {
            await createSchedulerJob(payload.data);
            message("任务已创建", { type: "success" });
          }
          done();
          onDone();
        } catch (err) {
          message(warnText(err, "保存失败"), { type: "error" });
        }
      });
    }
  });
}

/** 删除自定义任务，二次确认。 */
export async function confirmDeleteJob(row: SchedulerJobRow, onDone: () => void) {
  try {
    await ElMessageBox.confirm(
      `确认删除任务「${row.name}」吗？历史执行记录会保留，删除后不可恢复。`,
      "删除定时任务",
      {
        confirmButtonText: "删除",
        cancelButtonText: "取消",
        type: "warning",
        draggable: true,
        closeOnClickModal: false,
        closeOnPressEscape: false
      }
    );
  } catch {
    return;
  }

  try {
    await deleteSchedulerJob(row.id);
    message("任务已删除", { type: "success" });
    onDone();
  } catch (err) {
    message(warnText(err, "删除失败"), { type: "error" });
  }
}

/** 手动触发任务，可填写 JSON 参数。 */
export function openTriggerDialog(row: SchedulerJobRow, onDone: () => void) {
  const triggerRef = ref();
  dialogService.open({
    ...DIALOG_OPTIONS,
    title: "手动触发任务",
    width: "560px",
    contentRenderer: () => h(TriggerDialog, { ref: triggerRef, job: row }),
    beforeSure: done => {
      const payload = triggerRef.value?.getPayload();
      if (!payload.ok) {
        message(payload.message, { type: "error" });
        return;
      }
      triggerSchedulerJob(row.id, payload.payload)
        .then(res => {
          done();
          message(res?.msg || "已提交执行", { type: "success" });
          onDone();
        })
        .catch(err => message(warnText(err, "触发失败"), { type: "error" }));
    }
  });
}

/** 打开执行记录详情。 */
export function openRunDetail(row: SchedulerRunRow) {
  dialogService.open({
    ...DIALOG_OPTIONS,
    title: "执行记录详情",
    width: "820px",
    hideFooter: true,
    contentRenderer: () => h(RunDetail, { row })
  });
}

/** 重试失败的执行记录。 */
export async function confirmRetryRun(row: SchedulerRunRow, onDone: () => void) {
  try {
    await ElMessageBox.confirm(
      `确认重试任务「${row.jobName || row.jobCode}」的这次执行吗？会生成一条新的执行记录。`,
      "重试执行",
      {
        confirmButtonText: "重试",
        cancelButtonText: "取消",
        type: "warning",
        draggable: true,
        closeOnClickModal: false,
        closeOnPressEscape: false
      }
    );
  } catch {
    return;
  }

  try {
    const res = await retrySchedulerRun(row.id);
    message(res?.msg || "已提交重试", { type: "success" });
    onDone();
  } catch (err) {
    message(warnText(err, "重试失败"), { type: "error" });
  }
}

/** 取消待执行（或本实例执行中）的记录。 */
export async function confirmCancelRun(row: SchedulerRunRow, onDone: () => void) {
  const running = row.status === SchedulerRunStatus.Running;
  try {
    await ElMessageBox.confirm(
      running
        ? "该执行正在本实例上运行，取消会中断任务，确认继续吗？"
        : "确认取消这条待执行记录吗？",
      "取消执行",
      {
        confirmButtonText: "取消执行",
        cancelButtonText: "返回",
        type: "warning",
        draggable: true,
        closeOnClickModal: false,
        closeOnPressEscape: false
      }
    );
  } catch {
    return;
  }

  try {
    await cancelSchedulerRun(row.id);
    message("已取消", { type: "success" });
    onDone();
  } catch (err) {
    message(warnText(err, "取消失败"), { type: "error" });
  }
}

/** 清理历史执行记录。 */
export function openCleanupRuns(onDone: () => void) {
  const cleanupRef = ref();
  dialogService.open({
    ...DIALOG_OPTIONS,
    title: "清理执行记录",
    width: "520px",
    props: { formInline: { olderThanDays: 30 } },
    contentRenderer: ({ options }) =>
      h(CleanupForm, { ref: cleanupRef, formInline: options.props.formInline }),
    beforeSure: done => {
      const formEl = cleanupRef.value?.getRef();
      formEl?.validate(async (valid: boolean) => {
        if (!valid) return;
        const data = cleanupRef.value?.getFormData() as { olderThanDays: number };
        try {
          await ElMessageBox.confirm(
            `将删除 ${data.olderThanDays} 天之前已结束的执行记录，删除后不可恢复，确认继续吗？`,
            "确认清理",
            {
              confirmButtonText: "清理",
              cancelButtonText: "取消",
              type: "warning",
              draggable: true,
              closeOnClickModal: false,
              closeOnPressEscape: false
            }
          );
        } catch {
          return;
        }

        try {
          const res = await cleanupSchedulerRuns(data.olderThanDays);
          const deleted = (res?.data as { deleted?: number } | undefined)?.deleted ?? 0;
          done();
          message(`已清理 ${deleted} 条执行记录`, { type: "success" });
          onDone();
        } catch (err) {
          message(warnText(err, "清理失败"), { type: "error" });
        }
      });
    }
  });
}
