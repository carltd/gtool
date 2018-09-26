package commands

import (
	"flag"
	"strings"
	"github.com/carltd/gtool/utils"
)

var ErrUseError = "Use gtool -help for a list"

var AdapterCommands = []*Command{}


func Register(cmd *Command) {
	if cmd == nil {
		utils.Output("gtool command : Register command is nil.", utils.Error)
	}
	AdapterCommands = append(AdapterCommands, cmd)
}

type Command struct {
	Run     func(cmd *Command, args []string) int
	Usage   string
	Use     string
	Options string

	Flag flag.FlagSet
}

// Name returns the command's name: the first word in the Usage line.
func (c *Command) Name() string {
	name := c.Usage
	i := strings.IndexAny(name, " ")
	if i >= 0 {
		name = name[:i]
	}
	return name
}
