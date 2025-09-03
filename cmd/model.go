package cmd

import (
	"fmt"

	"github.com/HanmaDevin/schlama/config"
	"github.com/spf13/cobra"
)

// modelCmd represents the model command
var modelCmd = &cobra.Command{
	Use:   "model",
	Short: "Show the currently selected model.",
	Long:  `Show the currently selected model.`,
	Run: func(cmd *cobra.Command, args []string) {
		body := config.ReadConfig()
		out := fmt.Sprintf(Green("[Msg]")+" Current Model: %s", body.Model)
		fmt.Println(out)
	},
}

func init() {
	rootCmd.AddCommand(modelCmd)
}
