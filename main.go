package main

import (
	"os"

	"github.com/k24d/ssm-vault/cli"
	"gopkg.in/alecthomas/kingpin.v2"
)

// Version is provided at compile time
var Version = "dev"

func main() {
	run(os.Args[1:], os.Exit)
}

func run(args []string, exit func(int)) {
	app := kingpin.New(
		`ssm-vault`,
		`Secret management with Amazon SSM Parameter Store.`,
	)

	app.ErrorWriter(os.Stderr)
	app.Writer(os.Stdout)
	app.Version(Version)
	app.Terminate(exit)

	cli.ConfigureGlobals(app)
	cli.ConfigureClipboardCommand(app)
	cli.ConfigureDeleteCommand(app)
	cli.ConfigureExecCommand(app)
	cli.ConfigureListCommand(app)
	cli.ConfigureReadCommand(app)
	cli.ConfigureRenameCommand(app)
	cli.ConfigureRenderCommand(app)
	cli.ConfigureTreeCommand(app)
	cli.ConfigureWriteCommand(app)

	kingpin.MustParse(app.Parse(args))
}
