package cmd

import (
	"github.com/HanmaDevin/schlama/chat"
	"github.com/spf13/cobra"
)

// chatCmd represents the chat command
var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Chat with local LLMs",
	Long:  `Opens your browser with a chat interface for local LLMs.`,
	Run: func(cmd *cobra.Command, args []string) {
		chat.Start()
	},
}

func init() {
	rootCmd.AddCommand(chatCmd)
}
