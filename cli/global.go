package cli

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"gopkg.in/alecthomas/kingpin.v2"
)

var GlobalFlags struct {
	Debug bool
}

func ConfigureGlobals(app *kingpin.Application) {
	app.Flag("debug", "Enable debug mode.").
		BoolVar(&GlobalFlags.Debug)
}

func NewSsmClient() *ssm.SSM {
	sess := session.Must(session.NewSession())
	if GlobalFlags.Debug {
		return ssm.New(sess, aws.NewConfig().WithLogLevel(aws.LogDebugWithHTTPBody))
	} else {
		return ssm.New(sess)
	}
}
