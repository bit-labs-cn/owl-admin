package main

import (
	"bit-labs.cn/owl"
	admin "bit-labs.cn/owl-admin/app"
)

func main() {
	var subApps = []owl.SubApp{
		&admin.SubAppAdmin{},
	}
	owl.NewApp(subApps...).WebShell()
}
