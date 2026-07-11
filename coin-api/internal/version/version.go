package version

// Version is set at link time via -ldflags; fallback for local go run.
var Version = "0.1.0-dev"
