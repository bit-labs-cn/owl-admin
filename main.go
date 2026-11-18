package main

import (
	"bit-labs.cn/owl"
	admin "bit-labs.cn/owl-admin/app"
	notification "bit-labs.cn/owl-admin/app-notification"
)

func main() {
	var subApps = []owl.SubApp{
		&admin.SubAppAdmin{},
		&notification.SubAppNotification{},
	}
	owl.NewApp(subApps...).WebShell()
}
