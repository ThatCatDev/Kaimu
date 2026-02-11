package commands

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kaimu",
	Short: "Kaimu - Project management tool for software teams",
	Long:  `Kaimu is a project management tool similar to Jira or Linear, built with Go and GraphQL.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
