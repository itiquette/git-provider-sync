// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

// Static errors for err113 compliance.
var (
	ErrRequiredTestFileMissing   = errors.New("required test file missing")
	ErrCheckingTestFiles         = errors.New("error checking test files")
	ErrFoundTestFiles            = errors.New("found")
	ErrInvalidTestFileNaming     = errors.New("invalid test file naming")
	ErrFailedToCreateOutputDir   = errors.New("failed to create output directory")
	ErrFailedToCreateResultsFile = errors.New("failed to create results file")
)

// TestSuite represents a collection of related tests.
type TestSuite struct {
	Name        string
	Description string
	Commands    []TestCommand
	Timeout     time.Duration
	Parallel    bool
}

// TestCommand represents a single test command.
type TestCommand struct {
	Name     string
	Command  string
	Args     []string
	Dir      string
	Timeout  time.Duration
	Optional bool // If true, failure won't fail the entire suite
}

// TestResult represents the result of running a test command.
type TestResult struct {
	Command  TestCommand
	Success  bool
	Duration time.Duration
	Output   string
	Error    error
	ExitCode int
}

// TestRunner manages and executes test suites.
type TestRunner struct {
	verbose     bool
	parallel    bool
	timeout     time.Duration
	coverage    bool
	integration bool
	benchmark   bool
	dryRun      bool
	outputDir   string
}

// NewTestRunner creates a new test runner with default settings.
func NewTestRunner() *TestRunner {
	return &TestRunner{
		verbose:     false,
		parallel:    true,
		timeout:     10 * time.Minute,
		coverage:    false,
		integration: false,
		benchmark:   false,
		dryRun:      false,
		outputDir:   "test-results",
	}
}

// GetTestSuites returns all available test suites.
func (tr *TestRunner) GetTestSuites() ([]TestSuite, error) {
	baseDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	return []TestSuite{
		{
			Name:        "unit",
			Description: "Fast unit tests with mocks and isolated components",
			Timeout:     5 * time.Minute,
			Parallel:    true,
			Commands: []TestCommand{
				{
					Name:    "domain-sync-tests",
					Command: "go",
					Args:    []string{"test", "-v", "-count=1", "./internal/domain/sync"},
					Dir:     baseDir,
					Timeout: 2 * time.Minute,
				},
				{
					Name:    "domain-entities-tests",
					Command: "go",
					Args:    []string{"test", "-v", "-count=1", "./internal/domain/entities"},
					Dir:     baseDir,
					Timeout: 1 * time.Minute,
				},
				{
					Name:    "domain-validation-tests",
					Command: "go",
					Args:    []string{"test", "-v", "-count=1", "./internal/domain/validation"},
					Dir:     baseDir,
					Timeout: 1 * time.Minute,
				},
				{
					Name:    "configuration-tests",
					Command: "go",
					Args:    []string{"test", "-v", "-count=1", "./internal/configuration"},
					Dir:     baseDir,
					Timeout: 1 * time.Minute,
				},
				{
					Name:    "model-configuration-tests",
					Command: "go",
					Args:    []string{"test", "-v", "-count=1", "./internal/model/configuration"},
					Dir:     baseDir,
					Timeout: 1 * time.Minute,
				},
				{
					Name:    "print-command-tests",
					Command: "go",
					Args:    []string{"test", "-v", "-count=1", "./cmd/printcmd"},
					Dir:     baseDir,
					Timeout: 1 * time.Minute,
				},
			},
		},
		{
			Name:        "integration",
			Description: "Integration tests with real implementations and temporary directories",
			Timeout:     15 * time.Minute,
			Parallel:    false,
			Commands: []TestCommand{
				{
					Name:    "domain-sync-integration",
					Command: "go",
					Args:    []string{"test", "-tags=integration", "-v", "-count=1", "./internal/domain/sync"},
					Dir:     baseDir,
					Timeout: 10 * time.Minute,
				},
				{
					Name:    "gogit-adapter-integration",
					Command: "go",
					Args:    []string{"test", "-tags=integration", "-v", "-count=1", "./internal/adapters/repository/gogit"},
					Dir:     baseDir,
					Timeout: 8 * time.Minute,
				},
			},
		},
		{
			Name:        "enhanced",
			Description: "Enhanced tests with advanced scenarios",
			Timeout:     20 * time.Minute,
			Parallel:    false,
			Commands: []TestCommand{
				{
					Name:    "enhanced-push-tests",
					Command: "go",
					Args:    []string{"test", "-v", "-count=1", "-run", "Enhanced", "./internal/domain/sync"},
					Dir:     baseDir,
					Timeout: 8 * time.Minute,
				},
				{
					Name:    "enhanced-integration-tests",
					Command: "go",
					Args:    []string{"test", "-tags=integration", "-v", "-count=1", "-run", "Enhanced", "./internal/domain/sync"},
					Dir:     baseDir,
					Timeout: 15 * time.Minute,
				},
			},
		},
		{
			Name:        "coverage",
			Description: "Unit tests with coverage analysis",
			Timeout:     8 * time.Minute,
			Parallel:    true,
			Commands: []TestCommand{
				{
					Name:    "coverage-analysis",
					Command: "go",
					Args:    []string{"test", "-cover", "-coverprofile=coverage.out", "./..."},
					Dir:     baseDir,
					Timeout: 5 * time.Minute,
				},
				{
					Name:     "coverage-report",
					Command:  "go",
					Args:     []string{"tool", "cover", "-html=coverage.out", "-o", "coverage.html"},
					Dir:      baseDir,
					Timeout:  1 * time.Minute,
					Optional: true,
				},
			},
		},
		{
			Name:        "race",
			Description: "Race condition detection tests",
			Timeout:     10 * time.Minute,
			Parallel:    true,
			Commands: []TestCommand{
				{
					Name:    "race-detection",
					Command: "go",
					Args:    []string{"test", "-race", "-count=1", "./..."},
					Dir:     baseDir,
					Timeout: 8 * time.Minute,
				},
			},
		},
		{
			Name:        "benchmark",
			Description: "Performance benchmark tests",
			Timeout:     15 * time.Minute,
			Parallel:    false,
			Commands: []TestCommand{
				{
					Name:    "domain-benchmarks",
					Command: "go",
					Args:    []string{"test", "-bench=.", "-benchmem", "./internal/domain/..."},
					Dir:     baseDir,
					Timeout: 10 * time.Minute,
				},
				{
					Name:    "adapter-benchmarks",
					Command: "go",
					Args:    []string{"test", "-tags=integration", "-bench=.", "-benchmem", "./internal/adapters/repository/gogit"},
					Dir:     baseDir,
					Timeout: 8 * time.Minute,
				},
			},
		},
		{
			Name:        "validation",
			Description: "Test structure and configuration validation",
			Timeout:     5 * time.Minute,
			Parallel:    true,
			Commands: []TestCommand{
				{
					Name:    "test-structure-validation",
					Command: "go",
					Args:    []string{"run", "test_runner.go", "--validate-only"},
					Dir:     baseDir,
					Timeout: 2 * time.Minute,
				},
				{
					Name:    "test-files-exist",
					Command: "find",
					Args:    []string{".", "-name", "*_test.go", "-type", "f"},
					Dir:     baseDir,
					Timeout: 30 * time.Second,
				},
			},
		},
		{
			Name:        "quality",
			Description: "Code quality and linting checks",
			Timeout:     8 * time.Minute,
			Parallel:    true,
			Commands: []TestCommand{
				{
					Name:    "golangci-lint",
					Command: "make",
					Args:    []string{"quality/golangcilint"},
					Dir:     baseDir,
					Timeout: 5 * time.Minute,
				},
				{
					Name:    "go-vet",
					Command: "go",
					Args:    []string{"vet", "./..."},
					Dir:     baseDir,
					Timeout: 2 * time.Minute,
				},
				{
					Name:    "staticcheck",
					Command: "staticcheck",
					Args:    []string{"./..."},
					Dir:     baseDir,
					Timeout: 3 * time.Minute,
				},
			},
		},
	}, nil
}

// RunTestSuite executes a single test suite.
func (tr *TestRunner) RunTestSuite(ctx context.Context, suite TestSuite) ([]TestResult, error) {
	fmt.Printf("\n🧪 Running test suite: %s\n", suite.Name) //nolint:forbidigo // CLI output
	fmt.Printf("Description: %s\n", suite.Description)     //nolint:forbidigo // CLI output
	fmt.Printf("⏱️  Timeout: %v\n", suite.Timeout)         //nolint:forbidigo // CLI output
	fmt.Printf("Commands: %d\n", len(suite.Commands))      //nolint:forbidigo // CLI output

	if tr.dryRun {
		fmt.Printf("DRY RUN - Would execute %d commands\n", len(suite.Commands)) //nolint:forbidigo // CLI output

		for _, cmd := range suite.Commands {
			fmt.Printf("   - %s: %s %s\n", cmd.Name, cmd.Command, strings.Join(cmd.Args, " ")) //nolint:forbidigo // CLI output
		}

		return []TestResult{}, nil
	}

	// Create context with timeout for the entire suite
	suiteCtx, cancel := context.WithTimeout(ctx, suite.Timeout)
	defer cancel()

	var results []TestResult

	var mutex sync.Mutex

	var waitGroup sync.WaitGroup

	// Execute commands
	if suite.Parallel && tr.parallel {
		// Run commands in parallel
		for _, cmd := range suite.Commands {
			waitGroup.Add(1)

			go func(command TestCommand) {
				defer waitGroup.Done()

				result := tr.runTestCommand(suiteCtx, command)

				mutex.Lock()

				results = append(results, result)

				mutex.Unlock()
			}(cmd)
		}

		waitGroup.Wait()
	} else {
		// Run commands sequentially
		for _, cmd := range suite.Commands {
			result := tr.runTestCommand(suiteCtx, cmd)
			results = append(results, result)

			// Stop on first failure for sequential tests (unless optional)
			if !result.Success && !cmd.Optional {
				break
			}
		}
	}

	return results, nil
}

// GenerateReport generates a test report.
func (tr *TestRunner) GenerateReport(results map[string][]TestResult) {
	fmt.Printf("\nTEST REPORT\n")                   //nolint:forbidigo // CLI output
	fmt.Printf("%s\n", "="+strings.Repeat("=", 50)) //nolint:forbidigo // CLI output

	totalSuites := len(results)
	totalCommands := 0
	successfulSuites := 0
	successfulCommands := 0
	totalDuration := time.Duration(0)

	for suiteName, suiteResults := range results {
		totalCommands += len(suiteResults)
		suiteSuccess := true
		suiteDuration := time.Duration(0)

		for _, result := range suiteResults {
			suiteDuration += result.Duration

			if !result.Success && !result.Command.Optional {
				suiteSuccess = false
			} else if result.Success {
				successfulCommands++
			}
		}

		totalDuration += suiteDuration

		status := "✓"
		if !suiteSuccess {
			status = "✗"
		} else {
			successfulSuites++
		}

		fmt.Printf("%s %s: %d/%d commands passed (%.2fs)\n", //nolint:forbidigo // CLI output
			status, suiteName,
			tr.countSuccessfulCommands(suiteResults), len(suiteResults),
			suiteDuration.Seconds())
	}

	fmt.Printf("\nSUMMARY\n")                                                      //nolint:forbidigo // CLI output
	fmt.Printf("Test Suites: %d/%d passed\n", successfulSuites, totalSuites)       //nolint:forbidigo // CLI output
	fmt.Printf("Test Commands: %d/%d passed\n", successfulCommands, totalCommands) //nolint:forbidigo // CLI output
	fmt.Printf("Total Duration: %.2fs\n", totalDuration.Seconds())                 //nolint:forbidigo // CLI output

	// Success rate
	suiteRate := float64(successfulSuites) / float64(totalSuites) * 100
	commandRate := float64(successfulCommands) / float64(totalCommands) * 100

	fmt.Printf("Suite Success Rate: %.1f%%\n", suiteRate)     //nolint:forbidigo // CLI output
	fmt.Printf("Command Success Rate: %.1f%%\n", commandRate) //nolint:forbidigo // CLI output

	// Overall status
	if successfulSuites == totalSuites {
		fmt.Printf("\n🎉 ALL TESTS PASSED!\n") //nolint:forbidigo // CLI output
	} else {
		fmt.Printf("\n!  SOME TESTS FAILED\n") //nolint:forbidigo // CLI output
	}
}

// ValidateTestStructure validates the test structure and configuration.
func (tr *TestRunner) ValidateTestStructure() error {
	fmt.Printf("Validating test structure...\n") //nolint:forbidigo // CLI output

	// Check for required test files
	requiredTests := []string{
		"internal/domain/sync/sync_test.go",
		"internal/domain/sync/push_test.go",
		"internal/domain/sync/fetch_test.go",
		"internal/domain/sync/mocks_test.go",
		"internal/domain/entities/repository_test.go",
		"internal/configuration/loader_test.go",
	}

	for _, testFile := range requiredTests {
		if _, err := os.Stat(testFile); errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("%w: %s", ErrRequiredTestFileMissing, testFile)
		}
	}

	// Check for test files in key directories
	testDirs := []string{
		"internal/domain",
		"internal/adapters/configuration",
		"cmd",
	}

	for _, dir := range testDirs {
		hasTests := tr.hasTestFiles(dir)

		if !hasTests {
			fmt.Printf("!  No test files found in %s\n", dir) //nolint:forbidigo // CLI output
		}
	}

	// Validate test file naming conventions
	err := tr.validateTestNaming()
	if err != nil {
		return err
	}

	fmt.Printf("✓ Test structure validation passed\n") //nolint:forbidigo // CLI output

	return nil
}

// SaveResults saves test results to files.
func (tr *TestRunner) SaveResults(results map[string][]TestResult) error {
	if tr.outputDir == "" {
		return nil
	}

	err := os.MkdirAll(tr.outputDir, 0750)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrFailedToCreateOutputDir, err)
	}

	// Save detailed results
	for suiteName, suiteResults := range results {
		filename := filepath.Join(tr.outputDir, suiteName+"-results.txt")

		file, err := os.Create(filename) // #nosec G304
		if err != nil {
			return fmt.Errorf("%w: %w", ErrFailedToCreateResultsFile, err)
		}

		// Write results to file
		err = tr.writeResultsToFile(file, suiteName, suiteResults)
		closeErr := file.Close()

		if err != nil {
			return fmt.Errorf("failed to write results: %w", err)
		}

		if closeErr != nil {
			return fmt.Errorf("failed to close results file: %w", closeErr)
		}
	}

	fmt.Printf("Results saved to %s/\n", tr.outputDir) //nolint:forbidigo // CLI output

	return nil
}

// writeResultsToFile writes test results to a file.
func (tr *TestRunner) writeResultsToFile(file *os.File, suiteName string, suiteResults []TestResult) error {
	writer := bufio.NewWriter(file)

	// Write header
	header := fmt.Sprintf("Test Suite: %s\nTimestamp: %s\nCommands: %d\n\n",
		suiteName, time.Now().Format(time.RFC3339), len(suiteResults))
	if _, err := writer.WriteString(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write results
	for _, result := range suiteResults {
		if err := tr.writeResult(writer, result); err != nil {
			return fmt.Errorf("failed to write result: %w", err)
		}
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush writer: %w", err)
	}

	return nil
}

// writeResult writes a single test result.
func (tr *TestRunner) writeResult(writer *bufio.Writer, result TestResult) error {
	output := fmt.Sprintf("Command: %s\nSuccess: %v\nDuration: %v\nExit Code: %d\n",
		result.Command.Name, result.Success, result.Duration, result.ExitCode)

	if result.Output != "" {
		output += fmt.Sprintf("Output:\n%s\n", result.Output)
	}

	if result.Error != nil {
		output += fmt.Sprintf("Error: %v\n", result.Error)
	}

	output += fmt.Sprintf("\n%s\n\n", strings.Repeat("-", 50))

	if _, err := writer.WriteString(output); err != nil {
		return fmt.Errorf("failed to write result: %w", err)
	}

	return nil
}

// RunTestCommand executes a single test command.
func (tr *TestRunner) runTestCommand(ctx context.Context, testCmd TestCommand) TestResult {
	start := time.Now()

	if tr.verbose {
		fmt.Printf("  ▶️  Running: %s (%s %s)\n", testCmd.Name, testCmd.Command, strings.Join(testCmd.Args, " ")) //nolint:forbidigo // CLI output
	}

	// Create command context with timeout
	cmdCtx, cancel := context.WithTimeout(ctx, testCmd.Timeout)
	defer cancel()

	// Create command
	cmd := exec.CommandContext(cmdCtx, testCmd.Command, testCmd.Args...) // #nosec G204
	if testCmd.Dir != "" {
		cmd.Dir = testCmd.Dir
	}

	// Capture output
	output, err := cmd.CombinedOutput()
	duration := time.Since(start)

	result := TestResult{
		Command:  testCmd,
		Duration: duration,
		Output:   string(output),
		Error:    err,
	}

	// Determine exit code
	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			result.ExitCode = exitError.ExitCode()
		} else {
			result.ExitCode = -1
		}

		result.Success = false
	} else {
		result.ExitCode = 0
		result.Success = true
	}

	// Log result
	if result.Success {
		if tr.verbose {
			fmt.Printf("  ✓ %s completed in %v\n", testCmd.Name, duration) //nolint:forbidigo // CLI output
		}
	} else {
		fmt.Printf("  ✗ %s failed in %v (exit code: %d)\n", testCmd.Name, duration, result.ExitCode) //nolint:forbidigo // CLI output

		if tr.verbose && result.Output != "" {
			fmt.Printf("     Output: %s\n", strings.TrimSpace(result.Output)) //nolint:forbidigo // CLI output
		}
	}

	return result
}

// CountSuccessfulCommands counts successful commands in a suite.
func (tr *TestRunner) countSuccessfulCommands(results []TestResult) int {
	count := 0

	for _, result := range results {
		if result.Success {
			count++
		}
	}

	return count
}

// HasTestFiles checks if directory has test files.
func (tr *TestRunner) hasTestFiles(dir string) bool {
	err := filepath.WalkDir(dir, func(path string, _ fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(path, "_test.go") {
			return ErrFoundTestFiles // Use error to break early
		}

		return nil
	})

	return err != nil
}

// ValidateTestNaming validates test file naming conventions.
func (tr *TestRunner) validateTestNaming() error {
	testPattern := regexp.MustCompile(`^.*_test\.go$`)

	if err := filepath.WalkDir(".", func(path string, dirEntry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if strings.Contains(path, "_test.go") && !testPattern.MatchString(dirEntry.Name()) {
			return fmt.Errorf("%w: %s", ErrInvalidTestFileNaming, path)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("failed to walk directory for test naming validation: %w", err)
	}

	return nil
}

// CommandLineFlags holds all command line flag values.
type commandLineFlags struct {
	verbose     bool
	parallel    bool
	timeout     time.Duration
	coverage    bool
	integration bool
	benchmark   bool
	dryRun      bool
	outputDir   string
	suites      string
	validate    bool
}

// ParseCommandLineFlags parses and returns command line flags.
func parseCommandLineFlags() commandLineFlags {
	var flags commandLineFlags

	flag.BoolVar(&flags.verbose, "verbose", false, "Enable verbose output")
	flag.BoolVar(&flags.parallel, "parallel", true, "Enable parallel execution where possible")
	flag.DurationVar(&flags.timeout, "timeout", 20*time.Minute, "Global timeout for all tests")
	flag.BoolVar(&flags.coverage, "coverage", false, "Include coverage analysis")
	flag.BoolVar(&flags.integration, "integration", false, "Include integration tests")
	flag.BoolVar(&flags.benchmark, "benchmark", false, "Include benchmark tests")
	flag.BoolVar(&flags.dryRun, "dry-run", false, "Show what would be executed without running")
	flag.StringVar(&flags.outputDir, "output", "test-results", "Output directory for results")
	flag.StringVar(&flags.suites, "suites", "", "Comma-separated list of test suites to run (default: all)")
	flag.BoolVar(&flags.validate, "validate-only", false, "Only validate test structure")

	flag.Parse()

	return flags
}

// SetupTestRunner creates and configures a TestRunner instance.
func setupTestRunner(flags commandLineFlags) *TestRunner {
	runner := NewTestRunner()
	runner.verbose = flags.verbose
	runner.parallel = flags.parallel
	runner.timeout = flags.timeout
	runner.coverage = flags.coverage
	runner.integration = flags.integration
	runner.benchmark = flags.benchmark
	runner.dryRun = flags.dryRun
	runner.outputDir = flags.outputDir

	return runner
}

// SelectTestSuites determines which test suites should be run based on flags.
func selectTestSuites(runner *TestRunner, flags commandLineFlags) ([]TestSuite, error) {
	allSuites, err := runner.GetTestSuites()
	if err != nil {
		return nil, fmt.Errorf("failed to get test suites: %w", err)
	}

	if flags.suites != "" {
		return selectSpecificSuites(allSuites, flags.suites), nil
	}

	return filterSuitesByFlags(allSuites, flags), nil
}

// SelectSpecificSuites returns the requested test suites by name.
func selectSpecificSuites(allSuites []TestSuite, suiteNames string) []TestSuite {
	requestedSuites := strings.Split(suiteNames, ",")

	suiteMap := make(map[string]TestSuite)
	for _, suite := range allSuites {
		suiteMap[suite.Name] = suite
	}

	var suitesToRun []TestSuite

	for _, name := range requestedSuites {
		name = strings.TrimSpace(name)
		if suite, exists := suiteMap[name]; exists {
			suitesToRun = append(suitesToRun, suite)
		} else {
			log.Fatalf("Unknown test suite: %s", name)
		}
	}

	return suitesToRun
}

// FilterSuitesByFlags filters test suites based on command line flags.
func filterSuitesByFlags(allSuites []TestSuite, flags commandLineFlags) []TestSuite {
	var suitesToRun []TestSuite

	for _, suite := range allSuites {
		include := true

		if suite.Name == "integration" && !flags.integration {
			include = false
		}

		if suite.Name == "benchmark" && !flags.benchmark {
			include = false
		}

		if suite.Name == "coverage" && !flags.coverage {
			include = false
		}

		if include {
			suitesToRun = append(suitesToRun, suite)
		}
	}

	return suitesToRun
}

// PrintStartupInfo displays startup information to the user.
func printStartupInfo(suitesToRun []TestSuite, dryRun bool) {
	fmt.Printf("Git Provider Sync Test Runner\n")                //nolint:forbidigo // CLI output
	fmt.Printf("Started: %s\n", time.Now().Format(time.RFC3339)) //nolint:forbidigo // CLI output
	fmt.Printf("🧪 Test Suites: %d\n", len(suitesToRun))          //nolint:forbidigo // CLI output

	if dryRun {
		fmt.Printf("DRY RUN MODE - No tests will be executed\n") //nolint:forbidigo // CLI output
	}
}

// ExecuteTestSuites runs all the selected test suites.
func executeTestSuites(runner *TestRunner, suitesToRun []TestSuite, flags commandLineFlags) map[string][]TestResult {
	ctx, cancel := context.WithTimeout(context.Background(), flags.timeout)
	defer cancel()

	results := make(map[string][]TestResult)

	for _, suite := range suitesToRun {
		suiteResults, err := runner.RunTestSuite(ctx, suite)
		if err != nil {
			log.Printf("Error running suite %s: %v", suite.Name, err)
		}

		results[suite.Name] = suiteResults
	}

	return results
}

// GenerateOutput creates reports and saves results if not in dry run mode.
func generateOutput(runner *TestRunner, results map[string][]TestResult, flags commandLineFlags) {
	if !flags.dryRun {
		runner.GenerateReport(results)

		if err := runner.SaveResults(results); err != nil {
			log.Printf("Error saving results: %v", err)
		}
	}
}

func main() {
	flags := parseCommandLineFlags()
	runner := setupTestRunner(flags)

	// Validate test structure if requested
	if flags.validate {
		if err := runner.ValidateTestStructure(); err != nil {
			log.Fatalf("Test structure validation failed: %v", err)
		}

		return
	}

	// Determine which test suites to run
	suitesToRun, err := selectTestSuites(runner, flags)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)

		os.Exit(1)
	}

	// Print startup information
	printStartupInfo(suitesToRun, flags.dryRun)

	// Execute test suites
	results := executeTestSuites(runner, suitesToRun, flags)

	// Generate final output
	generateOutput(runner, results, flags)

	fmt.Printf("\nTest runner completed at %s\n", time.Now().Format(time.RFC3339)) //nolint:forbidigo // CLI output
}
