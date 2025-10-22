package cli

import (
	"emkai/go-cli-gui/monday"
	"fmt"
	"os"
)

type Command struct {
	Command string
	Args    []string
}

type CLI struct {
	command Command
	config  *monday.Config
}

func NewCLI() *CLI {
	config, err := monday.LoadConfig(monday.GetConfigPath())
	if err != nil {
		fmt.Println("Error loading config:", err)
		return nil
	}
	c := &CLI{
		config: config,
	}
	c.ReadCommand()
	return c
}

func (c *CLI) ReadCommand() Command {
	if len(os.Args) < 2 {
		return Command{
			Command: "help",
			Args:    []string{},
		}
	}
	c.command.Command = os.Args[1]
	c.command.Args = os.Args[2:]
	return c.command
}
