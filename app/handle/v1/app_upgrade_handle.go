package v1

import (
	"strconv"

	"bit-labs.cn/owl-admin/app/service"
	"bit-labs.cn/owl/provider/router"
	"github.com/gin-gonic/gin"
)

var _ router.Handler = (*AppUpgradeHandle)(nil)

type AppUpgradeHandle struct {
	appVersionSvc *service.AppVersionService
}

func NewAppUpgradeHandle(appVersionSvc *service.AppVersionService) *AppUpgradeHandle {
	return &AppUpgradeHandle{appVersionSvc: appVersionSvc}
}

func (i AppUpgradeHandle) ModuleName() (en string, zh string) {
	return "app", "APP"
}

// @Summary		获取最新版本
// @Description	获取可用于升级的最新版本信息
// @Tags			APP
// @Produce		json
// @Param			apkType	query		int	false	"安装包类型：1-ios、2-android"
// @Success		200		{object}	router.Resp	"操作成功"
// @Router			/api/v1/app/upgrade [GET]
func (i AppUpgradeHandle) Upgrade(ctx *gin.Context) {
	var apkType *int32
	if v := ctx.Query("apkType"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			router.BadRequest(ctx, "参数绑定失败")
			return
		}
		t := int32(n)
		apkType = &t
	}

	latest, err := i.appVersionSvc.Latest(ctx.Request.Context(), apkType)
	if err != nil {
		router.Fail(ctx, err)
		return
	}
	router.Success(ctx, latest)
}
