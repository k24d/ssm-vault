// +build !windows

package cli

import (
	"os"
	"os/exec"
	"strings"

	"github.com/aws/aws-sdk-go/service/ssm"
	"golang.org/x/sys/unix"
)

func ExecCommand(input ExecCommandInput) error {
	svc := NewSsmClient()

	// ensure "/" at the end
	path := input.Path
	if path != "" && !strings.HasSuffix(path, "/") {
		input.Path += "/"
	}

	// environment variables
	env := environ(os.Environ())
	index := func (env environ, key string) int {
		key_eq := key + "="
		for i, v := range env {
			if strings.HasPrefix(v, key_eq) {
				return i
			}
		}
		return -1
	}

	flag := true
	getParams := &ssm.GetParametersByPathInput{
		Path: &input.Path,
		Recursive: &flag,
		WithDecryption: &flag,
	}
	err := svc.GetParametersByPathPages(getParams,
		func (page *ssm.GetParametersByPathOutput, lastPage bool) bool {
			for _, p := range page.Parameters {
				if !input.Safe || *p.Type != "SecureString" {
					key, val := GetEnv(input, p)
					i := index(env, key)
					if i == -1 {
						// new environment variable
						env = append(env, key + "=" + val)
					} else {
						// replace when -f is given
						if input.Overwrite {
							env = append(append(env[:i], env[i+1:]...), key + "=" + val)
						}
					}
				}
			}
			return true
		})
	if err != nil {
		return err
	}

	return Execute(input.Command, input.Args, env)
}

func GetEnv(input ExecCommandInput, p *ssm.Parameter) (string, string) {
	name := *p.Name

	// remove path prefix
	if input.Path != "" && strings.HasPrefix(name, input.Path) {
		name = name[len(input.Path):]
	}

	// remove leading "/"
	if strings.HasPrefix(name, "/") {
		name = strings.TrimLeft(name, "/")
	}

	// replace symbols by "_"
	name = strings.Replace(name, "/", "_", -1)
	name = strings.Replace(name, "-", "_", -1)
	name = strings.Replace(name, ".", "_", -1)

	// capitalize
	name = strings.ToUpper(name)

	// variable
	value := *p.Value

	return name, value
}

func Execute(command string, args []string, env environ) error {
	argv0, err := exec.LookPath(command)
	if err != nil {
		return err
	}

	argv := make([]string, 0, 1+len(args))
	argv = append(argv, command)
	argv = append(argv, args...)

	return unix.Exec(argv0, argv, env)
}
