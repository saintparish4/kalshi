package main

import (
	"fmt"
	"runtime"
)

// BuildInfo contains build information
type BuildInfo struct {
	Version   string `json:"version"`
	BuildTime string `json:"build_time"`
	GitCommit string `json:"git_commit"`
	GoVersion string `json:"go_version"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
}

// GetBuildInfo returns build information
func GetBuildInfo() *BuildInfo {
	return &BuildInfo{
		Version:   version,
		BuildTime: buildTime,
		GitCommit: gitCommit,
		GoVersion: runtime.Version(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}
}

// GetVersionString returns formatted version string
func GetVersionString() string {
	return fmt.Sprintf("Kalshi API Gateway v%s (built %s, commit %s)", 
		version, buildTime, gitCommit)
}