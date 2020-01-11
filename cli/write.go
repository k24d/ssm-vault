package cli

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/service/ssm"
	"golang.org/x/crypto/ssh/terminal"
	"gopkg.in/alecthomas/kingpin.v2"
)

type WriteCommandInput struct {
	Name      string
	Value     string
	Type      string
	String    bool
	Overwrite bool
}

func ConfigureWriteCommand(app *kingpin.Application) {
	input := WriteCommandInput{}

	cmd := app.Command("write", "Write a parameter value. (alias 'put')")
	cmd.Alias("put")

	cmd.Arg("name", "Parameter name.").
		Required().
		StringVar(&input.Name)

	cmd.Flag("string", "Store as String (PLAIN TEXT).").
		Short('s').
		BoolVar(&input.String)

	cmd.Flag("overwrite", "Overwrite an existing value.").
		Short('f').
		BoolVar(&input.Overwrite)

	cmd.Action(func(c *kingpin.ParseContext) error {
		app.FatalIfError(ParseInput(&input), "")
		app.FatalIfError(WriteCommand(input), "")
		return nil
	})
}

func ParseInput(input *WriteCommandInput) error {
	if terminal.IsTerminal(int(os.Stdin.Fd())) {
		// read from the terminal
		if input.String {
			reader := bufio.NewReader(os.Stdin)
			fmt.Printf("Enter text: ")
			value, err := reader.ReadString('\n')
			if err != nil {
				return err
			}
			input.Value = strings.TrimRight(value, "\n")
		} else {
			fmt.Printf("Enter secret: ")
			value, err := terminal.ReadPassword(int(os.Stdin.Fd()))
			fmt.Println("")
			if err != nil {
				return err
			}
			input.Value = string(value)
		}
	} else {
		// read from stdin
		buf, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		input.Value = string(buf)
	}

	if input.String {
		input.Type = "String"
	} else {
		input.Type = "SecureString"
	}
	return nil
}

func WriteCommand(input WriteCommandInput) error {
	svc := NewSsmClient()

	params := &ssm.PutParameterInput{
		Name: &input.Name,
		Value: &input.Value,
		Type: &input.Type,
		Overwrite: &input.Overwrite,
	}
	_, err := svc.PutParameter(params)
	if err != nil {
		return err
	}

	return nil
}
