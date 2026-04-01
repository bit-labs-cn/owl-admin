package route

import (
	"time"

	"bit-labs.cn/owl"
	"bit-labs.cn/owl-admin/app/handle/oauth"
	v1 "bit-labs.cn/owl-admin/app/handle/v1"
	middleware2 "bit-labs.cn/owl-admin/app/middleware"
	"bit-labs.cn/owl-admin/app/provider/jwt"
	"bit-labs.cn/owl-admin/app/service"
	"bit-labs.cn/owl/contract/foundation"
	"bit-labs.cn/owl/contract/log"
	"bit-labs.cn/owl/provider/router"
	"bit-labs.cn/owl/provider/router/middleware"
	"bit-labs.cn/owl/provider/storage"
	"github.com/casbin/casbin/v3"
	"github.com/gin-gonic/gin"
)

var userMenu, roleMenu, menuMenu, apiMenu, deptMenu, dictMenu, positionMenu, userGroupMenu *router.Menu

func InitMenu() []*router.Menu {
	return []*router.Menu{
		{
			Path: "/system/rbac",
			Name: "SystemRbac",
			Meta: router.Meta{
				Title:      "用户权限",
				Icon:       "ep:user",
				ShowParent: true,
				ShowLink:   true,
			},
			MenuType: router.MenuTypeDir,
			Children: []*router.Menu{
				userMenu,
				roleMenu,
				userGroupMenu,
				deptMenu,
				positionMenu,
			},
		},
		{
			Path: "/system/logs",
			Name: "SystemLogs",
			Meta: router.Meta{
				Title:      "日志管理",
				Icon:       "ep:monitor",
				ShowParent: true,
				ShowLink:   true,
			},
			MenuType: router.MenuTypeDir,
			Children: []*router.Menu{
				{
					Path: "/system/login-log/index",
					Name: "SystemLoginLog",
					Meta: router.Meta{
						Title:      "登录日志",
						Icon:       "material-symbols-light:login-outline-rounded",
						ShowParent: true,
						ShowLink:   true,
					},
					MenuType: router.MenuTypeMenu,
				},
				{
					Path: "/system/operation-log/index",
					Name: "SystemMonitorOperationLog",
					Meta: router.Meta{
						Title:      "操作日志",
						Icon:       "twemoji:hammer-and-wrench",
						ShowParent: true,
						ShowLink:   true,
					},
					MenuType: router.MenuTypeMenu,
				},
			},
		},
		{
			Path: "/system",
			Name: "System",
			Meta: router.Meta{
				Title:      "系统管理",
				Icon:       "ep:setting",
				ShowParent: true,
				ShowLink:   true,
			},
			MenuType: router.MenuTypeDir,
			Children: []*router.Menu{
				menuMenu,
				apiMenu,
				dictMenu,
			},
		},
	}
}

func InitApi(app foundation.Application, appName string) {

	err := app.Invoke(func(
		userHandle *v1.UserHandle,
		roleHandle *v1.RoleHandle,
		userGroupHandle *v1.UserGroupHandle,
		apiHandle *v1.ApiHandle,
		fileHandle *storage.FileHandle,
		menuHandle *v1.MenuHandle,
		dictHandle *v1.DictHandle,
		deptHandle *v1.DeptHandle,
		positionHandle *v1.PositionHandle,
		areaHandle *v1.AreaHandle,
		logHandle *v1.LogHandle,
		appUpgradeHandle *v1.AppUpgradeHandle,
		enforcer *casbin.SyncedEnforcer,
		oauthHandle *oauth.Handle,
		engine *gin.Engine,
		jwtSvc *jwt.JWTService,
		logService *service.LogService,
		log log.Logger,
	) {

		gv1 := engine.Group("/api/v1", middleware2.PermissionCheck(enforcer, jwtSvc))
		gv1.Use(middleware2.OperationLog(logService))

		// file
		{
			r := router.NewRouteInfoBuilder(appName, fileHandle, gv1, router.MenuOption{
				ComponentName: "SystemFile",
				Path:          "/system/file/index",
				Icon:          "ep:upload",
			})
			r.Post("/files/upload", router.AccessAuthenticated, fileHandle.Upload).Name("上传文件").WithoutOperateLog().Build()
		}

		// user
		{
			r := router.NewRouteInfoBuilder(appName, userHandle, gv1, router.MenuOption{
				ComponentName: "SystemUser",
				Path:          "/system/user/index",
				Icon:          "ep:user",
			})

			r.Use(middleware.RateLimiter(time.Second*1, 2)).Post("/users/login", router.AccessPublic, userHandle.Login).Name("用户登录").WithoutOperateLog().Build()

			r.Put("/users/me/password", router.AccessAuthenticated, userHandle.ChangePassword).Name("修改我的密码").Build()
			r.Get("/users/me/menus", router.AccessAuthenticated, userHandle.GetMyMenus).Name("我的菜单").Build()
			r.Get("/users/me/permissions", router.AccessAuthenticated, userHandle.GetMyPermissions).Name("我的权限").Build()
			r.Get("/users/me", router.AccessAuthenticated, userHandle.Me).Name("我的信息").Build()

			r.Post("/users", router.AccessAuthorized, userHandle.Create).Name("创建用户").Build()
			r.Delete("/users/:id", router.AccessAuthorized, userHandle.Delete).Name("删除用户").Build()

			r.Put("/users/:id", router.AccessAuthorized, userHandle.Update).Deps(
				[]router.Dep{
					{Handler: userHandle, Method: userHandle.Detail},
				}...,
			).Name("更新用户").Build()

			r.Put("/users/:id/status", router.AccessAuthorized, userHandle.ChangeStatus).Name("启用，禁用用户").Build()

			r.Get("/users", router.AccessAuthorized, userHandle.Retrieve).Deps(
				[]router.Dep{
					{Handler: deptHandle, Method: deptHandle.Retrieve},
				}...,
			).Name("分页获取用户").Build()

			r.Get("/users/:id", router.AccessAuthorized, userHandle.Detail).Name("获取用户详情").Build()
			r.Put("/users/:id/reset", router.AccessSuperAdmin, userHandle.ResetPassword).Name("重置用户密码").Build()
			r.Put("/users/:id/avatar", router.AccessAuthorized, userHandle.ChangeAvatar).Name("修改用户头像").Build()

			r.Post("/users/:id/roles", router.AccessAuthorized, userHandle.AssignRolesToUser).Deps(
				[]router.Dep{
					{Handler: userHandle, Method: userHandle.GetRoleIdsByUserId},
					{Handler: roleHandle, Method: roleHandle.RoleOptions},
				}...,
			).Name("分配角色给用户").Build()

			r.Get("/users/:id/roles", router.AccessAuthorized, userHandle.GetRoleIdsByUserId).Name("获取用户角色").Build()

			userMenu = r.GetMenu()
		}

		// role
		{
			r := router.NewRouteInfoBuilder(appName, roleHandle, gv1, router.MenuOption{
				ComponentName: "SystemRole",
				Path:          "/system/role/index",
				Icon:          "fa-solid:users",
			})

			r.Post("/roles", router.AccessAuthorized, roleHandle.Create).Name("创建角色").Build()
			r.Delete("/roles/:id", router.AccessAuthorized, roleHandle.Delete).Name("删除角色").Build()
			r.Put("/roles/:id", router.AccessAuthorized, roleHandle.Update).Name("更新角色").Build()
			r.Put("/roles/:id/status", router.AccessAuthorized, roleHandle.ChangeStatus).Name("禁用，启用角色").Build()

			r.Get("/roles", router.AccessAuthorized, roleHandle.Retrieve).Name("角色列表").Build()
			r.Get("/roles/:id", router.AccessAuthorized, roleHandle.Detail).Name("获取角色详情").Build()
			r.Get("/roles/:id/menu-ids", router.AccessAuthorized, roleHandle.GetRoleMenuIDs).Name("获取角色拥有的菜单").Build()
			r.Get("/roles/:id/users", router.AccessAuthorized, roleHandle.GetUsersByRoleID).Name("获取角色下的用户").Build()

			r.Get("/roles-options", router.AccessAuthenticated, roleHandle.RoleOptions).Name("所有角色(id,name)").Build()

			deps := []router.Dep{
				{Handler: roleHandle, Method: roleHandle.GetRoleMenuIDs},
				{Handler: menuHandle, Method: menuHandle.Assignable},
			}
			r.Put("/roles/:id/menus", router.AccessAuthorized, roleHandle.AssignMenusToRole).Deps(deps...).Name("分配菜单给角色").Build()
			roleMenu = r.GetMenu()
		}

		// user-group
		{
			r := router.NewRouteInfoBuilder(appName, userGroupHandle, gv1, router.MenuOption{
				ComponentName: "SystemUserGroup",
				Path:          "/system/user-group/index",
				Icon:          "ri:group-line",
			})

			r.Post("/user-groups", router.AccessAuthorized, userGroupHandle.Create).Name("创建用户组").Build()
			r.Delete("/user-groups/:id", router.AccessAuthorized, userGroupHandle.Delete).Name("删除用户组").Build()
			r.Put("/user-groups/:id", router.AccessAuthorized, userGroupHandle.Update).Name("更新用户组").Build()
			r.Put("/user-groups/:id/status", router.AccessAuthorized, userGroupHandle.ChangeStatus).Name("启用，禁用用户组").Build()

			r.Get("/user-groups", router.AccessAuthorized, userGroupHandle.Retrieve).Name("用户组列表").Build()
			r.Get("/user-groups-options", router.AccessAuthenticated, userGroupHandle.Options).Name("所有用户组(id,name)").Build()
			r.Get("/user-groups/:id/users", router.AccessAuthorized, userGroupHandle.GetUsersByGroupID).Name("获取用户组下的用户").Build()
			r.Get("/user-groups/:id/user-ids", router.AccessAuthorized, userGroupHandle.GetUserIDsByGroupID).Name("获取用户组的用户ID列表").Build()
			r.Put("/user-groups/:id/users", router.AccessAuthorized, userGroupHandle.AssignUsersToGroup).Name("为用户组分配用户").Build()

			userGroupMenu = r.GetMenu()
		}

		// api(permission)
		{
			r := router.NewRouteInfoBuilder(appName, apiHandle, gv1, router.MenuOption{
				ComponentName: "SystemApi",
				Path:          "/system/api/index",
				Icon:          "ep:user",
			})
			r.Get("/api", router.AccessAuthorized, apiHandle.GetAll).Name("查询所有接口").Build()

			apiMenu = r.GetMenu()
		}

		// menu
		{
			r := router.NewRouteInfoBuilder(appName, menuHandle, gv1, router.MenuOption{
				ComponentName: "SystemMenu",
				Path:          "/system/menu/index",
				Icon:          "ep:menu",
			})

			r.Get("/menus/assignable", router.AccessAuthorized, menuHandle.Assignable).Name("查询可分配的菜单").Description("查询可分配的菜单（包含按钮）").Build()
			r.Get("/menus", router.AccessAuthorized, menuHandle.Menus).Name("菜单列表").Build()

			menuMenu = r.GetMenu()
		}

		// dictionary
		{
			r := router.NewRouteInfoBuilder(appName, dictHandle, gv1, router.MenuOption{
				ComponentName: "SystemDict",
				Path:          "/system/dict/index",
				Icon:          "ep:menu",
			})

			r.Post("/dict", router.AccessSuperAdmin, dictHandle.Create).Name("创建字典").Build()
			r.Delete("/dict/:id", router.AccessSuperAdmin, dictHandle.Delete).Name("删除字典").Build()
			r.Put("/dict/:id", router.AccessSuperAdmin, dictHandle.Update).Name("更新字典").Build()
			r.Get("/dict", router.AccessSuperAdmin, dictHandle.Retrieve).Name("字典列表").Build()
			r.Get("/dict/types/:type/items", router.AccessAuthenticated, dictHandle.GetItemsByType).Name("按类型获取字典项").Description("仅需登录，返回启用字典项").Build()

			r.Post("/dict/:id/item", router.AccessSuperAdmin, dictHandle.CreateItem).Name("新增字典项").Build()
			r.Put("/dict/:id/item/:itemID", router.AccessSuperAdmin, dictHandle.UpdateItem).Name("更新字典项").Build()
			r.Get("/dict/:id/item", router.AccessSuperAdmin, dictHandle.RetrieveItems).Name("获取字典列表").Build()
			r.Delete("/dict/:id/item/:itemID", router.AccessSuperAdmin, dictHandle.DeleteItem).Name("删除字典项").Build()

			dictMenu = r.GetMenu()
		}

		// area
		{
			r := router.NewRouteInfoBuilder(appName, areaHandle, gv1, router.MenuOption{
				ComponentName: "SystemArea",
				Path:          "/system/area/index",
				Icon:          "ep:location",
			})
			r.Post("/areas/all", router.AccessAuthenticated, areaHandle.RetrieveAll).Name("查询省市区").Description("查询所有省市区数据（平铺，不构建树）").Build()
		}

		// dept
		{
			r := router.NewRouteInfoBuilder(appName, deptHandle, gv1, router.MenuOption{
				ComponentName: "Dept",
				Path:          "/system/dept/index",
				Icon:          "ep:menu",
			})

			r.Post("/dept", router.AccessAuthorized, deptHandle.Create).Name("新增部门").Build()
			r.Delete("/dept/:id", router.AccessAuthorized, deptHandle.Delete).Name("删除部门").Build()
			r.Put("/dept/:id", router.AccessAuthorized, deptHandle.Update).Name("更新部门").Build()
			r.Get("/dept", router.AccessAuthorized, deptHandle.Retrieve).Name("获取部门列表").Build()
			r.Get("/dept/:id/users", router.AccessAuthorized, roleHandle.GetRoleMenuIDs).Name("获取部门下的用户").Build()

			deptMenu = r.GetMenu()
		}

		// position
		{
			r := router.NewRouteInfoBuilder(appName, positionHandle, gv1, router.MenuOption{
				ComponentName: "SystemPosition",
				Path:          "/system/position/index",
				Icon:          "ep:user",
			})

			r.Post("/position", router.AccessAuthorized, positionHandle.Create).Name("创建岗位").Build()
			r.Delete("/position/:id", router.AccessAuthorized, positionHandle.Delete).Name("删除岗位").Build()
			r.Put("/position/:id", router.AccessAuthorized, positionHandle.Update).Name("更新岗位").Build()
			r.Put("/position/:id/status", router.AccessAuthorized, positionHandle.ChangeStatus).Name("修改岗位状态").Build()
			r.Get("/position", router.AccessAuthorized, positionHandle.Retrieve).Name("岗位列表").Build()
			r.Get("/position-options", router.AccessAuthenticated, positionHandle.Options).Name("所有岗位(id,name)").Build()

			positionMenu = r.GetMenu()
		}
		// monitor
		{
			r := router.NewRouteInfoBuilder(appName, logHandle, gv1, router.MenuOption{
				ComponentName: "SystemMonitor",
				Path:          "/system/monitor/index",
				Icon:          "ep:monitor",
			})
			r.Post("/monitor/login-logs", router.AccessAuthorized, logHandle.LoginLogs).Name("登录日志").WithoutOperateLog().Build()
			r.Post("/monitor/operation-logs", router.AccessAuthorized, logHandle.OperationLogs).Name("操作日志").WithoutOperateLog().Build()
		}

		// app upgrade
		{
			r := router.NewRouteInfoBuilder(appName, appUpgradeHandle, gv1, router.MenuOption{})
			r.Get("/app/upgrade", router.AccessPublic, appUpgradeHandle.Upgrade).Name("获取最新版本").Build()
		}
		// oauth

		{
			r := router.NewRouteInfoBuilder(appName, oauthHandle, gv1, router.MenuOption{})
			r.Get("/oauth/:provider/login", router.AccessPublic, oauthHandle.Login).Name("第三方登录").WithoutOperateLog().Build()
			r.Get("/oauth/:provider/callback", router.AccessPublic, oauthHandle.Callback).Name("第三方登录回调").WithoutOperateLog().Build()
		}
	})
	owl.PanicIf(err)
}
