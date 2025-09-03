package cmd

import (
	"fmt"
	"regexp"

	"github.com/HanmaDevin/schlama/ollama"
	"github.com/spf13/cobra"
)

// rmCmd represents the rm command
var rmCmd = &cobra.Command{
	Use:   "rm",
	Short: "Remove a model.",
	Long:  `Remove a model.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println(Red("[Error]") + " Please provide the name of the model to remove.")
			return
		} else if len(args) > 1 {
			fmt.Println(Red("[Error]") + " Too many arguments provided. Please provide only one model to remove.")
			return
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
				fmt.Printf("%s Model %s not found locally. Cannot remove a model that does not exist.\n", Red("[Error]"), model)
				return
			} else {
				fmt.Printf("%s Removing model %s...\n", Yellow("[Hint]"), model)
				err := ollama.RemoveModel(model)
				if err != nil {
					fmt.Println(Red("[Error] ") + err.Error())
					return
				}
				fmt.Printf("%s Model %s removed successfully.\n", Green("[Msg]"), model)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(rmCmd)
}
