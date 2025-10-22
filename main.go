package main

import (
	"fmt"
	"os"

	"emkai/go-cli-gui/cli"
)

func main() {
	c := cli.NewCLI()
	if c == nil {
		fmt.Println("Error creating CLI")
		os.Exit(1)
	}
	c.HandleCommand()
}
