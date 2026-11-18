import { ElMessageBox } from "element-plus";

export async function confirmAction(content: string, title: string) {
  try {
    await ElMessageBox.confirm(content, title, {
      confirmButtonText: "确定",
      cancelButtonText: "取消",
      type: "warning",
      draggable: true,
      closeOnClickModal: false,
      closeOnPressEscape: false
    });
    return true;
  } catch {
    return false;
  }
}
