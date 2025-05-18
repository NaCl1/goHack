package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var (
	testargs string
)

func test(url string) {
	fmt.Println(url)
}

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "test of goHack",
	Run: func(cmd *cobra.Command, args []string) {
		test(testargs)
	},
}

func init() {
	testCmd.Flags().StringVarP(&testargs, "testargs", "t", "", "testargs")
	rootCmd.AddCommand(testCmd)
}
