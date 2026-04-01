# UI 模式：表格、弹窗与表单

本篇说明本包内**表格**（PureTableBar、pure-table）、**弹窗**（addDialog、contentRenderer）、**表单校验闭环**（getRef、beforeSure）的标准用法。

## 概念说明

- **表格**：使用 `@bit-labs.cn/owl-ui` 的 `PureTableBar` 与 `pure-table`，列配置来自 `createColumns` 返回的数组，操作列用 `slot: "operation"`，在模板中写按钮。
- **弹窗**：使用 `addDialog`（来自 `@bit-labs.cn/owl-ui/components/ReDialog`），通过 `contentRenderer: () => h(FormComponent, { ref, formInline: null })` 渲染表单组件，`props.formInline` 由 addDialog 的 `props.formInline` 传入，用于新增/编辑的初始值。
- **表单校验**：表单子组件通过 `defineExpose({ getRef })` 暴露 `getRef()` 返回 el-form 的 ref；父页在 `beforeSure(done, { options })` 中取 `formRef.value.getRef().validate(valid => { if (valid) { /* 调 API */ chores(); done(); } })`，先校验再提交，提交成功后 `done()` 关闭弹窗并可选刷新列表。

## 代码入口

- 表格与弹窗示例：`src/views/position/index.vue`（PureTableBar、pure-table、addDialog、contentRenderer、beforeSure）
- 表单暴露 getRef：`src/views/position/PositionForm.vue`（getRef、defineExpose）
- 列定义：`src/views/position/columns.tsx`（createColumns、cellRenderer、slot: "operation"）

## 标准步骤（新增带表单的 CRUD 页）

1. 在 index.vue 中 `addDialog({ title, props: { formInline }, contentRenderer: () => h(XxxForm, { ref: xxxFormRef, formInline: null }), beforeSure: (done, { options }) => { const FormRef = xxxFormRef.value.getRef(); const curData = options.props.formInline; FormRef.validate(valid => { if (valid) { api.create/update(curData).then(() => { done(); onSearch(); }); } }); } })`。
2. 在 Form.vue 中接收 `formInline`，用 `defineExpose({ getRef })` 暴露表单 ref，校验规则写在 Form 内或 rules.ts。
3. 表格列用 columns.tsx 的 createColumns，操作列用 slot 在 index.vue 里写编辑/删除按钮。

## 注意事项

- `formInline` 在 addDialog 的 props 中传入，用于区分新增（空/默认值）与编辑（行数据）；contentRenderer 里可传 `formInline: null`，实际数据从 options.props.formInline 取。
- 提交前必须通过 FormRef.validate，否则可能提交非法数据；done() 应在请求成功后再调，以关闭弹窗。

## 完成定义

- 读者能正确使用 PureTableBar、addDialog、contentRenderer 与 getRef/validate/beforeSure 完成增删改查弹窗闭环。
