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
	if coreCount < 2 {
		return 0, fmt.Errorf(
			"au moins deux cœurs physiques ou P-cores sont nécessaires : %d",
			coreCount,
		)
	}

	// ceil(coreCount * 0.80), calculé avec des entiers :
	// ceil(4 * coreCount / 5) == (4*coreCount + 4) / 5.
	workerCount := (4*coreCount + 4) / 5

	// Conserver systématiquement un cœur disponible.
	workerCount = min(workerCount, coreCount-1)

	return workerCount, nil
}
