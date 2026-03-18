package middleware

import (
	"bytes"
	"io"
	"time"

	"bit-labs.cn/owl/provider/router"

	"bit-labs.cn/owl-admin/app/model"
	"bit-labs.cn/owl-admin/app/service"
	"github.com/gin-gonic/gin"
)

func OperationLog(logSvc *service.LogService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "GET" {
			c.Next()
			return
		}

		start := time.Now()
		var body string
		if c.Request.Body != nil {
			b, _ := io.ReadAll(c.Request.Body)
			body = string(b)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(b))
		}
		c.Next()
		cost := time.Since(start).Milliseconds()
		status := c.Writer.Status()
		var user *model.User
		if v, ok := c.Get("user"); ok {
			user = v.(*model.User)
		}
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		uType := "user"
		uId := 0
		uName := ""
		if user != nil {
			if user.IsSuperAdmin {
				uType = "super_admin"
			}
			uId = int(user.ID)
			uName = user.Username
		}

		var apiName string

		for _, r := range router.GetAllRoutes() {
			if r.Path == path && r.Method == c.Request.Method {
				if !r.ShouldOperateLog() {
					return
				}
				apiName = r.Name
				break
			}
		}

		if apiName == "" {
			apiName = "未命名接口"
		}

		_ = logSvc.CreateOperationLog(c.Request.Context(), &service.CreateOperationLogReq{
			UserId:    uId,
			UserName:  uName,
			UserType:  uType,
			ApiName:   apiName,
			Method:    c.Request.Method,
			Path:      path,
			Status:    status,
			CostMs:    int(cost),
			Ip:        c.ClientIP(),
			UserAgent: c.GetHeader("User-Agent"),
			ReqBody:   body,
		})
	}
}
