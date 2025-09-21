<!--
SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors

SPDX-License-Identifier: CC0-1.0

-->

# Functional Patterns in Git Provider Sync

Functional programming patterns in the domain layer.

## 1. Immutable Results with Functional Options

### Traditional Mutable Approach

```go
// Old way - mutating state
results := NewResults(false)
results.AddResult(result1)  // Mutates
results.AddResult(result2)  // Mutates
results.Complete()           // Mutates
```

### New Functional Approach

```go
// New way - immutable with functional options
results := sync.NewFunctionalResults(
    sync.WithDryRun(true),
    sync.WithStartTime(time.Now()),
)

// Chain operations - each returns new instance
results = results.
    WithResult(result1).
    WithResult(result2).
    WithTotalSources(5).
    WithTotalMirrors(3).
    WithCompletion()

// Or use builder pattern
results = sync.NewResultBuilder().
    DryRun(true).
    AddResult(result1).
    AddResult(result2).
    TotalSources(5).
    BuildCompleted()
```

## 2. Pure Validation Functions with Composition

### Traditional Validation

```go
// Old way - validation mixed with building
func (b RepositoryBuilder) Build() (Repository, error) {
    if b.repo.httpsURL == "" {  // Validation embedded
        return Repository{}, errors.New("URL required")
    }
    if len(b.repo.name) > 100 {  // More validation
        return Repository{}, errors.New("name too long")
    }
    // ... more checks
}
```

### New Functional Validation

```go
// Pure validators - no side effects
validator := validation.Compose(
    validation.NotEmpty,
    validation.MaxLength(100),
    validation.AlphaNumeric,
    validation.NoLeadingTrailing(".-"),
)

// Or use functional options
validator := validation.CreateValidator(
    validation.WithMaxLength(100),
    validation.WithMinLength(3),
    validation.WithPattern(`^[a-z]+$`),
    validation.WithReservedWords("admin", "root"),
)

// Or use builder pattern
validator := validation.NewValidatorBuilder().
    Required().
    Max(100).
    Min(3).
    Matches(`^[a-z]+$`).
    Custom(myCustomValidator).
    Build()

// Apply validation
if err := validator(name); err != nil {
    return err
}
```

## 3. Composable Filter Predicates

### Pure, Composable Filters

```go
// Define predicates
publicAndActive := filter.All(
    filter.IsPublic,
    filter.NotArchived,
    filter.ActiveWithin(30 * 24 * time.Hour),
)

importantRepos := filter.Any(
    filter.NamePrefix("core-"),
    filter.NameContains("critical"),
    filter.HasName("main-service"),
)

// Compose complex filters
finalFilter := filter.All(
    publicAndActive,
    importantRepos,
    filter.Not(filter.IsFork),
)

// Apply filter - pure function, returns new slice
filtered := filter.Filter(repositories, finalFilter)

// Or partition into two groups
matching, notMatching := filter.Partition(repositories, finalFilter)
```

## Benefits of These Patterns

1. **Testability** - Pure functions need no mocks
2. **Composability** - Small functions combine into complex behavior
3. **Immutability** - Prevents accidental mutation bugs
4. **Clarity** - Clear input/output relationships
5. **Reusability** - Functions can be reused across contexts

## Key Principles Applied

- **Pure Functions**: No side effects, deterministic output
- **Immutability**: Return new values instead of mutating
- **Composition**: Build complex behavior from simple functions
- **Functional Options**: Flexible configuration without overloading
- **Builder Pattern**: Fluent API for complex construction

## Idiomatic Go

These patterns stay idiomatic to Go:
- Standard error handling (no Result/Option types)
- Interfaces where appropriate
- No complex FP abstractions
- Performance-conscious (copy only when beneficial)
- Clear, readable code

## Usage Examples

### Example 1: Repository Validation

```go
// Provider-specific validation
err := validation.GitHubRepositoryName("my-repo")

// Custom validation pipeline
projectValidator := validation.Compose(
    validation.NotEmpty,
    validation.MaxLength(50),
    validation.Pattern(`^project-[a-z]+$`),
    func(s string) error {
        // Custom business rule
        if !strings.Contains(s, "-") {
            return errors.New("project name must contain hyphen")
        }
        return nil
    },
)
```

### Example 2: Results Aggregation

```go
// Functional results aggregation
results := sync.NewFunctionalResults(sync.WithDryRun(false))

for _, repo := range repositories {
    syncResult := syncRepository(repo)
    results = results.WithResult(syncResult)
}

finalResults := results.
    WithTotalRepositories(len(repositories)).
    WithCompletion()
```

### Example 3: Dynamic Filtering

```go
// Build filter based on options
var predicates []filter.Predicate

if !opts.IncludeForks {
    predicates = append(predicates, filter.NotFork)
}

if !opts.IncludeArchived {
    predicates = append(predicates, filter.NotArchived)
}

if opts.ActiveDays > 0 {
    duration := time.Duration(opts.ActiveDays) * 24 * time.Hour
    predicates = append(predicates, filter.ActiveWithin(duration))
}

// Combine all predicates
finalFilter := filter.All(predicates...)
filtered := filter.Filter(repos, finalFilter)
```

## Migration Strategy

These patterns are added alongside existing code:
1. New code uses functional patterns
2. Existing code migrates gradually
3. Both approaches work together
4. No breaking changes required

Functional patterns improve the codebase without requiring a rewrite, maintaining backward compatibility while improving testability and composability.
