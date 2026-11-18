package service

import (
	"context"

	"github.com/spf13/cast"
)

func currentPrincipal(ctx context.Context) (Principal, error) {
	id := cast.ToUint(ctx.Value("user_id"))
	if id == 0 {
		return Principal{}, NotAllowed("未登录")
	}
	return Principal{Type: 1, ID: id}, nil
}
