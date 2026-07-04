package main

import (
	"fmt"
	"os"

	"github.com/halqme/blackboard/internal/cli"
)

func main() {
	if err := cli.Run(os.Args[1:]); err != nil {
		if e, ok := err.(cli.ExitError); ok {
			fmt.Fprintln(os.Stderr, e.Message)
			os.Exit(e.Code)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
