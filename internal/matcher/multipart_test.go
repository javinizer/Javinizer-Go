package matcher

import (
	"testing"
)

func TestDetectPartSuffix(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantNum int
		wantSuf string
	}{
		{"IPX-535-pt1", "IPX-535", 1, "-pt1"},
		{"IPX-535PT2", "IPX-535", 2, "-pt2"},
		{"IPX-535-part1", "IPX-535", 1, "-part1"},
		{"IPX-535part2", "IPX-535", 2, "-part2"},
		{"MDB-087A", "MDB-087", 1, "-A"},
		{"MDB-087-b", "MDB-087", 2, "-B"},
		{"ABP-123c", "ABP-123", 3, "-C"},
		{"ABC-123", "ABC-123", 0, ""},
		{"IPX-535 pt1", "IPX-535", 1, "-pt1"},
		{"IPX-535_part3", "IPX-535", 3, "-part3"},
		{"IPX-535-D", "IPX-535", 4, "-D"},
		{"IPX-535-Z", "IPX-535", 26, "-Z"},
		{"IPX-535 no suffix", "IPX-535", 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			num, suf := DetectPartSuffix(tt.name, tt.id)
			if num != tt.wantNum {
				t.Errorf("PartNumber: got %d, want %d", num, tt.wantNum)
			}
			if suf != tt.wantSuf {
				t.Errorf("PartSuffix: got %q, want %q", suf, tt.wantSuf)
			}
		})
	}
}
