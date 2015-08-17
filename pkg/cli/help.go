// Copyright (C) 2015 Scaleway. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE.md file.

package cli

import (
	"os"
	"text/template"

	"github.com/scaleway/scaleway-cli/vendor/github.com/Sirupsen/logrus"
)

// CmdHelp is the 'scw help' command
var CmdHelp = &Command{
	Exec:        nil,
	UsageLine:   "help [COMMAND]",
	Description: "help of the scw command line",
	Help: `
Help prints help information about scw and its commands.

By default, help lists available commands with a short description.
When invoked with a command name, it prints the usage and the help of
the command.
`,
}

func init() {
	// break dependency loop
	CmdHelp.Exec = runHelp
	CmdHelp.Flag.BoolVar(&helpHelp, []string{"h", "-help"}, false, "Print usage")
}

// Flags
var helpHelp bool // -h, --help flag

var helpTemplate = `Usage: scw [OPTIONS] COMMAND [arg...]

Interact with Scaleway from the command line.

Options:
 -h, --help=false             Print usage
 -D, --debug=false            Enable debug mode
 -V, --verbose=false          Enable verbose mode
 -q, --quiet=false            Enable quiet mode
 --api-endpoint=APIEndPoint   Set the API endpoint
 --sensitive=false            Show sensitive data in outputs, i.e. API Token/Organization
 -v, --version=false          Print version information and quit

Commands:
{{range .}}{{if not .Hidden}}    {{.Name | printf "%-9s"}} {{.Description}}
{{end}}{{end}}
Run 'scw COMMAND --help' for more information on a command.
`

func runHelp(cmd *Command, args []string) error {
	if waitHelp {
		cmd.PrintUsage()
	}
	if len(args) > 1 {
		cmd.PrintShortUsage()
	}

	if len(args) == 1 {
		name := args[0]
		for _, command := range Commands {
			if command.Name() == name {
				command.PrintUsage()
			}
		}
		logrus.Fatalf("Unknown help topic `%s`.  Run 'scw help'.", name)
	} else {
		t := template.New("top")
		template.Must(t.Parse(helpTemplate))
		if err := t.Execute(os.Stdout, Commands); err != nil {
			return err
		}
	}
	return nil
}