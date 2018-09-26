package commands

import (
	"fmt"
	"github.com/carltd/gtool/utils"
	"html/template"
	"os"
)

var usage = `gtool is a development command-line tool for carltd.
---------------------------------------------------------
Usage:

	gtool command [arguments]

The commands are:
`
var usageTemplate = `	{{.Name}}	{{.Use | printf "%s"}}
	help	This help page
`

var usageHelpTemplate = `
Use "gtool help [command]" for more information about a command.
`

var helpTemplate = `
usage: {{.Usage | printf "%s"}}
{{.Use | printf "%s"}}
{{.Options | printf "%s"}}
`

func Usage() {
	fmt.Println(usage)
	for _, cmd := range AdapterCommands {
		data := template.FuncMap{"Name": cmd.Name(), "Use": cmd.Use}
		utils.Tmpl(usageTemplate, data, os.Stderr)
	}
	fmt.Println(usageHelpTemplate)
}

func Help(args []string) {
	if len(args) == 0 {
		Usage()
		return
	}

	if len(args) != 1 {
		utils.Output("Too many arguments.", utils.Error)
		return
	}

	arg := args[0]
	for _, c := range AdapterCommands {
		if c.Name() == arg {
			utils.Tmpl(helpTemplate, c, os.Stderr)
			return
		}
	}

	utils.Output("Unknown help topic", utils.Error)
}
