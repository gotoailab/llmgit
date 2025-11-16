package commands

import (
	"fmt"
	"runtime"
)

var (
	// Version is the version of the application (set via ldflags)
	Version = "dev"
	// BuildDate is the date when the binary was built (set via ldflags)
	BuildDate = "unknown"
	// GitCommit is the git commit hash (set via ldflags)
	GitCommit = "none"
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

// HandleVersion handles the version command
func HandleVersion() {
	fmt.Print(GetVersionInfo())
}

