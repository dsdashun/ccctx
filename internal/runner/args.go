package runner

import (
	"fmt"
	"strings"
)

func validateFlagValue(name, value string) error {
	if strings.Contains(value, "\n") {
		return fmt.Errorf("%s value cannot contain newline", name)
	}
	return nil
}

// ExtractFlags extracts --model and --small-fast-model from args before the -- separator.
// Extracted flags are removed from the returned remaining args.
func ExtractFlags(args []string) (model, smallFastModel string, remaining []string, err error) {
	sepIdx := len(args)
	for i, a := range args {
		if a == "--" {
			sepIdx = i
			break
		}
	}

	remaining = make([]string, 0, len(args))
	preSep := args[:sepIdx]
	i := 0
	for i < len(preSep) {
		switch preSep[i] {
		case "--model":
			if i+1 >= len(preSep) {
				return "", "", []string{}, fmt.Errorf("--model requires a value")
			}
			if err := validateFlagValue("--model", preSep[i+1]); err != nil {
				return "", "", []string{}, err
			}
			model = preSep[i+1]
			i += 2
		case "--small-fast-model":
			if i+1 >= len(preSep) {
				return "", "", []string{}, fmt.Errorf("--small-fast-model requires a value")
			}
			if err := validateFlagValue("--small-fast-model", preSep[i+1]); err != nil {
				return "", "", []string{}, err
			}
			smallFastModel = preSep[i+1]
			i += 2
		default:
			remaining = append(remaining, preSep[i])
			i++
		}
	}

	if sepIdx < len(args) {
		remaining = append(remaining, args[sepIdx:]...)
	}

	return model, smallFastModel, remaining, nil
}

// WantsHelp checks if --help or -h appears before -- in args.
func WantsHelp(args []string) bool {
	for _, a := range args {
		if a == "--" {
			break
		}
		if a == "--help" || a == "-h" {
			return true
		}
	}
	return false
}

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
			if strings.HasPrefix(contextArgs[0], "-") {
				return "", nil, false, fmt.Errorf("flag-like argument '%s' not allowed in provider position", contextArgs[0])
			}
			return contextArgs[0], targetArgs, false, nil
		}
		return "", targetArgs, true, nil
	}

	if len(args) == 0 {
		return "", []string{}, true, nil
	}

	if strings.HasPrefix(args[0], "-") {
		return "", nil, false, fmt.Errorf("flag-like argument '%s' not allowed in provider position", args[0])
	}
	return args[0], args[1:], false, nil
}
