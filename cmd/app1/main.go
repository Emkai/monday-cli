package main

import (
	"fmt"
	"os"
	"monday-cli/cli"
)

func main() {
	fmt.Println("Starting Monday CLI...")
	c := cli.NewCLI()
	if c == nil {
		fmt.Println("Error creating CLI")
		os.Exit(1)
	}
	fmt.Println("CLI created successfully")
	cmd := cli.Command{}
	cmd.Command = "tasks"
	cmd.Args = make([]string, 1)
	cmd.Args[0] = "f"
	c.SetCommand(cmd)
	c.HandleCommand()
}
