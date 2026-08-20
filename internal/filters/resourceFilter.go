package filters

import (
	"fmt"
	"regexp"
	"slices"
)

// ResourceFilterOptions holds the raw, uncompiled filter values as supplied by CLI flags
// or config files. Compile it once before filtering many resource names.
type ResourceFilterOptions struct {
	IncludeNames      []string
	IncludeNamesRegex []string
	ExcludeNames      []string
	ExcludeNamesRegex []string
}

// IsSet checks if the filter options are set
func (filter *ResourceFilterOptions) IsSet() bool {
	return len(filter.IncludeNames) > 0 || len(filter.IncludeNamesRegex) > 0 || len(filter.ExcludeNames) > 0 || len(filter.ExcludeNamesRegex) > 0
}

// namePatterns is a list of resource name regex patterns after compilation, split at the
// first pattern that failed to compile: usable holds the patterns before it, which are
// matched in order, and err is the error it produced, reported only once matching a name
// has got past all of usable. Patterns behind it are unreachable and so are dropped, which
// is how they behaved when ShouldInclude compiled patterns on the fly.
type namePatterns struct {
	usable []*regexp.Regexp
	err    error
}

// isSet reports whether any pattern was supplied, valid or not.
func (patterns namePatterns) isSet() bool {
	return len(patterns.usable) > 0 || patterns.err != nil
}

// match reports whether name matches any of the usable patterns, and errors if name gets
// past all of them and an invalid pattern is waiting behind them.
func (patterns namePatterns) match(name string) (bool, error) {
	for _, re := range patterns.usable {
		if re.MatchString(name) {
			return true, nil
		}
	}
	return false, patterns.err
}

// CompiledResourceFilter is a ResourceFilterOptions with its regex patterns compiled once,
// so the result can be reused across many resource names without re-compiling per name.
// It is immutable once Compile returns, and is therefore safe for concurrent use.
type CompiledResourceFilter struct {
	includeNames    []string
	includePatterns namePatterns
	excludeNames    []string
	excludePatterns namePatterns
}

// Compile pre-compiles the include and exclude regex patterns of a filter, so the result
// can be reused across many resource names without re-compiling per name.
//
// Compiling cannot fail: an invalid pattern is kept as the error it produced and is
// reported by ShouldInclude when matching a name reaches that pattern. Reporting it here
// instead would reject filters that never reach it, such as a name settled by
// ExcludeNames, or an invalid pattern behind one that already matched.
func (filter *ResourceFilterOptions) Compile() *CompiledResourceFilter {
	return &CompiledResourceFilter{
		includeNames:    filter.IncludeNames,
		includePatterns: compileNamesRegex(filter.IncludeNamesRegex, "include"),
		excludeNames:    filter.ExcludeNames,
		excludePatterns: compileNamesRegex(filter.ExcludeNamesRegex, "exclude"),
	}
}

// compileNamesRegex compiles a list of resource name regex patterns, stopping at the first
// pattern that fails to compile and keeping its error.
// kind is the name of the filter operation (include, exclude) and is only used in errors.
func compileNamesRegex(patterns []string, kind string) namePatterns {
	if len(patterns) == 0 {
		return namePatterns{}
	}
	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, pattern := range patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return namePatterns{
				usable: compiled,
				err:    fmt.Errorf("invalid %s name regex pattern %s: %v", kind, pattern, err),
			}
		}
		compiled = append(compiled, re)
	}
	return namePatterns{usable: compiled}
}

// IsSet checks if the filter options are set
func (filter *CompiledResourceFilter) IsSet() bool {
	return len(filter.includeNames) > 0 || filter.includePatterns.isSet() || len(filter.excludeNames) > 0 || filter.excludePatterns.isSet()
}

// ShouldInclude checks if a name should be included or not according to the filter options
// the filter should only be used for one operation (include, exclude)
//
// An invalid pattern is reported only when matching this name reaches it, so a name that a
// literal or an earlier pattern already settled is filtered without error.
func (filter *CompiledResourceFilter) ShouldInclude(name string) (bool, error) {
	if len(filter.excludeNames) > 0 || filter.excludePatterns.isSet() {
		if slices.Contains(filter.excludeNames, name) {
			return false, nil
		}
		excluded, err := filter.excludePatterns.match(name)
		if err != nil {
			return false, err
		}
		return !excluded, nil
	} else if len(filter.includeNames) > 0 || filter.includePatterns.isSet() {
		// inclusion
		if slices.Contains(filter.includeNames, name) {
			return true, nil
		}

		included, err := filter.includePatterns.match(name)
		if err != nil {
			return false, err
		}
		return included, nil
	}
	return true, nil
}
