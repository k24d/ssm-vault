package cli

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ssm"
	"gopkg.in/alecthomas/kingpin.v2"
)

type ListCommandInput struct {
	Format string
}

func ConfigureListCommand(app *kingpin.Application) {
	input := ListCommandInput{}

	cmd := app.Command("list", "List parameters. (alias 'ls')")
	cmd.Alias("ls")

	cmd.Flag("format", "Output format (line or json).").
		Default("line").
		StringVar(&input.Format)

	cmd.Action(func(c *kingpin.ParseContext) error {
		app.FatalIfError(ListCommand(input), "")
		return nil
	})
}

func ListCommand(input ListCommandInput) error {
	svc := NewSsmClient()

	params := &ssm.DescribeParametersInput{}
	err := svc.DescribeParametersPages(params,
		func (page *ssm.DescribeParametersOutput, lastPage bool) bool {
			for _, p := range page.Parameters {
				if input.Format == "json" {
					fmt.Printf("%v\n", p)
				} else {
					fmt.Printf("%s\n", *p.Name)
				}
			}
			return true
		})
	if err != nil {
		return err
	}

	return nil
}
