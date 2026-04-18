package runner

import "fmt"

// ParseArgs parses command arguments in a command-agnostic way.
// It returns the provider name, forwarded target arguments, whether TUI should be used,
// and any error from ambiguous argument combinations.
func ParseArgs(args []string) (provider string, targetArgs []string, useTUI bool, err error) {
	separatorIndex := -1
	for i, arg := range args {
		if arg == "--" {
			separatorIndex = i
			break
		}
	}

	if separatorIndex != -1 {
		contextArgs := args[:separatorIndex]
		targetArgs = args[separatorIndex+1:]

		if len(contextArgs) > 1 {
			return "", nil, false, fmt.Errorf("at most one argument allowed before --")
		}
		if len(contextArgs) == 1 {
			return contextArgs[0], targetArgs, false, nil
		}
		return "", targetArgs, true, nil
	}

	if len(args) == 0 {
		return "", []string{}, true, nil
	}

	return args[0], args[1:], false, nil
}
