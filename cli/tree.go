package cli

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/xlab/treeprint"
	"gopkg.in/alecthomas/kingpin.v2"
)

type TreeCommandInput struct {
}

func ConfigureTreeCommand(app *kingpin.Application) {
	input := TreeCommandInput{}

	cmd := app.Command("tree", "Show parameters in tree format.")

	cmd.Action(func(c *kingpin.ParseContext) error {
		app.FatalIfError(TreeCommand(input), "")
		return nil
	})
}

func TreeCommand(input TreeCommandInput) error {
	svc := NewSsmClient()
	tree := treeprint.New()
	nodes := make(map[string]treeprint.Tree)

	params := &ssm.DescribeParametersInput{}
	err := svc.DescribeParametersPages(params,
		func (page *ssm.DescribeParametersOutput, lastPage bool) bool {
			for _, p := range page.Parameters {
				parent := tree
				names := strings.Split(*p.Name, "/")
				for idx, name := range names {
					if idx == len(names) - 1 {
						if *p.Type == "SecureString" {
							parent.AddNode(name + "\U0001F510 (" + *p.KeyId + ")")
						} else {
							parent.AddNode(name)
						}
					} else if idx > 0 {
						path := strings.Join(names[0 : idx + 1], "/")
						if nodes[path] == nil {
							if idx == 1 {
								nodes[path] = parent.AddBranch("/" + name + "/")
							} else {
								nodes[path] = parent.AddBranch(name + "/")
							}
						}
						parent = nodes[path]
					}
				}
			}
			return true
		})
	if err != nil {
		return err
	}

	fmt.Println(tree.String())
	return nil
}
