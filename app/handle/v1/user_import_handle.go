package v1

import (
	"fmt"
	"net/url"

	"bit-labs.cn/owl-admin/app/service"
	"bit-labs.cn/owl/provider/router"
	"github.com/gin-gonic/gin"
)

// @Summary		下载用户导入模板
// @Description	下载批量导入用户的 Excel 模板（含填写说明、可选角色与部门路径）
// @Tags			用户管理
// @Produce		application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Success		200	{file}		file	"Excel 模板"
// @Failure		500	{object}	router.Resp	"服务器内部错误"
// @Router			/api/v1/users/import-template [GET]
func (i *UserHandle) ImportTemplate(ctx *gin.Context) {
	data, name, ctype, err := i.userSvc.BuildUserImportTemplate(ctx.Request.Context())
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	ctx.Header("Content-Disposition", `attachment; filename="`+name+`"; filename*=UTF-8''`+url.PathEscape(name))
	ctx.Data(200, ctype, data)
}

// @Summary		批量导入用户
// @Description	上传 Excel 按模板列创建用户；有效行入库，错误行返回明细（不含密码）
// @Tags			用户管理
// @Accept			multipart/form-data
// @Produce		json
// @Param			file	formData	file					true	"Excel 文件"
// @Success		200		{object}	router.Resp				"操作成功"
// @Failure		400		{object}	router.Resp				"参数错误"
// @Failure		500		{object}	router.Resp				"服务器内部错误"
// @Router			/api/v1/users/import [POST]
func (i *UserHandle) Import(ctx *gin.Context) {
	fh, err := ctx.FormFile("file")
	if err != nil {
		router.Fail(ctx, service.ImportFileInvalid("请上传 Excel 文件"))
		return
	}
	if fh.Size > service.UserImportMaxBytes {
		router.Fail(ctx, service.ImportTooLarge())
		return
	}
	f, err := fh.Open()
	if err != nil {
		router.Fail(ctx, service.ImportFileInvalid("无法读取上传文件"))
		return
	}
	defer f.Close()

	data, err := service.ReadAllLimited(f, service.UserImportMaxBytes)
	if err != nil {
		router.Fail(ctx, err)
		return
	}

	result, err := i.userSvc.ImportUsers(ctx.Request.Context(), fh.Filename, data)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.SuccessWithMsg(ctx, fmt.Sprintf("导入完成：成功 %d 人，失败 %d 人", result.Created, result.Failed), result)
}

// @Summary		导出用户导入失败明细
// @Description	将导入失败的行号、账号、原因导出为 Excel
// @Tags			用户管理
// @Accept			json
// @Produce		application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Param			request	body		service.ExportUserImportErrorsReq	true	"失败明细"
// @Success		200		{file}		file	"错误明细 Excel"
// @Failure		400		{object}	router.Resp	"参数错误"
// @Failure		500		{object}	router.Resp	"服务器内部错误"
// @Router			/api/v1/users/import-errors/export [POST]
func (i *UserHandle) ExportImportErrors(ctx *gin.Context) {
	var req service.ExportUserImportErrorsReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	data, name, ctype, err := i.userSvc.BuildUserImportErrorsExcel(&req)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	ctx.Header("Content-Disposition", `attachment; filename="`+name+`"; filename*=UTF-8''`+url.PathEscape(name))
	ctx.Data(200, ctype, data)
}
