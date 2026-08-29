//go:build windows

package main

import (
	"bytes"
	"fmt"
	"math"
	"os/exec"
	"strconv"
	"strings"
)

type CPUCoreInfo struct {
	PhysicalCores    int
	PerformanceCores int
}

func detectRendererWorkerCount() (int, CPUCoreInfo, error) {
	info, err := queryCPUCoreInfo()
	if err != nil {
		return 0, CPUCoreInfo{}, err
	}

	cores := info.PhysicalCores
	if info.PerformanceCores > 0 {
		cores = info.PerformanceCores
	}

	workers, err := rendererWorkerCount(cores)
	if err != nil {
		return 0, CPUCoreInfo{}, err
	}

	return workers, info, nil
}

func queryCPUCoreInfo() (CPUCoreInfo, error) {
	const script = "$cpus = Get-CimInstance -ClassName Win32_Processor; $physical = ($cpus | Measure-Object -Property NumberOfCores -Sum).Sum; $performance = ($cpus | Measure-Object -Property NumberOfPerformanceCores -Sum).Sum; if ($null -eq $performance) { $performance = 0 }; Write-Output \"$physical,$performance\""

	command := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", script)
	output, err := command.Output()
	if err != nil {
		return CPUCoreInfo{}, fmt.Errorf("détection des cores CPU via PowerShell/CIM : %w", err)
	}

	fields := strings.Split(strings.TrimSpace(string(bytes.TrimSpace(output))), ",")
	if len(fields) != 2 {
		return CPUCoreInfo{}, fmt.Errorf("sortie PowerShell inattendue pour la détection CPU : %q", string(output))
	}

	physical, err := strconv.Atoi(strings.TrimSpace(fields[0]))
	if err != nil || physical < 1 {
		return CPUCoreInfo{}, fmt.Errorf("nombre de cores physiques invalide %q", fields[0])
	}

	performance, err := strconv.Atoi(strings.TrimSpace(fields[1]))
	if err != nil || performance < 0 {
		return CPUCoreInfo{}, fmt.Errorf("nombre de P-cores invalide %q", fields[1])
	}

	return CPUCoreInfo{PhysicalCores: physical, PerformanceCores: performance}, nil
}

func rendererWorkerCount(coreCount int) (int, error) {
	if coreCount < 1 {
		return 0, fmt.Errorf("nombre de cores invalide : %d", coreCount)
	}

	workers := int(math.Round(float64(coreCount) * 0.80))
	if workers < 1 {
		workers = 1
	}
	return workers, nil
}
