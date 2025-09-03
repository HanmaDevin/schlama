# Schlama Chat Application

Schlama is a modern chat application built with Go (Golang) that provides a sleek and user-friendly interface. It focuses as a CLI tool but offers a web-interface for GUI lovers.

## Features

### CLI

The CLI makes it easy to chat with local models, install new ones and add files to them. Also includes a simple subset of the already available Ollama commands such as 'show' and 'rm'.

### Web APP

The web app is just a chat interface for those who like GUIs.
It can select local models and you are able to add files to your prompt.

## Prerequisites

- Go 1.20 or later
- Ollama 

## Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/HanmaDevin/projects.git $HOME
   cd projects/schlama
   ```

2. Build the application:

   ```bash
   make build
   ```
   On Windows:

    ```bash
    make build_win
    ```

3. Run the application:

   ```bash
   make run
   ```

   On Windows:

    ```bash
    make run_win
    ```

Or make sure you have the '\$GOPATH' variable set and in '\$PATH' then just:

```bash
make install
```

After that you should be able to use it by just typing 'schlama' in the command line.

## Usage

### Web Application

- First start the application with:

    ```bash
    ./bin/schlama chat
    ```

- Access the application in your browser at `http://localhost:8080`.
- Use the dropdown menu to select a model.
- Enter your message in the text input and click "Send".
- Upload files using the file input above the text box.

### Command-Line Tools

- **Get Help**:

    ```bash
    ./bin/schlama -h or ./bin/schlama --help
    ```

- **List Models**:

  ```bash
  ./bin/schlama list
  ```

- **Show Local Models**:

  ```bash
  ./bin/schlama list --local
  ```

- **Select Model**:

  ```bash
  ./bin/schlama select <model>
  ```

- **Send Prompt**:

  ```bash
  ./bin/schlama prompt "Your message here"
  ```

- **Send Prompt with File**:

  ```bash
  ./bin/schlama prompt "Your message here" --file /path/to/file
  ```

- **Send Prompt with Directory content**:

  ```bash
  ./bin/schlama prompt "Your message here" --directory /path/to/directory
  ```

- **Send Prompt with Images content**:

  Use absolute paths or relative paths from the current working directory. Multiple images can be specified by separating them with commas.
  Environment variables can be used to specify paths, e.g., `$HOME/image.jpg`.

  ```bash
  ./bin/schlama prompt "Your message here" --images /path/to/image,/path/to/another/image,...
  ``` 

- **Install Model**:

  ```bash
  ./bin/schlama pull <model>
  ```

- **Uninstall Model**:

  ```bash
  ./bin/schlama rm <model>
  ```
- **Show Model Info**:

  ```bash
  ./bin/schlama show <model>
  ```
- **Start interactive Shell**:

    ```bash
    ./bin/schlama run <model>
    ```

## Contributing

Contributions are welcome! Please fork the repository and submit a pull request.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
