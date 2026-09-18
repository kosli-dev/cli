package main

import (
	"fmt"
	"slices"
	"strings"
)

// sortedTagPairs renders a tags map as "key=value" pairs ordered by key. The
// order matters because tags reach the printers as a map, whose iteration order
// would otherwise leak into the table output. A missing or non-map value yields
// no pairs, since the API omits "tags" entirely when there are none.
func sortedTagPairs(rawTags any) []string {
	tags, ok := rawTags.(map[string]any)
	if !ok || len(tags) == 0 {
		return nil
	}
	keys := make([]string, 0, len(tags))
	for key := range tags {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	pairs := make([]string, 0, len(tags))
	for _, key := range keys {
		pairs = append(pairs, fmt.Sprintf("%s=%v", key, tags[key]))
	}
	return pairs
}

// formatTags renders a tags map as "[key=value], [key=value]" ordered by key,
// or "" when there are no tags. Detail views substitute "None" for the empty
// string; list columns leave it blank. Repo and control tables use the
// unbracketed pairs from sortedTagPairs instead.
func formatTags(rawTags any) string {
	pairs := sortedTagPairs(rawTags)
	bracketed := make([]string, 0, len(pairs))
	for _, pair := range pairs {
		bracketed = append(bracketed, "["+pair+"]")
	}
	return strings.Join(bracketed, ", ")
}
