package cli

import (
	"io"
	"os"
)

var (
	cliIn  io.Reader = os.Stdin
	cliOut io.Writer = os.Stdout
)
