package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"bit-labs.cn/owl-admin/app/provider/jwt"

	"bit-labs.cn/owl/provider/router"
	"bit-labs.cn/owl/utils"
	"github.com/casbin/casbin/v3"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

func PermissionCheck(enforcer *casbin.SyncedEnforcer, jwtService *jwt.JWTService) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		url := ctx.Request.URL.Path
		method := ctx.Request.Method

		routes := router.GetAllRoutes()

		var findRoute bool
		for _, api := range routes {
			if api.Method == method && utils.UrlIsEq(url, api.Path) {
				findRoute = true
				accessLevel := api.AccessLevel
				if accessLevel == router.AccessPublic {
					ctx.Next()
					return
				}
				if accessLevel == router.AccessSuperAdmin || accessLevel == router.AccessAuthorized || accessLevel == router.AccessAuthenticated {

					token := ctx.Request.Header.Get("Authorization")

					user, err := jwtService.ParseToken(strings.Replace(token, "Bearer ", "", -1))
					if err != nil {
						_ = ctx.AbortWithError(http.StatusUnauthorized, errors.New("未授权的访问"))
						return
					}

					ctx.Set("user", user)
					reqCtx := ctx.Request.Context()
					reqCtx = context.WithValue(reqCtx, "user_id", user.ID)
					reqCtx = context.WithValue(reqCtx, "username", user.Username)
					reqCtx = context.WithValue(reqCtx, "nickname", user.Nickname)
					reqCtx = context.WithValue(reqCtx, "roles", user.Roles)
					ctx.Request = ctx.Request.WithContext(reqCtx)
					if user.IsSuperAdmin {
						ctx.Next()
						return
					}
					// 只有系统管理员才能访问
					if accessLevel == router.AccessSuperAdmin && !user.IsSuperAdmin {
						_ = ctx.AbortWithError(http.StatusForbidden, errors.New("未授权的访问"))
						return
					}

					// 需要登录，token 有效则认为已经登录
					if accessLevel == router.AccessAuthenticated {
						ctx.Next()
						return
					}

					// 需要授权
					if accessLevel == router.AccessAuthorized {
						permissionKey := api.Permission
						enforcer.LoadPolicy()
						can, err := enforcer.Enforce(cast.ToString(user.ID), permissionKey)
						if err != nil {
							_ = ctx.AbortWithError(http.StatusInternalServerError, err)
							return
						}
						if !can {
							_ = ctx.AbortWithError(http.StatusForbidden, errors.New("未授权的访问"))
							return
						}
					}
				}
			}
		}
		if !findRoute {
			_ = ctx.AbortWithError(http.StatusNotFound, errors.New("未找到匹配的路由"))
		}
	}
}
