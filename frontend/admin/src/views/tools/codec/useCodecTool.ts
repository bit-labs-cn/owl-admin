import { ref } from "vue";
import { ElMessage } from "element-plus";
import dayjs, { type Dayjs } from "dayjs";
import { md5 } from "./md5";

export const codecTools = [
  { id: "json", label: "JSON" },
  { id: "base64", label: "Base64" },
  { id: "url", label: "URL" },
  { id: "time", label: "时间戳" },
  { id: "digest", label: "摘要" },
  { id: "uuid", label: "UUID" }
] as const;

export type CodecToolId = (typeof codecTools)[number]["id"];

export function useCodecTool() {
  const active = ref<CodecToolId>("json");
  const jsonInput = ref("");
  const jsonOutput = ref("");
  const jsonError = ref("");
  const base64Input = ref("");
  const base64Output = ref("");
  const base64Error = ref("");
  const urlInput = ref("");
  const urlOutput = ref("");
  const urlError = ref("");
  const stamp = ref("");
  const localTime = ref("");
  const secondsText = ref("");
  const millisText = ref("");
  const timeError = ref("");
  const digestInput = ref("");
  const md5Text = ref("");
  const shaText = ref("");
  const digestError = ref("");
  const uuids = ref<string[]>([]);

  function formatJson(space?: number) {
    try {
      jsonOutput.value = JSON.stringify(JSON.parse(jsonInput.value), null, space);
      jsonError.value = "";
    } catch {
      jsonError.value = "JSON 格式不正确";
    }
  }

  function encodeBase64() {
    base64Output.value = bytesToBase64(new TextEncoder().encode(base64Input.value));
    base64Error.value = "";
  }

  function decodeBase64() {
    try {
      const binary = atob(base64Input.value.replace(/\s/g, ""));
      const bytes = Uint8Array.from(binary, char => char.charCodeAt(0));
      base64Output.value = new TextDecoder().decode(bytes);
      base64Error.value = "";
    } catch {
      base64Error.value = "不是有效的 Base64";
    }
  }

  function encodeUrl() {
    urlOutput.value = encodeURIComponent(urlInput.value);
    urlError.value = "";
  }

  function decodeUrl() {
    try {
      urlOutput.value = decodeURIComponent(urlInput.value);
      urlError.value = "";
    } catch {
      urlError.value = "不是有效的 URL 编码";
    }
  }

  function fromStamp() {
    const raw = stamp.value.trim();
    if (!raw) {
      timeError.value = "请输入时间戳";
      return;
    }
    const value = Number(raw);
    if (!Number.isFinite(value)) {
      timeError.value = "时间戳不是数字";
      return;
    }
    const millis = Math.abs(value) < 1e11 ? value * 1000 : value;
    const parsed = dayjs(millis);
    if (!parsed.isValid()) {
      timeError.value = "无法解析这个时间戳";
      return;
    }
    fillTime(parsed);
  }

  function fromLocal() {
    if (!localTime.value) {
      timeError.value = "请选择时间";
      return;
    }
    const parsed = dayjs(localTime.value);
    if (!parsed.isValid()) {
      timeError.value = "无法解析这个时间";
      return;
    }
    stamp.value = String(Math.trunc(parsed.valueOf() / 1000));
    fillTime(parsed);
  }

  function fillTime(parsed: Dayjs) {
    localTime.value = parsed.format("YYYY-MM-DD HH:mm:ss");
    secondsText.value = String(Math.trunc(parsed.valueOf() / 1000));
    millisText.value = String(parsed.valueOf());
    timeError.value = "";
  }

  async function digest() {
    md5Text.value = md5(digestInput.value);
    if (!globalThis.crypto?.subtle) {
      shaText.value = "";
      digestError.value = "当前页面无法计算 SHA-256";
      return;
    }
    const buffer = await crypto.subtle.digest("SHA-256", new TextEncoder().encode(digestInput.value));
    shaText.value = [...new Uint8Array(buffer)].map(item => item.toString(16).padStart(2, "0")).join("");
    digestError.value = "";
  }

  function makeUuids(count: number) {
    if (!globalThis.crypto?.randomUUID) {
      ElMessage.error("当前页面无法生成 UUID");
      return;
    }
    uuids.value = Array.from({ length: count }, () => crypto.randomUUID());
  }

  async function copyText(text: string) {
    if (!text) return;
    try {
      await navigator.clipboard.writeText(text);
      ElMessage.success("已复制");
    } catch {
      ElMessage.error("复制失败");
    }
  }

  return {
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
  };
}

function bytesToBase64(bytes: Uint8Array) {
  let binary = "";
  bytes.forEach(byte => {
    binary += String.fromCharCode(byte);
  });
  return btoa(binary);
}
