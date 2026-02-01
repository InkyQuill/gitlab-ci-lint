package version

// Version information (injected at build time).
// Release builds inject Version, Commit, BuildDate via ldflags.
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)
