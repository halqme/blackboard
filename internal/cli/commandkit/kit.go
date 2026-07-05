package commandkit

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
)

var (
	In  io.Reader = os.Stdin
	Out io.Writer = os.Stdout
)

type ExitError struct {
	Code    int
	Message string
}

func (e ExitError) Error() string { return e.Message }

func Usage(msg string) error              { return ExitError{Code: 2, Message: msg} }
func ConfigErr(msg string) error          { return ExitError{Code: 3, Message: msg} }
func TaskNotFound(msg string) error       { return ExitError{Code: 4, Message: msg} }
func InvalidTransition(msg string) error  { return ExitError{Code: 5, Message: msg} }
func ArtifactValidation(msg string) error { return ExitError{Code: 11, Message: msg} }
func LockConflict(msg string) error       { return ExitError{Code: 12, Message: msg} }

func Has(args []string, flag string) bool {
	return slices.Contains(args, flag)
}

func Value(args []string, flag string) string {
	for i, a := range args {
		if a == flag && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}

func PrintJSON(v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Fprintln(Out, string(b))
	return nil
}

func PromptYesNo(question string) (bool, error) {
	fmt.Fprintf(Out, "%s [y/N]: ", question)
	r := bufio.NewReader(In)
	line, err := r.ReadString('\n')
	if err != nil && len(line) == 0 {
		return false, err
	}
	answer := strings.ToLower(strings.TrimSpace(line))
	return answer == "y" || answer == "yes", nil
}
