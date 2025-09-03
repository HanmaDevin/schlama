package cmd

import (
	"os"
	"regexp"
	"strings"

	"github.com/HanmaDevin/schlama/config"
	"github.com/HanmaDevin/schlama/ollama"
	"github.com/chzyer/readline"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run an interactive shell session",
	Long:  `Run an interactive shell seesion.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println(Green(">>> [Msg]") + " Starting interactive shell session...")
		if len(args) == 0 {
			cmd.Println(Red(">>> [Error]") + " Please provide a model as an argument.")
			return
		} else {
			runInteractiveShell(args[0])
		}
	},
}

func runInteractiveShell(model string) {
	cfg := config.ReadConfig()

	l, err := readline.NewEx(&readline.Config{
		Prompt:          Cyan(">>> "),
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		println(Red(">>> [Error]")+"Failed to create readline instance:", err.Error())
		return
	}
	defer l.Close()
	l.CaptureExitSignal()

	// Check if the model is present in the local models
	nameReg := regexp.MustCompile(`[\w\-]+\d?\.?\d?`)
	name := nameReg.FindString(model)

	labelReg := regexp.MustCompile(`:\w+\.?(\w+)?`)
	label := labelReg.FindString(model)
	if label == "" {
		label = ":latest"
	}

	model = name + label

	for !ollama.IsModelPresent(model) {
		println(Red(">>> [Error]") + "Model not found. Please ensure the model is downloaded and available locally")
		line, err := l.Readline()
		if err == readline.ErrInterrupt || line == "exit" {
			println(Green(">>> [Msg]") + "Exiting interactive shell session. Bye!")
			return
		}

		nameReg := regexp.MustCompile(`[\w\-]+\d?\.?\d?`)
		name := nameReg.FindString(line)

		labelReg := regexp.MustCompile(`:\w+\.?(\w+)?`)
		label := labelReg.FindString(line)
		if label == "" {
			label = ":latest"
		}

		model = name + label
	}
	cfg.Model = model
	msg := ollama.Message{
		Role: "user",
	}

	println(Yellow(">>> [Hint]") + " Type 'help' or '?' for available flags.")
	println(Cyan(">>>")+" Hello, how can I assist you:", model)

	for {
		line, err := l.Readline()
		if err == readline.ErrInterrupt || line == "exit" {
			println(Green(">>> [Msg]") + " Exiting interactive shell session.")
			return
		}
		if len(line) == 0 {
			continue
		}

		if line == "help" || line == "?" {
			println(Yellow(">>>") + " Ask something, or use the following commands:")
			println(Yellow(">>>") + " --file <path> - Read the content of a file")
			println(Yellow(">>>") + " --directory <path> - Read the content of a directory")
			println(Yellow(">>>") + " --images <path> - Read the content of images")
			println(Yellow(">>>") + " Type 'exit' to exit the interactive shell session.")
			continue
		}

		if strings.Contains(line, "--") {
			parts := strings.Split(line, "--")
			msg.Content = parts[0]
			path := strings.Split(parts[1], " ")
			if len(path) <= 1 {
				println(Red(">>> [Error]") + " Please provide a valid path after the command.")
				println(Yellow(">>> [Hint]") + " Use 'help' or '?' to see available commands.")
				continue
			}
			switch path[0] {
			case "file":
				println(Yellow(">>> [Hint]")+" Reading file:", path[1])
				file, err := os.ReadFile(path[1])
				if err != nil {
					println(Red(">>> [Error]")+" Not able to read the specified file:", err.Error())
					continue
				}
				msg.Content += "\n" + string(file)
			case "directory":
				println(Yellow(">>> [Hint]")+" Reading directory:", path[1])
				data, err := GetDirContent(path[1])
				if err != nil {
					println(Red(">>> [Error]")+" Not able to read the specified directory:", err.Error())
					continue
				}
				msg.Content += "\n" + data
			case "images":
				println(Yellow(">>> [Hint]")+" Reading images:", path[1])
				encoded, err := EncodeImageToBase64(path[1])
				if err != nil {
					println(Red(">>> [Error]")+" Not able to read the specified image:", err.Error())
					continue
				}
				msg.Images = append(msg.Images, encoded)
			default:
				println(Red(">>> [Error]")+" Unknown command:", path[0])
				println(Yellow(">>> [Hint]") + " Use 'help' or '?' to see available commands.")
				continue
			}
		} else {
			msg.Content = line
		}

		// building up context
		cfg.Messages = append(cfg.Messages, msg)
		resp, err := ollama.GetResponse(cfg)
		if err != nil {
			println(Red(">>> [Error]")+" Failed to get response from Ollama:", err.Error())
			continue
		}

		// building up context
		cfg.Messages = append(cfg.Messages, ollama.Message{
			Role:    "assistant",
			Content: resp,
		})

		println("\n" + resp)
	}

}

func init() {
	rootCmd.AddCommand(runCmd)
}
