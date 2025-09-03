package cmd

import (
	"fmt"
	"regexp"

	"github.com/HanmaDevin/schlama/config"
	"github.com/HanmaDevin/schlama/ollama"
	"github.com/spf13/cobra"
)

// selectCmd represents the select command
var selectCmd = &cobra.Command{
	Use:   "select",
	Short: "Select which model to chat with.",
	Long:  `This command sets the model to chat with.`,
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
				fmt.Println(Red("[Error]") + " Model not found. Make sure to pull the model first using 'schlama pull <model_name>' command.")
				return
			}

			cfg := config.Config{
				Model: model,
			}
			config.WriteConfig(cfg)
			out := fmt.Sprintf("%s Current Model: %s", Green("[Msg]"), cfg.Model)
			fmt.Println(out)
		}
	},
}

func init() {
	rootCmd.AddCommand(selectCmd)
}
