// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package testutil

import (
	"context"
	"testing"
)

// TestContext provides a complete test environment with all helpers.
//
//nolint:containedctx // Test utility needs to store context for test lifecycle
type TestContext struct {
	T       testing.TB
	FS      *AferoTestFS       // Virtual filesystem (Afero MemMapFs)
	Fixture *TestFixture       // OS-based fixture (for tests needing real FS)
	Config  *TestConfigFixture // Configuration helpers
	Mocks   *StandardMocks     // Standard mock objects
	Context context.Context    // Context for operations
}

// UnitTest sets up a unit test environment with virtual filesystem.
// This is the recommended setup for most unit tests.
// Uses in-memory filesystem for ultra-fast operations.
func UnitTest(t *testing.T) *TestContext {
	t.Helper()

	// Verify test isolation
	RequireTestEnvironment(t)

	return &TestContext{
		T:       t,
		FS:      NewMemFS(t),
		Config:  NewTestConfigFixture(t),
		Mocks:   NewStandardMocks(t),
		Context: context.Background(),
	}
}

// IntegrationTest sets up an integration test environment.
// Uses real filesystem operations but in isolated temp directories.
func IntegrationTest(t *testing.T) *TestContext {
	t.Helper()

	// Verify test isolation
	RequireTestEnvironment(t)

	fixture := NewTestFixture(t)

	return &TestContext{
		T:       t,
		FS:      NewOsFS(t),
		Fixture: fixture,
		Config:  NewTestConfigFixture(t),
		Mocks:   NewStandardMocks(t),
		Context: context.Background(),
	}
}

// BenchmarkTest sets up a benchmark test environment.
// Optimized for performance testing with minimal overhead.
func BenchmarkTest(b *testing.B) *TestContext {
	b.Helper()

	return &TestContext{
		T:       b,
		FS:      NewMemFS(b), // Use in-memory for benchmarks
		Config:  NewTestConfigFixture(b),
		Mocks:   nil, // Mocks not typically used in benchmarks
		Context: context.Background(),
	}
}

// ParallelTest sets up a test that can run in parallel.
// Note: Cannot use t.Setenv() in parallel tests.
func ParallelTest(t *testing.T) *TestContext {
	t.Helper()
	t.Parallel()

	// Use in-memory filesystem for parallel safety
	return &TestContext{
		T:       t,
		FS:      NewMemFS(t),
		Config:  NewTestConfigFixture(t),
		Mocks:   NewStandardMocks(t),
		Context: context.Background(),
	}
}

// Cleanup performs any necessary cleanup.
// This is called automatically by testing.TB cleanup.
func (tc *TestContext) Cleanup() {
	tc.T.Helper()

	// Assert mock expectations if mocks were used
	if tc.Mocks != nil {
		if t, ok := tc.T.(*testing.T); ok {
			tc.Mocks.AssertExpectations(t)
		}
	}
}

// WriteConfig writes a configuration file and returns its path.
func (tc *TestContext) WriteConfig(content string) string {
	tc.T.Helper()

	if tc.FS != nil {
		return tc.FS.WriteConfig(content)
	} else if tc.Fixture != nil {
		return tc.Fixture.WriteConfig(content)
	}

	tc.T.Fatal("No filesystem available")

	return ""
}

// CreateFile creates a file with content and returns its path.
func (tc *TestContext) CreateFile(name, content string) string {
	tc.T.Helper()

	if tc.FS != nil {
		return tc.FS.CreateFile(name, content)
	} else if tc.Fixture != nil {
		return tc.Fixture.WriteFile(name, content)
	}

	tc.T.Fatal("No filesystem available")

	return ""
}

// CreateDir creates a directory and returns its path.
func (tc *TestContext) CreateDir(name string) string {
	tc.T.Helper()

	if tc.FS != nil {
		tc.FS.CreateDir(name)

		return tc.FS.Path(name)
	} else if tc.Fixture != nil {
		return tc.Fixture.Path(name)
	}

	tc.T.Fatal("No filesystem available")

	return ""
}

// QuickTest provides the absolute minimal test setup.
// Use this for simple tests that don't need filesystem or mocks.
//
//nolint:containedctx // Test utility needs to store context for test lifecycle
type QuickTest struct {
	T       testing.TB
	Context context.Context
}

// Quick creates a minimal test environment.
func Quick(tb testing.TB) *QuickTest {
	tb.Helper()

	return &QuickTest{
		T:       tb,
		Context: context.Background(),
	}
}

// TestType represents the type of test being run.
type TestType int

const (
	// TypeUnit represents a unit test.
	TypeUnit TestType = iota
	// TypeIntegration represents an integration test.
	TypeIntegration
	// TypeBenchmark represents a benchmark test.
	TypeBenchmark
	// TypeFuzz represents a fuzz test.
	TypeFuzz
)

// TestSetup provides a unified setup function based on test type.
func TestSetup(tb testing.TB, testType TestType) *TestContext {
	tb.Helper()

	switch testType {
	case TypeUnit:
		if test, ok := tb.(*testing.T); ok {
			return UnitTest(test)
		}
	case TypeIntegration:
		if test, ok := tb.(*testing.T); ok {
			return IntegrationTest(test)
		}
	case TypeBenchmark:
		if bench, ok := tb.(*testing.B); ok {
			return BenchmarkTest(bench)
		}
	case TypeFuzz:
		if test, ok := tb.(*testing.T); ok {
			// Fuzz tests should use minimal setup
			return UnitTest(test)
		}
	}

	// Fallback to unit test setup
	if test, ok := tb.(*testing.T); ok {
		return UnitTest(test)
	}

	tb.Fatal("Unsupported test type")

	return nil
}

// TestWith provides a fluent API for test setup.
type TestWith struct {
	t   testing.TB
	ctx *TestContext
}

// With starts a fluent test setup.
func With(tb testing.TB) *TestWith {
	tb.Helper()

	return &TestWith{
		t:   tb,
		ctx: nil,
	}
}

// MemoryFS uses in-memory filesystem.
func (tw *TestWith) MemoryFS() *TestWith {
	tw.ctx = &TestContext{
		T:       tw.t,
		FS:      NewMemFS(tw.t),
		Context: context.Background(),
	}

	return tw
}

// OsFS uses OS-backed filesystem.
func (tw *TestWith) OsFS() *TestWith {
	tw.ctx = &TestContext{
		T:       tw.t,
		FS:      NewOsFS(tw.t),
		Context: context.Background(),
	}

	return tw
}

// Mocks adds standard mocks.
func (tw *TestWith) Mocks() *TestWith {
	if tw.ctx == nil {
		tw.ctx = &TestContext{
			T:       tw.t,
			Context: context.Background(),
		}
	}

	if t, ok := tw.t.(*testing.T); ok {
		tw.ctx.Mocks = NewStandardMocks(t)
	}

	return tw
}

// Config adds configuration helpers.
func (tw *TestWith) Config() *TestWith {
	if tw.ctx == nil {
		tw.ctx = &TestContext{
			T:       tw.t,
			Context: context.Background(),
		}
	}

	tw.ctx.Config = NewTestConfigFixture(tw.t)

	return tw
}

// Parallel marks the test as parallel.
func (tw *TestWith) Parallel() *TestWith {
	if t, ok := tw.t.(*testing.T); ok {
		t.Parallel()
	}

	return tw
}

// Build returns the configured test context.
func (tw *TestWith) Build() *TestContext {
	if tw.ctx == nil {
		// Default to unit test setup
		if t, ok := tw.t.(*testing.T); ok {
			tw.ctx = UnitTest(t)
		} else {
			tw.t.Fatal("Cannot build test context")
		}
	}

	return tw.ctx
}

// Example usage patterns:

// ExampleUnitTest shows how to use the unit test helper.
func ExampleUnitTest() {
	// In your test:
	// func TestSomething(t *testing.T) {
	//     ctx := testutil.UnitTest(t)
	//     defer ctx.Cleanup()
	//
	//     // Write files to virtual filesystem
	//     ctx.FS.WriteFileString("test.txt", "content")
	//
	//     // Use mocks
	//     ctx.Mocks.Logger.On("Info", mock.Anything, "test", mock.Anything).Once()
	//
	//     // Your test logic here
	// }
}

// ExampleFluentAPI shows how to use the fluent API.
func ExampleFluentAPI() {
	// In your test:
	// func TestSomething(t *testing.T) {
	//     ctx := testutil.With(t).
	//         MemoryFS().
	//         Mocks().
	//         Config().
	//         Parallel().
	//         Build()
	//
	//     // Your test logic here
	// }
}

// ExampleIntegrationTest shows integration test setup.
func ExampleIntegrationTest() {
	// In your test (with //go:build integration tag):
	// func TestIntegration(t *testing.T) {
	//     ctx := testutil.IntegrationTest(t)
	//     defer ctx.Cleanup()
	//
	//     // Uses real filesystem in temp directory
	//     configPath := ctx.WriteConfig(testutil.StandardTestConfigs.GitHub)
	//
	//     // Your integration test logic here
	// }
}

// TestScenario represents a test scenario with setup and assertions.
type TestScenario struct {
	Name    string
	Setup   func(*TestContext)
	Run     func(*TestContext) error
	Assert  func(*TestContext, error)
	Cleanup func(*TestContext)
}

// RunScenarios executes multiple test scenarios.
func RunScenarios(t *testing.T, scenarios []TestScenario) {
	t.Helper()

	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(t *testing.T) {
			ctx := UnitTest(t)
			defer ctx.Cleanup()

			// Setup
			if scenario.Setup != nil {
				scenario.Setup(ctx)
			}

			// Run
			var err error
			if scenario.Run != nil {
				err = scenario.Run(ctx)
			}

			// Assert
			if scenario.Assert != nil {
				scenario.Assert(ctx, err)
			}

			// Cleanup
			if scenario.Cleanup != nil {
				scenario.Cleanup(ctx)
			}
		})
	}
}
