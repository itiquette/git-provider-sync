// SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors
// SPDX-License-Identifier: EUPL-1.2

// Package domain defines core domain errors and constants for git-provider-sync
package domain

import (
	"context"
	"errors"
	"os"
	"strings"
)

// Common static errors used throughout the application.
var (
	ErrConfigurationNotFound   = errors.New("failed to find a configuration")
	ErrConfigurationLoad       = errors.New("error loading config")
	ErrConfigurationUnmarshal  = errors.New("error unmarshalling yaml config")
	ErrConfigurationValidation = errors.New("failed to validate configuration")
	ErrNetworkTimeout          = errors.New("network timeout while fetching configuration")
	ErrInvalidProvider         = errors.New("unsupported provider")
	ErrMissingOwner            = errors.New("no owner configured")
	ErrMissingDomain           = errors.New("no domain configured")
	ErrRepositoryNotFound      = errors.New("repository not found")
	ErrAuthenticationFailed    = errors.New("authentication failed")
	ErrPermissionDenied        = errors.New("permission denied")
	ErrInvalidConfiguration    = errors.New("invalid configuration")
	ErrUnsupportedOperation    = errors.New("unsupported operation")
	ErrInvalidFormat           = errors.New("invalid format")

	// Repository errors.
	ErrRepositoryNameEmpty         = errors.New("repository name cannot be empty")
	ErrRepositoryNameTooLong       = errors.New("repository name too long (max 100 characters)")
	ErrRepositoryNameTooLongGitLab = errors.New("repository name too long (max 255 characters)")
	ErrRepositoryNameInvalidChars  = errors.New("repository name contains invalid characters")
	ErrRepositoryNameInvalid       = errors.New("invalid repository name")
	ErrRepositoryNil               = errors.New("repository is nil")
	ErrProjectNil                  = errors.New("project is nil")
	ErrProjectNameEmpty            = errors.New("project name cannot be empty")
	ErrProjectNameTooLong          = errors.New("project name too long (max 100 characters)")
	ErrProjectNamePeriodBounds     = errors.New("project name cannot start or end with a period")
	ErrProjectNameHyphenBounds     = errors.New("project name cannot start or end with a hyphen")
	ErrProjectNameInvalidChar      = errors.New("project name contains invalid character")
	ErrProjectNameReserved         = errors.New("project name is reserved")
	ErrRepositoryNoID              = errors.New("created repository has no ID")
	ErrRepositoryNoName            = errors.New("repository has no name")
	ErrRepositoryNoOwner           = errors.New("repository has no owner")
	ErrRepositoryInBothLists       = errors.New("repository appears in both include and exclude lists")

	// Git operation errors.
	ErrCannotCheckoutBare   = errors.New("cannot checkout branch in bare repository")
	ErrCannotPullBare       = errors.New("cannot pull in bare repository")
	ErrMaxCountReached      = errors.New("reached max count")
	ErrSSHAgentNotSupported = errors.New("SSH agent auth not supported by go-git adapter")
	ErrSSHKeyRequired       = errors.New("SSH key auth requires either SSHKeyPath or SSHKey")
	ErrInvalidSSHKey        = errors.New("invalid SSH private key")
	ErrUnsupportedAuthType  = errors.New("unsupported auth type")

	// Configuration errors.
	ErrInvalidConfig        = errors.New("invalid configuration detected")
	ErrBothIncludeExclude   = errors.New("cannot specify both include and exclude patterns")
	ErrInvalidURL           = errors.New("invalid URL format")
	ErrProtocolNotSupported = errors.New("protocol not supported")

	// Mirror errors.
	ErrMirrorFailed = errors.New("mirror operation failed")
	ErrCloneFailed  = errors.New("clone operation failed")
	ErrPushFailed   = errors.New("push operation failed")
	ErrFetchFailed  = errors.New("fetch operation failed")

	// Generic operation errors.
	ErrOperationFailed = errors.New("operation failed")
	ErrInvalidState    = errors.New("invalid state")
	ErrResourceBusy    = errors.New("resource is busy")
	ErrTimeout         = errors.New("operation timed out")

	// Directory repository errors.
	ErrBranchOpsNotSupported       = errors.New("branch operations not supported for directory repositories")
	ErrRemoteOpsNotSupported       = errors.New("remote operations not supported for directory repositories")
	ErrPushOpsNotSupported         = errors.New("push operations not supported for directory repositories")
	ErrNoSourceURLConfigured       = errors.New("no source URL configured for fetch")
	ErrFetchNotSupportedForURLType = errors.New("fetch not supported for URL type")
	ErrTagOpsNotSupported          = errors.New("tag operations not supported for directory repositories")
	ErrDiffOpsNotSupported         = errors.New("diff operations not supported for directory repositories")
	ErrTargetPathAlreadyExists     = errors.New("target path already exists")
	ErrTargetPathDoesNotExist      = errors.New("target path does not exist")

	// Git repository errors.
	ErrNotGitRepository = errors.New("not a git repository")
	ErrWorkingDirEmpty  = errors.New("failed to run git command, workingDir was empty")

	// HTTP client errors.
	ErrRequestTimeoutMustBePositive            = errors.New("request timeout must be positive")
	ErrIdleConnectionTimeoutMustBePositive     = errors.New("idle connection timeout must be positive")
	ErrDialTimeoutMustBePositive               = errors.New("dial timeout must be positive")
	ErrKeepAliveMustBePositive                 = errors.New("keep alive must be positive")
	ErrMaxIdleConnectionsMustBePositive        = errors.New("max idle connections must be positive")
	ErrMaxIdleConnectionsPerHostMustBePositive = errors.New("max idle connections per host must be positive")
	ErrMaxConnectionsPerHostMustBePositive     = errors.New("max connections per host must be positive")

	// Archive repository errors.
	ErrPullNotSupportedArchive      = errors.New("pull is not supported for archive repositories")
	ErrPushNotSupportedArchive      = errors.New("push is not supported for archive repositories")
	ErrAddFilesNotSupportedArchive  = errors.New("adding files is not supported for archive repositories")
	ErrCommitNotSupportedArchive    = errors.New("committing is not supported for archive repositories")
	ErrRemoteURLNotSupportedArchive = errors.New("setting remote URLs is not supported for archive repositories")
	ErrFilePathOutsideRepository    = errors.New("file path is outside repository")

	// Bulk operation errors.
	ErrBulkProtectionFailed = errors.New("protection failed for multiple repositories")
	ErrBulkRemovalFailed    = errors.New("protection removal failed for multiple repositories")

	// Adapter errors.
	ErrAdapterNotInitialized = errors.New("adapter not initialized")
	ErrNotYetImplemented     = errors.New("operation not yet implemented")

	// Clone/Push option errors.
	ErrRepositoryNameRequired = errors.New("repository name is required")
	ErrSourceURLRequired      = errors.New("source URL is required")
	ErrNoCloneURLs            = errors.New("repository has no clone URLs")
	ErrTargetURLRequired      = errors.New("target URL is required")
	ErrRefspecRequired        = errors.New("at least one refspec is required")
	ErrRemoteNameRequired     = errors.New("remote name is required")

	// Remote errors.
	ErrOriginRemoteNotFound       = errors.New("origin remote not found")
	ErrUpstreamRemoteNotFound     = errors.New("GPSUPSTREAM remote not found after creation")
	ErrRepositoryCreationDisabled = errors.New("repository does not exist and creation is disabled")
	ErrNamespaceNotFound          = errors.New("namespace not found")

	// Path/filesystem errors.
	ErrPathNotDirectory = errors.New("path exists but is not a directory")
	ErrPathIsDirectory  = errors.New("path exists but is a directory")

	// System errors.
	ErrSystemCriticalIssues = errors.New("system has critical issues")

	// Validation errors.
	ErrUnsupportedConnectivityType = errors.New("unsupported connectivity validation type")
	ErrUnsupportedFileSystemType   = errors.New("unsupported file system validation type")
	ErrHTTPRequestFailed           = errors.New("HTTP request returned non-2xx/3xx status")
	ErrProviderAPIServerError      = errors.New("provider API returned server error")

	// Configuration validation errors.
	ErrNoSyncConfigurations         = errors.New("no git provider sync configurations found")
	ErrNoSyncConfigInEnvironment    = errors.New("environment has no sync configurations")
	ErrSSHRewriteBothRequired       = errors.New("if either SSH URL rewrite parameter is specified, both must be provided")
	ErrInvalidVisibilitySetting     = errors.New("invalid visibility setting")
	ErrSSHAuthSockNotSet            = errors.New("SSH_AUTH_SOCK environment variable not set")
	ErrSSHAgentNoKeys               = errors.New("SSH agent is running but has no keys")
	ErrSSHCommandMustStartWithSSH   = errors.New("SSH command must start with 'ssh'")
	ErrDomainRequired               = errors.New("domain is required")
	ErrDomainShouldNotIncludeScheme = errors.New("domain should not include protocol scheme")

	// Entity validation errors.
	ErrMirrorNameEmpty             = errors.New("mirror name cannot be empty")
	ErrMirrorNameRequired          = errors.New("mirror name is required")
	ErrOwnerRequiredForProvider    = errors.New("owner is required for provider")
	ErrDirectoryPathMustBeAbsolute = errors.New("directory path must be absolute")

	// Factory errors.
	ErrUnsupportedGitImplementation      = errors.New("unsupported git implementation")
	ErrUnsupportedProvider               = errors.New("unsupported provider")
	ErrUnsupportedProviderType           = errors.New("unsupported provider type")
	ErrUserNameRequired                  = errors.New("user name is required")
	ErrUserEmailRequired                 = errors.New("user email is required")
	ErrMaxConcurrentMustBePositive       = errors.New("max concurrent must be positive")
	ErrImplementationNotAvailable        = errors.New("implementation not available")
	ErrUnknownImplementation             = errors.New("unknown implementation")
	ErrCountMustBePositive               = errors.New("count must be positive")
	ErrCertificateDirectoryNotExist      = errors.New("certificate directory does not exist")
	ErrOwnerRequired                     = errors.New("owner is required")
	ErrAuthenticationRequired            = errors.New("authentication is required")
	ErrAuthenticationRequiredForProvider = errors.New("authentication required for provider")
	ErrInvalidDomainFormat               = errors.New("invalid domain format")
	ErrProviderTypeRequired              = errors.New("provider type is required")

	// Mirror operation errors.
	ErrValidationFailed                = errors.New("validation failed")
	ErrUnknownEffectType               = errors.New("unknown effect type")
	ErrEffectTypeShouldNotReachHandler = errors.New("effect type should not reach utility effects handler")
	ErrCloneEffectMissingURL           = errors.New("clone effect missing url parameter")
	ErrCloneEffectMissingPath          = errors.New("clone effect missing local_path parameter")
	ErrCloneEffectMissingAuth          = errors.New("clone effect missing auth parameter")
	ErrCreateRepoMissingName           = errors.New("create repository effect missing name parameter")
	ErrCreateRepoMissingOwner          = errors.New("create repository effect missing owner parameter")
	ErrPathRequiredForProvider         = errors.New("path is required for provider")
	ErrPushEffectMissingPath           = errors.New("push effect missing local_path parameter")
	ErrPushEffectMissingAuth           = errors.New("push effect missing auth parameter")
	ErrUpdateEffectMissingRepo         = errors.New("update effect missing repository parameter")
	ErrUpdateEffectMissingParam        = errors.New("update effect missing parameter")
	ErrUpdateTopicsMissingRepo         = errors.New("update topics effect missing repository parameter")
	ErrUpdateTopicsMissingTopics       = errors.New("update topics effect missing topics parameter")
	ErrCleanupMissingPath              = errors.New("cleanup effect missing local_path parameter")

	// Validation errors.
	ErrFieldRequired         = errors.New("field is required")
	ErrValueNotInAllowedList = errors.New("value not in allowed list")

	// Sync operation errors.
	ErrActiveSinceAfterInactiveSince    = errors.New("active since time cannot be after inactive since time")
	ErrMinSizeGreaterThanMaxSize        = errors.New("minimum size cannot be greater than maximum size")
	ErrRepositoryNameInvalidForProvider = errors.New("repository name is invalid for provider")
	ErrPushToProviderFailed             = errors.New("push to provider failed")
	ErrUnknownProtectionOperation       = errors.New("unknown protection operation")
	ErrProviderNoProtectionSupport      = errors.New("provider does not support branch protection")
	ErrFailedToAuthenticateProvider     = errors.New("failed to authenticate with provider")
	ErrCloneFailedRepoNotFound          = errors.New("clone failed: repository not found")
	ErrRemoteMismatch                   = errors.New("mismatch in GPSUPSTREAM vs origin remote")
	ErrNoEnvironmentsConfigured         = errors.New("no environments configured")
	ErrConfigurationValidationFailed    = errors.New("configuration validation failed")
	ErrCLIOptionRetrievalFailed         = errors.New("failed to retrieve or type-assert CLIOption from context")
	ErrTempDirectoryNotFound            = errors.New("temporary directory path not found in context or is empty")
	ErrTestNotFound                     = errors.New("not found")
	ErrTestAuthenticationFailed         = errors.New("push failed: authentication failed")
	ErrInvalidRepositoryPath            = errors.New("invalid repository path")
	ErrInvalidRemoteURL                 = errors.New("invalid remote URL")
	ErrRemoteNotFound                   = errors.New("remote not found")
	ErrRepositoryCorrupted              = errors.New("repository is corrupted")
	ErrPathNotFound                     = errors.New("path not found")
	ErrInsufficientPermissions          = errors.New("insufficient permissions")
)

// Exit codes used by the application following Unix conventions:
// 0: Success
// 1: General operational errors
// 2: Configuration/usage errors
// 3: Network/connectivity errors
// 4: Permission/authentication errors
// 5: Resource not found errors

// Logger defines a minimal logger interface to avoid import cycles.
type Logger interface {
	Error(ctx context.Context, msg string, fields map[string]any)
	Info(ctx context.Context, msg string, fields map[string]any)
}

// ErrorHandler provides pure functional error handling without side effects.
type ErrorHandler struct {
	logger Logger
}

// NewErrorHandler creates a new error handler.
func NewErrorHandler(logger Logger) ErrorHandler {
	return ErrorHandler{logger: logger}
}

// HandleFatalError handles fatal errors and returns appropriate exit information
// is a pure function that returns what should be done rather than doing it.
func (h ErrorHandler) HandleFatalError(ctx context.Context, err error) (bool, int, string) {
	if err == nil {
		return false, 0, ""
	}

	h.logger.Error(ctx, "A fatal error occurred", map[string]any{
		"error": err.Error(),
	})

	userFriendlyMessage := h.createUserFriendlyMessage(err)
	if userFriendlyMessage != "" {
		h.logger.Info(ctx, userFriendlyMessage, nil)
	}

	exitCode := h.mapErrorToExitCode(err)

	return true, exitCode, err.Error()
}

// MapErrorToExitCode maps specific error types to appropriate exit codes.
func (h ErrorHandler) mapErrorToExitCode(err error) int {
	errMsg := err.Error()

	switch {
	case strings.Contains(errMsg, "configuration"),
		strings.Contains(errMsg, "config"),
		strings.Contains(errMsg, "flag"):
		return 2 // Configuration errors
	case strings.Contains(errMsg, "network"),
		strings.Contains(errMsg, "connection"),
		strings.Contains(errMsg, "timeout"):
		return 3 // Network errors
	case strings.Contains(errMsg, "permission"),
		strings.Contains(errMsg, "unauthorized"),
		strings.Contains(errMsg, "forbidden"),
		strings.Contains(errMsg, "authentication"):
		return 4 // Permission/auth errors
	case strings.Contains(errMsg, "not found"),
		strings.Contains(errMsg, "does not exist"):
		return 5 // Resource not found
	default:
		return 1 // General operation errors
	}
}

// CreateUserFriendlyMessage creates user-friendly messages for specific error types
// is a pure function with no side effects.
func (h ErrorHandler) createUserFriendlyMessage(err error) string {
	errMsg := err.Error()

	switch {
	case strings.Contains(errMsg, "non-fast-forward update"):
		return "A fast-forward update to target failed. The target may have diverged from the original. Consider using the --force-push option or resolve it manually."
	case strings.Contains(errMsg, "flag accessed but not defined"):
		return "Reading a flag value failed. " + errMsg
	default:
		return ""
	}
}

// ExitIfError is a helper that calls os.Exit if the error is fatal
// is only used at the application boundary (main function).
func ExitIfError(ctx context.Context, logger Logger, err error) {
	if err == nil {
		return
	}

	handler := NewErrorHandler(logger)
	shouldExit, exitCode, _ := handler.HandleFatalError(ctx, err)

	if shouldExit {
		os.Exit(exitCode)
	}
}
