package main

import (
	"fmt"
	"os"
)

type Config struct {
	Strategy    string
	BaseBranch  string
	IssueConfig string
}

func main() {
	err := run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	os.Exit(0)
}

func run() error {

	return nil
}

// strategy
//   - github flow
// main branch
//   - master
