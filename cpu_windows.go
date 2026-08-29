//go:build windows

package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	relationProcessorCore = 0

	systemLogicalProcessorInformationExHeaderSize = 8

	processorRelationshipFlagsOffset      = 8
	processorRelationshipEfficiencyOffset = 9
)

// CPUCoreInfo contains the physical and performance-core counts detected on Windows.
type CPUCoreInfo struct {
	PhysicalCores    int
	PerformanceCores int
}

func detectRendererWorkerCount() (int, CPUCoreInfo, error) {
	info, err := queryCPUCoreInfo()
	if err != nil {
		return 0, CPUCoreInfo{}, err
	}

	coreCount := info.PhysicalCores
	if info.PerformanceCores > 0 {
		coreCount = info.PerformanceCores
	}

	workerCount, err := rendererWorkerCount(coreCount)
	if err != nil {
		return 0, CPUCoreInfo{}, err
	}

	return workerCount, info, nil
}

func queryCPUCoreInfo() (CPUCoreInfo, error) {
	kernel32 := windows.NewLazySystemDLL("kernel32.dll")
	proc := kernel32.NewProc("GetLogicalProcessorInformationEx")

	var requiredLength uint32

	ok, _, callErr := proc.Call(
		uintptr(relationProcessorCore),
		0,
		uintptr(unsafe.Pointer(&requiredLength)),
	)

	// Le premier appel est attendu en erreur avec ERROR_INSUFFICIENT_BUFFER :
	// il sert uniquement à connaître la taille du buffer nécessaire.
	if ok != 0 {
		return CPUCoreInfo{}, errors.New(
			"GetLogicalProcessorInformationEx a réussi sans retourner de taille de buffer",
		)
	}

	if requiredLength == 0 {
		return CPUCoreInfo{}, fmt.Errorf(
			"GetLogicalProcessorInformationEx n'a retourné aucune taille : %w",
			callErr,
		)
	}

	if callErr != windows.ERROR_INSUFFICIENT_BUFFER {
		return CPUCoreInfo{}, fmt.Errorf(
			"GetLogicalProcessorInformationEx (taille buffer) : %w",
			callErr,
		)
	}

	buffer := make([]byte, requiredLength)

	ok, _, callErr = proc.Call(
		uintptr(relationProcessorCore),
		uintptr(unsafe.Pointer(&buffer[0])),
		uintptr(unsafe.Pointer(&requiredLength)),
	)
	if ok == 0 {
		return CPUCoreInfo{}, fmt.Errorf(
			"GetLogicalProcessorInformationEx : %w",
			callErr,
		)
	}

	return parseProcessorCoreRecords(buffer[:requiredLength])
}

func parseProcessorCoreRecords(buffer []byte) (CPUCoreInfo, error) {
	info := CPUCoreInfo{}
	classCounts := make(map[byte]int)

	for offset := 0; offset < len(buffer); {
		remaining := len(buffer) - offset

		if remaining < systemLogicalProcessorInformationExHeaderSize {
			return CPUCoreInfo{}, fmt.Errorf(
				"enregistrement CPU Windows tronqué à l'offset %d",
				offset,
			)
		}

		relationship := binary.LittleEndian.Uint32(buffer[offset:])
		recordLength := int(binary.LittleEndian.Uint32(buffer[offset+4:]))

		if relationship != relationProcessorCore {
			return CPUCoreInfo{}, fmt.Errorf(
				"relation CPU inattendue à l'offset %d : %d",
				offset,
				relationship,
			)
		}

		// PROCESSOR_RELATIONSHIP commence immédiatement après :
		//
		// SYSTEM_LOGICAL_PROCESSOR_INFORMATION_EX {
		//     LOGICAL_PROCESSOR_RELATIONSHIP Relationship; // 4 bytes
		//     DWORD Size;                                 // 4 bytes
		//     PROCESSOR_RELATIONSHIP Processor;           // offset 8
		// }
		//
		// PROCESSOR_RELATIONSHIP {
		//     BYTE Flags;           // offset +8
		//     BYTE EfficiencyClass; // offset +9
		//     ...
		// }
		const minCoreRecordLength = processorRelationshipEfficiencyOffset + 1

		if recordLength < minCoreRecordLength || recordLength > remaining {
			return CPUCoreInfo{}, fmt.Errorf(
				"taille d'enregistrement CPU invalide à l'offset %d : %d",
				offset,
				recordLength,
			)
		}

		efficiencyClass := buffer[offset+processorRelationshipEfficiencyOffset]

		// RelationProcessorCore retourne un record par cœur physique actif,
		// pas par thread logique.
		info.PhysicalCores++
		classCounts[efficiencyClass]++

		offset += recordLength
	}

	if info.PhysicalCores == 0 {
		return CPUCoreInfo{}, errors.New(
			"GetLogicalProcessorInformationEx n'a retourné aucun cœur physique",
		)
	}

	// EfficiencyClass est non nul uniquement sur un CPU à cœurs hétérogènes.
	// Une valeur supérieure représente un cœur intrinsèquement plus performant.
	//
	// On ne déclare donc des P-cores que si :
	// - plusieurs classes existent réellement ;
	// - la classe la plus performante est non nulle.
	if len(classCounts) > 1 {
		var highestClass byte

		for class := range classCounts {
			if class > highestClass {
				highestClass = class
			}
		}

		if highestClass > 0 {
			info.PerformanceCores = classCounts[highestClass]
		}
	}

	return info, nil
}

func rendererWorkerCount(coreCount int) (int, error) {
	if coreCount < 2 {
		return 0, fmt.Errorf(
			"au moins deux cœurs physiques ou P-cores sont nécessaires : %d",
			coreCount,
		)
	}

	workerCount := int(math.Ceil(float64(coreCount) * 0.80))

	// Ne jamais affecter tous les cœurs retenus au renderer :
	// on laisse au minimum un P-core / cœur physique au système et au jeu.
	workerCount = min(workerCount, coreCount-1)

	return workerCount, nil
}
