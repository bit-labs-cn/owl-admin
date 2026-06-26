package middleware

import (
	"context"
	"errors"
	"strings"
	"time"

	"bit-labs.cn/owl-admin/app/repository"
	"bit-labs.cn/owl-admin/app/service"
	"bit-labs.cn/owl/provider/captcha"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const loginTrustedDeviceTTL = 7 * 24 * time.Hour

// LoginCaptchaGuard 登录风险验证码中间件：新设备、IP 段变化或信任过期时要求图形验证码。
func LoginCaptchaGuard(userSvc *service.UserService, trustedDeviceRepo repository.TrustedDeviceRepositoryInterface, captchaSvc *captcha.Service) gin.HandlerFunc {
	need := func(c *gin.Context, payload captcha.GuardPayload) (bool, error) {
		if payload.Username == "" {
			return false, nil
		}

		user, err := userSvc.GetUserByName(c.Request.Context(), payload.Username)
		if err != nil {
			if errors.Is(err, service.ErrLogin) {
				return false, nil
			}
			return false, err
		}

		return evaluateLoginRisk(
			c.Request.Context(),
			trustedDeviceRepo,
			user.ID,
			payload.DeviceID,
			c.ClientIP(),
		), nil
	}

	return captcha.Guard(captchaSvc, need)
}

func evaluateLoginRisk(ctx context.Context, repo repository.TrustedDeviceRepositoryInterface, userID uint, deviceID, ip string) bool {
	if deviceID == "" {
		return true
	}

	device, err := repo.WithContext(ctx).FindByUserAndDevice(userID, deviceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return true
		}
		return true
	}

	if !sameIPSegment(device.LastIP, ip) {
		return true
	}

	if time.Since(device.VerifiedAt) > loginTrustedDeviceTTL {
		return true
	}

	return false
}

func sameIPSegment(a, b string) bool {
	if a == b {
		return true
	}
	partsA := strings.Split(a, ".")
	partsB := strings.Split(b, ".")
	if len(partsA) == 4 && len(partsB) == 4 {
		return partsA[0] == partsB[0] && partsA[1] == partsB[1] && partsA[2] == partsB[2]
	}
	return false
}
