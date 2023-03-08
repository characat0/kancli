package cmd

import (
	"github.com/spf13/cobra"
	"kancli/cmd/kancli"
	"os"
)

var Root = &cobra.Command{
	Short: "kancli",
	RunE: func(_ *cobra.Command, _ []string) error {
		return kancli.Run()
	},
}

func Execute() {
	err := Root.Execute()
	if err != nil {
		os.Exit(1)
	}
}
