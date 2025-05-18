package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:  "goHack",
	Long: "使用goHack进行安全测试",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("test cobra!")
	},
}

var (
	proxy string
)

func init() {
	rootCmd.Flags().StringVarP(&proxy, "proxy", "p", "", "设置代理服务器")
}
func Execute() error {
	return rootCmd.Execute()

}
