package main

import (
	"errors"
	"fmt"
	"os"

	gitstart "github.com/ykpythemind/git-start"
)

func main() {
	err := run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	os.Exit(0)
}

func run(args []string) error {
	debug := false
	if os.Getenv("DEBUG") != "" {
		debug = true
		fmt.Println("debug")
	}

	if len(args) == 1 {
		// todo: showhelp
		return errors.New("arg is invalid. issue is required")
	}

	cli, err := gitstart.NewCLI(debug, os.Stdin, os.Stdout, os.Stderr, "")
	if err != nil {
		return err
	}

	if args[1] == "pr" || args[1] == "pull-request" {
		return cli.RunPRCommand()
	}

	if len(args) != 2 {
		return errors.New("invalid arguments")
	}

	issuable := args[1]
	return cli.RunStartCommand(issuable)
}
