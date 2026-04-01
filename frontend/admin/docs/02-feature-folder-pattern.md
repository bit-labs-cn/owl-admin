# 功能目录模式

**适用范围**：本篇适用于在 **owl-admin-ui 包内** 的 `src/views/` 下新增业务模块。若要新建**独立前端子系统包**（与 owl-admin-ui 平级的独立 npm 包），请参考 **owl-ui** 的 `docs/03-subsystem-contract.md` 与 `docs/08-minimal-subsystem-template.md`。

本篇说明在 `src/views/` 下新增一个业务模块时的**标准目录与文件结构**，以及各文件职责。

## 概念说明

- **标准 CRUD 模块**：以单表列表 + 弹窗表单为主，采用 `index.vue`（页面壳） + `useXList.ts`（列表状态与请求） + `columns.tsx`（表格列） + `Form.vue`（弹窗表单） + `types.ts` / `rules.ts`（类型与校验规则）。
- **模板参考**：`position` 为最简标准模板；`dept` 为带树形/层级的模板；`user`、`role` 为更复杂（多弹窗、树选择等）；`dict` 为特殊的内联编辑模式，不作为默认模板。

## 代码入口（标准模板）

- 标准 CRUD 参考：`src/views/position/index.vue`、`usePositionList.ts`、`columns.tsx`、`PositionForm.vue`、`types.ts`
- 层级 CRUD 参考：`src/views/dept/`、`src/views/user/`（左树右表或树选择）

## 标准文件清单与职责

| 文件 | 职责 |
|------|------|
| `index.vue` | 页面壳：defineOptions({ name })、搜索表单、PureTableBar + pure-table、调用 useXList、打开 addDialog、传 formInline、beforeSure 中取 FormRef.validate 再请求 |
| `useXList.ts` 或 `useXList.tsx` | 列表状态：form、dataList、loading、pagination、switchLoadMap、onSearch、resetForm、handleSizeChange、handleCurrentChange、onChange（状态切换），内部调用对应 API |
| `columns.tsx` | 表格列定义：createColumns({ switchLoadMap, onChange })，返回列数组，含 cellRenderer（如 el-switch）、formatter、slot: "operation" |
| `Form.vue` | 弹窗表单：props formInline、ref 暴露 getRef() 返回 el-form 引用、rules、提交时由父页在 beforeSure 中校验并调用 API |
| `types.ts` | 表单/行数据类型、接口请求参数类型 |
| `rules.ts` | 可选，Element Plus 校验规则，也可写在 Form.vue 内 |

## 标准步骤（新增一个 CRUD 模块）

1. 在 `src/views/` 下新建目录，如 `notice/`。
2. 新增 `index.vue`，设置 `defineOptions({ name: "SystemNotice" })`（与后端菜单 name 一致）。
3. 新增 `useNoticeList.ts`，封装 form、dataList、pagination、onSearch、API 调用。
4. 新增 `columns.tsx`，`createColumns` 返回列配置。
5. 新增 `NoticeForm.vue`，接收 `formInline`，暴露 `getRef()`。
6. 在 `index.vue` 中用 `addDialog` + `contentRenderer: () => h(NoticeForm, ...)`，`beforeSure` 内 `FormRef.validate` 后调用 create/update API。
7. 在 `src/api/` 下新增或扩展 API 类（如 notice.ts），使用 `http.request`。
8. 后端需配置对应菜单（path/component 为 `/system/notice/index`，name 为 `SystemNotice`）。

## 注意事项

- 表单组件必须暴露 `getRef()`，父页在 `beforeSure` 中通过 `formRef.value.getRef().validate()` 校验后再提交。
- 各模块 API 请求风格可能略有差异（GET params vs POST body、分页字段名），新增时参考同模块或 position 的写法，保持与后端一致。

## 完成定义

- 读者能按 position 模式复制出一套新模块的文件结构并完成列表、弹窗、校验与 API 对接。
