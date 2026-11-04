# 标准模板与 AI 自检清单

本篇明确**推荐照抄的模板**、**仅作参考的复杂模板**、**不要照抄的实现**，以及 AI 生成代码时的自检项。

## 标准模板（推荐）

- **简单 CRUD**：`src/views/position/`（index.vue + usePositionList.ts + columns.tsx + PositionForm.vue + types）。新增单表 CRUD 时直接以此为模板复制并改名为新模块。
- **层级/树形**：`src/views/dept/`（树表 + 父子选择）、`src/views/user/`（左树右表）为进阶参考，非首选模板。

## 复杂模板（仅作参考）

- **角色与权限分配**：`src/views/role/`（表格 + 权限树侧栏），逻辑较多，可参考但不必一比一复制。
- **字典 dict**：`src/views/dict/` 使用 controller/smart_table 等内联编辑、拖拽排序，属于**特殊模式**，不要作为默认 CRUD 模板；新增普通业务模块时勿照抄 dict 结构。
- **菜单管理 menu**：`src/views/menu/` 用于维护菜单元数据与 auths，其表单与列表与标准 CRUD 有差异，且部分操作可能未完全对接后端持久化，参考时注意。

## 不要照抄的实现

- **menu 模块**：部分成功提示或界面逻辑可能与后端 API 未完全对接，仅作菜单字段与 auths 的参考。
- **api、login-log、operation-log**：多为只读列表，无新增/编辑表单，不要当成标准 CRUD 模板。
- **调试代码**：如 setTimeout 模拟 loading、零散的 console.log，不要当作规范写入新代码。
- **系统配置**：`src/views/setting/` 按接口返回的子系统、分组和配置项渲染，不要在页面里写死配置键，也不要做成可增删的参数表。左边是子系统、一级分组和可选的二级分组，最多这三层，分组名称旁边显示分组键。右边编辑当前叶子分组，配置项标题旁边显示配置键。其他页面用 `settingValue("admin.site.siteName")` 读取（`@bit-labs.cn/owl-ui/config/settings`，`useSetting.ts` 再导出），路径是子系统、分组和配置键，有二级分组时再多一段。公开项由 owl-ui 启动时请求 `GET /api/v1/setting/public` 写入；非公开项要等设置页加载过完整配置。不要在页面里再请求公开配置，也不要自己拆接口树。不要改回默认的 el-tree 加左标签表单。`image` 用选择文件，不要改成地址输入框。
- **系统工具**不是 CRUD：`src/views/tools/password/` 菜单 name `SystemPasswordTool`，path `/system/tools/password/index`；`src/views/tools/server/` 菜单 name `SystemServerMonitor`，path `/system/tools/server/index`；`src/views/tools/codec/` 菜单 name `SystemCodecTool`，path `/system/tools/codec/index`。视图目录必须和菜单 path 去掉 `/system` 前缀后一致，否则页面是空白。编码转换在浏览器内完成，不要为它补空接口。

## AI 自检清单

- [ ] 新页面已设置 `defineOptions({ name })`，且与后端菜单项 `name` 完全一致。
- [ ] 新页面在 `src/views/新模块/` 下，后端菜单的 path/component 为 `/system/新模块/index`（或与 viewModulesPathPrefix 一致）。
- [ ] 表单子组件已暴露 `getRef()`，父页在 addDialog 的 beforeSure 中先 validate 再请求、再 done()。
- [ ] API 方法使用 `http.request`，路径与后端一致；分页/筛选参数与后端约定一致。
- [ ] 未在 `src/routes/index.ts` 中为业务页添加静态路由；未照抄 dict 的复杂内联编辑结构作为默认模板。
- [ ] 类型（id/status 等）与后端一致，不混用 string/number 导致请求错误。

## 完成定义

- 读者与 AI 能选中正确模板、避开错误参考，并按自检清单减少常见错误。
