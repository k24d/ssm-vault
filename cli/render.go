package cli

import (
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"text/template"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ssm"
	"gopkg.in/alecthomas/kingpin.v2"
)

type RenderCommandInput struct {
	Path       string
	InputFile  string
	OutputFile *os.File
	OutputMode string
}

func ConfigureRenderCommand(app *kingpin.Application) {
	input := RenderCommandInput{}

	cmd := app.Command("render", "Render template.")

	cmd.Arg("template", "Template file.").
		ExistingFileVar(&input.InputFile)

	cmd.Flag("path", "Path prefix.").
		Short('p').
		StringVar(&input.Path)

	cmd.Flag("out", "Output to a file.").
		Short('o').
		OpenFileVar(&input.OutputFile, os.O_WRONLY|os.O_CREATE, 0600)

	cmd.Flag("mode", "Output file mode.").
		Short('m').
		Default("0600").
		StringVar(&input.OutputMode)

	cmd.Action(func(c *kingpin.ParseContext) error {
		app.FatalIfError(RenderCommand(input), "")
		return nil
	})
}

func RenderCommand(input RenderCommandInput) error {
	svc := NewSsmClient()

	var err error
	var templateName string
	var templateData []byte

	// read template
	if input.InputFile != "" {
		// from input file
		templateName = input.InputFile
		templateData, err = ioutil.ReadFile(input.InputFile)
	} else {
		// from stdin
		templateName = "<stdin>"
		templateData, err = ioutil.ReadAll(os.Stdin)
	}
	if err != nil {
		return err
	}

	// path
	path := input.Path
	if path != "" && !strings.HasSuffix(path, "/") {
		path += "/"
	}

	// build template
	temp, err := template.New(templateName).Funcs(template.FuncMap{
		// TODO: use bulk read (GetParameters) for speedup
		"aws_ssm_parameter": func(name string) (string, error) {
			if path != "" && !strings.HasPrefix(name, "/") {
				name = path + name
			}
			return GetParameter(svc, name)
		},
	}).Parse(string(templateData))
	if err != nil {
		return err
	}

	// render
	if input.OutputFile != nil {
		err = temp.Execute(input.OutputFile, nil)
	} else {
		err = temp.Execute(os.Stdout, nil)
	}
	if err != nil {
		return err
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

func GetParameter(svc *ssm.SSM, name string) (string, error) {
	flag := true
	params := &ssm.GetParameterInput{
		Name: &name,
		WithDecryption: &flag,
	}
	result, err := svc.GetParameter(params)
	if err != nil {
		// ignore ParameterNotFound error if -f is specified
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == ssm.ErrCodeParameterNotFound {
				return "", nil
			}
		}
		return "", err
	}

	return *result.Parameter.Value, nil
}
