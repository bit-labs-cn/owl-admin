# API 与数据请求

本篇说明本包内 API 层的组织方式、分页与筛选模式，以及请求风格差异的处理建议。

## 概念说明

- **API 层**：按领域在 `src/api/` 下建模块（如 position.ts、user.ts、role.ts、dept.ts、menu.ts、dict.ts、system.ts），每个模块导出一个类或对象，方法内使用 `@bit-labs.cn/owl-ui` 的 `http.request`。
- **分页**：列表接口通常为 GET + params（page、pageSize、筛选条件），返回格式与 `ResultTable` 约定一致（如 data.list、data.total）；部分接口为 POST + body，需与后端约定一致。
- **静默请求**：列表、下拉选项等不希望报错时全局提示的，可在 `http.request` 的 config 中传 `silentMessage: true`。

## 代码入口

- API 示例：`src/api/position.ts`（getPositions 用 params、createPosition/updatePosition 用 data）
- 通用类型：`@bit-labs.cn/owl-ui` 导出的 `Result`、`ResultTable` 等

## 标准步骤（新增 API）

1. 在 `src/api/` 下新增或扩展模块，从 `@bit-labs.cn/owl-ui/utils/http` 引入 `http`。
2. 定义方法：GET 列表用 `params`，POST/PUT 用 `data`；路径与后端一致（如 `/api/v1/notice`）。
3. 需要静默失败时传第三参 `{ silentMessage: true }`。
4. 在对应 useXList 或页面中调用，分页参数与后端字段名一致（如 page、pageSize、currentPage）。

## 注意事项

- 不同模块请求风格可能不同（如有的列表用 POST），新增时以该模块或 position 为参照，不盲目统一。
- id、status 等类型在 types 中可能为 string 或 number，与后端保持一致即可。

## 完成定义

- 读者能按现有风格在本包内新增 API 方法并在 composable 中正确调用。
