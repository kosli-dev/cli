package filters

import (
	"fmt"
	"regexp"
	"slices"
)

type ResourceFilterOptions struct {
	IncludeNames      []string
	IncludeNamesRegex []string
	ExcludeNames      []string
	ExcludeNamesRegex []string
}

// ShouldInclude checks if a name should be included or not according to the filter options
// the filter should only be used for one operation (include, exclude)
func (filter *ResourceFilterOptions) ShouldInclude(name string) (bool, error) {
	if len(filter.ExcludeNames) > 0 || len(filter.ExcludeNamesRegex) > 0 {
		if slices.Contains(filter.ExcludeNames, name) {
			return false, nil
		}
		for _, pattern := range filter.ExcludeNamesRegex {
			re, err := regexp.Compile(pattern)
			if err != nil {
				return false, fmt.Errorf("invalid exclude name regex pattern %s: %v", pattern, err)
			}
			if re.MatchString(name) {
				return false, nil
			}
		}
		return true, nil
	} else if len(filter.IncludeNames) > 0 || len(filter.IncludeNamesRegex) > 0 {
		// inclusion
		if slices.Contains(filter.IncludeNames, name) {
			return true, nil
		}

		for _, pattern := range filter.IncludeNamesRegex {
			re, err := regexp.Compile(pattern)
			if err != nil {
				return false, fmt.Errorf("invalid include name regex pattern %s: %v", pattern, err)
			}
			if re.MatchString(name) {
				return true, nil
			}
		}
		return false, nil
	}
	return true, nil
}
