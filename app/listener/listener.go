package listener

import (
	"bit-labs.cn/owl"
	"bit-labs.cn/owl-admin/app/event"
	"bit-labs.cn/owl-admin/app/service"
	"bit-labs.cn/owl/contract/foundation"
	"bit-labs.cn/owl/contract/log"
	"bit-labs.cn/owl/provider/router"
	"github.com/asaskevich/EventBus"
	"github.com/casbin/casbin/v2"
	"github.com/spf13/cast"
)

func Init(app foundation.Application) {
	err := app.Invoke(func(bus EventBus.Bus, enforcer casbin.IEnforcer, log log.Logger, menuRepo *router.MenuRepository) {
		bus.Subscribe(event.AssignRoleToUser, func(req *service.AssignRoleToUser) {
			userID := cast.ToString(req.UserID)
			var rules [][]string
			for _, roleID := range req.RoleIDs {
				rules = append(rules, []string{userID, roleID})
			}
			_, err := enforcer.RemoveFilteredGroupingPolicy(0, userID)
			log.Error(err)
			_, err = enforcer.AddGroupingPolicies(rules)
			log.Error(err)
		})

		bus.Subscribe(event.AssignMenuToRole, func(req *service.AssignMenuToRole) {
			roleID := cast.ToString(req.RoleID)
			permissions := menuRepo.GetPermissionsByMenuIDs(req.MenuIDs...)
			var rules [][]string
			for _, permission := range permissions {
				rules = append(rules, []string{roleID, permission})
			}

			_, err := enforcer.RemoveFilteredPolicy(0, roleID)
			if err != nil {
				log.Error(err)
			}
			_, err = enforcer.AddPolicies(rules)
			if err != nil {
				log.Error(err)
			}
		})
	})
	owl.PanicIf(err)
}
