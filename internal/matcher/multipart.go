package matcher

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	// Matches pt1, PT2, part1, PART2 with optional separators before/after
	reNumericPart = regexp.MustCompile(`(?i)(?:^|[-_\s])(?:(pt|part))[-_\s]?(\d{1,2})(?:$|[-_\s])`)
	// Strict letter-only remainder: optional sep + [a-z] + optional sep
	reLetterOnlyRemainder = regexp.MustCompile(`(?i)^\s*[-_\s]?([a-z])\s*$`)
)

// DetectPartSuffix parses the portion of filename after the first occurrence of id
// and returns (number, suffix) where number=0 means single-part and suffix is the
// normalized string to append to the base name (including leading dash).
func DetectPartSuffix(nameWithoutExt, id string) (int, string) {
	// Find the first occurrence of id case-insensitively to get the remainder
	lowerName := strings.ToLower(nameWithoutExt)
	lowerID := strings.ToLower(id)
	idx := strings.Index(lowerName, lowerID)

	remainder := nameWithoutExt
	if idx >= 0 {
		remainder = nameWithoutExt[idx+len(id):]
	}

	// Trim common separators/spaces around the remainder
	trimmed := strings.TrimSpace(remainder)

	// 1) Numeric parts: pt1 / part1 with optional dash/no-dash
	if m := reNumericPart.FindStringSubmatch(trimmed); len(m) == 3 {
		token := strings.ToLower(m[1]) // "pt" or "part"
		numStr := m[2]
		if n, err := strconv.Atoi(numStr); err == nil && n > 0 {
			return n, "-" + token + numStr
		}
	}

	// 2) Letter parts: single trailing letter (A/B/C/...) optionally separated by dash/underscore/space
	// Only accept when the remainder is just that letter (plus optional separators)
	if m := reLetterOnlyRemainder.FindStringSubmatch(trimmed); len(m) == 2 {
		letter := strings.ToUpper(m[1])
		n := int(letter[0]-'A') + 1
		if n >= 1 && n <= 26 {
			return n, "-" + letter
		}
	}

	// No recognizable part
	return 0, ""
}
