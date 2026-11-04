package v1

import (
	"bit-labs.cn/owl-admin/app/service"
	"bit-labs.cn/owl/provider/router"
	"github.com/gin-gonic/gin"
)

var _ router.Handler = (*PasswordToolHandle)(nil)

// PasswordToolHandle 提供随机密码和 bcrypt 计算。
type PasswordToolHandle struct {
	svc *service.PasswordToolService
}

// NewPasswordToolHandle 创建密码工具接口。
func NewPasswordToolHandle(svc *service.PasswordToolService) *PasswordToolHandle {
	return &PasswordToolHandle{svc: svc}
}

// ModuleName 返回权限前缀和菜单标题。
func (i *PasswordToolHandle) ModuleName() (string, string) {
	return "password", "密码生成"
}

// @Summary		生成密码
// @Description	按长度和字符种类生成随机密码，不保存
// @Tags			密码生成
// @Accept			json
// @Produce		json
// @Param			request	body		service.GeneratePasswordReq	true	"生成条件"
// @Success		200		{object}	router.Resp					"操作成功"
// @Failure		400		{object}	router.Resp					"参数错误"
// @Failure		500		{object}	router.Resp					"服务器内部错误"
// @Router			/api/v1/tools/passwords [POST]
func (i *PasswordToolHandle) Generate(ctx *gin.Context) {
	var req service.GeneratePasswordReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	result, err := i.svc.Generate(ctx.Request.Context(), &req)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, result)
}

// @Summary		计算密文
// @Description	用 bcrypt 计算密码密文，不写入操作日志
// @Tags			密码生成
// @Accept			json
// @Produce		json
// @Param			request	body		service.HashPasswordReq	true	"明文"
// @Success		200		{object}	router.Resp				"操作成功"
// @Failure		400		{object}	router.Resp				"参数错误"
// @Failure		500		{object}	router.Resp				"服务器内部错误"
// @Router			/api/v1/tools/password-hashes [POST]
func (i *PasswordToolHandle) Hash(ctx *gin.Context) {
	var req service.HashPasswordReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		router.Fail(ctx, err)
		return
	}
	result, err := i.svc.Hash(ctx.Request.Context(), &req)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, result)
}
