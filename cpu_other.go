//go:build !windows

package main

import "fmt"

// CPUCoreInfo contains the physical and performance-core counts reported by Windows.
type CPUCoreInfo struct {
	PhysicalCores    int
	PerformanceCores int
}

func detectRendererWorkerCount() (int, CPUCoreInfo, error) {
	return 0, CPUCoreInfo{}, fmt.Errorf("la détection des P-cores/cores physiques est uniquement prise en charge sous Windows")
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
