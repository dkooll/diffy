package diffy

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// SimpleLogger is a basic implementation of the Logger interface
type SimpleLogger struct{}

// Logf implements the Logger interface
func (l *SimpleLogger) Logf(format string, args ...any) {
	fmt.Printf(format+"\n", args...)
}

// ValidateProject validates an entire Terraform project
func ValidateProject(terraformRoot string) ([]ValidationFinding, error) {
	logger := &SimpleLogger{}

	// Use current directory if no root is specified
	if terraformRoot == "" {
		terraformRoot = "."
	}

	// Resolve absolute path
	absRoot, err := filepath.Abs(terraformRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path for %s: %w", terraformRoot, err)
	}

	// Run validation
	findings, err := ValidateTerraformProject(logger, absRoot)
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return findings, nil
}

// CreateValidationIssue creates a GitHub issue with validation findings
func CreateValidationIssue(ctx context.Context, terraformRoot string, findings []ValidationFinding) error {
	// Skip if no findings
	if len(findings) == 0 {
		return nil
	}

	// Get GitHub token
	ghToken := os.Getenv("GITHUB_TOKEN")
	if ghToken == "" {
		return fmt.Errorf("GITHUB_TOKEN environment variable not set")
	}

	// Get repository information
	gi := NewGitRepoInfo(terraformRoot)
	owner, repoName := gi.GetRepoInfo()
	if owner == "" || repoName == "" {
		return fmt.Errorf("could not determine repository info for GitHub issue creation")
	}

	// Create issue manager
	issueManager := NewGitHubIssueManager(owner, repoName, ghToken)

	// Create or update issue
	return issueManager.CreateOrUpdateIssue(ctx, findings)
}

// OutputFindings prints validation findings to stdout
func OutputFindings(findings []ValidationFinding) {
	if len(findings) == 0 {
		fmt.Println("No validation findings.")
		return
	}

	fmt.Printf("Found %d issues:\n", len(findings))

	for _, f := range findings {
		fmt.Println(FormatFinding(f))
	}
}
