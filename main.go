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
	c.HandleCommand()
}
