package cmd

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/HanmaDevin/schlama/config"
	"github.com/HanmaDevin/schlama/ollama"
	"github.com/spf13/cobra"
)

var file string
var directory string
var images []string

// promptCmd represents the prompt command
var promptCmd = &cobra.Command{
	Use:   "prompt",
	Short: "Prompt the model with a message.",
	Long:  `Makes an API call to localhost:11434/api/chat and outputs the response in a more readable fashion.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.Help()
		} else {
			body := config.ReadConfig()
			if body.Model == "" {
				fmt.Println(Yellow("[Hint]") + " No model selected. Please set a model using 'schlama select <model_name>'.")
				return
			}

			body.Messages[0].Content = args[0]

			var f []byte
			var err error
			if cmd.Flags().Changed("file") {
				fmt.Println(Yellow("[Hint]") + " Reading file: " + file)
				f, err = os.ReadFile(file)
				if err != nil {
					fmt.Println(Red("[Error]") + " Not able to read the specified file!")
					os.Exit(1)
				}
				body.Messages[0].Content += "\n" + string(f)
			}

			if cmd.Flags().Changed("directory") {
				fmt.Println(Yellow("[Hint]") + " Reading directory: " + directory)
				data, err := GetDirContent(directory)
				if err != nil {
					fmt.Println(Red("[Error]") + " Not able to read the specified directory!")
					os.Exit(1)
				}
				body.Messages[0].Content += "\n" + data
			}

			if cmd.Flags().Changed("images") {
				for _, imgPath := range images {
					fmt.Println(Yellow("[Hint]") + " Reading image: " + imgPath)
					encoded, err := EncodeImageToBase64(imgPath)
					if err != nil {
						fmt.Println(Red("[Error]") + " Not able to read the specified image!")
						os.Exit(1)
					}
					body.Messages[0].Images = append(body.Messages[0].Images, encoded)
				}
			}

			resp, err := ollama.GetResponse(body)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			ollama.PrintMarkdown(resp)
		}
	},
}

func EncodeImageToBase64(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}
	encoded := base64.StdEncoding.EncodeToString(data)
	return encoded, nil
}

func GetDirContent(root string) (string, error) {
	var sb strings.Builder
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			fmt.Println(Yellow("[Hint]") + " Reading file: " + path)
			sb.WriteString("File: " + filepath.Base(path) + "\n")
			sb.Write(content)
			sb.WriteString("\n\n")
		}
		return nil
	})
	return sb.String(), err
}

func init() {
	promptCmd.Flags().StringVarP(&file, "file", "f", "", "Prompt with file content")
	promptCmd.Flags().StringVarP(&directory, "directory", "d", "", "Prompt with directory content")
	promptCmd.Flags().StringSliceVarP(&images, "images", "i", nil, "Prompt with image content")
	rootCmd.AddCommand(promptCmd)
}
