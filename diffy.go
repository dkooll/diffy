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

// SchemaValidatorOptions contains options for schema validation
type SchemaValidatorOptions struct {
	TerraformRoot     string
	IncludeModules    bool
	CreateGitHubIssue bool
	Logger            Logger
	GitHubToken       string
	GitHubOwner       string
	GitHubRepo        string
	Silent            bool
}

// SchemaValidatorOption is a function that configures SchemaValidatorOptions
type SchemaValidatorOption func(*SchemaValidatorOptions)

// WithTerraformRoot sets the root directory for Terraform files
func WithTerraformRoot(path string) SchemaValidatorOption {
	return func(opts *SchemaValidatorOptions) {
		opts.TerraformRoot = path
	}
}

// // WithLogger sets a custom logger
func WithLogger(logger Logger) SchemaValidatorOption {
	return func(opts *SchemaValidatorOptions) {
		opts.Logger = logger
	}
}

// WithIncludeModules enables or disables validation of submodules
func WithIncludeModules(include bool) SchemaValidatorOption {
	return func(opts *SchemaValidatorOptions) {
		opts.IncludeModules = include
	}
}

// WithGitHubIssueCreation enables GitHub issue creation with the specified token and repository
func WithGitHubIssueCreation(token, owner, repo string) SchemaValidatorOption {
	return func(opts *SchemaValidatorOptions) {
		opts.CreateGitHubIssue = true
		opts.GitHubToken = token
		opts.GitHubOwner = owner
		opts.GitHubRepo = repo
	}
}

// WithGitHubIssueCreationFromEnv enables GitHub issue creation using environment variables
func WithGitHubIssueCreationFromEnv() SchemaValidatorOption {
	return func(opts *SchemaValidatorOptions) {
		opts.CreateGitHubIssue = true
		opts.GitHubToken = os.Getenv("GITHUB_TOKEN")
		// Let the GetRepoInfo method handle these later if not specified
	}
}

// WithSilent disables console output
func WithSilent(silent bool) SchemaValidatorOption {
	return func(opts *SchemaValidatorOptions) {
		opts.Silent = silent
	}
}

// ValidateSchema validates Terraform schema with the specified options
func ValidateSchema(options ...SchemaValidatorOption) ([]ValidationFinding, error) {
	// Default options
	opts := &SchemaValidatorOptions{
		TerraformRoot:  ".",
		IncludeModules: true,
		// Logger:          &SimpleLogger{},
		CreateGitHubIssue: false,
		Silent:            false,
	}

	// Apply options
	for _, option := range options {
		option(opts)
	}

	// Validate Terraform project
	findings, err := validateProject(opts)
	if err != nil {
		return nil, err
	}

	// Output findings to console if not silent
	if !opts.Silent {
		outputFindings(findings)
	}

	// Create GitHub issue if enabled
	if opts.CreateGitHubIssue && len(findings) > 0 {
		ctx := context.Background()
		if err := createGitHubIssue(ctx, opts, findings); err != nil {
			opts.Logger.Logf("Failed to create GitHub issue: %v", err)
		}
	}

	return findings, nil
}

// validateProject is the internal implementation of project validation
func validateProject(opts *SchemaValidatorOptions) ([]ValidationFinding, error) {
	// Resolve absolute path
	absRoot, err := filepath.Abs(opts.TerraformRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path for %s: %w", opts.TerraformRoot, err)
	}

	// Run validation on root directory
	rootFindings, err := ValidateTerraformSchemaInDirectory(opts.Logger, absRoot, "")
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	var allFindings []ValidationFinding
	allFindings = append(allFindings, rootFindings...)

	// Validate submodules if enabled
	if opts.IncludeModules {
		modulesDir := filepath.Join(absRoot, "modules")
		submodules, err := FindSubmodules(modulesDir)
		// if err != nil {
		// 	opts.Logger.Logf("Failed to find submodules in %s: %v", modulesDir, err)
		if !opts.Silent {
			fmt.Printf("Failed to find submodules in %s: %v\n", modulesDir, err)

		} else {
			for _, sm := range submodules {
				findings, err := ValidateTerraformSchemaInDirectory(opts.Logger, sm.Path, sm.Name)
				if err != nil {
					opts.Logger.Logf("Failed to validate submodule %s: %v", sm.Name, err)
					continue
				}
				allFindings = append(allFindings, findings...)
			}
		}
	}

	// Deduplicate findings
	deduplicatedFindings := DeduplicateFindings(allFindings)

	return deduplicatedFindings, nil
}

// outputFindings prints validation findings to stdout
func outputFindings(findings []ValidationFinding) {
	if len(findings) == 0 {
		fmt.Println("No validation findings.")
		return
	}

	fmt.Printf("Found %d issues:\n", len(findings))

	for _, f := range findings {
		fmt.Println(FormatFinding(f))
	}
}

// createGitHubIssue creates a GitHub issue with validation findings
func createGitHubIssue(ctx context.Context, opts *SchemaValidatorOptions, findings []ValidationFinding) error {
	// Skip if no findings
	if len(findings) == 0 {
		return nil
	}

	// Get GitHub token
	if opts.GitHubToken == "" {
		return fmt.Errorf("GitHub token not provided")
	}

	owner := opts.GitHubOwner
	repo := opts.GitHubRepo

	// If owner/repo not specified, try to determine from git
	if owner == "" || repo == "" {
		gi := NewGitRepoInfo(opts.TerraformRoot)
		owner, repo = gi.GetRepoInfo()
		if owner == "" || repo == "" {
			return fmt.Errorf("could not determine repository info for GitHub issue creation")
		}
	}

	// Create issue manager
	issueManager := NewGitHubIssueManager(owner, repo, opts.GitHubToken)

	// Create or update issue
	return issueManager.CreateOrUpdateIssue(ctx, findings)
}
