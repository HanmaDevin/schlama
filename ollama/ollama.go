package ollama

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/glamour"
	"github.com/schollz/progressbar/v3"
	"golang.org/x/net/html"
)

const ollama_api = "http://localhost:11434/api/chat"

type Message struct {
	Role    string   `json:"role"`
	Content string   `json:"content"`
	Images  []string `json:"images,omitempty"`
}

type Ollama struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type Response struct {
	Resp Message `json:"message"`
}

type PullResponse struct {
	Status    string `json:"status"`
	Total     int64  `json:"total"`
	Completed int64  `json:"completed"`
}

func NewOllama() *Ollama {
	return &Ollama{}
}

const bufferSize = 1024 * 1024 // 1 MB

func GetResponse(ollama *Ollama) (string, error) {
	body := new(bytes.Buffer)
	ollama.Stream = true // Enable streaming
	if err := json.NewEncoder(body).Encode(ollama); err != nil {
		return "", fmt.Errorf("failed to encode request: %w", err)
	}

	c := http.Client{Timeout: time.Minute * 10}
	resp, err := c.Post(ollama_api, "application/json", body)
	if err != nil {
		return "", fmt.Errorf("post request to ollama api failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama api returned status %d: %s", resp.StatusCode, string(b))
	}

	var aiResponse string
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, bufferSize), bufferSize)

	spinner := createSpinner(54, "Waiting for response")
	for scanner.Scan() {
		bts := scanner.Bytes()
		if len(bts) == 0 {
			continue
		}
		var response Response
		if err := json.Unmarshal(bts, &response); err != nil {
			return "", fmt.Errorf("failed to unmarshal response: %w", err)
		}
		if response.Resp.Content != "" {
			aiResponse += response.Resp.Content
		}
		spinner.Add(1)
	}
	spinner.Finish()

	return clean(aiResponse), nil
}

func PullModel(model string) error {
	models := ListModels()
	reg := regexp.MustCompile(`:\w+\.?(\w+)?`)
	modelname := reg.ReplaceAllString(model, "")
	for _, m := range models {
		if m.Name == modelname {
			m := map[string]string{
				"model": model,
			}
			var body = new(bytes.Buffer)
			if err := json.NewEncoder(body).Encode(m); err != nil {
				return fmt.Errorf("failed to encode model request: %w", err)
			}
			c := http.Client{Timeout: time.Minute * 10}
			resp, err := c.Post("http://localhost:11434/api/pull", "application/json", body)
			if err != nil {
				return fmt.Errorf("post request to ollama api failed: %w", err)
			}
			defer resp.Body.Close()

			// just for the total content length entry
			var pullResp PullResponse
			scanner := bufio.NewScanner(resp.Body)
			// inscrease the buffer size to handle larger responses
			scanner.Buffer(make([]byte, bufferSize), bufferSize)

			count := 0
			for scanner.Scan() {
				count++
				bts := scanner.Bytes()
				if len(bts) == 0 {
					continue
				}
				// second line contains the total content length
				if count == 2 {
					if err := json.Unmarshal(bts, &pullResp); err != nil {
						return fmt.Errorf("failed to unmarshal total content length: %w", err)
					}
					break
				}
			}

			bar := createPullProgressBar(pullResp.Total, model)

			for scanner.Scan() {
				bts := scanner.Bytes()
				if len(bts) == 0 {
					continue
				}
				var pullResponse PullResponse
				if err := json.Unmarshal(bts, &pullResponse); err != nil {
					return fmt.Errorf("failed to unmarshal pull response: %w", err)
				}

				bar.Set64(pullResponse.Completed)
				if pullResponse.Status == "success" {
					return nil
				}
			}
		}
	}
	return fmt.Errorf("Model %s not found in the list of available models.", model)
}

func IsOllamaRunning() bool {
	resp, err := http.Get("http://localhost:11434")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

func Show(model string) (string, error) {
	cmd := exec.Command("ollama", "show", model)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to run 'ollama show %s': %w", model, err)
	}

	lines := strings.Split(string(out), "\n")
	if len(lines) < 2 {
		return "", fmt.Errorf("no information found for model %s", model)
	}

	var info []string
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		info = append(info, line)
	}

	return strings.Join(info, "\n"), nil
}

func ListLocalModels() {
	cmd := exec.Command("ollama", "list")
	out, err := cmd.Output()
	if err != nil {
		message := "[Error] Could not run 'ollama list'!"
		fmt.Println(message)
		os.Exit(1)
	}

	lines := strings.Split(string(out), "\n")
	if len(lines) < 2 {
		message := "[Hint] No models found!"
		fmt.Println(message)
		return
	}

	var rows []string

	header := lines[0]
	if len(header) < 65 {
		header = header + strings.Repeat(" ", 65-len(header))
	}
	rows = append(rows, header)
	rows = append(rows, strings.Repeat("-", 65))

	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if len(line) < 65 {
			line = line + strings.Repeat(" ", 65-len(line))
		}
		rows = append(rows, line)
	}

	table := strings.Join(rows, "\n")
	fmt.Println(table)
}

func RemoveModel(model string) error {
	cmd := exec.Command("ollama", "rm", model)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to remove model %s: %s", model, string(out))
	}

	return nil
}

func IsModelPresent(model string) bool {
	cmd := exec.Command("ollama", "list")
	out, err := cmd.Output()
	if err != nil {
		fmt.Println("[Error] Could not run 'ollama list'!")
	}

	table := strings.Split(string(out), "\n")[1:] // Skip the header line
	for _, line := range table {
		present := strings.Contains(line, model)
		if present {
			return true
		}
	}
	return false
}

type ModelInfo struct {
	Name  string
	Sizes []string
}

func extractModels(n *html.Node) []ModelInfo {
	var models []ModelInfo

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "li" {
			hasModelAttr := false
			for _, attr := range n.Attr {
				if attr.Key == "x-test-model" {
					hasModelAttr = true
					break
				}
			}
			if hasModelAttr {
				var modelName string
				var sizes []string

				var findModelName func(*html.Node)
				findModelName = func(node *html.Node) {
					if node.Type == html.ElementNode && node.Data == "div" {
						for _, attr := range node.Attr {
							if attr.Key == "x-test-model-title" {
								for _, a := range node.Attr {
									if a.Key == "title" {
										modelName = a.Val
										break
									}
								}
							}
						}
					}
					for c := node.FirstChild; c != nil; c = c.NextSibling {
						findModelName(c)
					}
				}
				findModelName(n)

				var findSizes func(*html.Node)
				findSizes = func(node *html.Node) {
					if node.Type == html.ElementNode && node.Data == "span" {
						for _, attr := range node.Attr {
							if attr.Key == "x-test-size" {
								if node.FirstChild != nil && node.FirstChild.Type == html.TextNode {
									sizes = append(sizes, strings.TrimSpace(node.FirstChild.Data))
								}
							}
						}
					}
					for c := node.FirstChild; c != nil; c = c.NextSibling {
						findSizes(c)
					}
				}
				findSizes(n)

				if modelName != "" && len(sizes) > 0 {
					models = append(models, ModelInfo{
						Name:  modelName,
						Sizes: sizes,
					})
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)
	return models
}

func ListModels() []ModelInfo {
	c := http.Client{Timeout: time.Minute}
	resp, err := c.Get("https://ollama.com/library?sort=popular")
	if err != nil {
		fmt.Println("[Error] Could not get a response from https://ollama.com")
		os.Exit(1)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("[Error] Could not read response from https://ollama.com")
		os.Exit(1)
	}

	doc, err := html.Parse(bytes.NewReader(b))

	return extractModels(doc)
}

func PrintMarkdown(md string) {
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(100),
	)
	if err != nil {
		fmt.Println("[Error] Not able to create markdown renderer!")
		return
	}
	out, err := r.Render(md)
	if err != nil {
		fmt.Println("[Error] Not able to render markdown!")
		return
	}
	fmt.Fprint(os.Stdout, out)
}

func clean(s string) string {
	re := regexp.MustCompile(`(?s)<think>.*?</think>`)
	cleaned := re.ReplaceAllString(s, "")

	// Basic HTML to Markdown replacements
	replacements := []struct {
		old string
		new string
	}{
		{"<b>", "**"}, {"</b>", "**"},
		{"<strong>", "**"}, {"</strong>", "**"},
		{"<i>", "_"}, {"</i>", "_"},
		{"<em>", "_"}, {"</em>", "_"},
		{"<code>", "`"}, {"</code>", "`"},
		{"<pre>", "```\n"}, {"</pre>", "\n```"},
	}

	for _, r := range replacements {
		cleaned = strings.ReplaceAll(cleaned, r.old, r.new)
	}

	// Unescape HTML entities
	cleaned = html.UnescapeString(cleaned)

	return strings.TrimSpace(cleaned)
}

func StartSpinner(message string) func() {
	done := make(chan struct{})
	go func() {
		chars := []rune{'|', '/', '-', '\\'}
		i := 0
		for {
			select {
			case <-done:
				return
			default:
				fmt.Printf("\r%s %c", message, chars[i%len(chars)])
				time.Sleep(100 * time.Millisecond)
				i++
			}
		}
	}()
	return func() { close(done) }
}

func createPullProgressBar(total int64, model string) *progressbar.ProgressBar {
	bar := progressbar.NewOptions64(total,
		progressbar.OptionSetWriter(os.Stdout),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionShowTotalBytes(true),
		progressbar.OptionFullWidth(),
		progressbar.OptionOnCompletion(func() {
			fmt.Println()
		}),
		progressbar.OptionSetDescription("[cyan]"+fmt.Sprintf("[Msg] Pulling %s", model)+"[reset]"),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[yellow]=[reset]",
			SaucerHead:    "[yellow]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[blue][[reset]",
			BarEnd:        "[blue]][reset]",
		}))
	return bar
}

func createSpinner(spinner int, desc string) *progressbar.ProgressBar {
	bar := progressbar.NewOptions64(-1,
		progressbar.OptionSetWriter(os.Stdout),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionFullWidth(),
		progressbar.OptionClearOnFinish(),
		progressbar.OptionSetDescription("[cyan]"+desc+"[reset]"),
		progressbar.OptionSpinnerType(spinner),
	)
	return bar
}

func CreateTable(models []ModelInfo, limit int) string {
	var rows []string
	header := fmt.Sprintf("%-25s %-40s", "MODEL NAME", "SIZES")
	rows = append(rows, header)
	divider := fmt.Sprintf("%-25s %-40s", strings.Repeat("-", 25), strings.Repeat("-", 40))
	rows = append(rows, divider)
	for i, model := range models {
		if i >= limit {
			break
		}
		line := fmt.Sprintf("%-25s %-40s", model.Name, strings.Join(model.Sizes, ", "))
		rows = append(rows, line)
	}
	return strings.Join(rows, "\n")
}
