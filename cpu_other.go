//go:build !windows

package main

import (
	"fmt"
	"runtime"
)

// CPUCoreInfo contains the physical and performance-core counts detected on Windows.
type CPUCoreInfo struct {
	PhysicalCores    int
	PerformanceCores int
}

// detectRendererWorkerCount is a development-only fallback for non-Windows hosts.
// Windows builds use GetLogicalProcessorInformationEx in cpu_windows.go.
func detectRendererWorkerCount() (int, CPUCoreInfo, error) {
	coreCount := runtime.NumCPU()

	workerCount, err := rendererWorkerCount(coreCount)
	if err != nil {
		return 0, CPUCoreInfo{}, err
	}

	return workerCount, CPUCoreInfo{
		PhysicalCores: coreCount,
	}, nil
}

func rendererWorkerCount(coreCount int) (int, error) {
	if coreCount < 1 {
		return 0, fmt.Errorf("nombre de cœurs invalide : %d", coreCount)
	}

	workerCount := (coreCount*80 + 50) / 100
	if workerCount < 1 {
		workerCount = 1
	}

	return workerCount, nil
}
