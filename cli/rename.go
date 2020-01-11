package cli

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ssm"
	"gopkg.in/alecthomas/kingpin.v2"
)

type RenameCommandInput struct {
	Path string
	Dest string
}

func ConfigureRenameCommand(app *kingpin.Application) {
	input := RenameCommandInput{}

	cmd := app.Command("rename", "Rename a parameter. (alias 'mv')")
	cmd.Alias("mv")

	cmd.Arg("path", "Parameter path.").
		Required().
		StringVar(&input.Path)

	cmd.Arg("dest", "Destination path.").
		Required().
		StringVar(&input.Dest)

	cmd.Action(func(c *kingpin.ParseContext) error {
		app.FatalIfError(RenameCommand(input), "")
		return nil
	})
}

func RenameCommand(input RenameCommandInput) error {
	// remove the trailing / if any
	for strings.HasSuffix(input.Path, "/") {
		input.Path = strings.TrimRight(input.Path, "/")
	}
	for strings.HasSuffix(input.Dest, "/") {
		input.Dest = strings.TrimRight(input.Dest, "/")
	}

	if strings.HasPrefix(input.Path, "/") {
		err := RenameParameter(input)
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				if awsErr.Code() != ssm.ErrCodeParameterNotFound {
					return err
				}
			}
		}
		return RenamePath(input)
	} else {
		return RenameParameter(input)
	}
}

func RenameParameter(input RenameCommandInput) error {
	svc := NewSsmClient()

	flag := true
	params := &ssm.GetParameterInput{
		Name: &input.Path,
		WithDecryption: &flag,
	}
	result, err := svc.GetParameter(params)
	if err != nil {
		return err
	}

	err = Rename(svc, result.Parameter, input.Dest)
	if err != nil {
		return err
	}

	return nil
}

func RenamePath(input RenameCommandInput) error {
	svc := NewSsmClient()

	flag := true
	getParams := &ssm.GetParametersByPathInput{
		Path: &input.Path,
		Recursive: &flag,
		WithDecryption: &flag,
	}
	err := svc.GetParametersByPathPages(getParams,
		func(page *ssm.GetParametersByPathOutput, lastPage bool) bool {
			for _, p := range page.Parameters {
				name := *p.Name
				if strings.HasPrefix(name, input.Path) {
					dest := input.Dest + name[len(input.Path):]
					err := Rename(svc, p, dest)
					if err != nil {
						fmt.Println(err.Error())
					}
				}
			}
			return true
		})
	if err != nil {
		return err
	}

	return nil
}

func Rename(svc *ssm.SSM, parameter *ssm.Parameter, dest string) error {
	// TODO: copy more attributes:
	// - Description
	// - KeyId
	// - Policies
	// - Tags
	// - Tier

	putParams := &ssm.PutParameterInput{
		Name: &dest,
		Type: parameter.Type,
		Value: parameter.Value,
	}
	_, err := svc.PutParameter(putParams)
	if err != nil {
		return err
	}

	deleteParams := &ssm.DeleteParameterInput{
		Name: parameter.Name,
	}
	_, err = svc.DeleteParameter(deleteParams)
	if err != nil {
		return err
	}

	fmt.Printf("%s -> %s\n", *parameter.Name, dest)
	return nil
}
