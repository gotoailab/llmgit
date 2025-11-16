package main

import (
	"fmt"
	"runtime"
)

var (
	// Version is the version of the application
	Version = "v0.0.1"
	// BuildDate is the date when the binary was built
	BuildDate = "unknown"
	// GitCommit is the git commit hash
	GitCommit = "unknown"
	// GoVersion is the Go version used to build
	GoVersion = runtime.Version()
)

// GetVersionInfo returns formatted version information
func GetVersionInfo() string {
	return fmt.Sprintf(`llmgit %s

Build Information:
  Build Date: %s
  Git Commit: %s
  Go Version: %s
  Platform:   %s/%s
`, Version, BuildDate, GitCommit, GoVersion, runtime.GOOS, runtime.GOARCH)
}

