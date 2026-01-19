package matcher

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	// Matches pt1, PT2, part1, PART2 with optional separators before/after
	reNumericPart = regexp.MustCompile(`(?i)(?:^|[-_\s])(?:(pt|part))[-_\s]?(\d{1,2})(?:$|[-_\s])`)
	// Matches plain numbers: -1, -2, _3, etc. (common multi-part pattern)
	rePlainNumber = regexp.MustCompile(`^[-_\s]?(\d{1,2})$`)
	// Strict letter-only remainder: optional sep + [a-z] + optional sep
	reLetterOnlyRemainder = regexp.MustCompile(`(?i)^\s*[-_\s]?([a-z])\s*$`)
)

// Pattern type constants for multipart detection
const (
	// PatternExplicit indicates explicit multipart patterns (pt1, part2, -1, -2)
	// These are always considered multipart without directory context validation.
	PatternExplicit = "explicit"
	// PatternLetter indicates ambiguous single-letter patterns (A, B, C)
	// These need directory context validation to confirm multipart status.
	PatternLetter = "letter"
	// PatternNone indicates no multipart pattern detected
	PatternNone = ""
)

// DetectPartSuffix parses the portion of filename after the first occurrence of id
// and returns (number, suffix, patternType) where:
//   - number: 0 for single-part, 1..N for part index
//   - suffix: normalized string to append to base name (including leading dash)
//   - patternType: "explicit" for unambiguous patterns, "letter" for ambiguous single-letter,
//     "" for no pattern detected
func DetectPartSuffix(nameWithoutExt, id string) (int, string, string) {
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

	// 1) Numeric parts: pt1 / part1 with optional dash/no-dash - EXPLICIT
	if m := reNumericPart.FindStringSubmatch(trimmed); len(m) == 3 {
		token := strings.ToLower(m[1]) // "pt" or "part"
		numStr := m[2]
		if n, err := strconv.Atoi(numStr); err == nil && n > 0 {
			return n, "-" + token + numStr, PatternExplicit
		}
	}

	// 2) Plain numbers: -1, -2, _3, etc. (common multi-part pattern like pred-151-1.mp4) - EXPLICIT
	if m := rePlainNumber.FindStringSubmatch(trimmed); len(m) == 2 {
		numStr := m[1]
		if n, err := strconv.Atoi(numStr); err == nil && n > 0 {
			return n, "-" + numStr, PatternExplicit
		}
	}

	// 3) Letter parts: single trailing letter (A/B/C/...) optionally separated by dash/underscore/space
	// Only accept when the remainder is just that letter (plus optional separators) - AMBIGUOUS
	// These need directory context validation to confirm multipart status.
	if m := reLetterOnlyRemainder.FindStringSubmatch(trimmed); len(m) == 2 {
		letter := strings.ToUpper(m[1])
		n := int(letter[0]-'A') + 1
		if n >= 1 && n <= 26 {
			return n, "-" + letter, PatternLetter
		}
	}

	// No recognizable part
	return 0, "", PatternNone
}
