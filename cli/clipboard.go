package cli

import (
	"fmt"
	"time"

	"github.com/atotto/clipboard"
	"github.com/aws/aws-sdk-go/service/ssm"
	"gopkg.in/alecthomas/kingpin.v2"
)

type ClipboardCommandInput struct {
	Name    string
	Timeout time.Duration
}

func ConfigureClipboardCommand(app *kingpin.Application) {
	input := ClipboardCommandInput{}

	cmd := app.Command("clipboard", "Copy a parameter value to clipboard. (alias 'c')")
	cmd.Alias("c")

	cmd.Arg("name", "Parameter name.").
		Required().
		StringVar(&input.Name)

	cmd.Action(func(c *kingpin.ParseContext) error {
		app.FatalIfError(ClipboardCommand(input), "")
		return nil
	})
}

func ClipboardCommand(input ClipboardCommandInput) error {
	svc := NewSsmClient()

	flag := true
	params := &ssm.GetParameterInput{
		Name: &input.Name,
		WithDecryption: &flag,
	}
	result, err := svc.GetParameter(params)
	if err != nil {
		return err
	}

	value := *result.Parameter.Value
	err = clipboard.WriteAll(value)
	if err != nil {
		return err
	}

	fmt.Printf("Copied to clipboard: %s\n", input.Name)
	return nil
}
