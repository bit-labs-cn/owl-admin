package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/mail"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"

	"bit-labs.cn/owl-admin/app/event"
	"bit-labs.cn/owl-admin/app/model"
	errContract "bit-labs.cn/owl/contract/errors"
	"github.com/spf13/cast"
	"github.com/xuri/excelize/v2"
)

const (
	UserImportMaxBytes = 5 << 20
	UserImportMaxRows  = 2000
	UserImportMaxErrs  = 100

	UserImportTemplateName    = "user-import-template.xlsx"
	UserImportErrorsFileName  = "user-import-errors.xlsx"
	userImportSource          = ""

	CodeImportFileInvalid = "USER_IMPORT_FILE_INVALID"
	CodeImportEmpty       = "USER_IMPORT_EMPTY"
	CodeImportTooLarge    = "USER_IMPORT_TOO_LARGE"
	CodeImportTooMany     = "USER_IMPORT_TOO_MANY"
)

var userImportHeaders = []string{
	"用户账号", "用户昵称", "初始密码", "手机号", "邮箱", "性别", "状态", "归属部门", "角色", "备注",
}

var userImportHeaderAliases = map[string]string{
	"用户账号": "username", "账号": "username", "用户名": "username", "username": "username",
	"用户昵称": "nickname", "昵称": "nickname", "nickname": "nickname", "nickName": "nickname",
	"初始密码": "password", "密码": "password", "password": "password",
	"手机号": "phone", "手机号码": "phone", "手机": "phone", "phone": "phone",
	"邮箱": "email", "email": "email",
	"性别": "sex", "sex": "sex",
	"状态": "status", "status": "status",
	"归属部门": "dept", "部门": "dept", "部门路径": "dept", "dept": "dept",
	"角色": "roles", "角色名称": "roles", "roles": "roles",
	"备注": "remark", "remark": "remark",
}

var phoneRegexp = regexp.MustCompile(`^1[3-9]\d{9}$`)

func ImportFileInvalid(msg string) *errContract.BizError {
	if msg == "" {
		msg = "导入文件无效"
	}
	return errContract.NewBizError(CodeImportFileInvalid, msg)
}

func ImportEmpty() *errContract.BizError {
	return errContract.NewBizError(CodeImportEmpty, "导入文件没有有效用户行")
}

func ImportTooLarge() *errContract.BizError {
	return errContract.NewBizError(CodeImportTooLarge, "导入文件不能超过 5MB")
}

func ImportTooMany() *errContract.BizError {
	return errContract.NewBizError(CodeImportTooMany, "单次最多导入 2000 个用户")
}

// ImportUserError 导入失败行（不包含密码）
type ImportUserError struct {
	Row      int    `json:"row"`
	Username string `json:"username"`
	Reason   string `json:"reason"`
}

// ImportUserResult 导入汇总
type ImportUserResult struct {
	Total   int               `json:"total"`
	Created int               `json:"created"`
	Failed  int               `json:"failed"`
	Errors  []ImportUserError `json:"errors"`
}

func (r *ImportUserResult) addErr(row int, username, reason string) {
	r.Failed++
	if len(r.Errors) >= UserImportMaxErrs {
		return
	}
	r.Errors = append(r.Errors, ImportUserError{Row: row, Username: username, Reason: reason})
}

type userImportDraft struct {
	Line     int
	Username string
	Nickname string
	Password string
	Phone    string
	Email    string
	SexRaw   string
	StatusRaw string
	DeptPath string
	RolesRaw string
	Remark   string
}

type userImportPrepared struct {
	line     int
	username string
	roleIDs  []string
	user     model.User
}

// BuildUserImportTemplate 生成用户导入 Excel 模板
func (i *UserService) BuildUserImportTemplate(ctx context.Context) ([]byte, string, string, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheet := "用户"
	if err := f.SetSheetName("Sheet1", sheet); err != nil {
		return nil, "", "", err
	}
	for col, h := range userImportHeaders {
		cell, err := excelize.CoordinatesToCellName(col+1, 1)
		if err != nil {
			return nil, "", "", err
		}
		if err := f.SetCellValue(sheet, cell, h); err != nil {
			return nil, "", "", err
		}
	}
	example := []string{"zhangsan", "张三", "Passw0rd", "13800138000", "zhangsan@example.com", "男", "启用", "总公司/研发部", "普通用户,访客", "示例行，导入前请删除或改成真实数据"}
	for col, v := range example {
		cell, err := excelize.CoordinatesToCellName(col+1, 2)
		if err != nil {
			return nil, "", "", err
		}
		if err := f.SetCellValue(sheet, cell, v); err != nil {
			return nil, "", "", err
		}
	}
	_ = f.SetColWidth(sheet, "A", "J", 16)
	_ = f.SetColWidth(sheet, "H", "I", 24)

	helpIdx, err := f.NewSheet("填写说明")
	if err != nil {
		return nil, "", "", err
	}
	helpName := f.GetSheetName(helpIdx)
	notes := []string{
		"请先下载本模板后填写，仅支持 .xlsx。",
		"必填列：用户账号、用户昵称、初始密码。",
		"用户账号 2～32 个字符，同一文件内不可重复，且不可与系统已有账号重复。",
		"初始密码 6～64 位；导入成功后以加密形式存储，响应与错误明细不会回显密码。",
		"性别填写「男」或「女」；留空默认为男。",
		"状态填写「启用」或「停用」；留空默认为启用。",
		"归属部门填写完整层级路径，用 / 分隔，例如：总公司/研发部；留空表示不绑定部门。",
		"角色填写角色名称，多个用英文逗号或中文逗号分隔；留空表示不绑定角色。",
		"填写的部门路径或角色名称必须在系统中已存在；部门名称有歧义（同名多条）时该行失败。",
		"单次最多 2000 行，文件不超过 5MB。导入前请删除示例行。",
	}
	for idx, line := range notes {
		cell, err := excelize.CoordinatesToCellName(1, idx+1)
		if err != nil {
			return nil, "", "", err
		}
		if err := f.SetCellValue(helpName, cell, line); err != nil {
			return nil, "", "", err
		}
	}
	_ = f.SetColWidth(helpName, "A", "A", 96)

	if err := i.fillImportReferenceSheets(ctx, f); err != nil {
		return nil, "", "", err
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, "", "", err
	}
	return buf.Bytes(), UserImportTemplateName, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", nil
}

// ExportUserImportErrorsReq 导出导入失败明细
type ExportUserImportErrorsReq struct {
	Errors []ImportUserError `json:"errors" validate:"required,min=1,dive" label:"错误明细"`
}

// BuildUserImportErrorsExcel 将导入失败明细导出为 Excel
func (i *UserService) BuildUserImportErrorsExcel(req *ExportUserImportErrorsReq) ([]byte, string, string, error) {
	if req == nil || len(req.Errors) == 0 {
		return nil, "", "", ImportFileInvalid("没有可导出的错误记录")
	}
	f := excelize.NewFile()
	defer f.Close()

	sheet := "导入失败"
	if err := f.SetSheetName("Sheet1", sheet); err != nil {
		return nil, "", "", err
	}
	headers := []string{"行号", "账号", "原因"}
	for col, h := range headers {
		cell, err := excelize.CoordinatesToCellName(col+1, 1)
		if err != nil {
			return nil, "", "", err
		}
		if err := f.SetCellValue(sheet, cell, h); err != nil {
			return nil, "", "", err
		}
	}
	for n, item := range req.Errors {
		row := n + 2
		_ = f.SetCellValue(sheet, fmt.Sprintf("A%d", row), item.Row)
		_ = f.SetCellValue(sheet, fmt.Sprintf("B%d", row), item.Username)
		_ = f.SetCellValue(sheet, fmt.Sprintf("C%d", row), item.Reason)
	}
	_ = f.SetColWidth(sheet, "A", "A", 10)
	_ = f.SetColWidth(sheet, "B", "B", 24)
	_ = f.SetColWidth(sheet, "C", "C", 48)

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, "", "", err
	}
	return buf.Bytes(), UserImportErrorsFileName, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", nil
}

func (i *UserService) fillImportReferenceSheets(ctx context.Context, f *excelize.File) error {
	roles, err := i.roleRepo.WithContext(ctx).ListAll()
	if err != nil {
		return err
	}
	roleIdx, err := f.NewSheet("可选角色")
	if err != nil {
		return err
	}
	roleSheet := f.GetSheetName(roleIdx)
	_ = f.SetCellValue(roleSheet, "A1", "角色名称")
	_ = f.SetCellValue(roleSheet, "B1", "角色编码")
	for n, role := range roles {
		_ = f.SetCellValue(roleSheet, fmt.Sprintf("A%d", n+2), role.Name)
		_ = f.SetCellValue(roleSheet, fmt.Sprintf("B%d", n+2), role.Code)
	}
	_ = f.SetColWidth(roleSheet, "A", "B", 24)

	depts, err := i.deptRepo.WithContext(ctx).ListAll()
	if err != nil {
		return err
	}
	paths := buildDeptPaths(depts)
	deptIdx, err := f.NewSheet("可选部门路径")
	if err != nil {
		return err
	}
	deptSheet := f.GetSheetName(deptIdx)
	_ = f.SetCellValue(deptSheet, "A1", "部门路径")
	row := 2
	for _, path := range paths {
		_ = f.SetCellValue(deptSheet, fmt.Sprintf("A%d", row), path)
		row++
	}
	_ = f.SetColWidth(deptSheet, "A", "A", 48)
	return nil
}

// ImportUsers 从 Excel 批量导入用户（有效行入库，错误行返回明细）
func (i *UserService) ImportUsers(ctx context.Context, filename string, data []byte) (*ImportUserResult, error) {
	if len(data) == 0 {
		return nil, ImportEmpty()
	}
	if len(data) > UserImportMaxBytes {
		return nil, ImportTooLarge()
	}
	ext := strings.ToLower(filepath.Ext(strings.TrimSpace(filename)))
	if ext != "" && ext != ".xlsx" && ext != ".xlsm" {
		return nil, ImportFileInvalid("仅支持 .xlsx 文件")
	}

	rows, err := readUserImportRows(data)
	if err != nil {
		return nil, err
	}
	drafts, err := parseUserImportRows(rows)
	if err != nil {
		return nil, err
	}
	if len(drafts) == 0 {
		return nil, ImportEmpty()
	}
	if len(drafts) > UserImportMaxRows {
		return nil, ImportTooMany()
	}

	l := i.locker.New()
	if err := l.Lock("user:import"); err != nil {
		return nil, err
	}
	defer l.Unlock()

	out := &ImportUserResult{Total: len(drafts), Errors: make([]ImportUserError, 0)}
	catalog, err := i.loadUserImportCatalog(ctx, drafts)
	if err != nil {
		return nil, err
	}

	seen := map[string]int{}
	prepared := make([]*userImportPrepared, 0, len(drafts))
	for _, d := range drafts {
		if prev, ok := seen[d.Username]; ok {
			out.addErr(d.Line, d.Username, fmt.Sprintf("文件内用户账号重复，与第 %d 行相同", prev))
			continue
		}
		seen[d.Username] = d.Line
		row, reason := prepareUserImportDraft(d, catalog)
		if reason != "" {
			out.addErr(d.Line, d.Username, reason)
			continue
		}
		prepared = append(prepared, row)
	}

	if len(prepared) == 0 {
		out.Failed = out.Total
		return out, nil
	}

	users := make([]*model.User, len(prepared))
	for idx := range prepared {
		users[idx] = &prepared[idx].user
	}
	if err := i.userRepo.WithContext(ctx).SaveBatch(users); err != nil {
		return nil, err
	}

	out.Created = len(prepared)
	out.Failed = out.Total - out.Created

	for _, p := range prepared {
		if len(p.roleIDs) == 0 {
			continue
		}
		i.eventBus.Publish(event.AssignRoleToUser, &AssignRoleToUser{
			UserID:  p.user.ID,
			RoleIDs: p.roleIDs,
		})
	}
	return out, nil
}

type userImportCatalog struct {
	existing map[string]struct{}
	roles    map[string]model.Role // name -> role（同名多条时标歧义）
	roleAmbig map[string]struct{}
	deptPaths map[string]model.Dept // path -> dept
	deptAmbig map[string]struct{}
}

func (i *UserService) loadUserImportCatalog(ctx context.Context, drafts []userImportDraft) (*userImportCatalog, error) {
	names := make([]string, 0, len(drafts))
	for _, d := range drafts {
		if d.Username != "" {
			names = append(names, d.Username)
		}
	}
	existing, err := i.userRepo.WithContext(ctx).ExistingUsernames(uniqueImportStrings(names), userImportSource)
	if err != nil {
		return nil, err
	}
	roles, err := i.roleRepo.WithContext(ctx).ListAll()
	if err != nil {
		return nil, err
	}
	roleMap := map[string]model.Role{}
	roleAmbig := map[string]struct{}{}
	for _, role := range roles {
		name := strings.TrimSpace(role.Name)
		if name == "" {
			continue
		}
		if _, ok := roleMap[name]; ok {
			roleAmbig[name] = struct{}{}
			continue
		}
		roleMap[name] = role
	}

	depts, err := i.deptRepo.WithContext(ctx).ListAll()
	if err != nil {
		return nil, err
	}
	deptPaths, deptAmbig := buildDeptPathMaps(depts)

	return &userImportCatalog{
		existing:  existing,
		roles:     roleMap,
		roleAmbig: roleAmbig,
		deptPaths: deptPaths,
		deptAmbig: deptAmbig,
	}, nil
}

func prepareUserImportDraft(d userImportDraft, cat *userImportCatalog) (*userImportPrepared, string) {
	if d.Username == "" {
		return nil, "用户账号必填"
	}
	if utf8.RuneCountInString(d.Username) < 2 || utf8.RuneCountInString(d.Username) > 32 {
		return nil, "用户账号长度须为 2～32 个字符"
	}
	if d.Nickname == "" {
		return nil, "用户昵称必填"
	}
	if utf8.RuneCountInString(d.Nickname) > 32 {
		return nil, "用户昵称不能超过 32 个字符"
	}
	if d.Password == "" {
		return nil, "初始密码必填"
	}
	if utf8.RuneCountInString(d.Password) < 6 || utf8.RuneCountInString(d.Password) > 64 {
		return nil, "初始密码长度须为 6～64 位"
	}
	if _, ok := cat.existing[d.Username]; ok {
		return nil, "用户账号已存在"
	}
	if d.Phone != "" && !phoneRegexp.MatchString(d.Phone) {
		return nil, "手机号格式不正确"
	}
	if d.Email != "" {
		if _, err := mail.ParseAddress(d.Email); err != nil {
			return nil, "邮箱格式不正确"
		}
	}
	sex, ok := parseImportSex(d.SexRaw)
	if !ok {
		return nil, "性别仅支持「男」或「女」"
	}
	status, ok := parseImportStatus(d.StatusRaw)
	if !ok {
		return nil, "状态仅支持「启用」或「停用」"
	}
	if utf8.RuneCountInString(d.Remark) > 255 {
		return nil, "备注不能超过 255 个字符"
	}

	user := model.User{
		Username: d.Username,
		Nickname: d.Nickname,
		Phone:    d.Phone,
		Email:    d.Email,
		Sex:      sex,
		Status:   status,
		Remark:   d.Remark,
		Source:   userImportSource,
	}
	user.SetPassword(d.Password)

	if d.DeptPath != "" {
		if _, ambig := cat.deptAmbig[d.DeptPath]; ambig {
			return nil, "归属部门路径存在歧义：" + d.DeptPath
		}
		dept, found := cat.deptPaths[d.DeptPath]
		if !found {
			return nil, "归属部门不存在：" + d.DeptPath
		}
		user.Depts = []model.Dept{dept}
	}

	roleIDs := make([]string, 0)
	if d.RolesRaw != "" {
		names := splitImportRoleNames(d.RolesRaw)
		if len(names) == 0 {
			return nil, "角色名称无效"
		}
		roles := make([]model.Role, 0, len(names))
		seenRole := map[string]struct{}{}
		for _, name := range names {
			if _, ambig := cat.roleAmbig[name]; ambig {
				return nil, "角色名称存在歧义：" + name
			}
			role, found := cat.roles[name]
			if !found {
				return nil, "角色不存在：" + name
			}
			id := cast.ToString(role.ID)
			if _, ok := seenRole[id]; ok {
				continue
			}
			seenRole[id] = struct{}{}
			roles = append(roles, role)
			roleIDs = append(roleIDs, id)
		}
		user.SetRoles(roles)
	}

	return &userImportPrepared{
		line:     d.Line,
		username: d.Username,
		roleIDs:  roleIDs,
		user:     user,
	}, ""
}

func parseUserImportRows(rows [][]string) ([]userImportDraft, error) {
	if len(rows) == 0 {
		return nil, ImportEmpty()
	}
	idx, err := mapUserImportHeader(rows[0])
	if err != nil {
		return nil, err
	}
	out := make([]userImportDraft, 0, len(rows)-1)
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if importRowEmpty(row) {
			continue
		}
		out = append(out, userImportDraft{
			Line:      i + 1,
			Username:  cellBy(row, idx, "username"),
			Nickname:  cellBy(row, idx, "nickname"),
			Password:  cellBy(row, idx, "password"),
			Phone:     cellBy(row, idx, "phone"),
			Email:     cellBy(row, idx, "email"),
			SexRaw:    cellBy(row, idx, "sex"),
			StatusRaw: cellBy(row, idx, "status"),
			DeptPath:  normalizeDeptPath(cellBy(row, idx, "dept")),
			RolesRaw:  cellBy(row, idx, "roles"),
			Remark:    cellBy(row, idx, "remark"),
		})
	}
	return out, nil
}

func mapUserImportHeader(header []string) (map[string]int, error) {
	idx := map[string]int{}
	for i, raw := range header {
		key := userImportHeaderAliases[normalizeUserImportHeader(raw)]
		if key == "" {
			continue
		}
		if _, ok := idx[key]; !ok {
			idx[key] = i
		}
	}
	for _, required := range []string{"username", "nickname", "password"} {
		if _, ok := idx[required]; !ok {
			label := map[string]string{"username": "用户账号", "nickname": "用户昵称", "password": "初始密码"}[required]
			return nil, ImportFileInvalid("缺少列：" + label)
		}
	}
	return idx, nil
}

func normalizeUserImportHeader(s string) string {
	s = strings.TrimSpace(strings.TrimSuffix(s, "*"))
	s = strings.ReplaceAll(s, " ", "")
	if _, ok := userImportHeaderAliases[s]; ok {
		return s
	}
	return strings.ToLower(s)
}

func parseImportSex(raw string) (int, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "男" || raw == "0" {
		return 0, true
	}
	if raw == "女" || raw == "1" {
		return 1, true
	}
	return 0, false
}

func parseImportStatus(raw string) (int, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "启用" || raw == "1" {
		return 1, true
	}
	if raw == "停用" || raw == "禁用" || raw == "0" {
		return 0, true
	}
	return 0, false
}

func splitImportRoleNames(raw string) []string {
	raw = strings.ReplaceAll(raw, "，", ",")
	raw = strings.ReplaceAll(raw, "、", ",")
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, p := range parts {
		name := strings.TrimSpace(p)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	return out
}

func normalizeDeptPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	path = strings.ReplaceAll(path, "\\", "/")
	parts := strings.Split(path, "/")
	clean := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		clean = append(clean, p)
	}
	return strings.Join(clean, "/")
}

func buildDeptPaths(depts []model.Dept) []string {
	paths, _ := buildDeptPathMaps(depts)
	out := make([]string, 0, len(paths))
	for path := range paths {
		out = append(out, path)
	}
	sort.Strings(out)
	return out
}

func buildDeptPathMaps(depts []model.Dept) (map[string]model.Dept, map[string]struct{}) {
	byID := map[uint]model.Dept{}
	for _, d := range depts {
		byID[d.ID] = d
	}
	pathOf := map[uint]string{}
	var resolve func(id uint) string
	resolve = func(id uint) string {
		if p, ok := pathOf[id]; ok {
			return p
		}
		d, ok := byID[id]
		if !ok {
			return ""
		}
		name := strings.TrimSpace(d.Name)
		parentID := uint(d.ParentId)
		if parentID == 0 || parentID == id {
			pathOf[id] = name
			return name
		}
		parentPath := resolve(parentID)
		if parentPath == "" {
			pathOf[id] = name
			return name
		}
		path := parentPath + "/" + name
		pathOf[id] = path
		return path
	}

	paths := map[string]model.Dept{}
	ambig := map[string]struct{}{}
	for _, d := range depts {
		path := resolve(d.ID)
		if path == "" {
			continue
		}
		if exist, ok := paths[path]; ok && exist.ID != d.ID {
			ambig[path] = struct{}{}
			continue
		}
		paths[path] = d
	}
	return paths, ambig
}

func readUserImportRows(data []byte) ([][]string, error) {
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return nil, ImportFileInvalid("无法解析 Excel 文件，请使用下载的模板")
	}
	defer f.Close()
	sheet := f.GetSheetName(0)
	if sheet == "" {
		return nil, ImportEmpty()
	}
	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, ImportFileInvalid("无法读取 Excel 工作表")
	}
	return rows, nil
}

func importRowEmpty(row []string) bool {
	for _, c := range row {
		if strings.TrimSpace(c) != "" {
			return false
		}
	}
	return true
}

func cellBy(row []string, idx map[string]int, key string) string {
	i, ok := idx[key]
	if !ok || i < 0 || i >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[i])
}

func uniqueImportStrings(in []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

// ReadAllLimited 读取上传内容并校验大小（供 handle 使用）
func ReadAllLimited(r io.Reader, max int64) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(r, max+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > max {
		return nil, ImportTooLarge()
	}
	return data, nil
}
