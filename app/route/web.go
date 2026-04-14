package route

import (
	"bit-labs.cn/owl"
	"bit-labs.cn/owl/contract/foundation"
	"github.com/gin-gonic/gin"
)

// InitWeb 注册静态 Web 路由。
// /storage 对应本地存储根目录 ./storage，用于访问上传文件。
func InitWeb(app foundation.Application) {
	err := app.Invoke(func(engine *gin.Engine) {
		engine.Static("/storage", "./storage")
	})
	owl.PanicIf(err)
}
