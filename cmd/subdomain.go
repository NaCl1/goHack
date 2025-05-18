package cmd

import "github.com/spf13/cobra"

var subdomainCmd = &cobra.Command{
	Use:   "subdomain",
	Short: "子域名搜集",
	Run: func(cmd *cobra.Command, args []string) {

	},
}
var (
	targets []string
)

func init() {
	subdomainCmd.PersistentFlags().StringSliceVarP(&targets, "target", "t", nil, "目标域名或ip")

	rootCmd.AddCommand(subdomainCmd)
}
