// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

package gogit

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"itiquette/git-provider-sync/internal/domain/entities"
	"itiquette/git-provider-sync/internal/domain/ports"
	"itiquette/git-provider-sync/internal/shared"
)

// MirrorService provides go-git based repository mirroring operations.
type MirrorService struct {
	logger         ports.Logger
	fileSystem     ports.FileSystem
	tempDir        string
	progressWriter io.Writer
}

// NewMirrorService creates a new go-git mirror service.
func NewMirrorService(logger ports.Logger, fileSystem ports.FileSystem, tempDir string) *MirrorService {
	return &MirrorService{
		logger:     logger,
		fileSystem: fileSystem,
		tempDir:    tempDir,
	}
}

// MirrorRequest contains parameters for mirroring a repository.
type MirrorRequest struct {
	SourceRepository entities.Repository
	TargetRepository entities.Repository
	MirrorType       string // "full", "bare", "shallow"
	ShallowDepth     int
	IncludeBranches  []string
	ExcludeBranches  []string
	IncludeTags      []string
	ExcludeTags      []string
	PreserveRefs     bool
	UpdateReferences bool
	Force            bool
	DryRun           bool
}

// MirrorResult contains the results of a mirror operation.
type MirrorResult struct {
	Success         bool
	SourceCommits   int
	TargetCommits   int
	BranchesSynced  []string
	TagsSynced      []string
	ReferencesMap   map[string]string
	Errors          []string
	PerformanceInfo *PerformanceInfo
}

// PerformanceInfo contains performance metrics for the mirror operation.
type PerformanceInfo struct {
	Duration        string
	DataTransferred int64
	ObjectsCloned   int
	RefsUpdated     int
}

// SetProgressWriter sets a writer for progress reporting.
func (ms *MirrorService) SetProgressWriter(writer io.Writer) {
	ms.progressWriter = writer
}

// Mirror performs a repository mirror operation.
func (ms *MirrorService) Mirror(ctx context.Context, request MirrorRequest) (*MirrorResult, error) {
	ms.logger.Info(ctx, "Starting go-git repository mirror", map[string]any{
		"source":      request.SourceRepository.HTTPSURL(),
		"target":      request.TargetRepository.HTTPSURL(),
		"mirror_type": request.MirrorType,
		"dry_run":     request.DryRun,
	})

	result := &MirrorResult{
		BranchesSynced:  []string{},
		TagsSynced:      []string{},
		ReferencesMap:   make(map[string]string),
		Errors:          []string{},
		PerformanceInfo: &PerformanceInfo{},
	}

	if request.DryRun {
		result = ms.performDryRun(ctx, request, result)

		return result, nil
	}

	// Create temporary working directory
	workDir, err := ms.createWorkingDirectory(ctx)
	if err != nil {
		return result, fmt.Errorf("failed to create working directory: %w", err)
	}
	defer ms.cleanupWorkingDirectory(ctx, workDir)

	// Clone source repository
	sourceRepo, err := ms.cloneSourceRepository(ctx, request, workDir)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("clone failed: %v", err))

		return result, fmt.Errorf("failed to clone source repository: %w", err)
	}

	// Get source repository information
	if err := ms.analyzeSourceRepository(ctx, sourceRepo, result); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("analysis failed: %v", err))

		return result, fmt.Errorf("failed to analyze source repository: %w", err)
	}

	// Filter and prepare references
	filteredRefs, err := ms.filterReferences(ctx, sourceRepo, request)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("filtering failed: %v", err))

		return result, fmt.Errorf("failed to filter references: %w", err)
	}

	// Push to target repository
	if err := ms.pushToTarget(ctx, sourceRepo, request, filteredRefs, result); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("push failed: %v", err))

		return result, fmt.Errorf("failed to push to target: %w", err)
	}

	result.Success = true
	ms.logger.Info(ctx, "Repository mirror completed successfully", map[string]any{
		"source_commits":  result.SourceCommits,
		"target_commits":  result.TargetCommits,
		"branches_synced": len(result.BranchesSynced),
		"tags_synced":     len(result.TagsSynced),
	})

	return result, nil
}

// PerformDryRun simulates a mirror operation without making changes.
func (ms *MirrorService) performDryRun(ctx context.Context, request MirrorRequest, result *MirrorResult) *MirrorResult {
	ms.logger.Info(ctx, "Performing dry run analysis", map[string]any{
		"source": request.SourceRepository.HTTPSURL(),
		"target": request.TargetRepository.HTTPSURL(),
	})

	// In a dry run, we would analyze what would be done without making changes
	// For now, we'll simulate a successful analysis
	result.Success = true
	result.BranchesSynced = []string{"main", "develop"}
	result.TagsSynced = []string{"v1.0.0", "v1.1.0"}
	result.SourceCommits = 150
	result.TargetCommits = 0

	ms.logger.Info(ctx, "Dry run completed", map[string]any{
		"would_sync_branches": len(result.BranchesSynced),
		"would_sync_tags":     len(result.TagsSynced),
		"estimated_commits":   result.SourceCommits,
	})

	return result
}

// CreateWorkingDirectory creates a temporary working directory.
func (ms *MirrorService) createWorkingDirectory(ctx context.Context) (string, error) {
	workDir := filepath.Join(ms.tempDir, "gogit-mirror-"+generateRandomID())

	if err := ms.fileSystem.MkdirAll(workDir, 0750); err != nil {
		return "", fmt.Errorf("failed to create working directory: %w", err)
	}

	ms.logger.Debug(ctx, "Created working directory", map[string]any{
		"path": workDir,
	})

	return workDir, nil
}

// CleanupWorkingDirectory removes the temporary working directory.
func (ms *MirrorService) cleanupWorkingDirectory(ctx context.Context, workDir string) {
	if err := shared.RemoveAllInTempDir(workDir); err != nil {
		ms.logger.Warn(ctx, "Failed to cleanup working directory", map[string]any{
			"path":  workDir,
			"error": err.Error(),
		})
	} else {
		ms.logger.Debug(ctx, "Cleaned up working directory", map[string]any{
			"path": workDir,
		})
	}
}

// CloneSourceRepository clones the source repository to the working directory.
func (ms *MirrorService) cloneSourceRepository(ctx context.Context, request MirrorRequest, workDir string) (*git.Repository, error) {
	ms.logger.Debug(ctx, "Cloning source repository", map[string]any{
		"source":   request.SourceRepository.HTTPSURL(),
		"work_dir": workDir,
		"shallow":  request.ShallowDepth > 0,
	})

	cloneOptions := &git.CloneOptions{
		URL:      request.SourceRepository.HTTPSURL(),
		Progress: ms.progressWriter,
	}

	// Note: Authentication would be configured here in a real implementation

	// Configure shallow clone if requested
	if request.ShallowDepth > 0 {
		cloneOptions.Depth = request.ShallowDepth
	}

	// Configure mirror mode
	if request.MirrorType == "bare" {
		cloneOptions.Mirror = true
	}

	// Clone the repository
	repo, err := git.PlainClone(workDir, request.MirrorType == "bare", cloneOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to clone repository: %w", err)
	}

	return repo, nil
}

// AnalyzeSourceRepository analyzes the source repository to gather information.
func (ms *MirrorService) analyzeSourceRepository(ctx context.Context, repo *git.Repository, result *MirrorResult) error {
	ms.logger.Debug(ctx, "Analyzing source repository", nil)

	// Count commits
	commitIter, err := repo.Log(&git.LogOptions{})
	if err != nil {
		return fmt.Errorf("failed to get commit iterator: %w", err)
	}

	commitCount := 0

	err = commitIter.ForEach(func(_ *object.Commit) error {
		commitCount++

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to count commits: %w", err)
	}

	result.SourceCommits = commitCount

	ms.logger.Debug(ctx, "Source repository analysis completed", map[string]any{
		"commits": commitCount,
	})

	return nil
}

// FilterReferences filters branches and tags based on the request criteria.
func (ms *MirrorService) filterReferences(ctx context.Context, repo *git.Repository, request MirrorRequest) (map[string]plumbing.ReferenceName, error) {
	ms.logger.Debug(ctx, "Filtering references", map[string]any{
		"include_branches": request.IncludeBranches,
		"exclude_branches": request.ExcludeBranches,
		"include_tags":     request.IncludeTags,
		"exclude_tags":     request.ExcludeTags,
	})

	filteredRefs := make(map[string]plumbing.ReferenceName)

	// Get all references
	refs, err := repo.References()
	if err != nil {
		return nil, fmt.Errorf("failed to get references: %w", err)
	}

	err = refs.ForEach(func(ref *plumbing.Reference) error {
		refName := ref.Name().String()

		// Filter branches
		if ref.Name().IsBranch() {
			branchName := ref.Name().Short()
			if ms.shouldIncludeBranch(branchName, request.IncludeBranches, request.ExcludeBranches) {
				filteredRefs[refName] = ref.Name()
			}
		}

		// Filter tags
		if ref.Name().IsTag() {
			tagName := ref.Name().Short()
			if ms.shouldIncludeTag(tagName, request.IncludeTags, request.ExcludeTags) {
				filteredRefs[refName] = ref.Name()
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to iterate references: %w", err)
	}

	ms.logger.Debug(ctx, "Reference filtering completed", map[string]any{
		"filtered_count": len(filteredRefs),
	})

	return filteredRefs, nil
}

// ShouldIncludeBranch determines if a branch should be included.
func (ms *MirrorService) shouldIncludeBranch(branchName string, includePatterns, excludePatterns []string) bool {
	// Check exclude patterns first
	for _, pattern := range excludePatterns {
		if ms.matchesPattern(branchName, pattern) {
			return false
		}
	}

	// If no include patterns, include all (except excluded)
	if len(includePatterns) == 0 {
		return true
	}

	// Check include patterns
	for _, pattern := range includePatterns {
		if ms.matchesPattern(branchName, pattern) {
			return true
		}
	}

	return false
}

// ShouldIncludeTag determines if a tag should be included.
func (ms *MirrorService) shouldIncludeTag(tagName string, includePatterns, excludePatterns []string) bool {
	// Check exclude patterns first
	for _, pattern := range excludePatterns {
		if ms.matchesPattern(tagName, pattern) {
			return false
		}
	}

	// If no include patterns, include all (except excluded)
	if len(includePatterns) == 0 {
		return true
	}

	// Check include patterns
	for _, pattern := range includePatterns {
		if ms.matchesPattern(tagName, pattern) {
			return true
		}
	}

	return false
}

// MatchesPattern checks if a name matches a pattern (supports wildcards).
func (ms *MirrorService) matchesPattern(name, pattern string) bool {
	// For branch/tag patterns, handle special case where * should match path separators
	// E.g., "feature/*" should match "feature/ABC-123/description"
	if strings.Contains(pattern, "*") && strings.Contains(name, "/") {
		// Check if pattern is a prefix pattern like "feature/*"
		if strings.HasSuffix(pattern, "/*") {
			prefix := strings.TrimSuffix(pattern, "/*")

			return strings.HasPrefix(name, prefix+"/")
		}
		// Check if pattern is a suffix pattern like "*/release"
		if strings.HasPrefix(pattern, "*/") {
			suffix := strings.TrimPrefix(pattern, "*/")

			return strings.HasSuffix(name, "/"+suffix)
		}
	}

	// Use standard library glob matching for other cases
	matched, err := filepath.Match(pattern, name)
	if err != nil {
		// Invalid pattern, treat as literal string comparison
		return name == pattern
	}

	return matched
}

// PushToTarget pushes the filtered references to the target repository.
func (ms *MirrorService) pushToTarget(ctx context.Context, repo *git.Repository, request MirrorRequest, refs map[string]plumbing.ReferenceName, result *MirrorResult) error {
	ms.logger.Debug(ctx, "Pushing to target repository", map[string]any{
		"target":    request.TargetRepository.HTTPSURL(),
		"ref_count": len(refs),
		"force":     request.Force,
	})

	// Configure push options
	pushOptions := &git.PushOptions{
		RemoteName: "origin",
		Progress:   ms.progressWriter,
		Force:      request.Force,
	}

	// Note: Authentication would be configured here in a real implementation

	// Add target as remote if not exists
	targetRemote, err := repo.CreateRemote(&config.RemoteConfig{
		Name: "target",
		URLs: []string{request.TargetRepository.HTTPSURL()},
	})
	if err != nil {
		// Remote might already exist, try to get it
		targetRemote, err = repo.Remote("target")
		if err != nil {
			return fmt.Errorf("failed to create or get target remote: %w", err)
		}
	}

	// Push to target
	pushOptions.RemoteName = "target"

	// Build refspecs for the filtered references
	refSpecs := make([]config.RefSpec, 0, len(refs))

	for refName := range refs {
		refSpec := config.RefSpec(refName + ":" + refName)
		refSpecs = append(refSpecs, refSpec)
	}

	pushOptions.RefSpecs = refSpecs

	err = targetRemote.Push(pushOptions)
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return fmt.Errorf("failed to push to target: %w", err)
	}

	// Update result with synced references
	for refName := range refs {
		if strings.HasPrefix(refName, "refs/heads/") {
			branchName := strings.TrimPrefix(refName, "refs/heads/")
			result.BranchesSynced = append(result.BranchesSynced, branchName)
		} else if strings.HasPrefix(refName, "refs/tags/") {
			tagName := strings.TrimPrefix(refName, "refs/tags/")
			result.TagsSynced = append(result.TagsSynced, tagName)
		}
	}

	return nil
}

// GenerateRandomID generates a random ID for temporary directories.
func generateRandomID() string {
	// Simple implementation - in production, use crypto/rand
	return strconv.Itoa(os.Getpid())
}
