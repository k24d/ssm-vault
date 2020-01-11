package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

 	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ssm"
	"gopkg.in/alecthomas/kingpin.v2"
)

type DeleteCommandInput struct {
	Name  string
	Force bool
}

func ConfigureDeleteCommand(app *kingpin.Application) {
	input := DeleteCommandInput{}

	cmd := app.Command("delete", "Delete a parameter. (alias 'rm')")
	cmd.Alias("rm")

	cmd.Arg("name", "Parameter name.").
		Required().
		StringVar(&input.Name)

	cmd.Flag("force", "Delete parameter without confirmation.").
		Short('f').
		BoolVar(&input.Force)

	cmd.Action(func(c *kingpin.ParseContext) error {
		app.FatalIfError(DeleteCommand(input), "")
		return nil
	})
}

func DeleteCommand(input DeleteCommandInput) error {
	svc := NewSsmClient()

	if !input.Force {
		// check the existence first
		getParams := &ssm.GetParameterInput{
			Name: &input.Name,
		}
		_, err := svc.GetParameter(getParams)
		if err != nil {
			return err
		}

		// confirmation
		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("Are you sure to delete %s (y/N)? ", input.Name)
		value, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		if strings.ToLower(value) != "y\n" {
			return nil
		}
	}

	// delete
	deleteParams := &ssm.DeleteParameterInput{
		Name: &input.Name,
	}
	_, err := svc.DeleteParameter(deleteParams)
	if err != nil {
		// ignore ParameterNotFound error if -f is specified
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == ssm.ErrCodeParameterNotFound && input.Force {
				return nil
			}
		}
		return err
	}

	return nil
}
