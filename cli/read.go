package cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/service/ssm"
	"gopkg.in/alecthomas/kingpin.v2"
)

type ReadCommandInput struct {
	Name       string
	OutputFile *os.File
	OutputMode string
}

func ConfigureReadCommand(app *kingpin.Application) {
	input := ReadCommandInput{}

	cmd := app.Command("read", "Read a parameter value. (alias 'get')")
	cmd.Alias("get")

	cmd.Arg("name", "Parameter name.").
		Required().
		StringVar(&input.Name)

	cmd.Flag("out", "Output to a file.").
		Short('o').
		OpenFileVar(&input.OutputFile, os.O_WRONLY|os.O_CREATE, 0600)

	cmd.Flag("mode", "Output file mode.").
		Short('m').
		Default("0600").
		StringVar(&input.OutputMode)

	cmd.Action(func(c *kingpin.ParseContext) error {
		app.FatalIfError(ReadCommand(input), "")
		return nil
	})
}

func ReadCommand(input ReadCommandInput) error {
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

	p := result.Parameter
	if input.OutputFile != nil {
		input.OutputFile.WriteString(*p.Value)
	} else {
		fmt.Println(*p.Value)
	}

	if input.OutputFile != nil && input.OutputMode != "" {
		mode, err := strconv.ParseInt(input.OutputMode, 8, 16)
		if err == nil {
			err = input.OutputFile.Chmod(os.FileMode(mode))
		}
	}
	if err != nil {
		return err
	}

	return nil
}
