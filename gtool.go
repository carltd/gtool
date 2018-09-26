package main

import (
	"flag"
	"os"
	"time"
	"runtime"
	"fmt"

	"github.com/carltd/gtool/commands"
	_ "github.com/carltd/gtool/commands/gorm"

)

const VERSION = "0.0.1"

func init() {
	version := flag.Bool("v", false, "Use -v <current version>")
	flag.Parse()
	// Show version
	if *version {
		fmt.Println("gtool version", VERSION, runtime.GOOS+"/"+runtime.GOARCH)
		os.Exit(0)
	}
}

func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		commands.Usage()
		return
	}

	// Help
	if args[0] == "help" {
		commands.Help(args[1:])
		return
	}

	for _, c := range commands.AdapterCommands {
		if c.Name() == args[0] && c.Run != nil {
			//args = c.Flag.Args()
			code := c.Run(c, args[1:])

			time.Sleep(1 * time.Millisecond)
			os.Exit(code)
		}
	}
}
