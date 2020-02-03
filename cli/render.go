package cli

import (
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"text/template"

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
		// ensure "/" at the end
		if input.Path != "" && !strings.HasSuffix(input.Path, "/") {
			input.Path += "/"
		}

		app.FatalIfError(RenderCommand(input), "")
		return nil
	})
}

func RenderCommand(input RenderCommandInput) error {
	svc := NewSsmClient()
	parameters := make(map[string]string, 0)

	file, data, err := ReadTemplate(input)
	if err != nil {
		return err
	}

	if err = ParseTemplate(input, file, data, parameters); err != nil {
		return err
	}

	if err = GetParameters(svc, parameters); err != nil {
		return err
	}

	if err = RenderTemplate(input, file, data, parameters); err != nil {
		return err
	}

	// file mode
	if input.OutputFile != nil && input.OutputMode != "" {
		if mode, err := strconv.ParseInt(input.OutputMode, 8, 16); err == nil {
			err = input.OutputFile.Chmod(os.FileMode(mode))
		}
	}
	if err != nil {
		return err
	}

	return nil
}

func ReadTemplate(input RenderCommandInput) (string, []byte, error) {
	if input.InputFile != "" {
		// from input file
		name := input.InputFile
		data, err := ioutil.ReadFile(input.InputFile)
		return name, data, err
	} else {
		// from stdin
		name := "<stdin>"
		data, err := ioutil.ReadAll(os.Stdin)
		return name, data, err
	}
}

func ParseTemplate(input RenderCommandInput, file string, data []byte, parameters map[string]string) error {
	t, err := template.New(file).Funcs(template.FuncMap{
		"aws_ssm_parameter": func(name string) string {
			if !strings.HasPrefix(name, "/") {
				name = input.Path + name
			}
			if _, ok := parameters[name]; !ok {
				parameters[name] = ""
			}
			return ""
		},
	}).Parse(string(data))

	if err == nil {
		err = t.Execute(ioutil.Discard, nil)
	}

	return err
}

func RenderTemplate(input RenderCommandInput, file string, data []byte, parameters map[string]string) error {
	t, err := template.New(file).Funcs(template.FuncMap{
		"aws_ssm_parameter": func(name string) string {
			if !strings.HasPrefix(name, "/") {
				name = input.Path + name
			}
			return parameters[name]
		},
	}).Parse(string(data))

	if err == nil {
		if input.OutputFile != nil {
			err = t.Execute(input.OutputFile, nil)
		} else {
			err = t.Execute(os.Stdout, nil)
		}
	}

	return err
}

func GetParameters(svc *ssm.SSM, parameters map[string]string) error {
	if len(parameters) == 0 {
		return nil
	}

	keys := make([]string, 0)
	for k := range parameters {
		keys = append(keys, k)
	}

	names := make([]*string, len(keys))
	for i, _ := range keys {
		names[i] = &keys[i]
	}

	for i := 0; i < len(names); i += 10 {
		flag := true
		params := &ssm.GetParametersInput{
			Names: names[i : Min(i + 10, len(names))],
			WithDecryption: &flag,
		}
		result, err := svc.GetParameters(params)
		if err != nil {
			return err
		}

		for _, v := range result.Parameters {
			parameters[*v.Name] = *v.Value
		}
	}

	return nil
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
