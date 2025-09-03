package cmd

import (
	"fmt"

	"github.com/HanmaDevin/schlama/ollama"
	"github.com/spf13/cobra"
)

var limit int
var local bool

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available models.",
	Long:  `List gets all the available models from ollama.com and displays them.`,
	Run: func(cmd *cobra.Command, args []string) {
		if local {
			ollama.ListLocalModels()
			return
		}
		models := ollama.ListModels()
		table := ollama.CreateTable(models, int(limit))
		fmt.Println(table)
	},
}

func init() {
	listCmd.Flags().IntVarP(&limit, "limit", "l", 25, "Limit the output.")
	listCmd.Flags().BoolVar(&local, "local", false, "List local models.")
	rootCmd.AddCommand(listCmd)
}
