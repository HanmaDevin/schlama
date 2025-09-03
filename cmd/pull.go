package cmd

import (
	"fmt"
	"regexp"

	"github.com/HanmaDevin/schlama/config"
	"github.com/HanmaDevin/schlama/ollama"
	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull a model.",
	Long:  `This command pulls a model from the Ollama server. If the model is already present, it will do nothing.`,
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
				fmt.Printf("%s Model not found locally. Pulling model %s...\n", Yellow("[Hint]"), model)
				err := ollama.PullModel(model)
				if err != nil {
					fmt.Println(Red("[Error] ") + err.Error())
					fmt.Println(Yellow("[Hint]") + " Here is a short list of available models:")
					models := ollama.ListModels()
					table := ollama.CreateTable(models, 25)
					fmt.Println(table)
					return
				}
				fmt.Printf("%s %s pulled successfully.\n", Yellow("[Hint]"), model)

				cfg := config.Config{
					Model: model,
				}
				config.WriteConfig(cfg)
				out := fmt.Sprintf("%s Current Model: %s", Green("[Msg]"), cfg.Model)
				fmt.Println(out)
				return
			} else {
				fmt.Println(Yellow("[Hint]") + " Model already present locally. No need to pull again.")
				return
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(pullCmd)
}
