package main

import (
	"fmt"
	"os"

	"github.com/dkooll/diffy"
)

func main() {
	findings, err := diffy.ValidateSchema(
		diffy.WithTerraformRoot("../module/"),
	)
	if err != nil {
		fmt.Printf("Validation error: %v\n", err)
		os.Exit(1)
	}

	if len(findings) > 0 {
		os.Exit(1) // Exit with error code if findings exist
	}
}


// package main
//
// import (
// 	"fmt"
// 	"os"
//
// 	"github.com/dkooll/diffy"
// )
//
// // CustomLogger implements the Logger interface
// type CustomLogger struct{}
//
// func (l *CustomLogger) Logf(format string, args ...any) {
// 	fmt.Printf("[TERRAFORM VALIDATOR] "+format+"\n", args...)
// }
//
// func main() {
// 	// Advanced usage with multiple options
// 	findings, err := diffy.ValidateSchema(
// 		diffy.WithTerraformRoot("../module/"),
// 		diffy.WithLogger(&CustomLogger{}),
// 		diffy.WithIncludeModules(true),
// 		diffy.WithGitHubIssueCreationFromEnv(),
// 		diffy.WithSilent(false),
// 	)
// 	if err != nil {
// 		fmt.Printf("Validation error: %v\n", err)
// 		os.Exit(1)
// 	}
//
// 	// The findings are returned so you can do additional processing
// 	if len(findings) > 0 {
// 		// Process findings further if needed
// 		for i, f := range findings {
// 			fmt.Printf("Finding %d: %s in %s\n", i+1, f.Name, f.Path)
// 		}
//
// 		// Exit with error code
// 		os.Exit(1)
// 	}
// }
