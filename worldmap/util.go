package worldmap

import "strings"

// containsFold reports whether s contains substr, case-insensitively.
// Matches the gazetteer's case-insensitive substring POI-type matching.
func containsFold(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
