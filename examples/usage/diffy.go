package main

import (
	// "context"
	"fmt"
	"os"

	"github.com/dkooll/diffy"
)

func main() {
	findings, err := diffy.ValidateProject("../module/")
	if err != nil {
		fmt.Printf("Validation error: %v\n", err)
		os.Exit(1)
	}

	// Output findings to console
	diffy.OutputFindings(findings)

	// Optionally create GitHub issues
	// ctx := context.Background()
	// if err := diffy.CreateValidationIssue(ctx, "path/to/terraform", findings); err != nil {
	// 	fmt.Printf("Failed to create GitHub issue: %v\n", err)
	// }
}
