import { computed, reactive, ref } from "vue";
import { ElMessage } from "element-plus";
import { passwordToolAPI } from "@bit-labs.cn/owl-admin-ui/api/password-tool";

export function usePasswordTool() {
  const form = reactive({
    length: 16,
    count: 5,
    upper: true,
    lower: true,
    digit: true,
    symbol: true,
    excludeAmbiguous: true
  });
  const passwords = ref<string[]>([]);
  const generating = ref(false);
  const plain = ref("");
  const hash = ref("");
  const hashing = ref(false);
  const charsetSelected = computed(() => form.upper || form.lower || form.digit || form.symbol);

  async function generate() {
    if (!charsetSelected.value) {
      ElMessage.error("请至少选择一种字符");
      return;
    }
    if (!form.length || !form.count) {
      ElMessage.error("请填写长度和条数");
      return;
    }
    generating.value = true;
    try {
      const res = await passwordToolAPI.generate({ ...form });
      passwords.value = res.data?.passwords ?? [];
    } finally {
      generating.value = false;
    }
  }

  async function hashPassword() {
    const password = plain.value;
    if (!password) {
      ElMessage.error("请输入明文");
      return;
    }
    if (new TextEncoder().encode(password).length > 72) {
      ElMessage.error("密码不能超过 72 个字节");
      return;
    }
    hashing.value = true;
    try {
      const res = await passwordToolAPI.hash({ password });
      hash.value = res.data?.hash ?? "";
    } finally {
      hashing.value = false;
    }
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

  function copyAll() {
    void copyText(passwords.value.join("\n"));
  }

  return {
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
  };
}
