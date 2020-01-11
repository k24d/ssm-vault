package cli

import (
	"os"

	"gopkg.in/alecthomas/kingpin.v2"
)

type ExecCommandInput struct {
	Path      string
	Overwrite bool
	Safe      bool
	Command   string
	Args      []string
}

type environ []string

func ConfigureExecCommand(app *kingpin.Application) {
	input := ExecCommandInput{}

	cmd := app.Command("exec", "Execute a command with environment variables.")

	cmd.Flag("path", "Path prefix.").
		Short('p').
		Default("/").
		StringVar(&input.Path)

	cmd.Flag("overwrite", "Overwrite existing environment variables.").
		Short('f').
		BoolVar(&input.Overwrite)

	cmd.Flag("safe", "Process only plain-text values.").
		BoolVar(&input.Safe)

	cmd.Arg("cmd", "Command to execute. (default: $SHELL)").
		Default(os.Getenv("SHELL")).
		StringVar(&input.Command)

	cmd.Arg("args", "Command arguments.").
		StringsVar(&input.Args)

	cmd.Action(func(c *kingpin.ParseContext) error {
		app.FatalIfError(ExecCommand(input), "")
		return nil
	})
}
