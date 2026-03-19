//go:build ignore
// +build ignore

package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

type pkgStats struct {
	Name  string
	Hit   int
	Miss  int
	Total int
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: go run by_pkg.go <coverage.out>")
		os.Exit(1)
	}

	file, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening profile: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		fmt.Fprintln(os.Stderr, "Error: empty coverage profile")
		os.Exit(1)
	}

	header := scanner.Text()
	if !strings.HasPrefix(header, "mode:") {
		fmt.Fprintf(os.Stderr, "Error: invalid header: %s\n", header)
		os.Exit(1)
	}

	stats := make(map[string]*pkgStats)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// Format: file:start.line-end.line statements counts
		parts := strings.Fields(line)
		if len(parts) != 3 {
			continue
		}

		location := parts[0]
		execCount := 0
		fmt.Sscanf(parts[2], "%d", &execCount)

		// Extract package path (remove filename.go)
		pkg := extractPackage(location)
		if stats[pkg] == nil {
			stats[pkg] = &pkgStats{Name: pkg}
		}

		stats[pkg].Total++
		if execCount > 0 {
			stats[pkg].Hit++
		} else {
			stats[pkg].Miss++
		}
	}

	// Convert to slice and sort by coverage percentage
	var pkgList []*pkgStats
	for _, s := range stats {
		pkgList = append(pkgList, s)
	}
	sort.Slice(pkgList, func(i, j int) bool {
		pctI := float64(pkgList[i].Hit) / float64(pkgList[i].Total)
		pctJ := float64(pkgList[j].Hit) / float64(pkgList[j].Total)
		return pctI < pctJ
	})

	// Print header
	fmt.Printf("%8s %-80s (%6s hit, %6s miss, %6s total)\n", "COVERAGE", "PACKAGE", "HIT", "MISS", "TOTAL")
	fmt.Println(strings.Repeat("=", 110))

	// Print stats
	for _, s := range pkgList {
		if s.Total > 0 {
			percent := float64(s.Hit) / float64(s.Total) * 100
			fmt.Printf("%8.2f%% %-80s (%6d hit, %6d miss, %6d total)\n",
				percent, truncate(s.Name, 76), s.Hit, s.Miss, s.Total)
		}
	}
}

func extractPackage(location string) string {
	// Remove .go suffix and everything after last /
	lastSlash := strings.LastIndex(location, "/")
	if lastSlash == -1 {
		return location
	}
	return location[:lastSlash]
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
