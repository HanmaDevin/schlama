package chat

import (
	"embed"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/charmbracelet/log"

	"github.com/HanmaDevin/schlama/config"
	"github.com/HanmaDevin/schlama/ollama"
)

//go:embed views/*.html
var views embed.FS
var t, _ = template.New("").ParseFS(views, "views/*.html")

var history = []ollama.Message{}

type data struct {
	CurrentModel string
	Prompt       string
	Resp         string
	Models       []string
	Error        string
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	cfg := config.ReadConfig()
	data := data{
		CurrentModel: cfg.Model,
	}

	models, err := getLocalModels()
	if err != nil {
		log.Error("Failed to get local models: " + err.Error())
		http.Error(w, "Failed to get local models: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data.Models = models

	if err := t.ExecuteTemplate(w, "index.html", data); err != nil {
		log.Error("Failed to render index template: " + err.Error())
		http.Error(w, "Something went wrong :(", http.StatusInternalServerError)
	}
}

func setModelHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Error("Failed to parse form: " + err.Error())
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	cfg := config.Config{
		Model: r.FormValue("model"),
	}
	log.Infof("Setting model to %s...", cfg.Model)
	if err := config.WriteConfig(cfg); err != nil {
		log.Error("Failed to write config: " + err.Error())
		http.Error(w, "Failed to write config: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func chatHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("Handling chat request...")
	if err := r.ParseMultipartForm(200 << 20); err != nil {
		log.Error("Failed to parse form: " + err.Error())
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	cfg := config.ReadConfig()

	data := data{
		CurrentModel: cfg.Model,
	}

	prompt := r.FormValue("prompt")
	if prompt == "" {
		log.Warn("Prompt cannot be empty")
		http.Error(w, "", http.StatusBadRequest)
		data.Error = "Prompt cannot be empty"
		return
	}
	for _, msg := range history {
		cfg.Messages = append(cfg.Messages, msg)
	}

	msg := ollama.Message{}
	msg.Role = "user"
	msg.Content = prompt

	form := r.MultipartForm
	for _, fileHeader := range form.File["files"] {
		file, err := fileHeader.Open()
		if err != nil {
			log.Error("Failed to open file: " + err.Error())
			http.Error(w, "Failed to open file: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer file.Close()
		content, err := io.ReadAll(file)
		if err != nil {
			log.Error("Failed to read file: " + err.Error())
			http.Error(w, "Failed to read file: "+err.Error(), http.StatusInternalServerError)
			return
		}
		switch fileHeader.Header.Get("Content-Type") {
		case "text/plain":
			log.Info("Received text file: " + fileHeader.Filename)
			msg.Content += "\n" + string(content)
		case "image/png", "image/jpeg", "image/gif":
			log.Info("Received image file: " + fileHeader.Filename)
			encoded := encodeImageToBase64(content)
			msg.Images = append(msg.Images, encoded)
		case "text/html":
			log.Info("Received HTML file: " + fileHeader.Filename)
			msg.Content += "\n" + string(content)
		case "application/json":
			log.Info("Received JSON file: " + fileHeader.Filename)
			msg.Content += "\n" + string(content)
		case "text/xml", "application/xml":
			log.Info("Received XML file: " + fileHeader.Filename)
			msg.Content += "\n" + string(content)
		case "application/x-shellscript":
			log.Info("Received shell script file: " + fileHeader.Filename)
			msg.Content += "\n" + string(content)
		default:
			log.Info("Received file with unknown content type: " + fileHeader.Header.Get("Content-Type"))
			msg.Content += "\n" + string(content)
		}
	}

	cfg.Messages = append(cfg.Messages, msg)
	resp, err := ollama.GetResponse(cfg)
	if err != nil {
		log.Error("Failed to get response from Ollama: " + err.Error())
		http.Error(w, "Failed to get response from Ollama: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data.Resp = resp
	data.Prompt = prompt

	history = append(history, msg)
	history = append(history, ollama.Message{
		Role:    "assistant",
		Content: resp,
	})

	if err := t.ExecuteTemplate(w, "response.html", data); err != nil {
		log.Error("Failed to render response template: " + err.Error())
		http.Error(w, "Failed to render response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func Start() {
	router := http.NewServeMux()
	router.HandleFunc("GET /", rootHandler)
	router.HandleFunc("POST /set-model", setModelHandler)
	router.HandleFunc("POST /chat", chatHandler)

	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	err := openURL("http://localhost:8080")
	if err != nil {
		fmt.Println("Failed to open browser: " + err.Error())
		os.Exit(1)
	}

	log.Info("Chat started in your browser at http://localhost:8080")
	server.ListenAndServe()
}

func openURL(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", url}
		return exec.Command(cmd, args...).Run()
	case "darwin":
		cmd = "open"
		args = []string{url}
		return exec.Command(cmd, args...).Run()
	default: // "linux", "freebsd", "openbsd", "netbsd"
		// Detect WSL by checking for /proc/sys/fs/binfmt_misc/WSLInterop
		if _, err := os.Stat("/proc/sys/fs/binfmt_misc/WSLInterop"); err == nil {
			log.Info("Detected WSL environment, please manually open the URL in your browser.")
			break
		} else {
			cmd = "xdg-open"
			args = []string{url}
			return exec.Command(cmd, args...).Run()
		}
	}

	return nil
}

func getLocalModels() ([]string, error) {
	cmd := exec.Command("ollama", "list")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(output), "\n")
	var models []string
	for _, line := range lines[1:] { // skip header
		fields := strings.Fields(line)
		if len(fields) > 0 {
			models = append(models, fields[0])
		}
	}
	return models, nil
}

func encodeImageToBase64(content []byte) string {
	encoded := base64.StdEncoding.EncodeToString(content)
	return encoded
}
