package main

import (
	"fmt"
	"os"
)

// Global vars moved to root.go

func main() {
	// Root command and all setup moved to root.go
	// Commands wired via root.go's init() function
	if err := Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// All command definitions extracted to separate files
// Shared helper functions remain below

// Shared helper functions used by multiple commands (sort, update)
