package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/HanmaDevin/schlama/config"
	"github.com/HanmaDevin/schlama/ollama"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var Green = color.New(color.FgGreen).SprintFunc()
var Red = color.New(color.FgRed).SprintFunc()
var Yellow = color.New(color.FgYellow).SprintFunc()
var Cyan = color.New(color.FgCyan).SprintFunc()

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "schlama",
	Short: "A better ollama user interface.",
	Long:  `Schlama is a CLI and a web-chat app, depending on what you perfer, which allows for easy communication with local LLMs. It allows file/directory input and images are also supported (Only works with multimodal models). Basically an easier way to chat with local LLMs and install new ones.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	if ollama.IsOllamaRunning() {
		var home, _ = os.UserHomeDir()
		var config_Path string = filepath.Dir(home + "/.config/schlama/")
		if _, err := os.Stat(config_Path); os.IsNotExist(err) {
			err := os.MkdirAll(config_Path, 0755)
			if err != nil {
				fmt.Println(Red("[Error] ") + "Creating config directory: '~/.config/schlama/config.yaml' did not work!")
				os.Exit(-1)
			}
		}

		if _, err := os.Stat(config_Path + "/config.yaml"); os.IsNotExist(err) {
			config.WriteConfig(config.Config{
				Model: "",
			})
		}
	} else {
		fmt.Println(Red("[Error] ") + "Ollama is not running.")
		fmt.Println(Red("[Error]") + " Please start ollama first.")
		fmt.Println(Yellow("[Hint]") + " You can start ollama with the command: 'ollama serve'")
		fmt.Println(Yellow("[Hint]") + " Or you can install ollama on linux with the command: 'curl -fsSL https://ollama.com/install.sh | sh'")
		fmt.Println(Yellow("[Hint]") + " Visis https://ollama.com/download for more information.")
		os.Exit(1)
	}
}
