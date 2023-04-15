package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/assistent-ai/client/chat"
	jess_cli "github.com/assistent-ai/client/cli"
	"github.com/assistent-ai/client/db"
	"github.com/assistent-ai/client/gpt"
	"github.com/assistent-ai/client/model"
	"github.com/urfave/cli/v2"
)

const Version = "2"

func main() {
	apiKeyFilePath := ""
	defaultFilePath := filepath.Join(os.Getenv("HOME"), ".open-ai.key")
	app := setupApp(&apiKeyFilePath, defaultFilePath)

	err := app.Run(os.Args)
	if err != nil {
		cli.Exit(err, 1)
	}
}

func setupApp(apiKeyFilePath *string, defaultFilePath string) *cli.App {
	app := cli.NewApp()
	app.Name = "jessica"
	app.Usage = "Jessica is an AI assistent."
	app.Version = Version
	app.Flags = defineGlobalFlags(apiKeyFilePath, defaultFilePath)

	app.Commands = defineCommands(apiKeyFilePath)

	return app
}

func defineGlobalFlags(apiKeyFilePath *string, defaultFilePath string) []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "key-file",
			Value:       defaultFilePath,
			Usage:       "Path to the text file containing the API key",
			Destination: apiKeyFilePath,
		},
	}
}

func defineCommands(apiKeyFilePath *string) []*cli.Command {
	return []*cli.Command{
		defineDialogCommand(apiKeyFilePath),
		defineFileCommand(apiKeyFilePath),
	}
}

func defineDialogCommand(apiKeyFilePath *string) *cli.Command {
	return &cli.Command{
		Name:   "dialog",
		Usage:  "Manage dialogs",
		Action: handleDialogAction(apiKeyFilePath),
		Flags:  dialogFlags(),
	}
}

func handleDialogAction(apiKeyFilePath *string) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		if c.Bool("list") {
			handleDialogList()
		} else if id := c.String("continue"); id != "" {
			handleDialogContinue(id, apiKeyFilePath)
		} else if id := c.String("show"); id != "" {
			handleDialogShow(id)
		} else if id := c.String("delete"); id != "" {
			handleDialogDelete(id)
		} else {
			return errors.New("no required key provided")
		}
		return nil
	}
}

func dialogFlags() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:    "list",
			Aliases: []string{"l"},
			Usage:   "list all dialogs",
		},
		&cli.StringFlag{
			Name:    "continue",
			Aliases: []string{"c"},
			Usage:   "continue dialog with the id",
		},
		&cli.StringFlag{
			Name:    "show",
			Aliases: []string{"s"},
			Usage:   "show dialog with the id",
		},
		&cli.StringFlag{
			Name:    "delete",
			Aliases: []string{"d"},
			Usage:   "delete dialog with the id",
		},
	}
}

func handleDialogList() {
	dialogIds, err := db.GetDialogIDs()
	if err != nil {
		cli.Exit(err, 1)
	} else {
		jess_cli.PrintDialogIDs(dialogIds)
	}
}

func handleDialogContinue(id string, apiKeyFilePath *string) {
	fmt.Println("Starting a new conversation...")
	ctx, err := initContext(*apiKeyFilePath)
	if err != nil {
		cli.Exit(err, 1)
	}
	err = chat.StartChat(id, ctx)
	if err != nil {
		cli.Exit(err, 1)
	}
}

func handleDialogShow(id string) {
	messages, err := db.GetMessagesByDialogID(id)
	if err != nil {
		cli.Exit(err, 1)
	} else {
		chat.ShowMessages(messages)
	}
}

func handleDialogDelete(id string) {
	err := db.RemoveMessagesByDialogId(id)
	if err != nil {
		cli.Exit(err, 1)
	}
}

func defineFileCommand(apiKeyFilePath *string) *cli.Command {
	return &cli.Command{
		Name:   "file",
		Usage:  "Process file input",
		Flags:  fileFlags(),
		Action: handleFileAction(apiKeyFilePath),
	}
}

func fileFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:     "input",
			Aliases:  []string{"i"},
			Usage:    "path to input file",
			Required: true,
		},
		&cli.StringFlag{
			Name:    "prompt",
			Aliases: []string{"p"},
			Usage:   "prompt input to pass with the file",
		},
		&cli.BoolFlag{
			Name:    "refactor",
			Aliases: []string{"r"},
			Usage:   "refactor file by applying best practices",
		},
	}
}

func handleFileAction(apiKeyFilePath *string) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		filePath := c.String("input")
		prompt := c.String("prompt")
		refactor := c.Bool("refactor")
		if refactor {
			prompt = "Refactor following file, extract code, de-duplicate, apply all best practices that you can think off that would be valuable here and would improve readability"
		}
		ctx, _ := initContext(*apiKeyFilePath)
		b, _ := os.ReadFile(filePath)
		input := model.FileInput{
			UserMessage: prompt,
			FileContent: string(b),
		}
		gptPrompt, _ := chat.GeneratePromptForFile(input)
		answer, _ := gpt.RandomMessage(gptPrompt, ctx)
		fmt.Println(answer)
		return nil
	}
}

func initContext(openAiKeyFilePath string) (*model.AppContext, error) {
	b, err := os.ReadFile(openAiKeyFilePath)
	if err != nil {
		return nil, err
	}
	return &model.AppContext{
		OpenAiKey: strings.ReplaceAll(string(b), "\n", ""),
	}, nil
}
