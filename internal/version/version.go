package version

import "fmt"

// These are set via -ldflags at build time.
var (
	GitSHA    = "dev"
	BuildTime = "unknown"
)

func String() string {
	return fmt.Sprintf("asih git=%s build_time=%s", GitSHA, BuildTime)
}
