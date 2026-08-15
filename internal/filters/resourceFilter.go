package filters

import (
	"fmt"
	"regexp"
	"slices"
	"sync"
)

type ResourceFilterOptions struct {
	IncludeNames      []string
	IncludeNamesRegex []string
	ExcludeNames      []string
	ExcludeNamesRegex []string

	excludeOnce     sync.Once
	excludeErr      error
	compiledExclude []*regexp.Regexp

	includeOnce     sync.Once
	includeErr      error
	compiledInclude []*regexp.Regexp
}

// IsSet checks if the filter options are set
func (filter *ResourceFilterOptions) IsSet() bool {
	return len(filter.IncludeNames) > 0 || len(filter.IncludeNamesRegex) > 0 || len(filter.ExcludeNames) > 0 || len(filter.ExcludeNamesRegex) > 0
}

// CompilePatterns pre-compiles all include and exclude regex patterns
func (filter *ResourceFilterOptions) CompilePatterns() error {
	if _, err := filter.compileExcludeRegexes(); err != nil {
		return err
	}
	if _, err := filter.compileIncludeRegexes(); err != nil {
		return err
	}
	return nil
}

func (filter *ResourceFilterOptions) compileExcludeRegexes() ([]*regexp.Regexp, error) {
	filter.excludeOnce.Do(func() {
		compiled := make([]*regexp.Regexp, 0, len(filter.ExcludeNamesRegex))
		for _, pattern := range filter.ExcludeNamesRegex {
			re, err := regexp.Compile(pattern)
			if err != nil {
				filter.excludeErr = fmt.Errorf("invalid exclude name regex pattern %s: %v", pattern, err)
				return
			}
			compiled = append(compiled, re)
		}
		filter.compiledExclude = compiled
	})
	return filter.compiledExclude, filter.excludeErr
}

func (filter *ResourceFilterOptions) compileIncludeRegexes() ([]*regexp.Regexp, error) {
	filter.includeOnce.Do(func() {
		compiled := make([]*regexp.Regexp, 0, len(filter.IncludeNamesRegex))
		for _, pattern := range filter.IncludeNamesRegex {
			re, err := regexp.Compile(pattern)
			if err != nil {
				filter.includeErr = fmt.Errorf("invalid include name regex pattern %s: %v", pattern, err)
				return
			}
			compiled = append(compiled, re)
		}
		filter.compiledInclude = compiled
	})
	return filter.compiledInclude, filter.includeErr
}

// ShouldInclude checks if a name should be included or not according to the filter options
// the filter should only be used for one operation (include, exclude)
func (filter *ResourceFilterOptions) ShouldInclude(name string) (bool, error) {
	if len(filter.ExcludeNames) > 0 || len(filter.ExcludeNamesRegex) > 0 {
		if slices.Contains(filter.ExcludeNames, name) {
			return false, nil
		}
		excludeRegexes, err := filter.compileExcludeRegexes()
		if err != nil {
			return false, err
		}
		for _, re := range excludeRegexes {
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

		includeRegexes, err := filter.compileIncludeRegexes()
		if err != nil {
			return false, err
		}
		for _, re := range includeRegexes {
			if re.MatchString(name) {
				return true, nil
			}
		}
		return false, nil
	}
	return true, nil
}
