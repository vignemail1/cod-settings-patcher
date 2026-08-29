//go:build !windows

package main

import (
	"fmt"
	"runtime"
)

// CPUCoreInfo contient les nombres de cœurs physiques et de cœurs performance.
type CPUCoreInfo struct {
	PhysicalCores    int
	PerformanceCores int
}

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
		return 0, fmt.Errorf("nombre de cores invalide : %d", coreCount)
	}

	workers := (coreCount*80 + 50) / 100
	if workers < 1 {
		workers = 1
	}
	return workers, nil
}
