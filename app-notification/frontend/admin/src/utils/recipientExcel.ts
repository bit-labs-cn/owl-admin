import type { ExternalRecipient } from "../views/send/useManualSend";

/** 外部人员导入模板的表头，顺序就是列顺序。 */
export const externalTemplateHeaders = ["称呼", "邮箱", "手机号"] as const;

/** 单次导入的行数上限，和系统用户导入保持一致的量级。 */
export const externalImportMaxRows = 2000;

/** 导入时被跳过的一行。 */
export interface ExternalImportError {
  row: number;
  reason: string;
}

/** 导入结果：能用的行和跳过原因，让人知道哪几行没进来。 */
export interface ExternalImportResult {
  rows: ExternalRecipient[];
  errors: ExternalImportError[];
  total: number;
}

const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

/**
 * exceljs 体积不小，只有真正导入或下载模板时才加载，
 * 免得每次打开接收人弹窗都要多下一个 Excel 解析包。
 */
async function loadExcelJS() {
  const mod = await import("exceljs");
  return mod.default ?? (mod as unknown as typeof import("exceljs"));
}

/** 表头归一：去掉空格并转小写，兼容中英文表头。 */
function normalizeHeader(value: string) {
  return value.replace(/\s+/g, "").toLowerCase();
}

const headerAliases: Record<string, "name" | "email" | "phone"> = {
  称呼: "name",
  姓名: "name",
  名字: "name",
  name: "name",
  邮箱: "email",
  邮件: "email",
  email: "email",
  手机号: "phone",
  手机: "phone",
  电话: "phone",
  phone: "phone",
  mobile: "phone"
};

/** 单元格取字符串，兼容富文本、公式结果和数字（手机号常被存成数字）。 */
function cellText(value: unknown): string {
  if (value === null || value === undefined) return "";
  if (value instanceof Date) return value.toISOString();
  if (typeof value === "object") {
    const record = value as Record<string, unknown>;
    if (typeof record.text === "string") return record.text.trim();
    if (record.result !== undefined && record.result !== null) return String(record.result).trim();
    if (Array.isArray(record.richText)) {
      return record.richText
        .map(item => (item as { text?: string }).text ?? "")
        .join("")
        .trim();
    }
    return "";
  }
  return String(value).trim();
}

/**
 * 解析外部人员 Excel。第一张工作表、第一行是表头，列顺序不限。
 * 没有邮箱的行进不了投递，直接记为错误；重复邮箱只保留第一条。
 */
export async function parseExternalWorkbook(file: File): Promise<ExternalImportResult> {
  const ExcelJS = await loadExcelJS();
  const workbook = new ExcelJS.Workbook();
  await workbook.xlsx.load(await file.arrayBuffer());
  const sheet = workbook.worksheets[0];
  if (!sheet) {
    throw new Error("文件里没有工作表");
  }

  const columns: Partial<Record<"name" | "email" | "phone", number>> = {};
  sheet.getRow(1).eachCell((cell, colNumber) => {
    const field = headerAliases[normalizeHeader(cellText(cell.value))];
    if (field && columns[field] === undefined) columns[field] = colNumber;
  });
  if (columns.email === undefined) {
    throw new Error(`表头缺少「邮箱」列，请使用导入模板（${externalTemplateHeaders.join(" / ")}）`);
  }

  const result: ExternalImportResult = { rows: [], errors: [], total: 0 };
  const seen = new Set<string>();
  const lastRow = sheet.rowCount;

  for (let rowNumber = 2; rowNumber <= lastRow; rowNumber++) {
    const row = sheet.getRow(rowNumber);
    const name = columns.name ? cellText(row.getCell(columns.name).value) : "";
    const email = columns.email ? cellText(row.getCell(columns.email).value) : "";
    const phone = columns.phone ? cellText(row.getCell(columns.phone).value) : "";
    if (!name && !email && !phone) continue;

    result.total++;
    if (result.total > externalImportMaxRows) {
      result.errors.push({ row: rowNumber, reason: `超过单次导入上限 ${externalImportMaxRows} 行` });
      break;
    }
    if (!email) {
      result.errors.push({ row: rowNumber, reason: "缺邮箱，外部人员只能按邮箱投递" });
      continue;
    }
    if (!emailPattern.test(email)) {
      result.errors.push({ row: rowNumber, reason: `邮箱格式不正确：${email}` });
      continue;
    }
    const key = email.toLowerCase();
    if (seen.has(key)) {
      result.errors.push({ row: rowNumber, reason: `邮箱重复：${email}` });
      continue;
    }
    seen.add(key);
    result.rows.push({ name: name || email, email, phone });
  }

  return result;
}

/** 生成并下载导入模板，表头加一行示例，列宽调到能看清。 */
export async function downloadExternalTemplate() {
  const ExcelJS = await loadExcelJS();
  const workbook = new ExcelJS.Workbook();
  const sheet = workbook.addWorksheet("外部人员");
  sheet.addRow([...externalTemplateHeaders]);
  sheet.addRow(["张三", "zhangsan@example.com", "13800000000"]);
  sheet.getRow(1).font = { bold: true };
  sheet.columns = [{ width: 18 }, { width: 32 }, { width: 18 }];

  const buffer = await workbook.xlsx.writeBuffer();
  const blob = new Blob([buffer], {
    type: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
  });
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = "通知外部人员导入模板.xlsx";
  link.click();
  URL.revokeObjectURL(url);
}
