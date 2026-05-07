package cmd

import (
	"os"

	"bit-labs.cn/owl/utils"
	"github.com/spf13/cobra"
)

var GenPwd = &cobra.Command{
	Use:   "admin:gen-pwd",
	Short: "生成密码",
	Run: func(cmd *cobra.Command, args []string) {
		utils.PrintBlue("密码：", os.Args[2])
		utils.PrintBlue("密文", utils.BcryptHash(os.Args[2]))
	},
}
