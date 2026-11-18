package subscribe

import (
	"context"

	"bit-labs.cn/owl-admin/app/event"
	"bit-labs.cn/owl-admin/app/service"
	contractEvent "bit-labs.cn/owl/contract/event"
	"bit-labs.cn/owl/contract/log"
	"bit-labs.cn/owl/provider/router"
	"github.com/casbin/casbin/v3"
	"github.com/spf13/cast"
)

// subscribeAssignMenuToRole 监听菜单分配给角色，把菜单权限同步成 Casbin 策略。
func subscribeAssignMenuToRole(
	bus contractEvent.EventBus,
	enforcer *casbin.SyncedEnforcer,
	menuRepo *router.MenuRepository,
	logger log.Logger,
) error {
	return bus.Subscribe(event.AssignMenuToRole, contractEvent.AsHandler(func(ctx context.Context, req *service.AssignMenuToRole) error {
		roleID := cast.ToString(req.RoleID)
		permissions := menuRepo.GetPermissionsByMenuIDs(req.MenuIDs...)
		rules := make([][]string, 0, len(permissions))
		for _, permission := range permissions {
			rules = append(rules, []string{roleID, permission})
		}
		_, _ = enforcer.RemoveFilteredPolicy(0, roleID)
		if _, err := enforcer.AddPolicies(rules); err != nil {
			logger.Error("Casbin AddPolicies failed: ", err)
			return err
		}
		return nil
	}))
}
