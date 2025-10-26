package cli

import (
	"fmt"
	"monday-cli/monday"
	"os"
)

type Flag struct {
	Flag  string
	Value string
}

type Command struct {
	Command string
	Args    []string
	Flags   []Flag
}

type CLI struct {
	command Command
	config  *monday.Config
}

func NewCLI() *CLI {
	fmt.Println("Loading config...")
	config, err := monday.LoadConfig(monday.GetConfigPath())
	if err != nil {
		fmt.Println("Error loading config:", err)
		return nil
	}
	fmt.Println("Config loaded successfully")
	c := &CLI{
		config: config,
	}
	fmt.Println("Reading command...")
	c.ReadCommand()
	fmt.Println("Command read successfully")
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
	args := []string{}
	var inString bool
	for _, arg := range os.Args[2:] {
		s := arg
		if arg[0] == '"' {
			inString = true
			s = s[1:]
		}
		if arg[len(arg)-1] == '"' {
			inString = false
			s = s[:len(s)-1]
		}
		if inString {
			args[len(args)-1] += " " + s
		} else {
			args = append(args, s)
		}
	}

	var skipNext bool
	for i, arg := range args {
		if skipNext {
			skipNext = false
			continue
		}
		if arg[0] == '-' {
			if i == len(args)-1 {
				fmt.Println("Error: Invalid flag: " + arg)
				os.Exit(1)
			}
			if args[i+1][0] == '-' {
				fmt.Println("Error: Invalid flag: " + arg)
				os.Exit(1)
			}
			c.command.Flags = append(c.command.Flags, Flag{Flag: arg, Value: args[i+1]})
			skipNext = true
		} else {
			c.command.Args = append(c.command.Args, arg)
		}
	}
	//PrintCommand(c.command)
	return c.command
}
