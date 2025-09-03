package cmd

import (
	"fmt"
	"regexp"

	"github.com/HanmaDevin/schlama/ollama"
	"github.com/spf13/cobra"
)

// showCmd represents the show command
var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show informataion about a model",
	Long:  `Show informataion about a model`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.Help()
		} else {
			// Check if the model is present in the local models
			nameReg := regexp.MustCompile(`[\w\-]+\d?\.?\d?`)
			name := nameReg.FindString(args[0])

			labelReg := regexp.MustCompile(`:\w+\.?(\w+)?`)
			label := labelReg.FindString(args[0])
			if label == "" {
				label = ":latest"
			}

			model := name + label

			if !ollama.IsModelPresent(model) {
				fmt.Println(Red("[Error]") + " Model not found. No information available.")
				return
			} else {
				info, err := ollama.Show(model)
				if err != nil {
					fmt.Println(Red("[Error]") + " Unable to retrieve model information: " + err.Error())
					return
				}
				fmt.Println(info)
				return
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(showCmd)
}
