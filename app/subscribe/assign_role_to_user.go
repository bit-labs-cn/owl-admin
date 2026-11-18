package subscribe

import (
	"context"

	"bit-labs.cn/owl-admin/app/event"
	"bit-labs.cn/owl-admin/app/service"
	contractEvent "bit-labs.cn/owl/contract/event"
	"bit-labs.cn/owl/contract/log"
	"github.com/casbin/casbin/v3"
	"github.com/spf13/cast"
)

// subscribeAssignRoleToUser 监听用户角色分配，把分组策略同步到 Casbin。
// 先按用户清空再批量写入，重复投递也会收敛到同一结果。
func subscribeAssignRoleToUser(bus contractEvent.EventBus, enforcer *casbin.SyncedEnforcer, logger log.Logger) error {
	return bus.Subscribe(event.AssignRoleToUser, contractEvent.AsHandler(func(ctx context.Context, req *service.AssignRoleToUser) error {
		userID := cast.ToString(req.UserID)
		rules := make([][]string, 0, len(req.RoleIDs))
		for _, roleID := range req.RoleIDs {
			rules = append(rules, []string{userID, roleID})
		}
		_, _ = enforcer.RemoveFilteredGroupingPolicy(0, userID)
		if _, err := enforcer.AddGroupingPolicies(rules); err != nil {
			logger.Error("Casbin AddGroupingPolicies failed: ", err)
			return err
		}
		return nil
	}))
}
