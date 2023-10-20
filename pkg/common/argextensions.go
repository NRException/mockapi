package common

import (
	"strings"
)

// Search arguments for an existing switch, returns true if it finds your switch, returns false otherwise.
func ArgSliceContains(args []string, switchTerm string) (switchTermMatched bool) {
	for _, i := range args {
		if i == switchTerm {
			return true
		}
	}
	return false
}

// Search arguments for a slice of switches, if one is found, returns true, returns flase otherwise.
func ArgSliceContainsInTerms(args []string, switchTerms []string) (switchTermMatched bool) {
	for _, switchTerm := range switchTerms {
		return ArgSliceContains(args, switchTerm)
	}
	return false
}

// Search arguments for an existing switch, returns true if it finds your switch, along with any trailing parameters, returns false,nil otherwise.
func ArgSliceSwitchParameters(args []string, switchTerm string) (switchTermMatched bool, trailingParameters []string) {
	params := []string{}
	for i, arg := range args {
		if arg == switchTerm {
			for p := i + 1; p < len(args); p++ {
				if !strings.Contains(args[p], "-") {
					params = append(params, args[p])
				}
			}
			break
		}
	}
	if len(params) > 0 {
		return true, params
	}
	return false, nil
}
