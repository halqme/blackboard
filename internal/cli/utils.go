package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

func has(args []string, flag string) bool {
	return slices.Contains(args, flag)
}

func value(args []string, flag string) string {
	for i, a := range args {
		if a == flag && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}

func printJSON(v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Fprintln(cliOut, string(b))
	return nil
}

func promptYesNo(question string) (bool, error) {
	fmt.Fprintf(cliOut, "%s [y/N]: ", question)
	r := bufio.NewReader(cliIn)
	line, err := r.ReadString('\n')
	if err != nil && len(line) == 0 {
		return false, err
	}
	answer := strings.ToLower(strings.TrimSpace(line))
	return answer == "y" || answer == "yes", nil
}
